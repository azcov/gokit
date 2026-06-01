package redis

import (
	"context"
	"time"

	"github.com/azcov/gokit/cache"
	"github.com/redis/go-redis/v9"
)

var _ cache.Cache = (*Redis)(nil)

type Redis struct {
	client *redis.Client
}

func New(opts *redis.Options) *Redis {
	return &Redis{client: redis.NewClient(opts)}
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (r *Redis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (r *Redis) Close() error {
	return r.client.Close()
}
