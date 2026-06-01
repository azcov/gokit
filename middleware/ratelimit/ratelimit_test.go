package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/middleware/ratelimit"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimit_AllowsRequests(t *testing.T) {
	mw := ratelimit.New(ratelimit.Config{RequestsPerSecond: 10, Burst: 10})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRateLimit_BlocksAfterBurst(t *testing.T) {
	mw := ratelimit.New(ratelimit.Config{RequestsPerSecond: 1, Burst: 2})
	h := mw(okHandler())

	for i := range 2 {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 after burst, got %d", rr.Code)
	}
}

func TestRateLimit_CustomKeyFunc(t *testing.T) {
	mw := ratelimit.New(ratelimit.Config{
		RequestsPerSecond: 1,
		Burst:             1,
		KeyFunc:           func(r *http.Request) string { return r.Header.Get("X-User-ID") },
	})
	h := mw(okHandler())

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-User-ID", "user-1")
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("user-1 first request: expected 200, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-User-ID", "user-2")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("user-2 first request: expected 200, got %d", rr2.Code)
	}
}

func TestRateLimit_DefaultConfig(t *testing.T) {
	mw := ratelimit.New(ratelimit.Config{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(okHandler()).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
