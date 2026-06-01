package mongo

import (
	"context"

	"github.com/azcov/gokit/docstore"
	"go.mongodb.org/mongo-driver/bson"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ docstore.Store = (*Mongo)(nil)

type Mongo struct {
	db *mongod.Database
}

type Config struct {
	URI      string `env:"URI" json:"uri" yaml:"uri"`
	Database string `env:"DATABASE" json:"database" yaml:"database"`
}

func New(ctx context.Context, cfg Config) (*Mongo, error) {
	client, err := mongod.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}
	return &Mongo{db: client.Database(cfg.Database)}, nil
}

func (m *Mongo) Get(ctx context.Context, collection, id string, dest any) error {
	return m.db.Collection(collection).FindOne(ctx, bson.M{"_id": id}).Decode(dest)
}

func (m *Mongo) Set(ctx context.Context, collection, id string, doc any) error {
	upsert := true
	_, err := m.db.Collection(collection).ReplaceOne(
		ctx,
		bson.M{"_id": id},
		doc,
		&options.ReplaceOptions{Upsert: &upsert},
	)
	return err
}

func (m *Mongo) Update(ctx context.Context, collection, id string, fields map[string]any) error {
	_, err := m.db.Collection(collection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": fields},
	)
	return err
}

func (m *Mongo) Delete(ctx context.Context, collection, id string) error {
	_, err := m.db.Collection(collection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *Mongo) List(ctx context.Context, collection string, q docstore.Query) ([]map[string]any, error) {
	filter := buildFilter(q.Filters)

	opts := options.Find()
	if q.OrderBy != "" {
		dir := 1
		if q.Desc {
			dir = -1
		}
		opts.SetSort(bson.D{{Key: q.OrderBy, Value: dir}})
	}
	if q.Limit > 0 {
		opts.SetLimit(int64(q.Limit))
	}

	cur, err := m.db.Collection(collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var docs []map[string]any
	for cur.Next(ctx) {
		row := map[string]any{}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		docs = append(docs, row)
	}
	return docs, cur.Err()
}

func (m *Mongo) Close() error {
	return m.db.Client().Disconnect(context.Background())
}

func buildFilter(filters []docstore.Filter) bson.M {
	f := bson.M{}
	for _, v := range filters {
		switch v.Op {
		case docstore.OpEq:
			f[v.Field] = v.Value
		case docstore.OpNe:
			f[v.Field] = bson.M{"$ne": v.Value}
		case docstore.OpLt:
			f[v.Field] = bson.M{"$lt": v.Value}
		case docstore.OpLte:
			f[v.Field] = bson.M{"$lte": v.Value}
		case docstore.OpGt:
			f[v.Field] = bson.M{"$gt": v.Value}
		case docstore.OpGte:
			f[v.Field] = bson.M{"$gte": v.Value}
		}
	}
	return f
}
