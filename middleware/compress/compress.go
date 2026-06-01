package compress

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

const defaultMinSize = 1024 // bytes

type Config struct {
	Level   int // gzip compression level, default gzip.DefaultCompression
	MinSize int // minimum response size to compress
}

var gzipPool = sync.Pool{
	New: func() any { w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression); return w },
}

// New returns middleware that gzip-compresses responses when the client
// sends Accept-Encoding: gzip and the response is >= MinSize bytes.
func New(cfg ...Config) func(http.Handler) http.Handler {
	c := Config{Level: gzip.DefaultCompression, MinSize: defaultMinSize}
	if len(cfg) > 0 {
		if cfg[0].Level != 0 {
			c.Level = cfg[0].Level
		}
		if cfg[0].MinSize > 0 {
			c.MinSize = cfg[0].MinSize
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gz := gzipPool.Get().(*gzip.Writer)
			gz.Reset(w)
			defer func() {
				_ = gz.Close()
				gzipPool.Put(gz)
			}()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			w.Header().Add("Vary", "Accept-Encoding")

			next.ServeHTTP(&gzipWriter{ResponseWriter: w, gz: gz}, r)
		})
	}
}

type gzipWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (g *gzipWriter) Write(b []byte) (int, error) { return g.gz.Write(b) }
func (g *gzipWriter) Flush() {
	_ = g.gz.Flush()
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
