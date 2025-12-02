package repository

import (
	"github.com/Masterminds/squirrel"
	"github.com/sskorolev/balun_microservices/lib/postgres"
)

// Repository - единый репозиторий для работы с БД auth сервиса
type Repository struct {
	tm postgres.TransactionManagerAPI
	sb squirrel.StatementBuilderType
}

// NewRepository создает новый экземпляр репозитория
func NewRepository(tm postgres.TransactionManagerAPI) *Repository {
	return &Repository{
		tm: tm,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
