package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sskorolev/balun_microservices/lib/postgres"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"users/internal/app/models"
	"users/internal/app/repository/user"
)

const saveUserApi = "[Repository][SaveUser]"

// SaveUser создает новый профиль пользователя
func (r *Repository) SaveUser(ctx context.Context, profile *models.UserProfile) (*models.UserProfile, error) {
	now := time.Now()
	profile.CreatedAt = &now
	profile.UpdatedAt = &now

	// Создаем строку для вставки в таблицу user_profiles
	row := user.FromModel(profile)

	// Собираем запрос для вставки профиля
	insertQuery := r.sb.Insert(user.UserProfilesTable).
		Columns(
			user.UserProfilesTableColumnID,
			user.UserProfilesTableColumnNickname,
			user.UserProfilesTableColumnBio,
			user.UserProfilesTableColumnAvatarURL,
			user.UserProfilesTableColumnCreatedAt,
			user.UserProfilesTableColumnUpdatedAt,
		).
		Values(row.ID, row.Nickname, row.Bio, row.AvatarURL, row.CreatedAt, row.UpdatedAt)

	// Получаем QueryEngine из контекста
	conn := r.tm.GetQueryEngine(ctx)

	// Выполняем вставку профиля
	if _, err := conn.Execx(ctx, insertQuery); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("%s: %w", saveUserApi, models.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("%s: %w", saveUserApi, postgres.ConvertPGError(err))
	}

	return profile, nil
}
