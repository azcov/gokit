package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestProvider(srv *httptest.Server) *GitHub {
	g := New(Config{ClientID: "cid", ClientSecret: "sec"})
	g.endpoint.TokenURL = srv.URL + "/token"
	g.endpoint.AuthURL = srv.URL + "/auth"
	g.userURL = srv.URL + "/user"
	g.emailURL = srv.URL + "/emails"
	g.httpClient = srv.Client()
	return g
}

func TestExchange_UsesPrimaryEmail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"at","token_type":"Bearer"}`))
		case "/user":
			// no public email -> falls back to /emails
			_, _ = w.Write([]byte(`{"id":42,"login":"octocat","name":"The Octocat","email":"","avatar_url":"http://a"}`))
		case "/emails":
			_, _ = w.Write([]byte(`[{"email":"sec@x.com","primary":false,"verified":true},{"email":"prim@x.com","primary":true,"verified":true}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	g := newTestProvider(srv)
	res, err := g.Exchange(context.Background(), "code", "https://app/cb")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if res.User.ID != "42" {
		t.Errorf("ID = %q", res.User.ID)
	}
	if res.User.Email != "prim@x.com" {
		t.Errorf("expected primary email, got %q", res.User.Email)
	}
	if !res.User.EmailVerified {
		t.Error("expected verified primary email")
	}
	if res.User.Meta["login"] != "octocat" {
		t.Errorf("login meta = %v", res.User.Meta["login"])
	}
}

func TestExchange_TokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	g := newTestProvider(srv)
	if _, err := g.Exchange(context.Background(), "bad", "https://app/cb"); err == nil {
		t.Error("expected token exchange error")
	}
}
