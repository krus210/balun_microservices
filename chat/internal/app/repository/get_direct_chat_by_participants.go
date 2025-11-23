package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sskorolev/balun_microservices/lib/postgres"

	"chat/internal/app/models"
	"chat/internal/app/repository/chat"
	"chat/internal/app/repository/chat_member"

	"github.com/Masterminds/squirrel"
)

// GetDirectChatByParticipants получает личный чат между двумя пользователями
func (r *Repository) GetDirectChatByParticipants(ctx context.Context, userID1, userID2 models.UserID) (*models.Chat, error) {
	const api = "[Repository][GetDirectChatByParticipants]"

	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Получаем chat_ids для первого пользователя
	getChatIDsUser1Query := r.sb.Select(chat_member.ChatMembersTableColumnChatID).
		From(chat_member.ChatMembersTable).
		Where(squirrel.Eq{chat_member.ChatMembersTableColumnUserID: userID1})

	var chatIDsUser1 []string
	if err := conn.Selectx(ctx, &chatIDsUser1, getChatIDsUser1Query); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	if len(chatIDsUser1) == 0 {
		return nil, nil
	}

	// Получаем chat_ids для второго пользователя
	getChatIDsUser2Query := r.sb.Select(chat_member.ChatMembersTableColumnChatID).
		From(chat_member.ChatMembersTable).
		Where(squirrel.Eq{chat_member.ChatMembersTableColumnUserID: userID2})

	var chatIDsUser2 []string
	if err := conn.Selectx(ctx, &chatIDsUser2, getChatIDsUser2Query); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	if len(chatIDsUser2) == 0 {
		return nil, nil
	}

	// Находим пересечение chat_ids (чаты, где участвуют оба пользователя)
	chatIDsSet := make(map[string]bool)
	for _, chatID := range chatIDsUser1 {
		chatIDsSet[chatID] = true
	}

	var candidateChatIDs []string
	for _, chatID := range chatIDsUser2 {
		if chatIDsSet[chatID] {
			candidateChatIDs = append(candidateChatIDs, chatID)
		}
	}

	// Если нет общих чатов, возвращаем nil
	if len(candidateChatIDs) == 0 {
		return nil, nil
	}

	// Для каждого чата проверяем, что в нем ровно 2 участника
	for _, chatID := range candidateChatIDs {
		// Подсчитываем количество участников в чате
		countQuery := r.sb.Select("COUNT(*)").
			From(chat_member.ChatMembersTable).
			Where(squirrel.Eq{chat_member.ChatMembersTableColumnChatID: chatID})

		var count int
		if err := conn.Getx(ctx, &count, countQuery); err != nil {
			return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
		}

		// Если в чате ровно 2 участника, это наш чат
		if count == 2 {
			// Получаем информацию о чате
			getChatQuery := r.sb.Select(chat.ChatsTableColumns...).
				From(chat.ChatsTable).
				Where(squirrel.Eq{chat.ChatsTableColumnID: chatID})

			var chatRow chat.Row
			if err := conn.Getx(ctx, &chatRow, getChatQuery); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
			}

			// Конвертируем в модель (ToModel уже инициализирует пустые слайсы)
			chatModel := chat.ToModel(&chatRow)

			// Загружаем участников
			getMembersQuery := r.sb.Select(chat_member.ChatMembersTableColumns...).
				From(chat_member.ChatMembersTable).
				Where(squirrel.Eq{chat_member.ChatMembersTableColumnChatID: chatID})

			var memberRows []chat_member.Row
			if err := conn.Selectx(ctx, &memberRows, getMembersQuery); err != nil {
				return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
			}

			// Заполняем список участников
			for _, memberRow := range memberRows {
				_, userID := chat_member.ToModel(&memberRow)
				chatModel.ParticipantIDs = append(chatModel.ParticipantIDs, userID)
			}

			return chatModel, nil
		}
	}

	// Если не нашли чат с ровно двумя участниками, возвращаем nil
	return nil, nil
}
