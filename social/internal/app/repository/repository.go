package repository

import (
	"social/internal/app/usecase"
	"social/pkg/postgres/transaction_manager"

	"github.com/Masterminds/squirrel"
)

// Проверка удовлетворению интерфейсу usecase.ChatRepository
var _ usecase.SocialRepository = (*Repository)(nil)

// Repository реализация usecase.SocialRepository
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
