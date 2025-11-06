package app

import (
	"context"
	"fmt"

	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/postgres"
)

// InitPostgres создает и инициализирует PostgreSQL connection pool
func InitPostgres(ctx context.Context, dbCfg config.DatabaseConfig) (*postgres.Connection, func(), error) {
	conn, _, err := postgres.New(ctx,
		postgres.WithHost(dbCfg.Host),
		postgres.WithPort(dbCfg.Port),
		postgres.WithDatabase(dbCfg.Name),
		postgres.WithUser(dbCfg.User),
		postgres.WithPassword(dbCfg.Password),
		postgres.WithSSLMode(dbCfg.SSLMode),
		postgres.WithMaxConnIdleTime(dbCfg.MaxConnIdleTime),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize postgres connection: %w", err)
	}

	cleanup := func() {
		conn.Close()
	}

	return conn, cleanup, nil
}

// InitTransactionManager создает transaction manager
func InitTransactionManager(conn *postgres.Connection) postgres.TransactionManagerAPI {
	return postgres.NewTransactionManager(conn)
}
