package auth

import (
	"context"
	"net/http"
	"strings"

	gkauth "github.com/azcov/gokit/auth"
	"github.com/azcov/gokit/response"
)

type ctxKey struct{}

// Config configures the auth middleware.
type Config struct {
	// Provider verifies JWT Bearer tokens.
	Provider gkauth.Provider
	// APIKeyFunc verifies API keys. Optional — if nil, API key auth is disabled.
	APIKeyFunc func(key string) (*gkauth.Claims, error)
	// APIKeyHeader is the header name for API keys. Defaults to "X-API-Key".
	APIKeyHeader string
	// Skipper, if set, skips auth for requests where it returns true.
	Skipper func(r *http.Request) bool
}

// New returns middleware that enforces authentication.
// It first checks for a Bearer token, then an API key (if configured).
func New(cfg Config) func(http.Handler) http.Handler {
	if cfg.APIKeyHeader == "" {
		cfg.APIKeyHeader = "X-API-Key"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Skipper != nil && cfg.Skipper(r) {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := extractClaims(r, cfg)
			if err != nil {
				response.Unauthorized(w, "UNAUTHORIZED", "invalid or missing credentials")
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, claims)))
		})
	}
}

// ClaimsFromContext returns the authenticated claims stored by this middleware.
func ClaimsFromContext(ctx context.Context) (*gkauth.Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*gkauth.Claims)
	return c, ok && c != nil
}

func extractClaims(r *http.Request, cfg Config) (*gkauth.Claims, error) {
	// 1. JWT Bearer token
	if cfg.Provider != nil {
		if token := bearerToken(r); token != "" {
			claims, err := cfg.Provider.Verify(r.Context(), token)
			if err == nil {
				return claims, nil
			}
		}
	}

	// 2. API key
	if cfg.APIKeyFunc != nil {
		key := r.Header.Get(cfg.APIKeyHeader)
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}
		if key != "" {
			return cfg.APIKeyFunc(key)
		}
	}

	return nil, errNoCredentials
}

func bearerToken(r *http.Request) string {
	v := r.Header.Get("Authorization")
	if strings.HasPrefix(v, "Bearer ") {
		return strings.TrimPrefix(v, "Bearer ")
	}
	return ""
}

var errNoCredentials = &authError{"no credentials provided"}

type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }
