package vector

import "context"

type Vector []float64

type Record struct {
	ID        string
	Vector    Vector
	Metadata  map[string]any
	Namespace string
}

type QueryOptions struct {
	TopK      int
	Filter    map[string]any
	Namespace string
}

type QueryResult struct {
	ID       string
	Score    float64
	Metadata map[string]any
}

type Config struct {
	APIKey    string `config:"api_key"`
	BaseURL   string `config:"base_url"`
	Dimension int    `config:"dimension"`
	Namespace string `config:"namespace"`
}

type Store interface {
	Upsert(ctx context.Context, records []Record) error
	Query(ctx context.Context, vec Vector, opts QueryOptions) ([]QueryResult, error)
	Delete(ctx context.Context, ids []string, namespace string) error
	Close() error
}
