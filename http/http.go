package http

import (
	"net/http"
	"time"
)

type ClientConfig struct {
	BaseURL    string        `config:"base_url"`
	Timeout    time.Duration `config:"timeout"`
	MaxRetries int           `config:"max_retries"`
	RetryDelay time.Duration `config:"retry_delay"`
	Headers    map[string]string
}

type Client interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body []byte) (*http.Response, error)
	Close() error
}
