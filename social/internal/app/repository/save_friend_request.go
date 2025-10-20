package repository

import (
	"context"
	"fmt"
	"time"

	"social/internal/app/models"
	"social/internal/app/repository/friend_request"
)

const saveFriendRequestApi = "[Repository][SaveFriendRequest]"

// SaveFriendRequest создает новую заявку в друзья в рамках транзакции
func (r *Repository) SaveFriendRequest(ctx context.Context, req *models.FriendRequest) (*models.FriendRequest, error) {
	now := time.Now()
	req.CreatedAt = &now
	req.UpdatedAt = &now

	// Создаем строку для вставки в таблицу friend_requests
	row := friend_request.FromModel(req)

	// Собираем запрос для вставки заявки
	insertQuery := r.sb.Insert(friend_request.FriendRequestsTable).
		Columns(
			friend_request.FriendRequestsTableColumnFromUserID,
			friend_request.FriendRequestsTableColumnToUserID,
			friend_request.FriendRequestsTableColumnStatus,
			friend_request.FriendRequestsTableColumnCreatedAt,
			friend_request.FriendRequestsTableColumnUpdatedAt,
		).
		Values(row.FromUserID, row.ToUserID, row.Status, row.CreatedAt, row.UpdatedAt).
		Suffix("RETURNING id")

	// Получаем QueryEngine из контекста транзакции
	conn := r.tm.GetQueryEngine(ctx)

	// Выполняем вставку заявки и получаем сгенерированный ID
	var requestID int64
	if err := conn.Getx(ctx, &requestID, insertQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", saveFriendRequestApi, ConvertPGError(err))
	}

	req.ID = models.FriendRequestID(requestID)

	return req, nil
}
