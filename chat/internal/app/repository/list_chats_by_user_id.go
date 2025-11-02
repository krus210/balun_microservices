package repository

import (
	"context"
	"fmt"

	"chat/internal/app/models"
	"chat/internal/app/repository/chat"
	"chat/internal/app/repository/chat_member"

	"lib/postgres"

	"github.com/Masterminds/squirrel"
)

// ListChatsByUserID получает список всех чатов, в которых участвует пользователь
func (r *Repository) ListChatsByUserID(ctx context.Context, userID models.UserID) ([]*models.Chat, error) {
	const api = "[Repository][ListChatsByUserID]"

	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Запрос для получения всех чатов пользователя
	// Добавляем префикс таблицы к именам колонок для корректного JOIN
	listChatsQuery := r.sb.
		Select(
			chat.ChatsTable+"."+chat.ChatsTableColumnID,
			chat.ChatsTable+"."+chat.ChatsTableColumnCreatedAt,
			chat.ChatsTable+"."+chat.ChatsTableColumnUpdatedAt,
		).
		From(chat.ChatsTable).
		InnerJoin(chat_member.ChatMembersTable + " ON " + chat.ChatsTable + "." + chat.ChatsTableColumnID + " = " + chat_member.ChatMembersTable + "." + chat_member.ChatMembersTableColumnChatID).
		Where(squirrel.Eq{chat_member.ChatMembersTable + "." + chat_member.ChatMembersTableColumnUserID: int64(userID)}).
		OrderBy(chat.ChatsTable + "." + chat.ChatsTableColumnUpdatedAt + " DESC")

	// Выполняем запрос
	var chatRows []chat.Row
	if err := conn.Selectx(ctx, &chatRows, listChatsQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Если чатов нет, возвращаем пустой список
	if len(chatRows) == 0 {
		return []*models.Chat{}, nil
	}

	// Собираем ID всех чатов для получения участников
	chatIDs := make([]int64, 0, len(chatRows))
	chatMap := make(map[models.ChatID]*models.Chat, len(chatRows))
	for _, chatRow := range chatRows {
		chatModel := chat.ToModel(&chatRow)
		// ToModel уже инициализирует пустые слайсы для участников и сообщений
		chatIDs = append(chatIDs, int64(chatModel.ID))
		chatMap[chatModel.ID] = chatModel
	}

	// Запрос для получения всех участников всех чатов одним запросом
	getMembersQuery := r.sb.Select(chat_member.ChatMembersTableColumns...).
		From(chat_member.ChatMembersTable).
		Where(squirrel.Eq{chat_member.ChatMembersTableColumnChatID: chatIDs})

	// Выполняем запрос для получения всех участников
	var memberRows []chat_member.Row
	if err := conn.Selectx(ctx, &memberRows, getMembersQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Группируем участников по чатам
	for _, memberRow := range memberRows {
		chatID, userID := chat_member.ToModel(&memberRow)
		if chatModel, ok := chatMap[chatID]; ok {
			chatModel.ParticipantIDs = append(chatModel.ParticipantIDs, userID)
		}
	}

	// Формируем результат в виде слайса
	result := make([]*models.Chat, 0, len(chatRows))
	for _, chatRow := range chatRows {
		chatID := models.ChatID(chatRow.ID)
		if chatModel, ok := chatMap[chatID]; ok {
			result = append(result, chatModel)
		}
	}

	return result, nil
}
