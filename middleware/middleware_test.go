package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/middleware"
)

func TestChain_Single(t *testing.T) {
	var order []int
	m := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 1)
			next.ServeHTTP(w, r)
		})
	}
	h := middleware.Chain(middleware.Middleware(m))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, 99)
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if len(order) != 2 || order[0] != 1 || order[1] != 99 {
		t.Errorf("unexpected execution order: %v", order)
	}
}

func TestChain_Order(t *testing.T) {
	var order []int
	makeMiddleware := func(n int) middleware.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, n)
				next.ServeHTTP(w, r)
			})
		}
	}

	chain := middleware.Chain(makeMiddleware(1), makeMiddleware(2), makeMiddleware(3))
	h := chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, 99)
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []int{1, 2, 3, 99}
	for i, v := range order {
		if v != want[i] {
			t.Errorf("position %d: expected %d, got %d", i, want[i], v)
		}
	}
}

func TestChain_Empty(t *testing.T) {
	called := false
	chain := middleware.Chain()
	h := chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Error("expected handler to be called")
	}
}
