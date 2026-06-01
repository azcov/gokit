package discord

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestProvider(srv *httptest.Server) *Discord {
	d := New(Config{ClientID: "cid", ClientSecret: "sec"})
	d.endpoint.TokenURL = srv.URL + "/token"
	d.endpoint.AuthURL = srv.URL + "/auth"
	d.userInfo = srv.URL + "/me"
	d.httpClient = srv.Client()
	return d
}

func TestExchange_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"at","token_type":"Bearer"}`))
		case "/me":
			_, _ = w.Write([]byte(`{"id":"d-1","username":"bob","discriminator":"1234","email":"bob@x.com","verified":true,"avatar":"abc"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	res, err := newTestProvider(srv).Exchange(context.Background(), "code", "https://app/cb")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if res.User.ID != "d-1" || res.User.Name != "bob#1234" {
		t.Errorf("user = %+v", res.User)
	}
	if !res.User.EmailVerified {
		t.Error("expected verified")
	}
}

func TestExchange_TokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	if _, err := newTestProvider(srv).Exchange(context.Background(), "bad", "https://app/cb"); err == nil {
		t.Error("expected error")
	}
}
