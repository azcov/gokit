package clerk

import (
	"context"
	"errors"
	"time"

	"github.com/azcov/gokit/auth"
	clerkSDK "github.com/clerk/clerk-sdk-go/v2"
	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"
	clerksession "github.com/clerk/clerk-sdk-go/v2/session"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
)

var _ auth.Provider = (*Clerk)(nil)
var _ auth.UserManager = (*Clerk)(nil)

type Config struct {
	SecretKey string `config:"secret_key"`
}

type Clerk struct{}

func New(cfg Config) *Clerk {
	clerkSDK.SetKey(cfg.SecretKey)
	return &Clerk{}
}

// Login is not supported — Clerk tokens are issued by Clerk's frontend SDK (SignIn).
func (c *Clerk) Login(_ context.Context, _ auth.LoginRequest) (*auth.LoginResponse, error) {
	return nil, errors.New("clerk: use the Clerk frontend SDK to sign in and obtain tokens")
}

// Logout revokes the session associated with the given session ID.
func (c *Clerk) Logout(ctx context.Context, sessionID string) error {
	_, err := clerksession.Revoke(ctx, &clerksession.RevokeParams{ID: sessionID})
	return err
}

func (c *Clerk) Verify(ctx context.Context, token string) (*auth.Claims, error) {
	claims, err := clerkjwt.Verify(ctx, &clerkjwt.VerifyParams{Token: token})
	if err != nil {
		return nil, err
	}
	ac := &auth.Claims{
		UserID:    claims.Subject,
		SessionID: claims.SessionID,
	}
	if claims.Expiry != nil {
		ac.ExpiresAt = time.Unix(*claims.Expiry, 0)
	}
	if claims.IssuedAt != nil {
		ac.IssuedAt = time.Unix(*claims.IssuedAt, 0)
	}
	return ac, nil
}

// CreateToken is not supported — tokens are issued by Clerk's frontend SDK.
func (c *Clerk) CreateToken(_ context.Context, _ auth.Claims) (string, error) {
	return "", errors.New("clerk: tokens are issued by Clerk's frontend SDK")
}

// RefreshToken is not supported — Clerk handles token refresh on the frontend.
func (c *Clerk) RefreshToken(_ context.Context, _ string) (*auth.TokenPair, error) {
	return nil, errors.New("clerk: token refresh is handled by Clerk's frontend SDK")
}

// RevokeToken revokes the Clerk session by session ID.
func (c *Clerk) RevokeToken(ctx context.Context, sessionID string) error {
	return c.Logout(ctx, sessionID)
}

// ── UserManager ───────────────────────────────────────────────────────────────

func (c *Clerk) GetUser(ctx context.Context, userID string) (*auth.User, error) {
	u, err := clerkuser.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return clerkUserToAuth(u), nil
}

func (c *Clerk) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	list, err := clerkuser.List(ctx, &clerkuser.ListParams{
		EmailAddresses: []string{email},
	})
	if err != nil {
		return nil, err
	}
	if len(list.Users) == 0 {
		return nil, errors.New("clerk: user not found")
	}
	return clerkUserToAuth(list.Users[0]), nil
}

func (c *Clerk) CreateUser(ctx context.Context, req auth.RegisterRequest) (*auth.User, error) {
	addrs := []string{req.Email}
	params := &clerkuser.CreateParams{
		EmailAddresses: &addrs,
		Password:       &req.Password,
	}
	if req.Name != "" {
		params.FirstName = &req.Name
	}
	u, err := clerkuser.Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return clerkUserToAuth(u), nil
}

func (c *Clerk) UpdateUser(ctx context.Context, userID string, req auth.UpdateUserRequest) (*auth.User, error) {
	params := &clerkuser.UpdateParams{}
	if req.Name != nil {
		params.FirstName = req.Name
	}
	u, err := clerkuser.Update(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	return clerkUserToAuth(u), nil
}

func (c *Clerk) DeleteUser(ctx context.Context, userID string) error {
	_, err := clerkuser.Delete(ctx, userID)
	return err
}

func (c *Clerk) ListUsers(ctx context.Context, opts auth.ListUsersOptions) ([]*auth.User, error) {
	params := &clerkuser.ListParams{}
	if opts.Limit > 0 {
		limit := int64(opts.Limit)
		params.Limit = &limit
	}
	if opts.Offset > 0 {
		offset := int64(opts.Offset)
		params.Offset = &offset
	}
	if opts.Query != "" {
		params.Query = &opts.Query
	}
	list, err := clerkuser.List(ctx, params)
	if err != nil {
		return nil, err
	}
	users := make([]*auth.User, len(list.Users))
	for i, u := range list.Users {
		users[i] = clerkUserToAuth(u)
	}
	return users, nil
}

func clerkUserToAuth(u *clerkSDK.User) *auth.User {
	user := &auth.User{ID: u.ID, Banned: u.Banned}
	if len(u.EmailAddresses) > 0 {
		user.Email = u.EmailAddresses[0].EmailAddress
		user.EmailVerified = u.EmailAddresses[0].Verification != nil &&
			u.EmailAddresses[0].Verification.Status == "verified"
	}
	if u.FirstName != nil {
		user.Name = *u.FirstName
	}
	if u.LastName != nil && *u.LastName != "" {
		if user.Name != "" {
			user.Name += " "
		}
		user.Name += *u.LastName
	}
	return user
}
