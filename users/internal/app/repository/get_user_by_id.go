package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"users/internal/app/models"
	"users/internal/app/repository/user"

	"github.com/Masterminds/squirrel"
)

const getUserByIDApi = "[Repository][GetUserByID]"

// GetUserByID получает профиль пользователя по ID
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.UserProfile, error) {
	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Запрос для получения информации о профиле
	getQuery := r.sb.Select(user.UserProfilesTableColumns...).
		From(user.UserProfilesTable).
		Where(squirrel.Eq{user.UserProfilesTableColumnID: id})

	// Выполняем запрос
	var row user.Row
	if err := conn.Getx(ctx, &row, getQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", getUserByIDApi, ConvertPGError(err))
	}

	// Конвертируем строку в модель
	return user.ToModel(&row), nil
}
