package github_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/azcov/gokit/auth/oauth/github"
)

func TestAuthorizeURL(t *testing.T) {
	p := github.New(github.Config{ClientID: "cid-123", ClientSecret: "secret-xyz"})
	raw := p.AuthorizeURL("https://app.example.com/callback", "state-abc")

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("AuthorizeURL produced an invalid URL: %v", err)
	}
	q := u.Query()
	if q.Get("client_id") != "cid-123" {
		t.Errorf("client_id = %q", q.Get("client_id"))
	}
	if q.Get("state") != "state-abc" {
		t.Errorf("state = %q", q.Get("state"))
	}
	if q.Get("redirect_uri") != "https://app.example.com/callback" {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
	if q.Get("response_type") != "code" {
		t.Errorf("response_type = %q", q.Get("response_type"))
	}
	if q.Get("scope") == "" {
		t.Error("expected default scopes to be set")
	}
}

func TestAuthorizeURL_CustomScopes(t *testing.T) {
	p := github.New(github.Config{ClientID: "cid", ClientSecret: "sec"})
	raw := p.AuthorizeURL("https://cb", "s", "custom.scope")
	if !strings.Contains(raw, "custom.scope") {
		t.Errorf("custom scope missing from %q", raw)
	}
}

func TestClose(t *testing.T) {
	// Exchange/RefreshOAuth hit live endpoints; AuthorizeURL is the unit-testable
	// surface. Compile-time interface check lives in the package source.
	_ = github.New(github.Config{})
}
