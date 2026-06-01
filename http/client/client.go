package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	gokithttp "github.com/azcov/gokit/http"
)

var _ gokithttp.Client = (*Client)(nil)

type Client struct {
	cfg   gokithttp.ClientConfig
	inner *http.Client
}

func New(cfg gokithttp.ClientConfig) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 500 * time.Millisecond
	}
	return &Client{
		cfg:   cfg,
		inner: &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
		body []byte
	)
	if req.Body != nil {
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("client: read body: %w", err)
		}
		_ = req.Body.Close()
	}

	delay := c.cfg.RetryDelay
	for attempt := range c.cfg.MaxRetries {
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
		for k, v := range c.cfg.Headers {
			req.Header.Set(k, v)
		}
		resp, err = c.inner.Do(req)
		if err == nil && !isRetryable(resp.StatusCode) {
			return resp, nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		if attempt < c.cfg.MaxRetries-1 {
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(delay):
				delay *= 2
			}
		}
	}
	return resp, err
}

func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("client: build get request: %w", err)
	}
	return c.Do(req)
}

func (c *Client) Post(url, contentType string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("client: build post request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.Do(req)
}

func (c *Client) Close() error { return nil }

func isRetryable(status int) bool {
	return status == http.StatusTooManyRequests ||
		status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}
