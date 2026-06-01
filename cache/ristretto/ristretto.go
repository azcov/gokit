package ristretto

import (
	"context"
	"time"

	"github.com/azcov/gokit/cache"
	"github.com/dgraph-io/ristretto"
)

var _ cache.Cache = (*Ristretto)(nil)

type Ristretto struct {
	c *ristretto.Cache
}

func New(cfg *ristretto.Config) (*Ristretto, error) {
	c, err := ristretto.NewCache(cfg)
	if err != nil {
		return nil, err
	}
	return &Ristretto{c: c}, nil
}

func (r *Ristretto) Get(_ context.Context, key string) ([]byte, error) {
	val, ok := r.c.Get(key)
	if !ok {
		return nil, nil
	}
	return val.([]byte), nil
}

func (r *Ristretto) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	r.c.SetWithTTL(key, value, int64(len(value)), ttl)
	r.c.Wait()
	return nil
}

func (r *Ristretto) Delete(_ context.Context, key string) error {
	r.c.Del(key)
	return nil
}

func (r *Ristretto) Exists(_ context.Context, key string) (bool, error) {
	_, ok := r.c.Get(key)
	return ok, nil
}

func (r *Ristretto) Close() error {
	r.c.Close()
	return nil
}
