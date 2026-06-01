package memory

import (
	"context"
	"sync"

	"github.com/azcov/gokit/pubsub"
)

var _ pubsub.PubSub = (*Memory)(nil)

type Memory struct {
	mu       sync.RWMutex
	handlers map[string][]pubsub.Handler
}

func New() *Memory {
	return &Memory{handlers: make(map[string][]pubsub.Handler)}
}

func (m *Memory) Publish(ctx context.Context, _ string, msg pubsub.Message) error {
	m.mu.RLock()
	handlers := m.handlers[msg.Topic]
	m.mu.RUnlock()
	for _, h := range handlers {
		if err := h(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (m *Memory) Subscribe(_ context.Context, topic string, handler pubsub.Handler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[topic] = append(m.handlers[topic], handler)
	return nil
}

func (m *Memory) Unsubscribe(topic string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.handlers, topic)
	return nil
}

func (m *Memory) Close() error { return nil }
