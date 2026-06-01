package posthog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/azcov/gokit/analytics"
)

var _ analytics.Tracker = (*PostHog)(nil)

const defaultEndpoint = "https://app.posthog.com"

type PostHog struct {
	cfg    analytics.Config
	client *http.Client
}

func New(cfg analytics.Config) *PostHog {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	return &PostHog{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type capturePayload struct {
	APIKey     string         `json:"api_key"`
	Event      string         `json:"event"`
	DistinctID string         `json:"distinct_id"`
	Properties map[string]any `json:"properties,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

type identifyPayload struct {
	APIKey     string         `json:"api_key"`
	Event      string         `json:"event"`
	DistinctID string         `json:"distinct_id"`
	Properties map[string]any `json:"properties,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

func (p *PostHog) Track(ctx context.Context, event analytics.Event) error {
	payload := capturePayload{
		APIKey:     p.cfg.APIKey,
		Event:      event.Name,
		DistinctID: event.UserID,
		Properties: event.Properties,
		Timestamp:  time.Now().UTC(),
	}
	return p.send(ctx, "/capture/", payload)
}

func (p *PostHog) Identify(ctx context.Context, trait analytics.Trait) error {
	payload := identifyPayload{
		APIKey:     p.cfg.APIKey,
		Event:      "$identify",
		DistinctID: trait.UserID,
		Properties: trait.Properties,
		Timestamp:  time.Now().UTC(),
	}
	return p.send(ctx, "/capture/", payload)
}

func (p *PostHog) Page(ctx context.Context, page analytics.PageView) error {
	props := make(map[string]any)
	for k, v := range page.Properties {
		props[k] = v
	}
	props["$current_url"] = page.URL
	props["name"] = page.Name
	payload := capturePayload{
		APIKey:     p.cfg.APIKey,
		Event:      "$pageview",
		DistinctID: page.UserID,
		Properties: props,
		Timestamp:  time.Now().UTC(),
	}
	return p.send(ctx, "/capture/", payload)
}

func (p *PostHog) Close() error { return nil }

func (p *PostHog) send(ctx context.Context, path string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("posthog: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.Endpoint+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("posthog: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("posthog: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("posthog: unexpected status %d", resp.StatusCode)
	}
	return nil
}
