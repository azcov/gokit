package tracing_test

import (
	"testing"

	"github.com/azcov/gokit/tracing"
)

func TestSpanOptionWithAttribute(t *testing.T) {
	cfg := &tracing.SpanConfig{}
	opt := tracing.WithAttribute("key", "value")
	opt(cfg)

	if len(cfg.Attributes) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(cfg.Attributes))
	}
	if cfg.Attributes[0].Key != "key" {
		t.Errorf("expected key 'key', got %q", cfg.Attributes[0].Key)
	}
	if cfg.Attributes[0].Value != "value" {
		t.Errorf("expected value 'value', got %v", cfg.Attributes[0].Value)
	}
}

func TestSpanOptionMultiple(t *testing.T) {
	cfg := &tracing.SpanConfig{}
	tracing.WithAttribute("a", 1)(cfg)
	tracing.WithAttribute("b", true)(cfg)
	tracing.WithAttribute("c", 3.14)(cfg)

	if len(cfg.Attributes) != 3 {
		t.Fatalf("expected 3 attributes, got %d", len(cfg.Attributes))
	}
}

func TestStatusCodeConstants(t *testing.T) {
	codes := []struct {
		code tracing.StatusCode
		want tracing.StatusCode
	}{
		{tracing.StatusUnset, 0},
		{tracing.StatusOK, 1},
		{tracing.StatusError, 2},
	}
	for _, tc := range codes {
		if tc.code != tc.want {
			t.Errorf("status code: got %d, want %d", tc.code, tc.want)
		}
	}
}
