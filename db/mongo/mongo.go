package mongo

import (
	"context"

	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo does not implement db.DB (MongoDB has no database/sql interface).
// Use DB() or Collection() for direct document access.
type Mongo struct {
	client *mongod.Client
	db     *mongod.Database
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
	return &Mongo{client: client, db: client.Database(cfg.Database)}, nil
}

func (m *Mongo) DB() *mongod.Database {
	return m.db
}

func (m *Mongo) Collection(name string) *mongod.Collection {
	return m.db.Collection(name)
}

func (m *Mongo) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

func (m *Mongo) Close() error {
	return m.client.Disconnect(context.Background())
}
