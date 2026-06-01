package auth

import "context"

type Claims struct {
	UserID string
	Email  string
	Roles  []string
	Meta   map[string]any
}

type Provider interface {
	Verify(ctx context.Context, token string) (*Claims, error)
	CreateToken(ctx context.Context, claims Claims) (string, error)
	RevokeToken(ctx context.Context, token string) error
}
