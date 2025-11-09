package repository

import (
	"context"
	"fmt"

	"github.com/sskorolev/balun_microservices/lib/postgres"

	"chat/internal/app/models"
	"chat/internal/app/repository/chat"
	"chat/internal/app/repository/chat_member"

	"github.com/Masterminds/squirrel"
)

// ListChatsByUserID получает список всех чатов, в которых участвует пользователь
func (r *Repository) ListChatsByUserID(ctx context.Context, userID models.UserID) ([]*models.Chat, error) {
	const api = "[Repository][ListChatsByUserID]"

	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Сначала получаем список chat_id для пользователя из chat_members
	getChatIDsQuery := r.sb.
		Select(chat_member.ChatMembersTableColumnChatID).
		From(chat_member.ChatMembersTable).
		Where(squirrel.Eq{chat_member.ChatMembersTableColumnUserID: userID})

	var chatIDs []string
	if err := conn.Selectx(ctx, &chatIDs, getChatIDsQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Если у пользователя нет чатов, возвращаем пустой список
	if len(chatIDs) == 0 {
		return []*models.Chat{}, nil
	}

	// Получаем информацию о чатах по списку ID
	listChatsQuery := r.sb.
		Select(chat.ChatsTableColumns...).
		From(chat.ChatsTable).
		Where(squirrel.Eq{chat.ChatsTableColumnID: chatIDs}).
		OrderBy(chat.ChatsTableColumnUpdatedAt + " DESC")

	// Выполняем запрос
	var chatRows []chat.Row
	if err := conn.Selectx(ctx, &chatRows, listChatsQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	// Если чатов нет, возвращаем пустой список
	if len(chatRows) == 0 {
		return []*models.Chat{}, nil
	}

	// Создаем мапу чатов для быстрого доступа
	chatMap := make(map[models.ChatID]*models.Chat, len(chatRows))
	for _, chatRow := range chatRows {
		chatModel := chat.ToModel(&chatRow)
		// ToModel уже инициализирует пустые слайсы для участников и сообщений
		chatMap[chatModel.ID] = chatModel
	}

	// Запрос для получения всех участников всех чатов одним запросом
	// Используем chatIDs, полученные из первого запроса
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
