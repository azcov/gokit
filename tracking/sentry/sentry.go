package sentry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/azcov/gokit/tracking"
)

var _ tracking.Tracker = (*Sentry)(nil)

type Sentry struct {
	cfg    tracking.Config
	client *http.Client
	dsn    *parsedDSN
}

type parsedDSN struct {
	storeURL string
	key      string
}

func parseDSN(dsn string) (*parsedDSN, error) {
	if dsn == "" {
		return nil, fmt.Errorf("sentry: empty DSN")
	}
	return &parsedDSN{storeURL: dsn, key: dsn}, nil
}

func New(cfg tracking.Config) (*Sentry, error) {
	dsn, err := parseDSN(cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &Sentry{
		cfg:    cfg,
		client: &http.Client{Timeout: 5 * time.Second},
		dsn:    dsn,
	}, nil
}

type sentryEvent struct {
	EventID     string            `json:"event_id"`
	Timestamp   string            `json:"timestamp"`
	Level       string            `json:"level"`
	Platform    string            `json:"platform"`
	Release     string            `json:"release,omitempty"`
	Environment string            `json:"environment,omitempty"`
	Message     *sentryMessage    `json:"message,omitempty"`
	Exception   *sentryException  `json:"exception,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type sentryMessage struct {
	Formatted string `json:"formatted"`
}

type sentryException struct {
	Values []sentryExceptionValue `json:"values"`
}

type sentryExceptionValue struct {
	Type       string       `json:"type"`
	Value      string       `json:"value"`
	Stacktrace *stacktrace  `json:"stacktrace,omitempty"`
}

type stacktrace struct {
	Frames []frame `json:"frames"`
}

type frame struct {
	Filename string `json:"filename"`
	Function string `json:"function"`
	Lineno   int    `json:"lineno"`
}

func (s *Sentry) CaptureError(ctx context.Context, err error, tags ...tracking.Tag) error {
	if err == nil {
		return nil
	}
	event := s.buildEvent(string(tracking.LevelError), tags...)
	event.Exception = &sentryException{
		Values: []sentryExceptionValue{
			{
				Type:       fmt.Sprintf("%T", err),
				Value:      err.Error(),
				Stacktrace: captureStack(2),
			},
		},
	}
	return s.send(ctx, event)
}

func (s *Sentry) CaptureMessage(ctx context.Context, msg string, level tracking.Level, tags ...tracking.Tag) error {
	event := s.buildEvent(string(level), tags...)
	event.Message = &sentryMessage{Formatted: msg}
	return s.send(ctx, event)
}

func (s *Sentry) Close() error { return nil }

func (s *Sentry) buildEvent(level string, tags ...tracking.Tag) *sentryEvent {
	tagMap := make(map[string]string, len(tags))
	for _, t := range tags {
		tagMap[t.Key] = t.Value
	}
	return &sentryEvent{
		EventID:     generateID(),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Level:       level,
		Platform:    "go",
		Release:     s.cfg.Release,
		Environment: s.cfg.Environment,
		Tags:        tagMap,
	}
}

func (s *Sentry) send(ctx context.Context, event *sentryEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("sentry: marshal event: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.dsn.storeURL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("sentry: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sentry-Auth", fmt.Sprintf("Sentry sentry_key=%s,sentry_version=7", s.dsn.key))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sentry: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sentry: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func captureStack(skip int) *stacktrace {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip+2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	var result []frame
	for {
		f, more := frames.Next()
		result = append(result, frame{
			Filename: f.File,
			Function: f.Function,
			Lineno:   f.Line,
		})
		if !more {
			break
		}
	}
	return &stacktrace{Frames: result}
}

func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
