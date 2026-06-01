package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/azcov/gokit/lock"
	"github.com/redis/go-redis/v9"
)

var _ lock.Lock = (*Redis)(nil)

var releaseLua = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`)

var extendLua = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("pexpire", KEYS[1], ARGV[2])
end
return 0
`)

type Redis struct {
	client *redis.Client
	mu     sync.Mutex
	held   map[string]string // key → token
}

func New(client *redis.Client) *Redis {
	return &Redis{client: client, held: make(map[string]string)}
}

func (r *Redis) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	tok := newToken()
	ok, err := r.client.SetNX(ctx, key, tok, ttl).Result()
	if err != nil || !ok {
		return false, err
	}
	r.mu.Lock()
	r.held[key] = tok
	r.mu.Unlock()
	return true, nil
}

func (r *Redis) Release(ctx context.Context, key string) error {
	r.mu.Lock()
	tok, ok := r.held[key]
	if ok {
		delete(r.held, key)
	}
	r.mu.Unlock()
	if !ok {
		return nil
	}
	return releaseLua.Run(ctx, r.client, []string{key}, tok).Err()
}

func (r *Redis) Extend(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	r.mu.Lock()
	tok, ok := r.held[key]
	r.mu.Unlock()
	if !ok {
		return false, nil
	}
	ms := ttl.Milliseconds()
	n, err := extendLua.Run(ctx, r.client, []string{key}, tok, ms).Int()
	return n == 1, err
}

func newToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
