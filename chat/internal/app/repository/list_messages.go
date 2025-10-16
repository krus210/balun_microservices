package repository

import (
	"context"
	"fmt"
	"strconv"

	"chat/internal/app/models"
	"chat/internal/app/repository/message"

	"github.com/Masterminds/squirrel"
)

// ListMessages получает список сообщений чата с cursor-based пагинацией
func (r *Repository) ListMessages(ctx context.Context, chatID models.ChatID, limit int64, cursor *string) (messages []*models.Message, nextCursor *string, err error) {
	const api = "[Repository][ListMessages]"

	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Собираем базовый запрос
	listMessagesQuery := r.sb.Select(message.MessagesTableColumns...).
		From(message.MessagesTable).
		Where(squirrel.Eq{message.MessagesTableColumnChatID: int64(chatID)}).
		OrderBy(message.MessagesTableColumnID + " DESC")

	// Если есть cursor, добавляем фильтр по ID
	if cursor != nil && *cursor != "" {
		cursorID, err := strconv.ParseInt(*cursor, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: invalid cursor: %w", api, err)
		}
		listMessagesQuery = listMessagesQuery.Where(squirrel.Lt{message.MessagesTableColumnID: cursorID})
	}

	// Запрашиваем limit + 1 сообщений, чтобы понять, есть ли еще данные
	listMessagesQuery = listMessagesQuery.Limit(uint64(limit + 1))

	// Выполняем запрос
	var messageRows []message.Row
	if err := conn.Selectx(ctx, &messageRows, listMessagesQuery); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", api, ConvertPGError(err))
	}

	// Если сообщений нет, возвращаем пустой список
	if len(messageRows) == 0 {
		return []*models.Message{}, nil, nil
	}

	// Определяем, есть ли еще сообщения
	hasMore := len(messageRows) > int(limit)

	// Если есть еще сообщения, обрезаем результат до limit
	if hasMore {
		messageRows = messageRows[:limit]
	}

	// Конвертируем строки в модели
	result := make([]*models.Message, 0, len(messageRows))
	for _, msgRow := range messageRows {
		msg := message.ToModel(&msgRow)
		if msg != nil {
			result = append(result, msg)
		}
	}

	// Если есть еще сообщения, устанавливаем nextCursor
	if hasMore && len(result) > 0 {
		lastID := strconv.FormatInt(int64(result[len(result)-1].ID), 10)
		nextCursor = &lastID
	}

	return result, nextCursor, nil
}
