package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

	// Сначала найдем ID чатов, где участвуют оба пользователя
	// Используем подзапрос с INNER JOIN
	findChatsQuery := r.sb.Select("cm1." + chat_member.ChatMembersTableColumnChatID).
		From(chat_member.ChatMembersTable + " cm1").
		InnerJoin(chat_member.ChatMembersTable + " cm2 ON cm1." + chat_member.ChatMembersTableColumnChatID + " = cm2." + chat_member.ChatMembersTableColumnChatID).
		Where(squirrel.Eq{
			"cm1." + chat_member.ChatMembersTableColumnUserID: int64(userID1),
			"cm2." + chat_member.ChatMembersTableColumnUserID: int64(userID2),
		})

	// Выполняем запрос для получения ID чатов-кандидатов
	var candidateRows []chat_member.Row
	if err := conn.Selectx(ctx, &candidateRows, findChatsQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", api, ConvertPGError(err))
	}

	// Если нет чатов с обоими участниками, возвращаем nil
	if len(candidateRows) == 0 {
		return nil, nil
	}

	// Извлекаем ID чатов из структуры
	candidateChatIDs := make([]int64, 0, len(candidateRows))
	for _, row := range candidateRows {
		candidateChatIDs = append(candidateChatIDs, row.ChatID)
	}

	// Для каждого чата проверяем, что в нем ровно 2 участника
	for _, chatID := range candidateChatIDs {
		// Подсчитываем количество участников в чате
		countQuery := r.sb.Select("COUNT(*)").
			From(chat_member.ChatMembersTable).
			Where(squirrel.Eq{chat_member.ChatMembersTableColumnChatID: chatID})

		var count int
		if err := conn.Getx(ctx, &count, countQuery); err != nil {
			return nil, fmt.Errorf("%s: %w", api, ConvertPGError(err))
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
				return nil, fmt.Errorf("%s: %w", api, ConvertPGError(err))
			}

			// Конвертируем в модель (ToModel уже инициализирует пустые слайсы)
			chatModel := chat.ToModel(&chatRow)

			// Загружаем участников
			getMembersQuery := r.sb.Select(chat_member.ChatMembersTableColumns...).
				From(chat_member.ChatMembersTable).
				Where(squirrel.Eq{chat_member.ChatMembersTableColumnChatID: chatID})

			var memberRows []chat_member.Row
			if err := conn.Selectx(ctx, &memberRows, getMembersQuery); err != nil {
				return nil, fmt.Errorf("%s: %w", api, ConvertPGError(err))
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
