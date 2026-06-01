package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/azcov/gokit/ai"
)

var _ ai.Provider = (*OpenAI)(nil)

const defaultBaseURL = "https://api.openai.com/v1"
const defaultModel = "gpt-4o-mini"
const defaultTimeout = 30 * time.Second

type OpenAI struct {
	cfg    ai.Config
	client *http.Client
}

func New(cfg ai.Config) *OpenAI {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	return &OpenAI{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
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
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type completionRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type completionResponse struct {
	Model   string             `json:"model"`
	Choices []completionChoice `json:"choices"`
	Usage   usageInfo          `json:"usage"`
	Error   *apiError          `json:"error,omitempty"`
}

type completionChoice struct {
	Text string `json:"text"`
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

func (o *OpenAI) Chat(ctx context.Context, messages []ai.Message) (*ai.Response, error) {
	msgs := make([]chatMessage, len(messages))
	for i, m := range messages {
		msgs[i] = chatMessage{Role: string(m.Role), Content: m.Content}
	}
	reqBody := chatRequest{Model: o.cfg.Model, Messages: msgs}
	var resp chatResponse
	if err := o.post(ctx, "/chat/completions", reqBody, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("openai: %s", resp.Error.Message)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty response")
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

func (o *OpenAI) Complete(ctx context.Context, prompt string) (*ai.Response, error) {
	return o.Chat(ctx, []ai.Message{{Role: ai.RoleUser, Content: prompt}})
}

func (o *OpenAI) Embed(ctx context.Context, text string) (*ai.Embedding, error) {
	reqBody := embedRequest{Model: "text-embedding-3-small", Input: text}
	var resp embedResponse
	if err := o.post(ctx, "/embeddings", reqBody, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("openai: %s", resp.Error.Message)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("openai: empty embedding response")
	}
	return &ai.Embedding{
		Vector: resp.Data[0].Embedding,
		Model:  resp.Model,
	}, nil
}

func (o *OpenAI) Close() error { return nil }

func (o *OpenAI) post(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("openai: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("openai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("openai: http: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("openai: decode response: %w", err)
	}
	return nil
}
