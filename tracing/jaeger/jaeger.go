package jaeger

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

// Jaeger supports receiving traces via the Zipkin-compatible HTTP endpoint
// (enabled with --collector.zipkin.host-port on the Jaeger collector) and
// via the Jaeger Thrift HTTP endpoint. This implementation uses the Zipkin
// v2 JSON format which Jaeger can ingest natively.
const defaultEndpoint = "http://localhost:9411"
const jaegerSpansPath = "/api/v2/spans"

// Tracer sends spans to a Jaeger Collector using the Zipkin v2 JSON protocol.
type Tracer struct {
	cfg    tracing.Config
	client *http.Client
}

func New(cfg tracing.Config) *Tracer {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Tracer{cfg: cfg, client: &http.Client{Timeout: timeout}}
}

func (t *Tracer) Start(ctx context.Context, spanName string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	scfg := &tracing.SpanConfig{}
	for _, o := range opts {
		o(scfg)
	}

	traceID := extractTraceID(ctx)
	if traceID == "" {
		traceID = newID128()
	}
	spanID := newID64()

	s := &span{
		tracer:       t,
		traceID:      traceID,
		spanID:       spanID,
		parentID:     extractSpanID(ctx),
		name:         spanName,
		localService: t.cfg.ServiceName,
		start:        time.Now(),
		tags:         make(map[string]string),
	}
	for _, a := range scfg.Attributes {
		s.tags[a.Key] = fmt.Sprint(a.Value)
	}

	ctx = contextWithTrace(ctx, traceID, spanID)
	return ctx, s
}

func (t *Tracer) Close() error { return nil }

func (t *Tracer) flush(s *span) {
	durationMicros := time.Since(s.start).Microseconds()

	payload := []jaegerSpan{
		{
			TraceID:       s.traceID,
			ID:            s.spanID,
			ParentID:      s.parentID,
			Name:          s.name,
			Timestamp:     s.start.UnixMicro(),
			Duration:      durationMicros,
			Tags:          s.tags,
			LocalEndpoint: endpoint{ServiceName: s.localService},
		},
	}
	if s.err != nil {
		payload[0].Tags["error"] = s.err.Error()
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, t.cfg.Endpoint+jaegerSpansPath, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

type jaegerSpan struct {
	TraceID       string            `json:"traceId"`
	ID            string            `json:"id"`
	ParentID      string            `json:"parentId,omitempty"`
	Name          string            `json:"name"`
	Timestamp     int64             `json:"timestamp"`
	Duration      int64             `json:"duration"`
	Tags          map[string]string `json:"tags,omitempty"`
	LocalEndpoint endpoint          `json:"localEndpoint,omitempty"`
}

type endpoint struct {
	ServiceName string `json:"serviceName,omitempty"`
}

type span struct {
	tracer       *Tracer
	traceID      string
	spanID       string
	parentID     string
	name         string
	localService string
	start        time.Time
	tags         map[string]string
	err          error
}

func (s *span) End() { s.tracer.flush(s) }

func (s *span) SetAttribute(key string, value any) {
	s.tags[key] = fmt.Sprint(value)
}

func (s *span) RecordError(err error) { s.err = err }

func (s *span) SetStatus(code tracing.StatusCode, description string) {
	if description != "" {
		s.tags["status"] = description
	}
	if code == tracing.StatusError {
		s.tags["error"] = "true"
	}
}

func (s *span) AddEvent(name string, attrs ...tracing.Attribute) {
	s.tags["event"] = name
	for _, a := range attrs {
		s.tags[a.Key] = fmt.Sprint(a.Value)
	}
}

type ctxKeyTraceID struct{}
type ctxKeySpanID struct{}

func contextWithTrace(ctx context.Context, traceID, spanID string) context.Context {
	ctx = context.WithValue(ctx, ctxKeyTraceID{}, traceID)
	ctx = context.WithValue(ctx, ctxKeySpanID{}, spanID)
	return ctx
}

func extractTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyTraceID{}).(string); ok {
		return v
	}
	return ""
}

func extractSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeySpanID{}).(string); ok {
		return v
	}
	return ""
}

func newID128() string { return fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64()) }
func newID64() string  { return fmt.Sprintf("%016x", rand.Uint64()) }
