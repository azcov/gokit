package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"

	gcs "cloud.google.com/go/storage"
	"github.com/azcov/gokit/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var _ storage.Storage = (*GCS)(nil)

type GCS struct {
	client *gcs.Client
	bucket string
}

type Config struct {
	Bucket  string                `config:"bucket"`
	Options []option.ClientOption `json:"-" yaml:"-"`
}

func New(ctx context.Context, cfg Config) (*GCS, error) {
	client, err := gcs.NewClient(ctx, cfg.Options...)
	if err != nil {
		return nil, err
	}
	return &GCS{client: client, bucket: cfg.Bucket}, nil
}

func (g *GCS) Put(ctx context.Context, key string, r io.Reader, opts storage.PutOptions) error {
	w := g.client.Bucket(g.bucket).Object(key).NewWriter(ctx)
	w.ContentType = opts.ContentType
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func (g *GCS) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return g.client.Bucket(g.bucket).Object(key).NewReader(ctx)
}

func (g *GCS) Delete(ctx context.Context, key string) error {
	return g.client.Bucket(g.bucket).Object(key).Delete(ctx)
}

func (g *GCS) Exists(ctx context.Context, key string) (bool, error) {
	_, err := g.client.Bucket(g.bucket).Object(key).Attrs(ctx)
	if errors.Is(err, gcs.ErrObjectNotExist) {
		return false, nil
	}
	return err == nil, err
}

func (g *GCS) URL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucket, key), nil
}

func (g *GCS) List(ctx context.Context, prefix string) ([]storage.Object, error) {
	var objects []storage.Object
	it := g.client.Bucket(g.bucket).Objects(ctx, &gcs.Query{Prefix: prefix})
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		objects = append(objects, storage.Object{
			Key:         attrs.Name,
			Size:        attrs.Size,
			ContentType: attrs.ContentType,
		})
	}
	return objects, nil
}

func (g *GCS) Close() error {
	return g.client.Close()
}
