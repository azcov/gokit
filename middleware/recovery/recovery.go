package recovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
)

type Options struct {
	// OnPanic is called when a panic is recovered. If nil, a JSON error is written.
	OnPanic func(w http.ResponseWriter, r *http.Request, recovered any, stack []byte)
}

// New returns a middleware that recovers from panics, logs the stack trace, and
// writes a 500 Internal Server Error to the client.
func New(opts ...Options) func(http.Handler) http.Handler {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.OnPanic == nil {
		opt.OnPanic = defaultOnPanic
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					opt.OnPanic(w, r, rec, debug.Stack())
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func defaultOnPanic(w http.ResponseWriter, _ *http.Request, recovered any, stack []byte) {
	fmt.Printf("panic recovered: %v\n%s\n", recovered, stack)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "an unexpected error occurred",
		},
	})
}
