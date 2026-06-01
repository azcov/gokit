package logger

import (
	"fmt"
	"net/http"
	"time"

	gklogger "github.com/azcov/gokit/logger"
	"github.com/azcov/gokit/middleware/requestid"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// New returns middleware that logs each request using the provided logger.
// It logs: method, path, status, latency, bytes written, and request ID.
func New(log gklogger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			latency := time.Since(start)

			fields := []gklogger.Field{
				gklogger.F("method", r.Method),
				gklogger.F("path", r.URL.Path),
				gklogger.F("status", rec.status),
				gklogger.F("latency", latency.String()),
				gklogger.F("bytes", rec.bytes),
				gklogger.F("request_id", requestid.FromContext(r.Context())),
			}

			msg := fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, rec.status)
			if rec.status >= 500 {
				log.Error(msg, fields...)
			} else if rec.status >= 400 {
				log.Warn(msg, fields...)
			} else {
				log.Info(msg, fields...)
			}
		})
	}
}
