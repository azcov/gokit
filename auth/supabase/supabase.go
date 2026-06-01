package supabase

import (
	"context"
	"errors"

	"github.com/azcov/gokit/auth"
	"github.com/golang-jwt/jwt/v5"
)

var _ auth.Provider = (*Supabase)(nil)

type Supabase struct {
	jwtSecret []byte
}

type Config struct {
	// JWTSecret is found in Supabase project settings → API → JWT Secret.
	JWTSecret string `env:"JWT_SECRET" json:"jwt_secret" yaml:"jwt_secret"`
}

func New(cfg Config) *Supabase {
	return &Supabase{jwtSecret: []byte(cfg.JWTSecret)}
}

type supabaseClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (s *Supabase) Verify(_ context.Context, token string) (*auth.Claims, error) {
	t, err := jwt.ParseWithClaims(token, &supabaseClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*supabaseClaims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid token")
	}
	return &auth.Claims{
		UserID: c.Subject,
		Email:  c.Email,
		Roles:  []string{c.Role},
	}, nil
}

func (s *Supabase) CreateToken(_ context.Context, _ auth.Claims) (string, error) {
	return "", errors.New("supabase: use the Supabase Auth API to issue tokens")
}

func (s *Supabase) RevokeToken(_ context.Context, _ string) error {
	return errors.New("supabase: token revocation requires the Supabase Auth API")
}
