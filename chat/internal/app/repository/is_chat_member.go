package repository

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/repository/chat_member"

	"lib/postgres"

	"github.com/Masterminds/squirrel"
)

// IsChatMember проверяет, является ли пользователь участником чата
func (r *Repository) IsChatMember(ctx context.Context, chatID models.ChatID, userID models.UserID) (bool, error) {
	const api = "[Repository][IsChatMember]"

	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Используем простой подсчет для проверки существования записи
	countQuery := r.sb.Select("COUNT(*)").
		From(chat_member.ChatMembersTable).
		Where(squirrel.Eq{
			chat_member.ChatMembersTableColumnChatID: int64(chatID),
			chat_member.ChatMembersTableColumnUserID: int64(userID),
		})

	// Выполняем запрос
	var count int
	if err := conn.Getx(ctx, &count, countQuery); err != nil {
		return false, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return count > 0, nil
}
