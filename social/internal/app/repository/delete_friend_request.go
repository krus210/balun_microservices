package repository

import (
	"context"
	"fmt"

	"lib/postgres"

	"social/internal/app/models"
	"social/internal/app/repository/friend_request"

	"github.com/Masterminds/squirrel"
)

const deleteFriendRequestApi = "[Repository][DeleteFriendRequest]"

// DeleteFriendRequest удаляет заявку в друзья по ID
func (r *Repository) DeleteFriendRequest(ctx context.Context, requestID models.FriendRequestID) error {
	// Собираем запрос для удаления заявки
	deleteQuery := r.sb.Delete(friend_request.FriendRequestsTable).
		Where(squirrel.Eq{friend_request.FriendRequestsTableColumnID: int64(requestID)})

	// Получаем QueryEngine из контекста
	conn := r.tm.GetQueryEngine(ctx)

	// Выполняем удаление
	if _, err := conn.Execx(ctx, deleteQuery); err != nil {
		return fmt.Errorf("%s: %w", deleteFriendRequestApi, postgres.ConvertPGError(err))
	}

	return nil
}
