package repository

import (
	"chat/internal/app/usecase"

	"lib/postgres"

	"github.com/Masterminds/squirrel"
)

// Проверка удовлетворению интерфейсу usecase.ChatRepository
var _ usecase.ChatRepository = (*Repository)(nil)

// Repository реализация usecase.ChatRepository
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
