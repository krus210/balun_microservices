package repository

import (
	tm "social/pkg/postgres/transaction_manager"

	"github.com/Masterminds/squirrel"
)

// Repository реализация usecase.OutboxRepository
type Repository struct {
	db tm.TransactionManagerAPI
	qb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(db tm.TransactionManagerAPI) *Repository {
	return &Repository{
		db: db,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
