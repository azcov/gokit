package cockroachdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/azcov/gokit/db"
)

var _ db.DB = (*CockroachDB)(nil)

// Config holds CockroachDB connection settings.
// Connection string format: postgresql://user:pass@host:26257/dbname?sslmode=verify-full
type Config struct {
	DSN          string `config:"dsn"`
	MaxOpenConns int    `config:"max_open_conns"`
	MaxIdleConns int    `config:"max_idle_conns"`
}

type CockroachDB struct {
	pool  *pgxpool.Pool
	sqlDB *sql.DB
}

func New(ctx context.Context, cfg Config) (*CockroachDB, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("cockroachdb: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("cockroachdb: ping: %w", err)
	}
	sqlDB := stdlib.OpenDBFromPool(pool)
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	return &CockroachDB{pool: pool, sqlDB: sqlDB}, nil
}

func (c *CockroachDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := c.sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cockroachdb: query: %w", err)
	}
	return rows, nil
}

func (c *CockroachDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return c.sqlDB.QueryRowContext(ctx, query, args...)
}

func (c *CockroachDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := c.sqlDB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cockroachdb: exec: %w", err)
	}
	return res, nil
}

func (c *CockroachDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	tx, err := c.sqlDB.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("cockroachdb: begin tx: %w", err)
	}
	return tx, nil
}

func (c *CockroachDB) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

func (c *CockroachDB) Close() error {
	c.pool.Close()
	return c.sqlDB.Close()
}
