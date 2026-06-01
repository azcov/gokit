package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	googleOAuth "golang.org/x/oauth2/google"

	"github.com/azcov/gokit/auth"
)

var _ auth.OAuthProvider = (*Google)(nil)

const userInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"

var defaultScopes = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
	"openid",
}

type Config struct {
	ClientID     string `config:"client_id"`
	ClientSecret string `config:"client_secret"`
}

type Google struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	// endpoint and userInfo are overridable so tests can point at a local
	// server; they default to Google's production endpoints.
	endpoint oauth2.Endpoint
	userInfo string
}

func New(cfg Config) *Google {
	return &Google{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		endpoint:     googleOAuth.Endpoint,
		userInfo:     userInfoURL,
	}
}

func (g *Google) AuthorizeURL(redirectURI, state string, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = defaultScopes
	}
	return g.config(redirectURI, scopes).AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

func (g *Google) Exchange(ctx context.Context, code, redirectURI string) (*auth.OAuthResult, error) {
	token, err := g.config(redirectURI, defaultScopes).Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google oauth: exchange: %w", err)
	}
	user, err := g.fetchUser(ctx, token.AccessToken)
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

func (g *Google) RefreshOAuth(ctx context.Context, refreshToken string) (*auth.OAuthResult, error) {
	src := g.config("", defaultScopes).TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	token, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("google oauth: refresh: %w", err)
	}
	return &auth.OAuthResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
		RawToken:     tokenToMap(token),
	}, nil
}

func (g *Google) config(redirectURI string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.clientID,
		ClientSecret: g.clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint:     g.endpoint,
	}
}

type googleUser struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (g *Google) fetchUser(ctx context.Context, accessToken string) (*auth.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.userInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("google oauth: user info request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google oauth: fetch user: %w", err)
	}
	defer resp.Body.Close()

	var u googleUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("google oauth: decode user: %w", err)
	}
	return &auth.User{
		ID:            u.Sub,
		Email:         u.Email,
		Name:          u.Name,
		EmailVerified: u.EmailVerified,
		Meta:          map[string]any{"picture": u.Picture},
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
