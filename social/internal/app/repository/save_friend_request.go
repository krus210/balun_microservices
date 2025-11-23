package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sskorolev/balun_microservices/lib/postgres"

	"github.com/google/uuid"

	"social/internal/app/models"
	"social/internal/app/repository/friend_request"
)

const saveFriendRequestApi = "[Repository][SaveFriendRequest]"

// SaveFriendRequest создает новую заявку в друзья в рамках транзакции
func (r *Repository) SaveFriendRequest(ctx context.Context, req *models.FriendRequest) (*models.FriendRequest, error) {
	// Генерируем UUID для новой заявки
	requestID := uuid.New().String()
	req.ID = models.FriendRequestID(requestID)

	now := time.Now()
	req.CreatedAt = &now
	req.UpdatedAt = &now

	// Создаем строку для вставки в таблицу friend_requests
	row := friend_request.FromModel(req)

	// Собираем запрос для вставки заявки
	insertQuery := r.sb.Insert(friend_request.FriendRequestsTable).
		Columns(
			friend_request.FriendRequestsTableColumnID,
			friend_request.FriendRequestsTableColumnFromUserID,
			friend_request.FriendRequestsTableColumnToUserID,
			friend_request.FriendRequestsTableColumnStatus,
			friend_request.FriendRequestsTableColumnCreatedAt,
			friend_request.FriendRequestsTableColumnUpdatedAt,
		).
		Values(row.ID, row.FromUserID, row.ToUserID, row.Status, row.CreatedAt, row.UpdatedAt)

	// Получаем QueryEngine из контекста транзакции
	conn := r.tm.GetQueryEngine(ctx)

	// Выполняем вставку заявки
	if _, err := conn.Execx(ctx, insertQuery); err != nil {
		return nil, fmt.Errorf("%s: %w", saveFriendRequestApi, postgres.ConvertPGError(err))
	}

	return req, nil
}
