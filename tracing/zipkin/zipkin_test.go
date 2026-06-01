package zipkin_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/tracing"
	"github.com/azcov/gokit/tracing/zipkin"
)

func TestNew(t *testing.T) {
	tr := zipkin.New(tracing.Config{ServiceName: "my-service"})
	if tr == nil {
		t.Fatal("expected non-nil tracer")
	}
}

func TestStart_ReturnsContextAndSpan(t *testing.T) {
	tr := zipkin.New(tracing.Config{ServiceName: "svc"})
	ctx, sp := tr.Start(context.Background(), "op")
	if ctx == nil {
		t.Fatal("nil context")
	}
	if sp == nil {
		t.Fatal("nil span")
	}
	sp.End()
}

func TestSpan_SetAttribute(t *testing.T) {
	tr := zipkin.New(tracing.Config{ServiceName: "svc"})
	_, sp := tr.Start(context.Background(), "op")
	sp.SetAttribute("foo", "bar")
	sp.SetAttribute("count", 42)
	sp.End()
}

func TestSpan_RecordError(t *testing.T) {
	tr := zipkin.New(tracing.Config{ServiceName: "svc"})
	_, sp := tr.Start(context.Background(), "op")
	sp.RecordError(errors.New("something broke"))
	sp.End()
}

func TestSpan_SetStatus(t *testing.T) {
	tr := zipkin.New(tracing.Config{ServiceName: "svc"})
	_, sp := tr.Start(context.Background(), "op")
	sp.SetStatus(tracing.StatusError, "failed")
	sp.End()

	_, sp2 := tr.Start(context.Background(), "ok-op")
	sp2.SetStatus(tracing.StatusOK, "")
	sp2.End()
}

func TestSpan_AddEvent(t *testing.T) {
	tr := zipkin.New(tracing.Config{ServiceName: "svc"})
	_, sp := tr.Start(context.Background(), "op")
	sp.AddEvent("checkpoint", tracing.Attribute{Key: "step", Value: 1})
	sp.End()
}

func TestSpan_Propagation(t *testing.T) {
	tr := zipkin.New(tracing.Config{ServiceName: "svc"})
	ctx, parent := tr.Start(context.Background(), "parent")
	_, child := tr.Start(ctx, "child")
	child.End()
	parent.End()
}

func TestFlush_SendsToZipkin(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected JSON content-type, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	tr := zipkin.New(tracing.Config{ServiceName: "svc", Endpoint: srv.URL})
	_, sp := tr.Start(context.Background(), "op")
	sp.End()

	if !called {
		t.Error("expected Zipkin server to be called")
	}
}

func TestClose(t *testing.T) {
	tr := zipkin.New(tracing.Config{})
	if err := tr.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestNew_DefaultEndpoint(t *testing.T) {
	tr := zipkin.New(tracing.Config{})
	if tr == nil {
		t.Fatal("nil tracer")
	}
}
