package db

import (
	"context"
	"database/sql"
)

// DB covers SQL databases. MongoDB does not implement this interface;
// use db/mongo directly for document store access.
type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Ping(ctx context.Context) error
	Close() error
}
