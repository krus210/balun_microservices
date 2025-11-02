package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/repository/chat"
	"chat/internal/app/repository/chat_member"

	"lib/postgres"

	"github.com/Masterminds/squirrel"
)

// GetChat получает чат по ID вместе с участниками
func (r *Repository) GetChat(ctx context.Context, chatID models.ChatID) (*models.Chat, error) {
	const api = "[Repository][GetChat]"

	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Запрос для получения информации о чате
	getChatQuery := r.sb.Select(chat.ChatsTableColumns...).
		From(chat.ChatsTable).
		Where(squirrel.Eq{chat.ChatsTableColumnID: int64(chatID)})

	// Выполняем запрос
	var chatRow chat.Row
	if err := conn.Getx(ctx, &chatRow, getChatQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Конвертируем строку в модель
	chatModel := chat.ToModel(&chatRow)

	// Запрос для получения участников чата
	getMembersQuery := r.sb.Select(chat_member.ChatMembersTableColumns...).
		From(chat_member.ChatMembersTable).
		Where(squirrel.Eq{chat_member.ChatMembersTableColumnChatID: int64(chatID)})

	// Выполняем запрос для получения участников
	var memberRows []chat_member.Row
	if err := conn.Selectx(ctx, &memberRows, getMembersQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Заполняем список участников
	chatModel.ParticipantIDs = make([]models.UserID, 0, len(memberRows))
	for _, memberRow := range memberRows {
		_, userID := chat_member.ToModel(&memberRow)
		chatModel.ParticipantIDs = append(chatModel.ParticipantIDs, userID)
	}

	return chatModel, nil
}
