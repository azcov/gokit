package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const Header = "X-Request-ID"

type ctxKey struct{}

// New returns middleware that ensures every request has an X-Request-ID header.
// If the incoming request already carries one it is reused; otherwise a new
// UUID v4 is generated. The ID is stored in the context and set on the response.
func New() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(Header)
			if id == "" {
				id = uuid.New().String()
			}
			w.Header().Set(Header, id)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
		})
	}
}

// FromContext retrieves the request ID stored by this middleware.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
