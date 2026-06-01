package antrophic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/azcov/gokit/ai"
)

var _ ai.Provider = (*Anthropic)(nil)

const (
	defaultBaseURL        = "https://api.anthropic.com/v1"
	defaultModel          = "claude-3-5-sonnet-20241022"
	defaultMaxTokens      = 1024
	defaultTimeout        = 30 * time.Second
	anthropicVersion      = "2023-06-01"
	anthropicVersionHeader = "anthropic-version"
	apiKeyHeader          = "x-api-key"
)

type Anthropic struct {
	cfg    ai.Config
	client *http.Client
}

func New(cfg ai.Config) *Anthropic {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	return &Anthropic{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

type messagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
	System    string    `json:"system,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Content []content `json:"content"`
	Usage   usage     `json:"usage"`
	Error   *apiError `json:"error,omitempty"`
}

type content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (a *Anthropic) Chat(ctx context.Context, messages []ai.Message) (*ai.Response, error) {
	var systemMsg string
	msgs := make([]message, 0, len(messages))
	for _, m := range messages {
		if m.Role == ai.RoleSystem {
			systemMsg = m.Content
			continue
		}
		msgs = append(msgs, message{Role: string(m.Role), Content: m.Content})
	}

	req := messagesRequest{
		Model:     a.cfg.Model,
		MaxTokens: defaultMaxTokens,
		Messages:  msgs,
		System:    systemMsg,
	}

	var resp messagesResponse
	if err := a.post(ctx, "/messages", req, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("anthropic: %s: %s", resp.Error.Type, resp.Error.Message)
	}
	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("anthropic: empty response")
	}

	return &ai.Response{
		Content: resp.Content[0].Text,
		Model:   resp.Model,
		Usage: ai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}, nil
}

func (a *Anthropic) Complete(ctx context.Context, prompt string) (*ai.Response, error) {
	return a.Chat(ctx, []ai.Message{{Role: ai.RoleUser, Content: prompt}})
}

// Embed is not supported by Anthropic — use a dedicated embedding model instead.
func (a *Anthropic) Embed(_ context.Context, _ string) (*ai.Embedding, error) {
	return nil, fmt.Errorf("anthropic: embedding not supported; use a dedicated embedding provider")
}

func (a *Anthropic) Close() error { return nil }

func (a *Anthropic) post(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("anthropic: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("anthropic: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(apiKeyHeader, a.cfg.APIKey)
	req.Header.Set(anthropicVersionHeader, anthropicVersion)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("anthropic: http: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("anthropic: decode: %w", err)
	}
	return nil
}
