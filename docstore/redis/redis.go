package redis

import (
	"context"
	"encoding/json"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/azcov/gokit/docstore"
)

var _ docstore.Store = (*Redis)(nil)

type Config struct {
	Addr      string `config:"addr"`
	Password  string `config:"password"`
	DB        int    `config:"db"`
	Namespace string `config:"namespace"`
}

// Redis stores documents as JSON strings with keys: {namespace}:{collection}:{id}
type Redis struct {
	client *goredis.Client
	ns     string
}

func New(cfg Config) *Redis {
	if cfg.Namespace == "" {
		cfg.Namespace = "docstore"
	}
	return &Redis{
		client: goredis.NewClient(&goredis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		ns: cfg.Namespace,
	}
}

func NewFromClient(client *goredis.Client, namespace string) *Redis {
	if namespace == "" {
		namespace = "docstore"
	}
	return &Redis{client: client, ns: namespace}
}

func (r *Redis) Set(ctx context.Context, collection, id string, doc any) error {
	b, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("docstore/redis: marshal: %w", err)
	}
	if err := r.client.Set(ctx, r.key(collection, id), b, 0).Err(); err != nil {
		return fmt.Errorf("docstore/redis: set %s/%s: %w", collection, id, err)
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, collection, id string, dest any) error {
	b, err := r.client.Get(ctx, r.key(collection, id)).Bytes()
	if err == goredis.Nil {
		return fmt.Errorf("docstore/redis: %s/%s: not found", collection, id)
	}
	if err != nil {
		return fmt.Errorf("docstore/redis: get %s/%s: %w", collection, id, err)
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("docstore/redis: unmarshal %s/%s: %w", collection, id, err)
	}
	return nil
}

func (r *Redis) Update(ctx context.Context, collection, id string, fields map[string]any) error {
	// Read-modify-write: load existing doc, merge fields, save.
	var existing map[string]any
	if err := r.Get(ctx, collection, id, &existing); err != nil {
		return err
	}
	for k, v := range fields {
		existing[k] = v
	}
	return r.Set(ctx, collection, id, existing)
}

func (r *Redis) Delete(ctx context.Context, collection, id string) error {
	if err := r.client.Del(ctx, r.key(collection, id)).Err(); err != nil {
		return fmt.Errorf("docstore/redis: delete %s/%s: %w", collection, id, err)
	}
	return nil
}

func (r *Redis) List(ctx context.Context, collection string, q docstore.Query) ([]map[string]any, error) {
	pattern := r.key(collection, "*")
	var keys []string
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("docstore/redis: scan: %w", err)
	}

	results := make([]map[string]any, 0, len(keys))
	for _, k := range keys {
		b, err := r.client.Get(ctx, k).Bytes()
		if err != nil {
			continue
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			continue
		}
		if matchesFilters(doc, q.Filters) {
			results = append(results, doc)
		}
	}
	if q.Limit > 0 && len(results) > q.Limit {
		results = results[:q.Limit]
	}
	return results, nil
}

func (r *Redis) Close() error { return r.client.Close() }

func (r *Redis) key(collection, id string) string {
	return r.ns + ":" + collection + ":" + id
}

func matchesFilters(doc map[string]any, filters []docstore.Filter) bool {
	for _, f := range filters {
		v, ok := doc[f.Field]
		if !ok {
			return false
		}
		switch f.Op {
		case docstore.OpEq:
			if v != f.Value {
				return false
			}
		case docstore.OpNe:
			if v == f.Value {
				return false
			}
		}
	}
	return true
}
