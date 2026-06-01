package basic

import (
	"context"
	"errors"
	"time"

	"github.com/azcov/gokit/auth"
	"github.com/golang-jwt/jwt/v5"
)

var _ auth.Provider = (*Basic)(nil)

type Basic struct {
	secret []byte
	ttl    time.Duration
}

type Config struct {
	Secret string        `env:"SECRET" json:"secret" yaml:"secret"`
	TTL    time.Duration `env:"TTL" json:"ttl" yaml:"ttl"`
}

func New(cfg Config) *Basic {
	return &Basic{secret: []byte(cfg.Secret), ttl: cfg.TTL}
}

type claims struct {
	jwt.RegisteredClaims
	UserID string   `json:"uid"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}

func (b *Basic) Verify(_ context.Context, token string) (*auth.Claims, error) {
	t, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return b.secret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid token")
	}
	return &auth.Claims{UserID: c.UserID, Email: c.Email, Roles: c.Roles}, nil
}

func (b *Basic) CreateToken(_ context.Context, c auth.Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(b.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: c.UserID,
		Email:  c.Email,
		Roles:  c.Roles,
	}).SignedString(b.secret)
}

// RevokeToken is a no-op for stateless JWTs; use an external blocklist if needed.
func (b *Basic) RevokeToken(_ context.Context, _ string) error { return nil }
