package opentelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oeltrace "go.opentelemetry.io/otel/trace"

	"github.com/azcov/gokit/tracing"
)

var _ tracing.Tracer = (*Tracer)(nil)

// Tracer wraps the global OTel tracer.
type Tracer struct {
	inner oeltrace.Tracer
}

// New returns a Tracer using the global OTel TracerProvider.
// Set a custom TracerProvider via otel.SetTracerProvider before calling New.
func New(cfg tracing.Config) *Tracer {
	name := cfg.ServiceName
	if name == "" {
		name = "gokit"
	}
	return &Tracer{inner: otel.GetTracerProvider().Tracer(name)}
}

func (t *Tracer) Start(ctx context.Context, spanName string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	cfg := &tracing.SpanConfig{}
	for _, o := range opts {
		o(cfg)
	}
	startOpts := []oeltrace.SpanStartOption{}
	if len(cfg.Attributes) > 0 {
		attrs := make([]attribute.KeyValue, len(cfg.Attributes))
		for i, a := range cfg.Attributes {
			attrs[i] = toAttr(a)
		}
		startOpts = append(startOpts, oeltrace.WithAttributes(attrs...))
	}

	ctx, otelSpan := t.inner.Start(ctx, spanName, startOpts...)
	return ctx, &span{inner: otelSpan}
}

func (t *Tracer) Close() error { return nil }

type span struct {
	inner oeltrace.Span
}

func (s *span) End() { s.inner.End() }

func (s *span) SetAttribute(key string, value any) {
	s.inner.SetAttributes(toAttrFromAny(key, value))
}

func (s *span) RecordError(err error) {
	s.inner.RecordError(err)
}

func (s *span) SetStatus(code tracing.StatusCode, description string) {
	switch code {
	case tracing.StatusOK:
		s.inner.SetStatus(codes.Ok, description)
	case tracing.StatusError:
		s.inner.SetStatus(codes.Error, description)
	default:
		s.inner.SetStatus(codes.Unset, description)
	}
}

func (s *span) AddEvent(name string, attrs ...tracing.Attribute) {
	kv := make([]attribute.KeyValue, len(attrs))
	for i, a := range attrs {
		kv[i] = toAttr(a)
	}
	s.inner.AddEvent(name, oeltrace.WithAttributes(kv...))
}

func toAttr(a tracing.Attribute) attribute.KeyValue {
	return toAttrFromAny(a.Key, a.Value)
}

func toAttrFromAny(key string, value any) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	default:
		return attribute.String(key, fmt.Sprint(v))
	}
}
