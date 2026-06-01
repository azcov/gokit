package statsd

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	mon "github.com/azcov/gokit/monitoring"
)

var _ mon.Metrics = (*Provider)(nil)

type Config struct {
	mon.Config
	// Addr is the StatsD UDP address. Defaults to "127.0.0.1:8125".
	Addr string
	// Prefix is prepended to every metric name.
	Prefix string
	// FlushInterval is unused for UDP (each observation is sent immediately).
}

// Provider sends metrics to a StatsD daemon over UDP.
type Provider struct {
	conn   net.Conn
	prefix string
}

func New(cfg Config) (*Provider, error) {
	addr := cfg.Addr
	if addr == "" {
		addr = "127.0.0.1:8125"
	}
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("statsd: dial %s: %w", addr, err)
	}
	return &Provider{conn: conn, prefix: cfg.Prefix}, nil
}

func (p *Provider) Counter(name, _ string) mon.Counter {
	return &statsdCounter{p: p, name: p.metricName(name)}
}

func (p *Provider) Gauge(name, _ string) mon.Gauge {
	return &statsdGauge{p: p, name: p.metricName(name)}
}

func (p *Provider) Histogram(name, _ string, _ []float64) mon.Histogram {
	return &statsdHistogram{p: p, name: p.metricName(name)}
}

func (p *Provider) Close() error { return p.conn.Close() }

func (p *Provider) send(line string) {
	_, _ = fmt.Fprint(p.conn, line)
}

func (p *Provider) metricName(name string) string {
	if p.prefix == "" {
		return name
	}
	return p.prefix + "." + name
}

// StatsD wire format: `metric_name[.tags]:value|type`
// Tags are appended as part of the name with dots for basic StatsD compatibility.

type statsdCounter struct {
	p    *Provider
	name string
}

func (c *statsdCounter) Add(_ context.Context, value int64, labels ...mon.Label) {
	c.p.send(fmt.Sprintf("%s:%d|c", taggedName(c.name, labels), value))
}

type statsdGauge struct {
	p    *Provider
	name string
}

func (g *statsdGauge) Set(_ context.Context, value float64, labels ...mon.Label) {
	g.p.send(fmt.Sprintf("%s:%g|g", taggedName(g.name, labels), value))
}

func (g *statsdGauge) Add(_ context.Context, value float64, labels ...mon.Label) {
	prefix := ""
	if value >= 0 {
		prefix = "+"
	}
	g.p.send(fmt.Sprintf("%s:%s%g|g", taggedName(g.name, labels), prefix, value))
}

type statsdHistogram struct {
	p    *Provider
	name string
}

func (h *statsdHistogram) Record(_ context.Context, value float64, labels ...mon.Label) {
	h.p.send(fmt.Sprintf("%s:%g|ms", taggedName(h.name, labels), value))
}

// taggedName appends labels as metric.key_value suffix (basic StatsD compat).
func taggedName(name string, labels []mon.Label) string {
	if len(labels) == 0 {
		return name
	}
	var sb strings.Builder
	sb.WriteString(name)
	for _, l := range labels {
		sb.WriteByte('.')
		sb.WriteString(l.Key)
		sb.WriteByte('_')
		sb.WriteString(strings.ReplaceAll(l.Value, " ", "_"))
	}
	return sb.String()
}
