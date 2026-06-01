package betterauth

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

var _ auth.Provider = (*BetterAuth)(nil)
var _ auth.UserManager = (*BetterAuth)(nil)

// Config holds Better Auth connection settings.
// Better Auth is framework-agnostic — this provider talks to its REST API
// and verifies JWTs using the shared secret.
type Config struct {
	// BaseURL is your Better Auth server URL, e.g. https://api.example.com
	BaseURL string `config:"base_url"`
	// Secret is the Better Auth secret (BETTER_AUTH_SECRET env var on the server).
	Secret string `config:"secret"`
	// Timeout for HTTP requests.
	Timeout time.Duration `config:"timeout"`
}

type BetterAuth struct {
	cfg    Config
	client *http.Client
}

func New(cfg Config) *BetterAuth {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &BetterAuth{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

type baClaims struct {
	jwt.RegisteredClaims
	Email     string `json:"email"`
	SessionID string `json:"sessionId"`
}

// Login authenticates via Better Auth's email+password endpoint.
func (b *BetterAuth) Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error) {
	body := map[string]string{"email": req.Email, "password": req.Password}
	var resp struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
		User         *struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"user"`
		Error string `json:"message"`
	}
	if err := b.post(ctx, "/api/auth/sign-in/email", nil, body, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New("betterauth: " + resp.Error)
	}
	lr := &auth.LoginResponse{
		Token:        resp.Token,
		RefreshToken: resp.RefreshToken,
	}
	if resp.User != nil {
		lr.User = &auth.User{ID: resp.User.ID, Email: resp.User.Email, Name: resp.User.Name}
	}
	return lr, nil
}

// Logout invalidates the current session.
func (b *BetterAuth) Logout(ctx context.Context, token string) error {
	var resp map[string]any
	return b.post(ctx, "/api/auth/sign-out", &token, nil, &resp)
}

// Verify validates a Better Auth JWT using the shared secret.
func (b *BetterAuth) Verify(_ context.Context, token string) (*auth.Claims, error) {
	t, err := jwt.ParseWithClaims(token, &baClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("betterauth: unexpected signing method")
		}
		return []byte(b.cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*baClaims)
	if !ok || !t.Valid {
		return nil, errors.New("betterauth: invalid token")
	}
	return &auth.Claims{
		UserID:    c.Subject,
		Email:     c.Email,
		SessionID: c.SessionID,
		IssuedAt:  c.IssuedAt.Time,
		ExpiresAt: c.ExpiresAt.Time,
	}, nil
}

// CreateToken is not supported — Better Auth issues tokens server-side.
func (b *BetterAuth) CreateToken(_ context.Context, _ auth.Claims) (string, error) {
	return "", errors.New("betterauth: tokens are issued by the Better Auth server")
}

// RefreshToken exchanges a refresh token for a new access token.
func (b *BetterAuth) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	body := map[string]string{"refreshToken": refreshToken}
	var resp struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
		Error        string `json:"message"`
	}
	if err := b.post(ctx, "/api/auth/token", nil, body, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New("betterauth: " + resp.Error)
	}
	return &auth.TokenPair{
		AccessToken:  resp.Token,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// RevokeToken revokes the session associated with the token.
func (b *BetterAuth) RevokeToken(ctx context.Context, token string) error {
	return b.Logout(ctx, token)
}

// ── UserManager ───────────────────────────────────────────────────────────────

func (b *BetterAuth) GetUser(ctx context.Context, userID string) (*auth.User, error) {
	var u baUser
	if err := b.get(ctx, "/api/auth/admin/users/"+userID, &u); err != nil {
		return nil, err
	}
	return baUserToAuth(&u), nil
}

func (b *BetterAuth) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	var u baUser
	if err := b.get(ctx, "/api/auth/admin/users?email="+email, &u); err != nil {
		return nil, err
	}
	return baUserToAuth(&u), nil
}

func (b *BetterAuth) CreateUser(ctx context.Context, req auth.RegisterRequest) (*auth.User, error) {
	body := map[string]any{
		"email":    req.Email,
		"password": req.Password,
		"name":     req.Name,
	}
	var u baUser
	if err := b.post(ctx, "/api/auth/sign-up/email", nil, body, &u); err != nil {
		return nil, err
	}
	return baUserToAuth(&u), nil
}

func (b *BetterAuth) UpdateUser(ctx context.Context, userID string, req auth.UpdateUserRequest) (*auth.User, error) {
	body := map[string]any{}
	if req.Name != nil {
		body["name"] = *req.Name
	}
	var u baUser
	if err := b.patch(ctx, "/api/auth/admin/users/"+userID, body, &u); err != nil {
		return nil, err
	}
	return baUserToAuth(&u), nil
}

func (b *BetterAuth) DeleteUser(ctx context.Context, userID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, b.cfg.BaseURL+"/api/auth/admin/users/"+userID, nil)
	if err != nil {
		return fmt.Errorf("betterauth: delete user: %w", err)
	}
	b.setHeaders(req, nil)
	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("betterauth: delete user: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func (b *BetterAuth) ListUsers(ctx context.Context, opts auth.ListUsersOptions) ([]*auth.User, error) {
	url := fmt.Sprintf("/api/auth/admin/users?limit=%d&offset=%d", opts.Limit, opts.Offset)
	if opts.Query != "" {
		url += "&search=" + opts.Query
	}
	var list []baUser
	if err := b.get(ctx, url, &list); err != nil {
		return nil, err
	}
	users := make([]*auth.User, len(list))
	for i := range list {
		users[i] = baUserToAuth(&list[i])
	}
	return users, nil
}

// ── internal ──────────────────────────────────────────────────────────────────

type baUser struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	EmailVerified bool      `json:"emailVerified"`
	Banned        bool      `json:"banned"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func baUserToAuth(u *baUser) *auth.User {
	return &auth.User{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		EmailVerified: u.EmailVerified,
		Banned:        u.Banned,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func (b *BetterAuth) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.cfg.BaseURL+path, nil)
	if err != nil {
		return fmt.Errorf("betterauth: get: %w", err)
	}
	b.setHeaders(req, nil)
	return b.do(req, out)
}

func (b *BetterAuth) post(ctx context.Context, path string, token *string, body, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("betterauth: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.cfg.BaseURL+path, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("betterauth: post: %w", err)
	}
	b.setHeaders(req, token)
	req.Header.Set("Content-Type", "application/json")
	return b.do(req, out)
}

func (b *BetterAuth) patch(ctx context.Context, path string, body, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("betterauth: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, b.cfg.BaseURL+path, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("betterauth: patch: %w", err)
	}
	b.setHeaders(req, nil)
	req.Header.Set("Content-Type", "application/json")
	return b.do(req, out)
}

func (b *BetterAuth) setHeaders(req *http.Request, token *string) {
	req.Header.Set("x-api-key", b.cfg.Secret)
	if token != nil {
		req.Header.Set("Authorization", "Bearer "+*token)
	}
}

func (b *BetterAuth) do(req *http.Request, out any) error {
	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("betterauth: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var e struct{ Message string `json:"message"` }
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if e.Message != "" {
			return errors.New("betterauth: " + e.Message)
		}
		return fmt.Errorf("betterauth: unexpected status %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("betterauth: decode: %w", err)
		}
	}
	return nil
}
