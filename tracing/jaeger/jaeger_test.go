package jaeger_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/tracing"
	"github.com/azcov/gokit/tracing/jaeger"
)

func TestNew(t *testing.T) {
	tr := jaeger.New(tracing.Config{ServiceName: "svc"})
	if tr == nil {
		t.Fatal("nil tracer")
	}
}

func TestStart(t *testing.T) {
	tr := jaeger.New(tracing.Config{ServiceName: "svc"})
	ctx, sp := tr.Start(context.Background(), "span-name")
	if ctx == nil || sp == nil {
		t.Fatal("expected non-nil context and span")
	}
	sp.End()
}

func TestSpan_WithOptions(t *testing.T) {
	tr := jaeger.New(tracing.Config{ServiceName: "svc"})
	_, sp := tr.Start(context.Background(), "op",
		tracing.WithAttribute("db", "postgres"),
		tracing.WithAttribute("query", "SELECT *"),
	)
	sp.SetAttribute("rows", 10)
	sp.End()
}

func TestSpan_RecordError(t *testing.T) {
	tr := jaeger.New(tracing.Config{ServiceName: "svc"})
	_, sp := tr.Start(context.Background(), "op")
	sp.RecordError(errors.New("db error"))
	sp.End()
}

func TestSpan_SetStatus(t *testing.T) {
	tr := jaeger.New(tracing.Config{ServiceName: "svc"})
	cases := []tracing.StatusCode{tracing.StatusOK, tracing.StatusError, tracing.StatusUnset}
	for _, code := range cases {
		_, sp := tr.Start(context.Background(), "op")
		sp.SetStatus(code, "desc")
		sp.End()
	}
}

func TestSpan_AddEvent(t *testing.T) {
	tr := jaeger.New(tracing.Config{ServiceName: "svc"})
	_, sp := tr.Start(context.Background(), "op")
	sp.AddEvent("retry", tracing.Attribute{Key: "attempt", Value: 2})
	sp.End()
}

func TestSpan_ChildPropagation(t *testing.T) {
	tr := jaeger.New(tracing.Config{ServiceName: "svc"})
	ctx, root := tr.Start(context.Background(), "root")
	ctx2, mid := tr.Start(ctx, "mid")
	_, leaf := tr.Start(ctx2, "leaf")
	leaf.End()
	mid.End()
	root.End()
}

func TestFlush_SendsToJaeger(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	tr := jaeger.New(tracing.Config{ServiceName: "svc", Endpoint: srv.URL})
	_, sp := tr.Start(context.Background(), "op")
	sp.End()

	if !called {
		t.Error("expected Jaeger to receive the span")
	}
}

func TestClose(t *testing.T) {
	tr := jaeger.New(tracing.Config{})
	if err := tr.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
