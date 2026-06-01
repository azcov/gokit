package memory

import (
	"context"
	"sync"
	"time"

	"github.com/azcov/gokit/cache"
)

var _ cache.Cache = (*Memory)(nil)

type entry struct {
	value     []byte
	expiresAt time.Time
}

type Memory struct {
	mu    sync.RWMutex
	store map[string]entry
}

func New() *Memory {
	return &Memory{store: make(map[string]entry)}
}

func (m *Memory) Get(_ context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.store[key]
	if !ok || (!e.expiresAt.IsZero() && time.Now().After(e.expiresAt)) {
		return nil, nil
	}
	return e.value, nil
}

func (m *Memory) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := entry{value: value}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}
	m.store[key] = e
	return nil
}

func (m *Memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, key)
	return nil
}

func (m *Memory) Exists(ctx context.Context, key string) (bool, error) {
	v, err := m.Get(ctx, key)
	return v != nil, err
}

func (m *Memory) Close() error { return nil }
