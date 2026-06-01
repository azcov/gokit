package tracing

import (
	"context"
	"time"
)

type StatusCode int

const (
	StatusUnset StatusCode = 0
	StatusOK    StatusCode = 1
	StatusError StatusCode = 2
)

type Attribute struct {
	Key   string
	Value any
}

type SpanConfig struct {
	Attributes []Attribute
}

type SpanOption func(*SpanConfig)

func WithAttribute(key string, value any) SpanOption {
	return func(c *SpanConfig) {
		c.Attributes = append(c.Attributes, Attribute{Key: key, Value: value})
	}
}

type Span interface {
	End()
	SetAttribute(key string, value any)
	RecordError(err error)
	SetStatus(code StatusCode, description string)
	AddEvent(name string, attrs ...Attribute)
}

type Config struct {
	ServiceName string        `config:"service_name"`
	Endpoint    string        `config:"endpoint"`
	SampleRate  float64       `config:"sample_rate"`
	Timeout     time.Duration `config:"timeout"`
}

type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span)
	Close() error
}
