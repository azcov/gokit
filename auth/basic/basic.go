package basic

import (
	"context"
	"errors"
	"time"

	"github.com/azcov/gokit/auth"
	"github.com/golang-jwt/jwt/v5"
)

var _ auth.Provider = (*Basic)(nil)

type Config struct {
	Secret string        `config:"secret"`
	TTL    time.Duration `config:"ttl"`
	// LoginFunc verifies the credentials in req and returns the user's claims.
	// It receives the full LoginRequest, so it can key on Email, Phone, Username
	// or Identifier and treat Password as a password, PIN, or one-time code.
	// Required for Login() to work.
	LoginFunc func(ctx context.Context, req auth.LoginRequest) (*auth.Claims, error)
}

type Basic struct {
	secret    []byte
	ttl       time.Duration
	loginFunc func(ctx context.Context, req auth.LoginRequest) (*auth.Claims, error)
}

func New(cfg Config) *Basic {
	ttl := cfg.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &Basic{
		secret:    []byte(cfg.Secret),
		ttl:       ttl,
		loginFunc: cfg.LoginFunc,
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID    string   `json:"uid"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	SessionID string   `json:"sid,omitempty"`
}

func (b *Basic) Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error) {
	if b.loginFunc == nil {
		return nil, errors.New("basic: LoginFunc not configured")
	}
	claims, err := b.loginFunc(ctx, req)
	if err != nil {
		return nil, err
	}
	token, err := b.CreateToken(ctx, *claims)
	if err != nil {
		return nil, err
	}
	return &auth.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(b.ttl),
	}, nil
}

func (b *Basic) Logout(_ context.Context, _ string) error {
	// Stateless JWT — add token to a blocklist externally if revocation is needed.
	return nil
}

func (b *Basic) Verify(_ context.Context, token string) (*auth.Claims, error) {
	t, err := jwt.ParseWithClaims(token, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("basic: unexpected signing method")
		}
		return b.secret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*jwtClaims)
	if !ok || !t.Valid {
		return nil, errors.New("basic: invalid token")
	}
	return &auth.Claims{
		UserID:    c.UserID,
		Email:     c.Email,
		Roles:     c.Roles,
		SessionID: c.SessionID,
		IssuedAt:  c.IssuedAt.Time,
		ExpiresAt: c.ExpiresAt.Time,
	}, nil
}

func (b *Basic) CreateToken(_ context.Context, c auth.Claims) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(b.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:    c.UserID,
		Email:     c.Email,
		Roles:     c.Roles,
		SessionID: c.SessionID,
	}).SignedString(b.secret)
}

func (b *Basic) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// For stateless JWT, the refresh token IS the token (re-sign with new expiry).
	claims, err := b.Verify(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	newToken, err := b.CreateToken(ctx, *claims)
	if err != nil {
		return nil, err
	}
	return &auth.TokenPair{
		AccessToken:  newToken,
		RefreshToken: newToken,
		ExpiresAt:    time.Now().Add(b.ttl),
	}, nil
}

func (b *Basic) RevokeToken(_ context.Context, _ string) error {
	// Stateless JWT — use an external blocklist for revocation.
	return nil
}
