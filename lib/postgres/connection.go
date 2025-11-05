package postgres

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxCommonAPI - pgx common api
type PgxCommonAPI interface {
	//
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	//
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	//
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// TransactionAPI - ...
type TransactionAPI interface {
	//
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (*Transaction, error)
	//
	Begin(ctx context.Context) (*Transaction, error)
}

// PgxCommonScanAPI улучшенный PgxCommonAPI
type PgxCommonScanAPI interface {
	// Getx - aka QueryRow
	Getx(ctx context.Context, dest any, sqlizer Sqlizer) error
	// Selectx - aka Query
	Selectx(ctx context.Context, dest any, sqlizer Sqlizer) error
	// Execx - aka Exec
	Execx(ctx context.Context, sqlizer Sqlizer) (pgconn.CommandTag, error)
}

// PgxExtendedAPI - ...
type PgxExtendedAPI interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// ConnectionAPI is a common database query interface.
// NOTE !
type ConnectionAPI interface {
	PgxCommonAPI
	PgxCommonScanAPI
	PgxExtendedAPI
	TransactionAPI
}

// Connection - postgres connection pool
type Connection struct {
	pool *pgxpool.Pool
}

var (
	_ PgxCommonAPI     = (*Connection)(nil)
	_ PgxCommonScanAPI = (*Connection)(nil)
	_ PgxExtendedAPI   = (*Connection)(nil)
	_ TransactionAPI   = (*Connection)(nil)
)

func (c *Connection) Close() error {
	c.pool.Close()
	return nil
}

// Query - pgx.Query
func (c *Connection) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return c.pool.Query(ctx, sql, args...)
}

// Exec - pgx.Exec
func (c *Connection) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return c.pool.Exec(ctx, sql, args...)
}

// QueryRow - pgx.QueryRow
func (c *Connection) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return c.pool.QueryRow(ctx, sql, args...)
}

// Begin - pgx.Begin
func (c *Connection) Begin(ctx context.Context) (*Transaction, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx}, nil
}

// BeginTx - pgx.BeginTx
func (c *Connection) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (*Transaction, error) {
	tx, err := c.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx}, nil
}

// SendBatch - pgx.SendBatch
func (c *Connection) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return c.pool.SendBatch(ctx, b)
}

// CopyFrom - pgx.CopyFrom
func (c *Connection) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return c.pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

// Sqlizer - something that can build sql query
type Sqlizer interface {
	ToSql() (sql string, args []any, err error)
}

// Getx - aka QueryRow
func (c *Connection) Getx(ctx context.Context, dest any, sqlizer Sqlizer) error {
	query, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}

	return pgxscan.Get(ctx, c.pool, dest, query, args...)
}

// Selectx - aka Query
func (c *Connection) Selectx(ctx context.Context, dest any, sqlizer Sqlizer) error {
	query, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}

	return pgxscan.Select(ctx, c.pool, dest, query, args...)
}

// Execx - aka Exec
func (c *Connection) Execx(ctx context.Context, sqlizer Sqlizer) (pgconn.CommandTag, error) {
	query, args, err := sqlizer.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	return c.pool.Exec(ctx, query, args...)
}
