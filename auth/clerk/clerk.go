package clerk

import (
	"context"
	"errors"

	"github.com/azcov/gokit/auth"
	"github.com/clerk/clerk-sdk-go/v2"
	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"
)

var _ auth.Provider = (*Clerk)(nil)

type Clerk struct{}

type Config struct {
	SecretKey string `env:"SECRET_KEY" json:"secret_key" yaml:"secret_key"`
}

func New(cfg Config) *Clerk {
	clerk.SetKey(cfg.SecretKey)
	return &Clerk{}
}

func (c *Clerk) Verify(ctx context.Context, token string) (*auth.Claims, error) {
	claims, err := clerkjwt.Verify(ctx, &clerkjwt.VerifyParams{Token: token})
	if err != nil {
		return nil, err
	}
	// Email is not in standard Clerk JWTs; add it via a custom JWT template if needed.
	return &auth.Claims{
		UserID: claims.Subject,
	}, nil
}

func (c *Clerk) CreateToken(_ context.Context, _ auth.Claims) (string, error) {
	return "", errors.New("clerk: tokens are issued by Clerk's frontend SDK")
}

func (c *Clerk) RevokeToken(_ context.Context, _ string) error {
	return errors.New("clerk: session revocation is managed by Clerk's dashboard or API")
}
