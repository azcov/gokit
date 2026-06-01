package mysql

import (
	"context"
	"database/sql"

	"github.com/azcov/gokit/db"
	_ "github.com/go-sql-driver/mysql"
)

var _ db.DB = (*MySQL)(nil)

type MySQL struct {
	db *sql.DB
}

func New(dsn string) (*MySQL, error) {
	d, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &MySQL{db: d}, nil
}

func (m *MySQL) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return m.db.QueryContext(ctx, query, args...)
}

func (m *MySQL) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return m.db.QueryRowContext(ctx, query, args...)
}

func (m *MySQL) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.db.ExecContext(ctx, query, args...)
}

func (m *MySQL) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}

func (m *MySQL) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

func (m *MySQL) Close() error {
	return m.db.Close()
}
