package repository

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/repository/chat_member"

	"lib/postgres"

	"github.com/Masterminds/squirrel"
)

// GetChatMembers получает список участников чата
func (r *Repository) GetChatMembers(ctx context.Context, chatID models.ChatID) ([]models.UserID, error) {
	const api = "[Repository][GetChatMembers]"

	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Запрос для получения всех участников чата
	getMembersQuery := r.sb.Select(chat_member.ChatMembersTableColumns...).
		From(chat_member.ChatMembersTable).
		Where(squirrel.Eq{chat_member.ChatMembersTableColumnChatID: int64(chatID)})

	// Выполняем запрос
	var memberRows []chat_member.Row
	if err := conn.Selectx(ctx, &memberRows, getMembersQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Если участников нет, возвращаем пустой список
	if len(memberRows) == 0 {
		return []models.UserID{}, nil
	}

	// Конвертируем строки в модели
	members := make([]models.UserID, 0, len(memberRows))
	for _, memberRow := range memberRows {
		_, userID := chat_member.ToModel(&memberRow)
		members = append(members, userID)
	}

	return members, nil
}
