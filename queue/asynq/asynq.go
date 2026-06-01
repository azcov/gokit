package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/azcov/gokit/queue"
	"github.com/hibiken/asynq"
)

// Client enqueues jobs via asynq (Redis-backed).
type Client struct {
	c *asynq.Client
}

var _ queue.Enqueuer = (*Client)(nil)

func NewClient(redisAddr string) *Client {
	return &Client{c: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})}
}

func (c *Client) Enqueue(ctx context.Context, job queue.Job, opts ...queue.Options) error {
	o := queue.Options{}
	if len(opts) > 0 {
		o = opts[0]
	}
	aopts := buildOpts(o)
	task := asynq.NewTask(job.Type, job.Payload)
	_, err := c.c.EnqueueContext(ctx, task, aopts...)
	return err
}

func (c *Client) Close() error {
	return c.c.Close()
}

// Server processes jobs via asynq.
type Server struct {
	srv      *asynq.Server
	mux      *asynq.ServeMux
	handlers map[string]queue.Handler
}

var _ queue.Worker = (*Server)(nil)

type ServerConfig struct {
	RedisAddr   string         `env:"REDIS_ADDRESS" json:"redis_address" yaml:"redis_address"`
	Concurrency int            `env:"CONCURRENCY" json:"concurrency" yaml:"concurrency"`
	Queues      map[string]int `json:"queues" yaml:"queues"`
}

func NewServer(cfg ServerConfig) *Server {
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 10
	}
	queues := cfg.Queues
	if queues == nil {
		queues = map[string]int{"default": 1}
	}
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues:      queues,
		},
	)
	return &Server{srv: srv, mux: asynq.NewServeMux(), handlers: make(map[string]queue.Handler)}
}

func (s *Server) Register(jobType string, handler queue.Handler) {
	s.handlers[jobType] = handler
	s.mux.HandleFunc(jobType, func(ctx context.Context, t *asynq.Task) error {
		return handler(ctx, queue.Job{Type: t.Type(), Payload: t.Payload()})
	})
}

func (s *Server) Start(_ context.Context) error {
	return s.srv.Run(s.mux)
}

func (s *Server) Stop() {
	s.srv.Shutdown()
}

func buildOpts(o queue.Options) []asynq.Option {
	var opts []asynq.Option
	if o.Queue != "" {
		opts = append(opts, asynq.Queue(o.Queue))
	}
	if o.MaxRetry > 0 {
		opts = append(opts, asynq.MaxRetry(o.MaxRetry))
	}
	if o.Delay > 0 {
		opts = append(opts, asynq.ProcessIn(o.Delay))
	}
	return opts
}

// Enqueue is a helper that JSON-encodes a payload struct before enqueuing.
func Enqueue[T any](ctx context.Context, c *Client, jobType string, payload T, opts ...queue.Options) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("asynq: marshal payload: %w", err)
	}
	return c.Enqueue(ctx, queue.Job{Type: jobType, Payload: b}, opts...)
}

// Decode is a helper that JSON-decodes a job payload inside a handler.
func Decode[T any](job queue.Job) (T, error) {
	var v T
	if err := json.Unmarshal(job.Payload, &v); err != nil {
		return v, fmt.Errorf("asynq: decode payload: %w", err)
	}
	return v, nil
}

var _ = time.Second // keep time import used via queue.Options
