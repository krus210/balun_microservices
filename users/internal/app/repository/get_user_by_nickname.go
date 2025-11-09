package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sskorolev/balun_microservices/lib/postgres"

	"users/internal/app/models"
	"users/internal/app/repository/user"

	"github.com/Masterminds/squirrel"
)

const getUserByNicknameApi = "[Repository][GetUserByNickname]"

// GetUserByNickname получает профиль пользователя по никнейму
func (r *Repository) GetUserByNickname(ctx context.Context, nickname string) (*models.UserProfile, error) {
	// Получаем QueryEngine из контекста (может быть транзакция или обычное соединение)
	conn := r.tm.GetQueryEngine(ctx)

	// Запрос для получения информации о профиле по никнейму
	getQuery := r.sb.Select(user.UserProfilesTableColumns...).
		From(user.UserProfilesTable).
		Where(squirrel.Eq{user.UserProfilesTableColumnNickname: nickname})

	// Выполняем запрос
	var row user.Row
	if err := conn.Getx(ctx, &row, getQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", getUserByNicknameApi, postgres.ConvertPGError(err))
	}

	// Конвертируем строку в модель
	return user.ToModel(&row), nil
}
