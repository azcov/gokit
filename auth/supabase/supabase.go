package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/azcov/gokit/auth"
	"github.com/golang-jwt/jwt/v5"
)

var _ auth.Provider = (*Supabase)(nil)

type Config struct {
	// JWTSecret is found in Supabase project settings → API → JWT Secret.
	JWTSecret string `config:"jwt_secret"`
	// ProjectURL is your Supabase project URL, e.g. https://abc.supabase.co
	ProjectURL string `config:"project_url"`
	// ServiceRoleKey is needed for admin user operations (optional).
	ServiceRoleKey string `config:"service_role_key"`
}

type Supabase struct {
	jwtSecret      []byte
	projectURL     string
	serviceRoleKey string
	client         *http.Client
}

func New(cfg Config) *Supabase {
	return &Supabase{
		jwtSecret:      []byte(cfg.JWTSecret),
		projectURL:     cfg.ProjectURL,
		serviceRoleKey: cfg.ServiceRoleKey,
		client:         &http.Client{Timeout: 10 * time.Second},
	}
}

type supabaseClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
	Phone string `json:"phone"`
}

// ── Provider ──────────────────────────────────────────────────────────────────

func (s *Supabase) Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error) {
	body := map[string]string{"email": req.Email, "password": req.Password}
	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error_description"`
	}
	if err := s.post(ctx, "/auth/v1/token?grant_type=password", body, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New("supabase: " + resp.Error)
	}
	return &auth.LoginResponse{
		Token:        resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}, nil
}

func (s *Supabase) Logout(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.projectURL+"/auth/v1/logout", nil)
	if err != nil {
		return fmt.Errorf("supabase: logout: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", s.serviceRoleKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("supabase: logout: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func (s *Supabase) Verify(_ context.Context, token string) (*auth.Claims, error) {
	t, err := jwt.ParseWithClaims(token, &supabaseClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("supabase: unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*supabaseClaims)
	if !ok || !t.Valid {
		return nil, errors.New("supabase: invalid token")
	}
	return &auth.Claims{
		UserID:    c.Subject,
		Email:     c.Email,
		Phone:     c.Phone,
		Roles:     []string{c.Role},
		IssuedAt:  c.IssuedAt.Time,
		ExpiresAt: c.ExpiresAt.Time,
	}, nil
}

func (s *Supabase) CreateToken(_ context.Context, _ auth.Claims) (string, error) {
	return "", errors.New("supabase: use the Supabase Auth API to issue tokens")
}

func (s *Supabase) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	body := map[string]string{"refresh_token": refreshToken}
	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error_description"`
	}
	if err := s.post(ctx, "/auth/v1/token?grant_type=refresh_token", body, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New("supabase: " + resp.Error)
	}
	return &auth.TokenPair{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}, nil
}

func (s *Supabase) RevokeToken(ctx context.Context, token string) error {
	return s.Logout(ctx, token)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *Supabase) post(ctx context.Context, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("supabase: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.projectURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("supabase: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.serviceRoleKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("supabase: http: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("supabase: decode: %w", err)
	}
	return nil
}
