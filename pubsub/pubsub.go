package pubsub

import "context"

type Message struct {
	ID      string
	Topic   string
	Payload []byte
	Meta    map[string]string
}

type Handler func(ctx context.Context, msg Message) error

type Publisher interface {
	Publish(ctx context.Context, topic string, msg Message) error
	Close() error
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler Handler) error
	Unsubscribe(topic string) error
	Close() error
}

type PubSub interface {
	Publisher
	Subscriber
}
