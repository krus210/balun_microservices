package repository

import (
	"context"
	"fmt"

	"lib/postgres"

	"notifications/internal/app/models"
)

func (r *Repository) SaveInboxMessage(ctx context.Context, msg *models.InboxMessage) error {
	const api = "inbox.Repository.SaveInboxMessage"

	row := FromModel(msg)

	qb := r.qb.Insert(InboxMessagesTable).
		Columns(InboxMessagesTableColumns...).
		Values(row.Values(InboxMessagesTableColumns...)...).
		Suffix("ON CONFLICT DO NOTHING")

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return nil
}
