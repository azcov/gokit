// Package otp implements one-time-password (passwordless) authentication.
//
// It is channel-agnostic: it imports neither email, sms, nor notification.
// You provide a Sender (any func that delivers the code) and a Store (any code
// store, e.g. cache.Cache). The SDK weight of your delivery channel stays in
// your application, not in this package.
//
//	o := otp.New(otp.Config{
//	    Store: myStore,
//	    Send: func(ctx context.Context, to, code string) error {
//	        return mailer.Send(ctx, email.Message{To: ..., Text: "Code: " + code})
//	    },
//	    Issuer: jwtProvider, // any auth.Provider for minting the final token
//	})
//
//	o.Request(ctx, "user@example.com")          // generate + send
//	resp, _ := o.Verify(ctx, "user@example.com", "123456", claims) // check + issue token
package otp

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/azcov/gokit/auth"
)

var (
	ErrCodeNotFound = errors.New("otp: no code issued for this identifier")
	ErrCodeMismatch = errors.New("otp: code does not match")
	ErrCodeExpired  = errors.New("otp: code expired")
)

// Sender delivers the code to the recipient over any channel.
type Sender func(ctx context.Context, to, code string) error

// Store persists codes keyed by identifier. cache.Cache satisfies a superset
// of this; a thin adapter or any custom store works too.
type Store interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// TokenIssuer mints the final auth token after a successful verification.
// Any auth.Provider satisfies this (via CreateToken).
type TokenIssuer interface {
	CreateToken(ctx context.Context, claims auth.Claims) (string, error)
}

type Config struct {
	Store  Store
	Send   Sender
	Issuer TokenIssuer // optional; if nil, Verify returns claims without a token

	// TTL is how long a code stays valid. Default 5 minutes.
	TTL time.Duration
	// Length is the number of digits in the code. Default 6.
	Length int
	// KeyPrefix namespaces store keys. Default "otp:".
	KeyPrefix string
}

type OTP struct {
	store     Store
	send      Sender
	issuer    TokenIssuer
	ttl       time.Duration
	length    int
	keyPrefix string
}

func New(cfg Config) (*OTP, error) {
	if cfg.Store == nil {
		return nil, errors.New("otp: Store is required")
	}
	if cfg.Send == nil {
		return nil, errors.New("otp: Send is required")
	}
	if cfg.TTL == 0 {
		cfg.TTL = 5 * time.Minute
	}
	if cfg.Length == 0 {
		cfg.Length = 6
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "otp:"
	}
	return &OTP{
		store:     cfg.Store,
		send:      cfg.Send,
		issuer:    cfg.Issuer,
		ttl:       cfg.TTL,
		length:    cfg.Length,
		keyPrefix: cfg.KeyPrefix,
	}, nil
}

// Request generates a code, stores it, and delivers it to identifier via Send.
func (o *OTP) Request(ctx context.Context, identifier string) error {
	code, err := generateCode(o.length)
	if err != nil {
		return fmt.Errorf("otp: generate: %w", err)
	}
	if err := o.store.Set(ctx, o.key(identifier), []byte(code), o.ttl); err != nil {
		return fmt.Errorf("otp: store: %w", err)
	}
	if err := o.send(ctx, identifier, code); err != nil {
		// Best-effort cleanup so a failed send doesn't leave a dangling code.
		_ = o.store.Delete(ctx, o.key(identifier))
		return fmt.Errorf("otp: send: %w", err)
	}
	return nil
}

// Verify checks the submitted code against the stored one. On success it deletes
// the code (single-use) and, if an Issuer is configured, mints a token from claims.
func (o *OTP) Verify(ctx context.Context, identifier, code string, claims auth.Claims) (*auth.LoginResponse, error) {
	want, err := o.store.Get(ctx, o.key(identifier))
	if err != nil || len(want) == 0 {
		return nil, ErrCodeNotFound
	}
	if subtle.ConstantTimeCompare(want, []byte(code)) != 1 {
		return nil, ErrCodeMismatch
	}
	// Single-use: invalidate immediately.
	_ = o.store.Delete(ctx, o.key(identifier))

	resp := &auth.LoginResponse{}
	if o.issuer != nil {
		token, err := o.issuer.CreateToken(ctx, claims)
		if err != nil {
			return nil, fmt.Errorf("otp: issue token: %w", err)
		}
		resp.Token = token
	}
	return resp, nil
}

func (o *OTP) key(identifier string) string { return o.keyPrefix + identifier }

// generateCode returns a cryptographically-random numeric string of n digits.
func generateCode(n int) (string, error) {
	const digits = "0123456789"
	buf := make([]byte, n)
	for i := range buf {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		buf[i] = digits[idx.Int64()]
	}
	return string(buf), nil
}
