package monitoring

import "context"

type Label struct {
	Key   string
	Value string
}

type Counter interface {
	Add(ctx context.Context, value int64, labels ...Label)
}

type Gauge interface {
	Set(ctx context.Context, value float64, labels ...Label)
	Add(ctx context.Context, value float64, labels ...Label)
}

type Histogram interface {
	Record(ctx context.Context, value float64, labels ...Label)
}

type Config struct {
	Endpoint    string `config:"endpoint"`
	ServiceName string `config:"service_name"`
}

type Metrics interface {
	Counter(name, description string) Counter
	Gauge(name, description string) Gauge
	Histogram(name, description string, buckets []float64) Histogram
	Close() error
}
