package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	PostgresUser     = os.Getenv("DB_USER")
	PostgresPassword = os.Getenv("DB_PASSWORD")
	PostgresHost     = os.Getenv("DB_HOST")
	PostgresPort     = os.Getenv("DB_PORT")
	PostgresDB       = os.Getenv("DB_NAME")
)

func DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		PostgresUser, PostgresPassword, PostgresHost, PostgresPort, PostgresDB,
	)
}
func NewPostgresConnection(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(DSN())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool connect: %w", err)
	}

	return pool, nil
}
