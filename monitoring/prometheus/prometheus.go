package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	mon "github.com/azcov/gokit/monitoring"
)

var _ mon.Metrics = (*Provider)(nil)

var defaultBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// Provider implements mon.Metrics and exposes a /metrics HTTP handler
// that serves Prometheus text format 0.0.4 — no external SDK required.
type Provider struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

type entryKind int

const (
	kindCounter   entryKind = iota
	kindGauge
	kindHistogram
)

type entry struct {
	name    string
	desc    string
	kind    entryKind
	buckets []float64 // histogram only

	mu     sync.Mutex
	values map[string]float64   // counter / gauge: labelKey → value
	hists  map[string]*histData // histogram: labelKey → data
}

type histData struct {
	counts []uint64
	sum    float64
	total  uint64
}

func New(_ mon.Config) *Provider {
	return &Provider{entries: make(map[string]*entry)}
}

func (p *Provider) Counter(name, desc string) mon.Counter {
	return &counter{e: p.getOrCreate(name, desc, kindCounter, nil)}
}

func (p *Provider) Gauge(name, desc string) mon.Gauge {
	return &gauge{e: p.getOrCreate(name, desc, kindGauge, nil)}
}

func (p *Provider) Histogram(name, desc string, buckets []float64) mon.Histogram {
	if len(buckets) == 0 {
		buckets = defaultBuckets
	}
	return &histogram{e: p.getOrCreate(name, desc, kindHistogram, buckets)}
}

func (p *Provider) Close() error { return nil }

// Handler returns an http.Handler that serves Prometheus text format on /metrics.
func (p *Provider) Handler() http.Handler { return p }

// ServeHTTP writes Prometheus text format. Register as: mux.Handle("/metrics", provider)
func (p *Provider) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	p.mu.RLock()
	names := make([]string, 0, len(p.entries))
	for n := range p.entries {
		names = append(names, n)
	}
	p.mu.RUnlock()
	sort.Strings(names)

	for _, name := range names {
		p.mu.RLock()
		e := p.entries[name]
		p.mu.RUnlock()

		e.mu.Lock()
		typStr := map[entryKind]string{kindCounter: "counter", kindGauge: "gauge", kindHistogram: "histogram"}[e.kind]
		fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s %s\n", e.name, e.desc, e.name, typStr)

		switch e.kind {
		case kindCounter, kindGauge:
			for _, k := range sortedStringKeys(e.values) {
				fmt.Fprintf(w, "%s%s %g\n", e.name, k, e.values[k])
			}
		case kindHistogram:
			for _, k := range sortedHistKeys(e.hists) {
				h := e.hists[k]
				for i, b := range e.buckets {
					fmt.Fprintf(w, "%s_bucket%s %d\n", e.name, withLabel(k, "le", fmt.Sprintf("%g", b)), h.counts[i])
				}
				fmt.Fprintf(w, "%s_bucket%s %d\n", e.name, withLabel(k, "le", "+Inf"), h.total)
				fmt.Fprintf(w, "%s_sum%s %g\n", e.name, k, h.sum)
				fmt.Fprintf(w, "%s_count%s %d\n", e.name, k, h.total)
			}
		}
		e.mu.Unlock()
		fmt.Fprintln(w)
	}
}

func (p *Provider) getOrCreate(name, desc string, kind entryKind, buckets []float64) *entry {
	p.mu.Lock()
	defer p.mu.Unlock()
	if e, ok := p.entries[name]; ok {
		return e
	}
	e := &entry{
		name:    name,
		desc:    desc,
		kind:    kind,
		buckets: buckets,
		values:  make(map[string]float64),
		hists:   make(map[string]*histData),
	}
	p.entries[name] = e
	return e
}

// --- counter ---

type counter struct{ e *entry }

var _ mon.Counter = (*counter)(nil)

func (c *counter) Add(_ context.Context, value int64, labels ...mon.Label) {
	k := labelKey(labels)
	c.e.mu.Lock()
	c.e.values[k] += float64(value)
	c.e.mu.Unlock()
}

// --- gauge ---

type gauge struct{ e *entry }

var _ mon.Gauge = (*gauge)(nil)

func (g *gauge) Set(_ context.Context, value float64, labels ...mon.Label) {
	k := labelKey(labels)
	g.e.mu.Lock()
	g.e.values[k] = value
	g.e.mu.Unlock()
}

func (g *gauge) Add(_ context.Context, value float64, labels ...mon.Label) {
	k := labelKey(labels)
	g.e.mu.Lock()
	g.e.values[k] += value
	g.e.mu.Unlock()
}

// --- histogram ---

type histogram struct{ e *entry }

var _ mon.Histogram = (*histogram)(nil)

func (h *histogram) Record(_ context.Context, value float64, labels ...mon.Label) {
	k := labelKey(labels)
	h.e.mu.Lock()
	hd, ok := h.e.hists[k]
	if !ok {
		hd = &histData{counts: make([]uint64, len(h.e.buckets))}
		h.e.hists[k] = hd
	}
	for i, b := range h.e.buckets {
		if value <= b {
			hd.counts[i]++
		}
	}
	hd.sum += value
	hd.total++
	h.e.mu.Unlock()
}

// --- helpers ---

func labelKey(labels []mon.Label) string {
	if len(labels) == 0 {
		return ""
	}
	sorted := make([]mon.Label, len(labels))
	copy(sorted, labels)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Key < sorted[j].Key })
	var b strings.Builder
	b.WriteByte('{')
	for i, l := range sorted {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `%s="%s"`, l.Key, l.Value)
	}
	b.WriteByte('}')
	return b.String()
}

func withLabel(existing, key, value string) string {
	kv := fmt.Sprintf(`%s="%s"`, key, value)
	if existing == "" {
		return "{" + kv + "}"
	}
	return existing[:len(existing)-1] + "," + kv + "}"
}

func sortedStringKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedHistKeys(m map[string]*histData) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
