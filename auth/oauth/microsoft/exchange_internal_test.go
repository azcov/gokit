package microsoft

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestProvider(srv *httptest.Server) *Microsoft {
	m := New(Config{ClientID: "cid", ClientSecret: "sec"})
	m.endpoint.TokenURL = srv.URL + "/token"
	m.endpoint.AuthURL = srv.URL + "/auth"
	m.userInfo = srv.URL + "/me"
	m.httpClient = srv.Client()
	return m
}

func TestExchange_FallsBackToUPN(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"at","token_type":"Bearer"}`))
		case "/me":
			// no mail -> falls back to userPrincipalName
			_, _ = w.Write([]byte(`{"id":"m-1","displayName":"Carol","mail":"","userPrincipalName":"carol@corp.com"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	res, err := newTestProvider(srv).Exchange(context.Background(), "code", "https://app/cb")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if res.User.Email != "carol@corp.com" {
		t.Errorf("expected UPN fallback, got %q", res.User.Email)
	}
	if res.User.Name != "Carol" {
		t.Errorf("name = %q", res.User.Name)
	}
}

func TestExchange_TokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()
	if _, err := newTestProvider(srv).Exchange(context.Background(), "bad", "https://app/cb"); err == nil {
		t.Error("expected error")
	}
}
