package ai

import (
	"context"
	"time"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Message struct {
	Role    Role
	Content string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type Response struct {
	Content string
	Model   string
	Usage   Usage
}

type Embedding struct {
	Vector []float64
	Model  string
}

type Config struct {
	APIKey  string        `config:"api_key"`
	Model   string        `config:"model"`
	BaseURL string        `config:"base_url"`
	Timeout time.Duration `config:"timeout"`
}

type Provider interface {
	Chat(ctx context.Context, messages []Message) (*Response, error)
	Complete(ctx context.Context, prompt string) (*Response, error)
	Embed(ctx context.Context, text string) (*Embedding, error)
	Close() error
}
