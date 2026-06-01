package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/azcov/gokit/lock"
)

var _ lock.Lock = (*Memory)(nil)

type entry struct {
	token     string
	expiresAt time.Time
}

type Memory struct {
	mu    sync.Mutex
	locks map[string]entry
	// token per key held by this instance
	held map[string]string
}

func New() *Memory {
	return &Memory{
		locks: make(map[string]entry),
		held:  make(map[string]string),
	}
}

func (m *Memory) Acquire(_ context.Context, key string, ttl time.Duration) (bool, error) {
	tok := newToken()
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.locks[key]; ok && time.Now().Before(e.expiresAt) {
		return false, nil
	}
	m.locks[key] = entry{token: tok, expiresAt: time.Now().Add(ttl)}
	m.held[key] = tok
	return true, nil
}

func (m *Memory) Release(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	tok, ok := m.held[key]
	if !ok {
		return nil
	}
	if e, exists := m.locks[key]; exists && e.token == tok {
		delete(m.locks, key)
	}
	delete(m.held, key)
	return nil
}

func (m *Memory) Extend(_ context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	tok, ok := m.held[key]
	if !ok {
		return false, nil
	}
	e, exists := m.locks[key]
	if !exists || e.token != tok {
		return false, nil
	}
	m.locks[key] = entry{token: tok, expiresAt: time.Now().Add(ttl)}
	return true, nil
}

func newToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
