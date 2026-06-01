package microsoft

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	microsoftOAuth "golang.org/x/oauth2/microsoft"

	"github.com/azcov/gokit/auth"
)

var _ auth.OAuthProvider = (*Microsoft)(nil)

const userInfoURL = "https://graph.microsoft.com/v1.0/me"

var defaultScopes = []string{"User.Read", "openid", "profile", "email", "offline_access"}

type Config struct {
	ClientID     string `config:"client_id"`
	ClientSecret string `config:"client_secret"`
	// TenantID scopes login to a specific Azure AD tenant.
	// Use "common" (default) to allow personal + work accounts.
	TenantID string `config:"tenant_id"`
}

type Microsoft struct {
	clientID     string
	clientSecret string
	tenantID     string
	httpClient   *http.Client
	endpoint     oauth2.Endpoint
	userInfo     string
}

func New(cfg Config) *Microsoft {
	if cfg.TenantID == "" {
		cfg.TenantID = "common"
	}
	return &Microsoft{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		tenantID:     cfg.TenantID,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		endpoint:     microsoftOAuth.AzureADEndpoint(cfg.TenantID),
		userInfo:     userInfoURL,
	}
}

func (m *Microsoft) AuthorizeURL(redirectURI, state string, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = defaultScopes
	}
	return m.config(redirectURI, scopes).AuthCodeURL(state)
}

func (m *Microsoft) Exchange(ctx context.Context, code, redirectURI string) (*auth.OAuthResult, error) {
	token, err := m.config(redirectURI, defaultScopes).Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("microsoft oauth: exchange: %w", err)
	}
	user, err := m.fetchUser(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	return &auth.OAuthResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
		User:         user,
		RawToken:     tokenToMap(token),
	}, nil
}

func (m *Microsoft) RefreshOAuth(ctx context.Context, refreshToken string) (*auth.OAuthResult, error) {
	src := m.config("", defaultScopes).TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	token, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("microsoft oauth: refresh: %w", err)
	}
	return &auth.OAuthResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
		RawToken:     tokenToMap(token),
	}, nil
}

func (m *Microsoft) config(redirectURI string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     m.clientID,
		ClientSecret: m.clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint:     m.endpoint,
	}
}

type msUser struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
}

func (m *Microsoft) fetchUser(ctx context.Context, accessToken string) (*auth.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.userInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft oauth: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("microsoft oauth: fetch user: %w", err)
	}
	defer resp.Body.Close()

	var u msUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("microsoft oauth: decode user: %w", err)
	}
	email := u.Mail
	if email == "" {
		email = u.UserPrincipalName
	}
	return &auth.User{
		ID:    u.ID,
		Email: email,
		Name:  u.DisplayName,
	}, nil
}

func tokenToMap(t *oauth2.Token) map[string]any {
	return map[string]any{
		"access_token":  t.AccessToken,
		"refresh_token": t.RefreshToken,
		"expiry":        t.Expiry,
		"token_type":    t.TokenType,
	}
}
