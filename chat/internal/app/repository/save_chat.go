package repository

import (
	"context"
	"fmt"
	"time"

	"chat/internal/app/models"
	"chat/internal/app/repository/chat"
	"chat/internal/app/repository/chat_member"

	"lib/postgres"
)

// SaveChat создает новый чат с участниками в рамках транзакции
func (r *Repository) SaveChat(ctx context.Context, chatModel *models.Chat) (*models.Chat, error) {
	const api = "[Repository][SaveChat]"

	now := time.Now()
	chatModel.CreatedAt = now
	chatModel.UpdatedAt = now

	// Создаем строку для вставки в таблицу chats
	row := chat.FromModel(chatModel)

	// Собираем запрос для вставки чата
	insertChatQuery := r.sb.Insert(chat.ChatsTable).
		Columns(chat.ChatsTableColumnCreatedAt, chat.ChatsTableColumnUpdatedAt).
		Values(row.CreatedAt, row.UpdatedAt).
		Suffix("RETURNING id")

	// Оборачиваем операции в транзакцию Read Committed
	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		// Получаем QueryEngine из контекста транзакции
		conn := r.tm.GetQueryEngine(txCtx)

		// Выполняем вставку чата и получаем сгенерированный ID
		var chatID int64
		if err := conn.Getx(txCtx, &chatID, insertChatQuery); err != nil {
			return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
		}

		chatModel.ID = models.ChatID(chatID)

		// Вставляем участников чата в таблицу chat_members
		if len(chatModel.ParticipantIDs) > 0 {
			insertMembersQuery := r.sb.Insert(chat_member.ChatMembersTable).
				Columns(chat_member.ChatMembersTableColumns...)

			for _, userID := range chatModel.ParticipantIDs {
				memberRow := chat_member.FromModel(chatModel.ID, userID)
				insertMembersQuery = insertMembersQuery.Values(memberRow.Values()...)
			}

			if _, err := conn.Execx(txCtx, insertMembersQuery); err != nil {
				return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
			}
		}

		// Возвращаем nil для COMMIT, любая ошибка выше вызовет ROLLBACK
		return nil
	})
	if err != nil {
		return nil, err
	}

	return chatModel, nil
}
