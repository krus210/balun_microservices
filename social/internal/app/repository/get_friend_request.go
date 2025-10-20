package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"social/internal/app/models"
	"social/internal/app/repository/friend_request"

	"github.com/Masterminds/squirrel"
)

const getFriendRequestApi = "[Repository][GetFriendRequest]"

// GetFriendRequest получает заявку в друзья по ID
func (r *Repository) GetFriendRequest(ctx context.Context, requestID models.FriendRequestID) (*models.FriendRequest, error) {
	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Запрос для получения информации о заявке
	getQuery := r.sb.Select(friend_request.FriendRequestsTableColumns...).
		From(friend_request.FriendRequestsTable).
		Where(squirrel.Eq{friend_request.FriendRequestsTableColumnID: int64(requestID)})

	// Выполняем запрос
	var row friend_request.Row
	if err := conn.Getx(ctx, &row, getQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", getFriendRequestApi, ConvertPGError(err))
	}

	// Конвертируем строку в модель
	return friend_request.ToModel(&row), nil
}
