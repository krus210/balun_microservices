package repository

import (
	"github.com/sskorolev/balun_microservices/lib/postgres"

	"users/internal/app/usecase"

	"github.com/Masterminds/squirrel"
)

// Проверка удовлетворению интерфейсу usecase.UsersRepository
var _ usecase.UsersRepository = (*Repository)(nil)

// Repository реализация usecase.UsersRepository
type Repository struct {
	tm postgres.TransactionManagerAPI
	sb squirrel.StatementBuilderType
}

// NewRepository конструктор Repository
func NewRepository(tm postgres.TransactionManagerAPI) *Repository {
	return &Repository{
		tm: tm,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
