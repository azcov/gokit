package queue

import (
	"context"
	"time"
)

type Job struct {
	Type    string
	Payload []byte
}

type Options struct {
	Queue    string
	MaxRetry int
	Delay    time.Duration
}

type Handler func(ctx context.Context, job Job) error

// Enqueuer submits jobs for background processing.
type Enqueuer interface {
	Enqueue(ctx context.Context, job Job, opts ...Options) error
	Close() error
}

// Worker processes jobs from the queue.
type Worker interface {
	Register(jobType string, handler Handler)
	Start(ctx context.Context) error
	Stop()
}
