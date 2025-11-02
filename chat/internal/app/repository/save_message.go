package repository

import (
	"context"
	"fmt"
	"time"

	"chat/internal/app/models"
	"chat/internal/app/repository/message"
)

// SaveMessage сохраняет новое сообщение в базе данных
func (r *Repository) SaveMessage(ctx context.Context, msg *models.Message) (*models.Message, error) {
	const api = "[Repository][SaveMessage]"

	now := time.Now()
	msg.CreatedAt = now
	msg.UpdatedAt = now

	// Создаем строку для вставки в таблицу messages
	row := message.FromModel(msg)

	// Собираем запрос для вставки сообщения
	insertMessageQuery := r.sb.Insert(message.MessagesTable).
		Columns(
			message.MessagesTableColumnText,
			message.MessagesTableColumnChatID,
			message.MessagesTableColumnOwnerID,
			message.MessagesTableColumnCreatedAt,
			message.MessagesTableColumnUpdatedAt,
		).
		Values(row.Text, row.ChatID, row.OwnerID, row.CreatedAt, row.UpdatedAt).
		Suffix("RETURNING id")

	// Оборачиваем операцию в транзакцию Read Committed
	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		// Получаем QueryEngine из контекста транзакции
		conn := r.tm.GetQueryEngine(txCtx)

		// Выполняем вставку сообщения и получаем сгенерированный ID
		var messageID int64
		if err := conn.Getx(txCtx, &messageID, insertMessageQuery); err != nil {
			return fmt.Errorf("%s: %w", api, ConvertPGError(err))
		}

		msg.ID = models.MessageID(messageID)

		// Возвращаем nil для COMMIT, любая ошибка выше вызовет ROLLBACK
		return nil
	})
	if err != nil {
		return nil, err
	}

	return msg, nil
}
