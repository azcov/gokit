package auth

import (
	"context"
	"time"
)

// ── core types ────────────────────────────────────────────────────────────────

type Claims struct {
	UserID    string
	Email     string
	Phone     string
	Roles     []string
	SessionID string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Meta      map[string]any
}

func (c *Claims) IsExpired() bool {
	return !c.ExpiresAt.IsZero() && time.Now().After(c.ExpiresAt)
}

type User struct {
	ID            string
	Email         string
	Phone         string
	Name          string
	Roles         []string
	EmailVerified bool
	PhoneVerified bool
	Banned        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Meta          map[string]any
}

type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	Meta      map[string]any
}

func (s *Session) IsExpired() bool {
	return !s.ExpiresAt.IsZero() && time.Now().After(s.ExpiresAt)
}

// ── request / response types ──────────────────────────────────────────────────

// LoginRequest carries credentials. Any one identifier (Email, Phone, Username,
// or the generic Identifier) plus a secret (Password — also used for PIN or OTP
// code) is enough. Providers decide which fields they honor.
type LoginRequest struct {
	Email      string
	Phone      string
	Username   string
	Identifier string // generic catch-all when none of the above fit
	Password   string // also holds a PIN or a one-time code
	Meta       map[string]any
}

// ID returns the first non-empty identifier in priority order:
// Email, Phone, Username, Identifier.
func (r LoginRequest) ID() string {
	switch {
	case r.Email != "":
		return r.Email
	case r.Phone != "":
		return r.Phone
	case r.Username != "":
		return r.Username
	default:
		return r.Identifier
	}
}

type LoginResponse struct {
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
	User         *User
}

type RegisterRequest struct {
	Email    string
	Phone    string
	Password string
	Name     string
	Roles    []string
	Meta     map[string]any
}

type UpdateUserRequest struct {
	Name  *string
	Phone *string
	Roles []string
	Meta  map[string]any
}

type ListUsersOptions struct {
	Limit  int
	Offset int
	Query  string // search by email/name
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// ── interfaces ────────────────────────────────────────────────────────────────

// Provider is the core auth interface — token lifecycle and basic authentication.
// All providers must implement this.
type Provider interface {
	// Login authenticates credentials and returns tokens.
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	// Logout invalidates the session/token.
	Logout(ctx context.Context, token string) error
	// Verify validates a token and returns the decoded claims.
	Verify(ctx context.Context, token string) (*Claims, error)
	// CreateToken issues a signed token for the given claims.
	CreateToken(ctx context.Context, claims Claims) (string, error)
	// RefreshToken exchanges a refresh token for a new token pair.
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	// RevokeToken permanently invalidates a token or session.
	RevokeToken(ctx context.Context, token string) error
}

// UserManager is an optional interface for providers that support user CRUD.
// Check with: um, ok := provider.(auth.UserManager)
type UserManager interface {
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, req RegisterRequest) (*User, error)
	UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, opts ListUsersOptions) ([]*User, error)
}

// PasswordManager is an optional interface for providers that support
// password-based operations.
type PasswordManager interface {
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error // sends reset email
	SetPassword(ctx context.Context, userID, newPassword string) error
}

// SessionManager is an optional interface for providers that expose sessions.
type SessionManager interface {
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	ListSessions(ctx context.Context, userID string) ([]*Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID string) error
}

// ── OAuth 2.0 ─────────────────────────────────────────────────────────────────

// OAuthResult holds the result of a successful OAuth exchange.
type OAuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
	Scopes       []string
	User         *User
	RawToken     map[string]any // full token response from the provider
}

// OAuthProvider handles the OAuth 2.0 authorization code flow (social login).
// Implement this to add Google, GitHub, Facebook etc. without a third-party
// auth service.
//
// Typical flow:
//
//	// 1. Redirect user
//	http.Redirect(w, r, provider.AuthorizeURL(callbackURL, state, "email", "profile"), 302)
//
//	// 2. Handle callback
//	result, err := provider.Exchange(ctx, r.URL.Query().Get("code"), callbackURL)
//	// result.User is populated, result.AccessToken ready to use
type OAuthProvider interface {
	// AuthorizeURL returns the URL to redirect the user to for authorization.
	// state should be a random value stored in a cookie/session to prevent CSRF.
	AuthorizeURL(redirectURI, state string, scopes ...string) string
	// Exchange converts the authorization code received at the callback into
	// tokens and populates User with the provider's profile info.
	Exchange(ctx context.Context, code, redirectURI string) (*OAuthResult, error)
	// RefreshOAuth exchanges a refresh token for a fresh access token.
	RefreshOAuth(ctx context.Context, refreshToken string) (*OAuthResult, error)
}
