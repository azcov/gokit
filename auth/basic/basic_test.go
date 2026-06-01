package basic_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/azcov/gokit/auth"
	"github.com/azcov/gokit/auth/basic"
)

func newProvider(loginFunc func(ctx context.Context, req auth.LoginRequest) (*auth.Claims, error)) *basic.Basic {
	return basic.New(basic.Config{
		Secret:    "test-secret",
		TTL:       time.Hour,
		LoginFunc: loginFunc,
	})
}

func TestCreateTokenAndVerify(t *testing.T) {
	p := newProvider(nil)
	ctx := context.Background()

	token, err := p.CreateToken(ctx, auth.Claims{
		UserID: "u1",
		Email:  "user@example.com",
		Roles:  []string{"admin"},
	})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	claims, err := p.Verify(ctx, token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.UserID != "u1" || claims.Email != "user@example.com" {
		t.Errorf("unexpected claims: %+v", claims)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Errorf("unexpected roles: %v", claims.Roles)
	}
}

func TestVerify_InvalidToken(t *testing.T) {
	p := newProvider(nil)
	if _, err := p.Verify(context.Background(), "garbage.token.here"); err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestLogin_AllIdentifierSecretCombos(t *testing.T) {
	// loginFunc accepts email/phone/username with password or pin.
	loginFunc := func(_ context.Context, req auth.LoginRequest) (*auth.Claims, error) {
		id := req.ID()
		if id == "" {
			return nil, errors.New("no identifier")
		}
		// "secret" must equal "correct" for any channel.
		if req.Password != "correct" {
			return nil, errors.New("bad secret")
		}
		return &auth.Claims{UserID: "u-" + id}, nil
	}

	cases := []struct {
		name string
		req  auth.LoginRequest
		ok   bool
	}{
		{"email+password", auth.LoginRequest{Email: "a@b.com", Password: "correct"}, true},
		{"phone+password", auth.LoginRequest{Phone: "+628123", Password: "correct"}, true},
		{"phone+pin", auth.LoginRequest{Phone: "+628123", Password: "correct"}, true},
		{"username+password", auth.LoginRequest{Username: "bob", Password: "correct"}, true},
		{"email+pin", auth.LoginRequest{Email: "a@b.com", Password: "correct"}, true},
		{"identifier+secret", auth.LoginRequest{Identifier: "did:xyz", Password: "correct"}, true},
		{"wrong secret", auth.LoginRequest{Email: "a@b.com", Password: "wrong"}, false},
		{"no identifier", auth.LoginRequest{Password: "correct"}, false},
	}

	p := newProvider(loginFunc)
	ctx := context.Background()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := p.Login(ctx, tc.req)
			if tc.ok {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				if resp.Token == "" {
					t.Error("expected a token")
				}
				// token must verify and carry the derived user id
				claims, err := p.Verify(ctx, resp.Token)
				if err != nil {
					t.Fatalf("verify issued token: %v", err)
				}
				if claims.UserID != "u-"+tc.req.ID() {
					t.Errorf("got UserID %q", claims.UserID)
				}
			} else if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestLogin_NoLoginFunc(t *testing.T) {
	p := newProvider(nil)
	if _, err := p.Login(context.Background(), auth.LoginRequest{Email: "x@y.z"}); err == nil {
		t.Error("expected error when LoginFunc is nil")
	}
}

func TestRefreshToken(t *testing.T) {
	p := newProvider(nil)
	ctx := context.Background()

	token, _ := p.CreateToken(ctx, auth.Claims{UserID: "u1"})
	pair, err := p.RefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("expected new access token")
	}
	if _, err := p.Verify(ctx, pair.AccessToken); err != nil {
		t.Errorf("refreshed token invalid: %v", err)
	}
}

func TestLogoutAndRevoke_NoOp(t *testing.T) {
	p := newProvider(nil)
	ctx := context.Background()
	if err := p.Logout(ctx, "any"); err != nil {
		t.Errorf("Logout: %v", err)
	}
	if err := p.RevokeToken(ctx, "any"); err != nil {
		t.Errorf("RevokeToken: %v", err)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ auth.Provider = (*basic.Basic)(nil)
}
