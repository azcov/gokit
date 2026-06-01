package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/azcov/gokit/ai"
)

var _ ai.Provider = (*OpenRouter)(nil)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1"
	defaultModel   = "openai/gpt-4o-mini"
	defaultTimeout = 60 * time.Second
)

// OpenRouter provides access to 100+ AI models through a single unified API.
// See https://openrouter.ai/docs for available model IDs.
type OpenRouter struct {
	cfg      ai.Config
	client   *http.Client
	referer  string
	appTitle string
}

// Config extends ai.Config with OpenRouter-specific optional fields.
type Config struct {
	ai.Config
	// Referer is sent as HTTP-Referer header (shown in openrouter.ai rankings).
	Referer string
	// AppTitle is sent as X-Title header (shown in openrouter.ai rankings).
	AppTitle string
}

func New(cfg Config) *OpenRouter {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	return &OpenRouter{
		cfg:      cfg.Config,
		client:   &http.Client{Timeout: cfg.Timeout},
		referer:  cfg.Referer,
		appTitle: cfg.AppTitle,
	}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   usageInfo    `json:"usage"`
	Error   *apiError    `json:"error,omitempty"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type usageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Data  []embedData `json:"data"`
	Model string      `json:"model"`
	Error *apiError   `json:"error,omitempty"`
}

type embedData struct {
	Embedding []float64 `json:"embedding"`
}

func (o *OpenRouter) Chat(ctx context.Context, messages []ai.Message) (*ai.Response, error) {
	msgs := make([]chatMessage, len(messages))
	for i, m := range messages {
		msgs[i] = chatMessage{Role: string(m.Role), Content: m.Content}
	}
	var resp chatResponse
	if err := o.post(ctx, "/chat/completions", chatRequest{Model: o.cfg.Model, Messages: msgs}, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("openrouter: %s", resp.Error.Message)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openrouter: empty response")
	}
	return &ai.Response{
		Content: resp.Choices[0].Message.Content,
		Model:   resp.Model,
		Usage: ai.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (o *OpenRouter) Complete(ctx context.Context, prompt string) (*ai.Response, error) {
	return o.Chat(ctx, []ai.Message{{Role: ai.RoleUser, Content: prompt}})
}

func (o *OpenRouter) Embed(ctx context.Context, text string) (*ai.Embedding, error) {
	// Use an embedding-capable model via OpenRouter.
	model := "openai/text-embedding-3-small"
	var resp embedResponse
	if err := o.post(ctx, "/embeddings", embedRequest{Model: model, Input: text}, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("openrouter: %s", resp.Error.Message)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("openrouter: empty embedding")
	}
	return &ai.Embedding{Vector: resp.Data[0].Embedding, Model: resp.Model}, nil
}

func (o *OpenRouter) Close() error { return nil }

func (o *OpenRouter) post(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("openrouter: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("openrouter: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)
	if o.referer != "" {
		req.Header.Set("HTTP-Referer", o.referer)
	}
	if o.appTitle != "" {
		req.Header.Set("X-Title", o.appTitle)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("openrouter: http: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("openrouter: decode: %w", err)
	}
	return nil
}
