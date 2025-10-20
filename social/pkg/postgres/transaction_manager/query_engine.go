package transaction_manager

import (
	"social/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

// QueryEngine is a common database query interface.
type QueryEngine interface {
	postgres.PgxCommonAPI
	postgres.PgxCommonScanAPI
	postgres.PgxExtendedAPI
}

// TxAccessMode is the transaction access mode (read write or read only)
type TxAccessMode = pgx.TxAccessMode

// Transaction access modes
const (
	ReadWrite = pgx.ReadWrite
	ReadOnly  = pgx.ReadOnly
)
