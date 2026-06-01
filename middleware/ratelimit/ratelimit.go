package ratelimit

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	// RequestsPerSecond is the sustained rate of requests allowed per key.
	RequestsPerSecond float64
	// Burst is the maximum number of requests allowed in a burst.
	Burst int
	// KeyFunc extracts the rate-limit key from the request (e.g. IP or user ID).
	// Defaults to the request's RemoteAddr.
	KeyFunc func(r *http.Request) string
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type limiter struct {
	cfg     Config
	mu      sync.Mutex
	buckets map[string]*bucket
}

// New returns a token-bucket rate-limiting middleware.
func New(cfg Config) func(http.Handler) http.Handler {
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = 10
	}
	if cfg.Burst <= 0 {
		cfg.Burst = int(cfg.RequestsPerSecond)
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(r *http.Request) string { return r.RemoteAddr }
	}
	l := &limiter{cfg: cfg, buckets: make(map[string]*bucket)}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.KeyFunc(r)
			if !l.allow(key) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"error": map[string]string{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "too many requests",
					},
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(l.cfg.Burst), lastRefill: time.Now()}
		l.buckets[key] = b
	}

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * l.cfg.RequestsPerSecond
	if b.tokens > float64(l.cfg.Burst) {
		b.tokens = float64(l.cfg.Burst)
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
