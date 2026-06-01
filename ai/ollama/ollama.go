package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/azcov/gokit/ai"
)

var _ ai.Provider = (*Ollama)(nil)

const (
	defaultBaseURL = "http://localhost:11434/v1"
	defaultModel   = "llama3"
	defaultTimeout = 120 * time.Second // local models can be slow
)

// Ollama connects to a locally-running Ollama instance via its OpenAI-compatible API.
// Run a model first: `ollama pull llama3`
type Ollama struct {
	cfg    ai.Config
	client *http.Client
}

func New(cfg ai.Config) *Ollama {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	return &Ollama{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Model   string     `json:"model"`
	Choices []choice   `json:"choices"`
	Usage   usageStats `json:"usage"`
	Error   *apiError  `json:"error,omitempty"`
}

type choice struct {
	Message chatMessage `json:"message"`
}

type usageStats struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type apiError struct {
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

func (o *Ollama) Chat(ctx context.Context, messages []ai.Message) (*ai.Response, error) {
	msgs := make([]chatMessage, len(messages))
	for i, m := range messages {
		msgs[i] = chatMessage{Role: string(m.Role), Content: m.Content}
	}
	var resp chatResponse
	if err := o.post(ctx, "/chat/completions", chatRequest{Model: o.cfg.Model, Messages: msgs}, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("ollama: %s", resp.Error.Message)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("ollama: empty response")
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

func (o *Ollama) Complete(ctx context.Context, prompt string) (*ai.Response, error) {
	return o.Chat(ctx, []ai.Message{{Role: ai.RoleUser, Content: prompt}})
}

func (o *Ollama) Embed(ctx context.Context, text string) (*ai.Embedding, error) {
	var resp embedResponse
	if err := o.post(ctx, "/embeddings", embedRequest{Model: o.cfg.Model, Input: text}, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("ollama: %s", resp.Error.Message)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("ollama: empty embedding response")
	}
	return &ai.Embedding{Vector: resp.Data[0].Embedding, Model: resp.Model}, nil
}

func (o *Ollama) Close() error { return nil }

func (o *Ollama) post(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("ollama: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("ollama: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// No auth header — Ollama runs locally without auth by default.

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama: http: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("ollama: decode: %w", err)
	}
	return nil
}
