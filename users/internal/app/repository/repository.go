package repository

import (
	"users/internal/app/usecase"
	"users/pkg/postgres/transaction_manager"

	"github.com/Masterminds/squirrel"
)

// Проверка удовлетворению интерфейсу usecase.UsersRepository
var _ usecase.UsersRepository = (*Repository)(nil)

// Repository реализация usecase.UsersRepository
type Repository struct {
	tm transaction_manager.TransactionManagerAPI
	sb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(tm transaction_manager.TransactionManagerAPI) *Repository {
	return &Repository{
		tm: tm,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
