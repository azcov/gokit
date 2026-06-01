package google

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestProvider points the token and userinfo endpoints at srv.
func newTestProvider(srv *httptest.Server) *Google {
	g := New(Config{ClientID: "cid", ClientSecret: "sec"})
	g.endpoint.TokenURL = srv.URL + "/token"
	g.endpoint.AuthURL = srv.URL + "/auth"
	g.userInfo = srv.URL + "/userinfo"
	g.httpClient = srv.Client()
	return g
}

func TestExchange_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"at-123","refresh_token":"rt-456","token_type":"Bearer","expires_in":3600}`))
		case "/userinfo":
			if got := r.Header.Get("Authorization"); got != "Bearer at-123" {
				t.Errorf("userinfo auth header = %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"sub":"u-1","email":"a@b.com","email_verified":true,"name":"Alice","picture":"http://img"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	g := newTestProvider(srv)
	res, err := g.Exchange(context.Background(), "auth-code", "https://app/callback")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if res.AccessToken != "at-123" || res.RefreshToken != "rt-456" {
		t.Errorf("tokens = %+v", res)
	}
	if res.User == nil || res.User.ID != "u-1" || res.User.Email != "a@b.com" {
		t.Errorf("user = %+v", res.User)
	}
	if !res.User.EmailVerified {
		t.Error("expected EmailVerified true")
	}
	if res.User.Meta["picture"] != "http://img" {
		t.Errorf("picture meta = %v", res.User.Meta["picture"])
	}
}

func TestExchange_TokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	g := newTestProvider(srv)
	if _, err := g.Exchange(context.Background(), "bad-code", "https://app/cb"); err == nil {
		t.Error("expected error on token exchange failure")
	}
}

func TestRefreshOAuth_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-at","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	g := newTestProvider(srv)
	res, err := g.RefreshOAuth(context.Background(), "rt-456")
	if err != nil {
		t.Fatalf("RefreshOAuth: %v", err)
	}
	if res.AccessToken != "new-at" {
		t.Errorf("access token = %q", res.AccessToken)
	}
}
