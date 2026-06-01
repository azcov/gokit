package facebook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	facebookOAuth "golang.org/x/oauth2/facebook"

	"github.com/azcov/gokit/auth"
)

var _ auth.OAuthProvider = (*Facebook)(nil)

const userInfoURL = "https://graph.facebook.com/me?fields=id,name,email,picture"

var defaultScopes = []string{"email", "public_profile"}

type Config struct {
	ClientID     string `config:"client_id"`
	ClientSecret string `config:"client_secret"`
}

type Facebook struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func New(cfg Config) *Facebook {
	return &Facebook{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *Facebook) AuthorizeURL(redirectURI, state string, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = defaultScopes
	}
	return f.config(redirectURI, scopes).AuthCodeURL(state)
}

func (f *Facebook) Exchange(ctx context.Context, code, redirectURI string) (*auth.OAuthResult, error) {
	token, err := f.config(redirectURI, defaultScopes).Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("facebook oauth: exchange: %w", err)
	}
	user, err := f.fetchUser(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	return &auth.OAuthResult{
		AccessToken: token.AccessToken,
		ExpiresAt:   token.Expiry,
		TokenType:   token.TokenType,
		User:        user,
		RawToken:    tokenToMap(token),
	}, nil
}

func (f *Facebook) RefreshOAuth(ctx context.Context, refreshToken string) (*auth.OAuthResult, error) {
	src := f.config("", defaultScopes).TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	token, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("facebook oauth: refresh: %w", err)
	}
	return &auth.OAuthResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
		RawToken:     tokenToMap(token),
	}, nil
}

func (f *Facebook) config(redirectURI string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     f.clientID,
		ClientSecret: f.clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint:     facebookOAuth.Endpoint,
	}
}

type fbUser struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture struct {
		Data struct{ URL string `json:"url"` } `json:"data"`
	} `json:"picture"`
}

func (f *Facebook) fetchUser(ctx context.Context, accessToken string) (*auth.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		userInfoURL+"&access_token="+accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("facebook oauth: build request: %w", err)
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("facebook oauth: fetch user: %w", err)
	}
	defer resp.Body.Close()

	var u fbUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("facebook oauth: decode user: %w", err)
	}
	return &auth.User{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Meta:  map[string]any{"picture": u.Picture.Data.URL},
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
