package timeout

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// New returns middleware that cancels the request context after d.
// If the handler does not finish in time, a 408 Request Timeout is written.
func New(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			done := make(chan struct{})
			tw := &timeoutWriter{ResponseWriter: w}

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				tw.flush(w)
			case <-ctx.Done():
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestTimeout)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"error": map[string]string{
						"code":    "REQUEST_TIMEOUT",
						"message": "request timed out",
					},
				})
			}
		})
	}
}

type timeoutWriter struct {
	http.ResponseWriter
	buf    []byte
	header http.Header
	status int
	wrote  bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	if !tw.wrote {
		tw.status = code
	}
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.buf = append(tw.buf, b...)
	return len(b), nil
}

func (tw *timeoutWriter) Header() http.Header {
	if tw.header == nil {
		tw.header = make(http.Header)
	}
	return tw.header
}

func (tw *timeoutWriter) flush(w http.ResponseWriter) {
	for k, vals := range tw.header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	if tw.status != 0 {
		w.WriteHeader(tw.status)
	}
	_, _ = w.Write(tw.buf)
}
