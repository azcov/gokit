package otp_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/azcov/gokit/auth"
	"github.com/azcov/gokit/auth/otp"
)

// memStore is a minimal in-memory otp.Store for tests.
type memStore struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newMemStore() *memStore { return &memStore{data: make(map[string][]byte)} }

func (m *memStore) Set(_ context.Context, k string, v []byte, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[k] = v
	return nil
}

func (m *memStore) Get(_ context.Context, k string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[k]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (m *memStore) Delete(_ context.Context, k string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, k)
	return nil
}

// stubIssuer mints a deterministic token.
type stubIssuer struct{}

func (stubIssuer) CreateToken(_ context.Context, c auth.Claims) (string, error) {
	return "token-" + c.UserID, nil
}

func TestNew_Validation(t *testing.T) {
	if _, err := otp.New(otp.Config{Send: func(context.Context, string, string) error { return nil }}); err == nil {
		t.Error("expected error when Store is nil")
	}
	if _, err := otp.New(otp.Config{Store: newMemStore()}); err == nil {
		t.Error("expected error when Send is nil")
	}
}

func TestRequestAndVerify_Success(t *testing.T) {
	store := newMemStore()
	var sentTo, sentCode string
	o, err := otp.New(otp.Config{
		Store:  store,
		Issuer: stubIssuer{},
		Send: func(_ context.Context, to, code string) error {
			sentTo, sentCode = to, code
			return nil
		},
		Length: 6,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()
	if err := o.Request(ctx, "user@example.com"); err != nil {
		t.Fatalf("Request: %v", err)
	}
	if sentTo != "user@example.com" {
		t.Errorf("sent to %q", sentTo)
	}
	if len(sentCode) != 6 {
		t.Errorf("expected 6-digit code, got %q", sentCode)
	}

	resp, err := o.Verify(ctx, "user@example.com", sentCode, auth.Claims{UserID: "u1"})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if resp.Token != "token-u1" {
		t.Errorf("got token %q", resp.Token)
	}
}

func TestVerify_WrongCode(t *testing.T) {
	store := newMemStore()
	o, _ := otp.New(otp.Config{
		Store: store,
		Send:  func(_ context.Context, _, _ string) error { return nil },
	})
	ctx := context.Background()
	_ = o.Request(ctx, "a@b.com")

	if _, err := o.Verify(ctx, "a@b.com", "000000", auth.Claims{}); !errors.Is(err, otp.ErrCodeMismatch) {
		t.Errorf("expected ErrCodeMismatch, got %v", err)
	}
}

func TestVerify_NoCodeIssued(t *testing.T) {
	o, _ := otp.New(otp.Config{
		Store: newMemStore(),
		Send:  func(_ context.Context, _, _ string) error { return nil },
	})
	if _, err := o.Verify(context.Background(), "nobody@x.com", "123456", auth.Claims{}); !errors.Is(err, otp.ErrCodeNotFound) {
		t.Errorf("expected ErrCodeNotFound, got %v", err)
	}
}

func TestVerify_SingleUse(t *testing.T) {
	store := newMemStore()
	var code string
	o, _ := otp.New(otp.Config{
		Store: store,
		Send:  func(_ context.Context, _, c string) error { code = c; return nil },
	})
	ctx := context.Background()
	_ = o.Request(ctx, "a@b.com")

	if _, err := o.Verify(ctx, "a@b.com", code, auth.Claims{}); err != nil {
		t.Fatalf("first verify should succeed: %v", err)
	}
	// second use must fail — code consumed
	if _, err := o.Verify(ctx, "a@b.com", code, auth.Claims{}); !errors.Is(err, otp.ErrCodeNotFound) {
		t.Errorf("expected single-use rejection, got %v", err)
	}
}

func TestRequest_SendFailureCleansUp(t *testing.T) {
	store := newMemStore()
	o, _ := otp.New(otp.Config{
		Store: store,
		Send:  func(_ context.Context, _, _ string) error { return errors.New("smtp down") },
	})
	ctx := context.Background()
	if err := o.Request(ctx, "a@b.com"); err == nil {
		t.Fatal("expected send error")
	}
	// code must not linger after a failed send
	if _, err := store.Get(ctx, "otp:a@b.com"); err == nil {
		t.Error("expected code to be cleaned up after send failure")
	}
}

func TestVerify_NoIssuer(t *testing.T) {
	store := newMemStore()
	var code string
	o, _ := otp.New(otp.Config{
		Store: store,
		Send:  func(_ context.Context, _, c string) error { code = c; return nil },
		// no Issuer
	})
	ctx := context.Background()
	_ = o.Request(ctx, "a@b.com")
	resp, err := o.Verify(ctx, "a@b.com", code, auth.Claims{UserID: "u1"})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if resp.Token != "" {
		t.Errorf("expected empty token without issuer, got %q", resp.Token)
	}
}
