package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/middleware/cors"
)

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefault_NoOrigin(t *testing.T) {
	mw := cors.Default()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(handler()).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("should not set CORS headers without Origin")
	}
}

func TestDefault_WithOrigin(t *testing.T) {
	mw := cors.Default()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	mw(handler()).ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("expected origin header, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Allow-Methods header")
	}
}

func TestPreflight_Options(t *testing.T) {
	mw := cors.Default()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	mw(handler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rr.Code)
	}
}

func TestNew_AllowedOrigin(t *testing.T) {
	mw := cors.New(cors.Config{
		AllowedOrigins: []string{"https://allowed.com"},
	})
	tests := []struct {
		origin  string
		wantSet bool
	}{
		{"https://allowed.com", true},
		{"https://blocked.com", false},
	}
	for _, tc := range tests {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", tc.origin)
		mw(handler()).ServeHTTP(rr, req)
		got := rr.Header().Get("Access-Control-Allow-Origin")
		if tc.wantSet && got == "" {
			t.Errorf("origin %q should be allowed", tc.origin)
		}
		if !tc.wantSet && got != "" {
			t.Errorf("origin %q should be blocked", tc.origin)
		}
	}
}

func TestNew_Credentials(t *testing.T) {
	mw := cors.New(cors.Config{
		AllowedOrigins:   []string{"https://app.com"},
		AllowCredentials: true,
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.com")
	mw(handler()).ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("expected credentials header")
	}
}

func TestNew_ExposedHeaders(t *testing.T) {
	mw := cors.New(cors.Config{
		AllowedOrigins: []string{"*"},
		ExposedHeaders: []string{"X-Request-ID"},
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.com")
	mw(handler()).ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Expose-Headers") != "X-Request-ID" {
		t.Error("expected exposed headers")
	}
}
