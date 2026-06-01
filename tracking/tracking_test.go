package tracking_test

import (
	"testing"

	"github.com/azcov/gokit/tracking"
)

func TestLevelConstants(t *testing.T) {
	cases := []struct {
		level tracking.Level
		want  string
	}{
		{tracking.LevelDebug, "debug"},
		{tracking.LevelInfo, "info"},
		{tracking.LevelWarning, "warning"},
		{tracking.LevelError, "error"},
		{tracking.LevelFatal, "fatal"},
	}
	for _, tc := range cases {
		if string(tc.level) != tc.want {
			t.Errorf("expected %q, got %q", tc.want, tc.level)
		}
	}
}

func TestTag(t *testing.T) {
	tag := tracking.Tag{Key: "env", Value: "production"}
	if tag.Key != "env" || tag.Value != "production" {
		t.Errorf("unexpected tag: %+v", tag)
	}
}

func TestConfig(t *testing.T) {
	cfg := tracking.Config{
		DSN:         "https://key@sentry.io/1",
		Environment: "staging",
		Release:     "v2.0.0",
	}
	if cfg.DSN == "" || cfg.Environment == "" || cfg.Release == "" {
		t.Error("config fields should not be empty")
	}
}
