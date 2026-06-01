package dragonfly

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/azcov/gokit/cache"
)

var _ cache.Cache = (*Dragonfly)(nil)

// Config holds Dragonfly connection settings.
// Dragonfly is Redis-compatible and uses the same wire protocol.
// Default port is 6379 (same as Redis).
type Config struct {
	Addr     string        `config:"addr"`
	Password string        `config:"password"`
	DB       int           `config:"db"`
	Timeout  time.Duration `config:"timeout"`
}

// Dragonfly wraps go-redis to connect to a Dragonfly instance.
// Drop-in replacement for cache/redis — use this when your infra runs Dragonfly.
type Dragonfly struct {
	client *goredis.Client
}

func New(cfg Config) *Dragonfly {
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:6379"
	}
	opts := &goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	if cfg.Timeout > 0 {
		opts.DialTimeout = cfg.Timeout
		opts.ReadTimeout = cfg.Timeout
		opts.WriteTimeout = cfg.Timeout
	}
	return &Dragonfly{client: goredis.NewClient(opts)}
}

func NewFromClient(client *goredis.Client) *Dragonfly {
	return &Dragonfly{client: client}
}

func (d *Dragonfly) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := d.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return nil, fmt.Errorf("dragonfly: %s: not found", key)
	}
	if err != nil {
		return nil, fmt.Errorf("dragonfly: get %s: %w", key, err)
	}
	return b, nil
}

func (d *Dragonfly) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := d.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("dragonfly: set %s: %w", key, err)
	}
	return nil
}

func (d *Dragonfly) Delete(ctx context.Context, key string) error {
	if err := d.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("dragonfly: delete %s: %w", key, err)
	}
	return nil
}

func (d *Dragonfly) Exists(ctx context.Context, key string) (bool, error) {
	n, err := d.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("dragonfly: exists %s: %w", key, err)
	}
	return n > 0, nil
}

func (d *Dragonfly) Close() error {
	return d.client.Close()
}
