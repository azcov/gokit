package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain composes multiple middlewares into a single one, applied in order.
func Chain(middlewares ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}
		return h
	}
}
