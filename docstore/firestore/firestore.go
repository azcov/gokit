package firestore

import (
	"context"
	"errors"

	fs "cloud.google.com/go/firestore"
	"github.com/azcov/gokit/docstore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var _ docstore.Store = (*Firestore)(nil)

type Firestore struct {
	client *fs.Client
}

type Config struct {
	ProjectID string                `config:"project_id"`
	Options   []option.ClientOption `json:"-" yaml:"-"`
}

func New(ctx context.Context, cfg Config) (*Firestore, error) {
	client, err := fs.NewClient(ctx, cfg.ProjectID, cfg.Options...)
	if err != nil {
		return nil, err
	}
	return &Firestore{client: client}, nil
}

func (f *Firestore) Get(ctx context.Context, collection, id string, dest any) error {
	snap, err := f.client.Collection(collection).Doc(id).Get(ctx)
	if err != nil {
		return err
	}
	return snap.DataTo(dest)
}

func (f *Firestore) Set(ctx context.Context, collection, id string, doc any) error {
	_, err := f.client.Collection(collection).Doc(id).Set(ctx, doc)
	return err
}

func (f *Firestore) Update(ctx context.Context, collection, id string, fields map[string]any) error {
	updates := make([]fs.Update, 0, len(fields))
	for k, v := range fields {
		updates = append(updates, fs.Update{Path: k, Value: v})
	}
	_, err := f.client.Collection(collection).Doc(id).Update(ctx, updates)
	return err
}

func (f *Firestore) Delete(ctx context.Context, collection, id string) error {
	_, err := f.client.Collection(collection).Doc(id).Delete(ctx)
	return err
}

func (f *Firestore) List(ctx context.Context, collection string, q docstore.Query) ([]map[string]any, error) {
	ref := f.client.Collection(collection)

	var cq fs.Query
	started := false
	for _, filter := range q.Filters {
		if !started {
			cq = ref.Where(filter.Field, filter.Op, filter.Value)
			started = true
		} else {
			cq = cq.Where(filter.Field, filter.Op, filter.Value)
		}
	}

	var iter *fs.DocumentIterator
	if started {
		if q.OrderBy != "" {
			dir := fs.Asc
			if q.Desc {
				dir = fs.Desc
			}
			cq = cq.OrderBy(q.OrderBy, dir)
		}
		if q.Limit > 0 {
			cq = cq.Limit(q.Limit)
		}
		iter = cq.Documents(ctx)
	} else {
		baseQ := ref.Query
		if q.OrderBy != "" {
			dir := fs.Asc
			if q.Desc {
				dir = fs.Desc
			}
			baseQ = baseQ.OrderBy(q.OrderBy, dir)
		}
		if q.Limit > 0 {
			baseQ = baseQ.Limit(q.Limit)
		}
		iter = baseQ.Documents(ctx)
	}
	defer iter.Stop()

	var docs []map[string]any
	for {
		snap, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		docs = append(docs, snap.Data())
	}
	return docs, nil
}

func (f *Firestore) Close() error {
	return f.client.Close()
}
