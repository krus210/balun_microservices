package repository

import (
	"lib/postgres"

	"github.com/Masterminds/squirrel"
)

type Repository struct {
	db postgres.TransactionManagerAPI
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(db postgres.TransactionManagerAPI) *Repository {
	return &Repository{
		db: db,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
