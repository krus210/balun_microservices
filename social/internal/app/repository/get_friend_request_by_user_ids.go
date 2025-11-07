package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sskorolev/balun_microservices/lib/postgres"

	"social/internal/app/models"
	"social/internal/app/repository/friend_request"

	"github.com/Masterminds/squirrel"
)

const GetFriendRequestByUserIDsApi = "[Repository][GetFriendRequestByUserIDs]"

// GetFriendRequestByUserIDs получает заявку в друзья между двумя пользователями
func (r *Repository) GetFriendRequestByUserIDs(ctx context.Context, fromUserID models.UserID, toUserID models.UserID) (*models.FriendRequest, error) {
	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Запрос для получения заявки между двумя пользователями
	getQuery := r.sb.Select(friend_request.FriendRequestsTableColumns...).
		From(friend_request.FriendRequestsTable).
		Where(squirrel.Eq{
			friend_request.FriendRequestsTableColumnFromUserID: string(fromUserID),
			friend_request.FriendRequestsTableColumnToUserID:   string(toUserID),
		})

	// Выполняем запрос
	var row friend_request.Row
	if err := conn.Getx(ctx, &row, getQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", GetFriendRequestByUserIDsApi, postgres.ConvertPGError(err))
	}

	// Конвертируем строку в модель
	return friend_request.ToModel(&row), nil
}
