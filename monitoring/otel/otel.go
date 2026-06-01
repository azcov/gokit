package otel

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	mon "github.com/azcov/gokit/monitoring"
)

var _ mon.Metrics = (*Provider)(nil)

// Provider wraps the global OTel MeterProvider.
type Provider struct {
	meter metric.Meter
}

// New creates a Provider using the global OTel MeterProvider.
// Call otel.SetMeterProvider before New to use a custom provider.
func New(cfg mon.Config) *Provider {
	name := cfg.ServiceName
	if name == "" {
		name = "gokit"
	}
	return &Provider{meter: otel.GetMeterProvider().Meter(name)}
}

func (p *Provider) Counter(name, description string) mon.Counter {
	c, _ := p.meter.Int64Counter(name, metric.WithDescription(description))
	return &counter{inst: c}
}

func (p *Provider) Gauge(name, description string) mon.Gauge {
	g, _ := p.meter.Float64UpDownCounter(name, metric.WithDescription(description))
	return &gauge{inst: g}
}

func (p *Provider) Histogram(name, description string, _ []float64) mon.Histogram {
	h, _ := p.meter.Float64Histogram(name, metric.WithDescription(description))
	return &histogram{inst: h}
}

func (p *Provider) Close() error { return nil }

type counter struct {
	inst metric.Int64Counter
}

func (c *counter) Add(ctx context.Context, value int64, labels ...mon.Label) {
	c.inst.Add(ctx, value, metric.WithAttributes(toAttrs(labels)...))
}

type gauge struct {
	inst    metric.Float64UpDownCounter
	mu      sync.Mutex
	current float64
}

func (g *gauge) Add(ctx context.Context, value float64, labels ...mon.Label) {
	g.mu.Lock()
	g.current += value
	g.mu.Unlock()
	g.inst.Add(ctx, value, metric.WithAttributes(toAttrs(labels)...))
}

func (g *gauge) Set(ctx context.Context, value float64, labels ...mon.Label) {
	g.mu.Lock()
	delta := value - g.current
	g.current = value
	g.mu.Unlock()
	g.inst.Add(ctx, delta, metric.WithAttributes(toAttrs(labels)...))
}

type histogram struct {
	inst metric.Float64Histogram
}

func (h *histogram) Record(ctx context.Context, value float64, labels ...mon.Label) {
	h.inst.Record(ctx, value, metric.WithAttributes(toAttrs(labels)...))
}

func toAttrs(labels []mon.Label) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, len(labels))
	for i, l := range labels {
		attrs[i] = attribute.String(l.Key, l.Value)
	}
	return attrs
}
