package datadog_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/tracing"
	"github.com/azcov/gokit/tracing/datadog"
)

func TestNew_DefaultEndpoint(t *testing.T) {
	tr := datadog.New(tracing.Config{ServiceName: "svc"})
	if tr == nil {
		t.Fatal("expected non-nil tracer")
	}
}

func TestStart_CreatesSpan(t *testing.T) {
	tr := datadog.New(tracing.Config{ServiceName: "svc"})
	ctx, sp := tr.Start(context.Background(), "test-span")
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if sp == nil {
		t.Fatal("expected non-nil span")
	}
	sp.End()
}

func TestSpan_Attributes(t *testing.T) {
	tr := datadog.New(tracing.Config{ServiceName: "svc"})
	_, sp := tr.Start(context.Background(), "op",
		tracing.WithAttribute("env", "test"),
		tracing.WithAttribute("version", 2),
	)
	sp.SetAttribute("user_id", "abc")
	sp.RecordError(errors.New("boom"))
	sp.SetStatus(tracing.StatusError, "something failed")
	sp.AddEvent("cache_miss", tracing.Attribute{Key: "key", Value: "session:1"})
	sp.End()
}

func TestSpan_SpanPropagation(t *testing.T) {
	tr := datadog.New(tracing.Config{ServiceName: "svc"})
	ctx, parent := tr.Start(context.Background(), "parent")
	ctx2, child := tr.Start(ctx, "child")
	_ = ctx2
	child.End()
	parent.End()
}

func TestFlush_SendsToAgent(t *testing.T) {
	var received []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		received = buf
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := datadog.New(tracing.Config{ServiceName: "test-svc", Endpoint: srv.URL})
	_, sp := tr.Start(context.Background(), "traced-op")
	sp.SetAttribute("env", "production")
	sp.End()

	if len(received) == 0 {
		t.Fatal("expected agent to receive payload")
	}
	var payload [][][]map[string]any
	if err := json.Unmarshal(received, &payload); err != nil {
		t.Fatalf("unexpected payload format: %v", err)
	}
}

func TestClose(t *testing.T) {
	tr := datadog.New(tracing.Config{})
	if err := tr.Close(); err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}
