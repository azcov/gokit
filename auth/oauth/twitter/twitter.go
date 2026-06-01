package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/azcov/gokit/auth"
)

var _ auth.OAuthProvider = (*Twitter)(nil)

// Twitter OAuth 2.0 (PKCE) endpoint.
var endpoint = oauth2.Endpoint{
	AuthURL:   "https://twitter.com/i/oauth2/authorize",
	TokenURL:  "https://api.twitter.com/2/oauth2/token",
	AuthStyle: oauth2.AuthStyleInHeader,
}

const userInfoURL = "https://api.twitter.com/2/users/me?user.fields=name,username,profile_image_url"

var defaultScopes = []string{"tweet.read", "users.read", "offline.access"}

type Config struct {
	ClientID     string `config:"client_id"`
	ClientSecret string `config:"client_secret"`
}

type Twitter struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func New(cfg Config) *Twitter {
	return &Twitter{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Twitter) AuthorizeURL(redirectURI, state string, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = defaultScopes
	}
	return t.config(redirectURI, scopes).AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", "challenge"),
		oauth2.SetAuthURLParam("code_challenge_method", "plain"),
	)
}

func (t *Twitter) Exchange(ctx context.Context, code, redirectURI string) (*auth.OAuthResult, error) {
	token, err := t.config(redirectURI, defaultScopes).Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", "challenge"),
	)
	if err != nil {
		return nil, fmt.Errorf("twitter oauth: exchange: %w", err)
	}
	user, err := t.fetchUser(ctx, token.AccessToken)
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

func (t *Twitter) RefreshOAuth(ctx context.Context, refreshToken string) (*auth.OAuthResult, error) {
	src := t.config("", defaultScopes).TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	token, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("twitter oauth: refresh: %w", err)
	}
	return &auth.OAuthResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
		RawToken:     tokenToMap(token),
	}, nil
}

func (t *Twitter) config(redirectURI string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     t.clientID,
		ClientSecret: t.clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint:     endpoint,
	}
}

type twitterResponse struct {
	Data struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Username        string `json:"username"`
		ProfileImageURL string `json:"profile_image_url"`
	} `json:"data"`
}

func (t *Twitter) fetchUser(ctx context.Context, accessToken string) (*auth.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("twitter oauth: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twitter oauth: fetch user: %w", err)
	}
	defer resp.Body.Close()

	var u twitterResponse
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("twitter oauth: decode user: %w", err)
	}
	return &auth.User{
		ID:   u.Data.ID,
		Name: u.Data.Name,
		Meta: map[string]any{
			"username":          u.Data.Username,
			"profile_image_url": u.Data.ProfileImageURL,
		},
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
