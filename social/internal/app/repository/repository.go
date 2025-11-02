package repository

import (
	"social/internal/app/usecase"

	"lib/postgres"

	"github.com/Masterminds/squirrel"
)

// Проверка удовлетворению интерфейсу usecase.SocialRepository
var _ usecase.SocialRepository = (*Repository)(nil)

// Repository реализация usecase.SocialRepository
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
