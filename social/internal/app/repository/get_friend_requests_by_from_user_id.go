package repository

import (
	"context"
	"fmt"
	"strconv"

	"social/internal/app/models"
	"social/internal/app/repository/friend_request"

	"github.com/Masterminds/squirrel"
)

const GetFriendRequestsByFromUserIDApi = "[Repository][GetFriendRequestsByFromUserID]"

// GetFriendRequestsByFromUserID получает список заявок в друзья, отправленных пользователем с cursor-based пагинацией
func (r *Repository) GetFriendRequestsByFromUserID(ctx context.Context, fromUserID models.UserID, limit *int64, cursor *string) (friends []*models.FriendRequest, nextCursor *string, err error) {
	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Устанавливаем лимит по умолчанию
	defaultLimit := int64(50)
	if limit == nil {
		limit = &defaultLimit
	}

	// Собираем базовый запрос
	listQuery := r.sb.Select(friend_request.FriendRequestsTableColumns...).
		From(friend_request.FriendRequestsTable).
		Where(squirrel.Eq{friend_request.FriendRequestsTableColumnFromUserID: int64(fromUserID)}).
		OrderBy(friend_request.FriendRequestsTableColumnID + " DESC")

	// Если есть cursor, добавляем фильтр по ID
	if cursor != nil && *cursor != "" {
		cursorID, err := strconv.ParseInt(*cursor, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: invalid cursor: %w", GetFriendRequestsByFromUserIDApi, err)
		}
		listQuery = listQuery.Where(squirrel.Lt{friend_request.FriendRequestsTableColumnID: cursorID})
	}

	// Запрашиваем limit + 1 записей, чтобы понять, есть ли еще данные
	listQuery = listQuery.Limit(uint64(*limit + 1))

	// Выполняем запрос
	var rows []friend_request.Row
	if err := conn.Selectx(ctx, &rows, listQuery); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", GetFriendRequestsByFromUserIDApi, ConvertPGError(err))
	}

	// Если записей нет, возвращаем пустой список
	if len(rows) == 0 {
		return []*models.FriendRequest{}, nil, nil
	}

	// Определяем, есть ли еще записи
	hasMore := len(rows) > int(*limit)

	// Если есть еще записи, обрезаем результат до limit
	if hasMore {
		rows = rows[:*limit]
	}

	// Конвертируем строки в модели
	result := make([]*models.FriendRequest, 0, len(rows))
	for _, row := range rows {
		req := friend_request.ToModel(&row)
		if req != nil {
			result = append(result, req)
		}
	}

	// Если есть еще записи, устанавливаем nextCursor
	if hasMore && len(result) > 0 {
		lastID := strconv.FormatInt(int64(result[len(result)-1].ID), 10)
		nextCursor = &lastID
	}

	return result, nextCursor, nil
}
