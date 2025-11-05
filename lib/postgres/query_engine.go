package postgres

import (
	"github.com/jackc/pgx/v5"
)

// QueryEngine is a common database query interface.
type QueryEngine interface {
	PgxCommonAPI
	PgxCommonScanAPI
	PgxExtendedAPI
}

// TxAccessMode is the transaction access mode (read write or read only)
type TxAccessMode = pgx.TxAccessMode

// Transaction access modes
const (
	ReadWrite = pgx.ReadWrite
	ReadOnly  = pgx.ReadOnly
)
