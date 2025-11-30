package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sskorolev/balun_microservices/lib/postgres"

	"auth/internal/app/models"
)

// CreateUser создает нового пользователя в БД
func (r *Repository) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	now := time.Now()
	user := &models.User{}

	insertQuery := r.sb.Insert("users").
		Columns("email", "password_hash", "created_at", "updated_at").
		Values(email, passwordHash, now, now).
		Suffix("RETURNING id, email, password_hash, created_at, updated_at")

	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		conn := r.tm.GetQueryEngine(txCtx)
		return conn.Getx(txCtx, user, insertQuery)
	})
	if err != nil {
		return nil, postgres.ConvertPGError(err)
	}

	return user, nil
}

// GetUserByEmail получает пользователя по email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	selectQuery := r.sb.Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		Where("email = ?", email)

	user := &models.User{}

	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		conn := r.tm.GetQueryEngine(txCtx)
		return conn.Getx(txCtx, user, selectQuery)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, postgres.ConvertPGError(err)
	}

	return user, nil
}

// GetUserByID получает пользователя по ID
func (r *Repository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	selectQuery := r.sb.Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		Where("id = ?", userID)

	user := &models.User{}

	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		conn := r.tm.GetQueryEngine(txCtx)
		return conn.Getx(txCtx, user, selectQuery)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, postgres.ConvertPGError(err)
	}

	return user, nil
}

// UpdateUser обновляет данные пользователя
func (r *Repository) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	updateQuery := r.sb.Update("users").
		Set("email", user.Email).
		Set("password_hash", user.PasswordHash).
		Set("updated_at", user.UpdatedAt).
		Where("id = ?", user.ID)

	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		conn := r.tm.GetQueryEngine(txCtx)
		_, err := conn.Execx(txCtx, updateQuery)
		return err
	})
	if err != nil {
		return postgres.ConvertPGError(err)
	}

	return nil
}

// SaveUser - для совместимости со старым интерфейсом
func (r *Repository) SaveUser(ctx context.Context, email, password string) (*models.User, error) {
	return r.CreateUser(ctx, email, password)
}
