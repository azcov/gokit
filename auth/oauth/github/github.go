package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"

	"github.com/azcov/gokit/auth"
)

var _ auth.OAuthProvider = (*GitHub)(nil)

const (
	userURL  = "https://api.github.com/user"
	emailURL = "https://api.github.com/user/emails"
)

var defaultScopes = []string{"read:user", "user:email"}

type Config struct {
	ClientID     string `config:"client_id"`
	ClientSecret string `config:"client_secret"`
}

type GitHub struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	endpoint     oauth2.Endpoint
	userURL      string
	emailURL     string
}

func New(cfg Config) *GitHub {
	return &GitHub{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		endpoint:     githubOAuth.Endpoint,
		userURL:      userURL,
		emailURL:     emailURL,
	}
}

func (g *GitHub) AuthorizeURL(redirectURI, state string, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = defaultScopes
	}
	return g.config(redirectURI, scopes).AuthCodeURL(state)
}

func (g *GitHub) Exchange(ctx context.Context, code, redirectURI string) (*auth.OAuthResult, error) {
	token, err := g.config(redirectURI, defaultScopes).Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("github oauth: exchange: %w", err)
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

func (g *GitHub) RefreshOAuth(ctx context.Context, refreshToken string) (*auth.OAuthResult, error) {
	src := g.config("", defaultScopes).TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	token, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("github oauth: refresh: %w", err)
	}
	return &auth.OAuthResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
		RawToken:     tokenToMap(token),
	}, nil
}

func (g *GitHub) config(redirectURI string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.clientID,
		ClientSecret: g.clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint:     g.endpoint,
	}
}

type githubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (g *GitHub) fetchUser(ctx context.Context, token string) (*auth.User, error) {
	var gu githubUser
	if err := g.get(ctx, token, g.userURL, &gu); err != nil {
		return nil, fmt.Errorf("github oauth: fetch user: %w", err)
	}

	email := gu.Email
	emailVerified := false
	if email == "" {
		var emails []githubEmail
		if err := g.get(ctx, token, g.emailURL, &emails); err == nil {
			for _, e := range emails {
				if e.Primary {
					email = e.Email
					emailVerified = e.Verified
					break
				}
			}
		}
	}

	return &auth.User{
		ID:            strconv.Itoa(gu.ID),
		Email:         email,
		Name:          gu.Name,
		EmailVerified: emailVerified,
		Meta:          map[string]any{"login": gu.Login, "avatar_url": gu.AvatarURL},
	}, nil
}

func (g *GitHub) get(ctx context.Context, token, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func tokenToMap(t *oauth2.Token) map[string]any {
	return map[string]any{
		"access_token":  t.AccessToken,
		"refresh_token": t.RefreshToken,
		"expiry":        t.Expiry,
		"token_type":    t.TokenType,
	}
}
