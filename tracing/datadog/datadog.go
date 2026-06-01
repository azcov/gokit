package datadog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/azcov/gokit/tracing"
)

var _ tracing.Tracer = (*Tracer)(nil)

const defaultAgentURL = "http://localhost:8126"
const ddAPIVersion = "/v0.4/traces"

// Tracer sends spans to a Datadog Agent via its HTTP API.
type Tracer struct {
	cfg     tracing.Config
	client  *http.Client
	service string
}

func New(cfg tracing.Config) *Tracer {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultAgentURL
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Tracer{
		cfg:     cfg,
		client:  &http.Client{Timeout: timeout},
		service: cfg.ServiceName,
	}
}

func (t *Tracer) Start(ctx context.Context, spanName string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	cfg := &tracing.SpanConfig{}
	for _, o := range opts {
		o(cfg)
	}

	traceID := extractTraceID(ctx)
	if traceID == 0 {
		traceID = rand.Uint64()
	}
	spanID := rand.Uint64()

	s := &span{
		tracer:     t,
		traceID:    traceID,
		spanID:     spanID,
		parentID:   extractSpanID(ctx),
		name:       spanName,
		service:    t.service,
		resource:   spanName,
		start:      time.Now(),
		meta:       make(map[string]string),
		statusCode: tracing.StatusUnset,
	}
	for _, a := range cfg.Attributes {
		s.meta[a.Key] = fmt.Sprint(a.Value)
	}

	ctx = contextWithTrace(ctx, traceID, spanID)
	return ctx, s
}

func (t *Tracer) Close() error { return nil }

func (t *Tracer) flush(s *span) {
	duration := time.Since(s.start).Nanoseconds()
	ddSpan := ddSpanPayload{
		TraceID:  s.traceID,
		SpanID:   s.spanID,
		ParentID: s.parentID,
		Name:     s.name,
		Resource: s.resource,
		Service:  s.service,
		Type:     "custom",
		Start:    s.start.UnixNano(),
		Duration: duration,
		Meta:     s.meta,
	}
	if s.err != nil {
		ddSpan.Error = 1
		ddSpan.Meta["error.message"] = s.err.Error()
		ddSpan.Meta["error.type"] = fmt.Sprintf("%T", s.err)
	}

	payload := [][][]ddSpanPayload{{{ddSpan}}}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, t.cfg.Endpoint+ddAPIVersion, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Datadog-Trace-Count", "1")
	resp, err := t.client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

type ddSpanPayload struct {
	TraceID  uint64            `json:"trace_id"`
	SpanID   uint64            `json:"span_id"`
	ParentID uint64            `json:"parent_id"`
	Name     string            `json:"name"`
	Resource string            `json:"resource"`
	Service  string            `json:"service"`
	Type     string            `json:"type"`
	Start    int64             `json:"start"`
	Duration int64             `json:"duration"`
	Error    int32             `json:"error"`
	Meta     map[string]string `json:"meta,omitempty"`
}

// span implements tracing.Span.
type span struct {
	tracer     *Tracer
	traceID    uint64
	spanID     uint64
	parentID   uint64
	name       string
	service    string
	resource   string
	start      time.Time
	meta       map[string]string
	err        error
	statusCode tracing.StatusCode
}

func (s *span) End() { s.tracer.flush(s) }

func (s *span) SetAttribute(key string, value any) {
	s.meta[key] = fmt.Sprint(value)
}

func (s *span) RecordError(err error) {
	s.err = err
}

func (s *span) SetStatus(code tracing.StatusCode, description string) {
	s.statusCode = code
	if description != "" {
		s.meta["status.description"] = description
	}
}

func (s *span) AddEvent(name string, attrs ...tracing.Attribute) {
	s.meta["event."+name] = name
	for _, a := range attrs {
		s.meta["event."+name+"."+a.Key] = fmt.Sprint(a.Value)
	}
}

// Context key types for trace propagation.
type ctxKeyTraceID struct{}
type ctxKeySpanID struct{}

func contextWithTrace(ctx context.Context, traceID, spanID uint64) context.Context {
	ctx = context.WithValue(ctx, ctxKeyTraceID{}, traceID)
	ctx = context.WithValue(ctx, ctxKeySpanID{}, spanID)
	return ctx
}

func extractTraceID(ctx context.Context) uint64 {
	if v, ok := ctx.Value(ctxKeyTraceID{}).(uint64); ok {
		return v
	}
	return 0
}

func extractSpanID(ctx context.Context) uint64 {
	if v, ok := ctx.Value(ctxKeySpanID{}).(uint64); ok {
		return v
	}
	return 0
}
