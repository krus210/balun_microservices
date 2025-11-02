package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"lib/postgres"

	"social/internal/app/models"
	"social/internal/app/repository/friend_request"

	"github.com/Masterminds/squirrel"
)

const updateFriendRequestApi = "[Repository][UpdateFriendRequest]"

// UpdateFriendRequest обновляет статус заявки в друзья
func (r *Repository) UpdateFriendRequest(ctx context.Context, requestID models.FriendRequestID, status models.FriendRequestStatus) (*models.FriendRequest, error) {
	now := time.Now()

	// Собираем запрос для обновления статуса заявки с RETURNING
	updateQuery := r.sb.Update(friend_request.FriendRequestsTable).
		Set(friend_request.FriendRequestsTableColumnStatus, int(status)).
		Set(friend_request.FriendRequestsTableColumnUpdatedAt, now).
		Where(squirrel.Eq{friend_request.FriendRequestsTableColumnID: int64(requestID)}).
		Suffix("RETURNING " + friend_request.FriendRequestsTableColumnID + ", " +
			friend_request.FriendRequestsTableColumnFromUserID + ", " +
			friend_request.FriendRequestsTableColumnToUserID + ", " +
			friend_request.FriendRequestsTableColumnStatus + ", " +
			friend_request.FriendRequestsTableColumnCreatedAt + ", " +
			friend_request.FriendRequestsTableColumnUpdatedAt)

	// Получаем QueryEngine из контекста
	conn := r.tm.GetQueryEngine(ctx)

	// Выполняем обновление и получаем обновленную запись
	var row friend_request.Row
	if err := conn.Getx(ctx, &row, updateQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", updateFriendRequestApi, postgres.ConvertPGError(err))
	}

	return friend_request.ToModel(&row), nil
}
