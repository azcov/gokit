package postgres

import (
	"context"
	"database/sql"

	"github.com/azcov/gokit/db"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var _ db.DB = (*Postgres)(nil)

type Postgres struct {
	db *sql.DB
}

func New(dsn string) (*Postgres, error) {
	d, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &Postgres{db: d}, nil
}

func (p *Postgres) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

func (p *Postgres) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

func (p *Postgres) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

func (p *Postgres) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, opts)
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *Postgres) Close() error {
	return p.db.Close()
}
