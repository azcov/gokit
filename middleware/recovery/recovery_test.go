package recovery_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/middleware/recovery"
)

func panicHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
}

func normalHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestRecovery_NoPanic(t *testing.T) {
	mw := recovery.New()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(normalHandler()).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRecovery_RecoversPanic(t *testing.T) {
	mw := recovery.New()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(panicHandler()).ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestRecovery_CustomOnPanic(t *testing.T) {
	called := false
	mw := recovery.New(recovery.Options{
		OnPanic: func(w http.ResponseWriter, r *http.Request, rec any, stack []byte) {
			called = true
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(panicHandler()).ServeHTTP(rr, req)
	if !called {
		t.Error("expected OnPanic to be called")
	}
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestRecovery_PanicWithNilValue(t *testing.T) {
	nilPanicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(nil)
	})
	mw := recovery.New()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(nilPanicHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
