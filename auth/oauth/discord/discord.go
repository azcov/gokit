package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/azcov/gokit/auth"
)

var _ auth.OAuthProvider = (*Discord)(nil)

var endpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

const userInfoURL = "https://discord.com/api/users/@me"

var defaultScopes = []string{"identify", "email"}

type Config struct {
	ClientID     string `config:"client_id"`
	ClientSecret string `config:"client_secret"`
}

type Discord struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func New(cfg Config) *Discord {
	return &Discord{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *Discord) AuthorizeURL(redirectURI, state string, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = defaultScopes
	}
	return d.config(redirectURI, scopes).AuthCodeURL(state)
}

func (d *Discord) Exchange(ctx context.Context, code, redirectURI string) (*auth.OAuthResult, error) {
	token, err := d.config(redirectURI, defaultScopes).Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("discord oauth: exchange: %w", err)
	}
	user, err := d.fetchUser(ctx, token.AccessToken)
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

func (d *Discord) RefreshOAuth(ctx context.Context, refreshToken string) (*auth.OAuthResult, error) {
	src := d.config("", defaultScopes).TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	token, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("discord oauth: refresh: %w", err)
	}
	return &auth.OAuthResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
		RawToken:     tokenToMap(token),
	}, nil
}

func (d *Discord) config(redirectURI string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     d.clientID,
		ClientSecret: d.clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint:     endpoint,
	}
}

type discordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Avatar        string `json:"avatar"`
}

func (d *Discord) fetchUser(ctx context.Context, accessToken string) (*auth.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("discord oauth: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discord oauth: fetch user: %w", err)
	}
	defer resp.Body.Close()

	var u discordUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("discord oauth: decode user: %w", err)
	}
	name := u.Username
	if u.Discriminator != "" && u.Discriminator != "0" {
		name += "#" + u.Discriminator
	}
	return &auth.User{
		ID:            u.ID,
		Email:         u.Email,
		Name:          name,
		EmailVerified: u.Verified,
		Meta:          map[string]any{"avatar": u.Avatar, "username": u.Username},
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
