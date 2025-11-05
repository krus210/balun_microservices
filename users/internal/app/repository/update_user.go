package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"lib/postgres"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"users/internal/app/models"
	"users/internal/app/repository/user"

	"github.com/Masterminds/squirrel"
)

const updateUserApi = "[Repository][UpdateUser]"

// UpdateUser обновляет профиль пользователя
func (r *Repository) UpdateUser(ctx context.Context, profile *models.UserProfile) (*models.UserProfile, error) {
	now := time.Now()
	profile.UpdatedAt = &now

	// Собираем запрос для обновления профиля с RETURNING
	updateQuery := r.sb.Update(user.UserProfilesTable).
		Where(squirrel.Eq{user.UserProfilesTableColumnID: profile.UserID}).
		Suffix("RETURNING " + user.UserProfilesTableColumnID + ", " +
			user.UserProfilesTableColumnNickname + ", " +
			user.UserProfilesTableColumnBio + ", " +
			user.UserProfilesTableColumnAvatarURL + ", " +
			user.UserProfilesTableColumnCreatedAt + ", " +
			user.UserProfilesTableColumnUpdatedAt)

	// Добавляем поля для обновления
	if profile.Nickname != "" {
		updateQuery = updateQuery.Set(user.UserProfilesTableColumnNickname, profile.Nickname)
	}
	updateQuery = updateQuery.Set(user.UserProfilesTableColumnBio, profile.Bio)
	updateQuery = updateQuery.Set(user.UserProfilesTableColumnAvatarURL, profile.AvatarURL)
	updateQuery = updateQuery.Set(user.UserProfilesTableColumnUpdatedAt, now)

	// Получаем QueryEngine из контекста
	conn := r.tm.GetQueryEngine(ctx)

	// Выполняем обновление и получаем обновленную запись
	var row user.Row
	if err := conn.Getx(ctx, &row, updateQuery); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("%s: %w", updateUserApi, models.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("%s: %w", updateUserApi, postgres.ConvertPGError(err))
	}

	return user.ToModel(&row), nil
}
