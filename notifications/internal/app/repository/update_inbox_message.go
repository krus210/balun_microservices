package repository

import (
	"context"
	"fmt"
	"lib/postgres"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// UpdateInboxMessageParams параметры для обновления inbox сообщения
type UpdateInboxMessageParams struct {
	ID          uuid.UUID
	Status      string
	Attempts    int
	ProcessedAt *time.Time // если nil - поле не обновляется
	LastError   *string
}

// UpdateInboxMessage обновляет статус и метаданные сообщения
func (r *Repository) UpdateInboxMessage(ctx context.Context, params UpdateInboxMessageParams) error {
	const api = "inbox.Repository.UpdateInboxMessage"

	qb := r.qb.Update(InboxMessagesTable).
		Set(InboxMessagesTableColumnStatus, params.Status).
		Set(InboxMessagesTableColumnAttempts, params.Attempts).
		Set(InboxMessagesTableColumnLastError, params.LastError).
		Where(squirrel.Eq{InboxMessagesTableColumnID: params.ID})

	// Обновляем processed_at только если передано значение
	if params.ProcessedAt != nil {
		qb = qb.Set(InboxMessagesTableColumnProcessedAt, *params.ProcessedAt)
	}

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return nil
}
