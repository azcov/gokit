package storage

import (
	"context"
	"io"
)

type Object struct {
	Key         string
	Size        int64
	ContentType string
	Metadata    map[string]string
}

type PutOptions struct {
	ContentType string
	Metadata    map[string]string
}

type Storage interface {
	Put(ctx context.Context, key string, r io.Reader, opts PutOptions) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	URL(ctx context.Context, key string) (string, error)
	List(ctx context.Context, prefix string) ([]Object, error)
	Close() error
}
