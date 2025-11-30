package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sskorolev/balun_microservices/lib/postgres"

	"auth/internal/app/models"
)

// CreateToken создает новый refresh token в БД
func (r *Repository) CreateToken(ctx context.Context, token *models.RefreshToken) error {
	insertQuery := r.sb.Insert("refresh_tokens").
		Columns("user_id", "token_hash", "jti", "device_id", "expires_at", "created_at").
		Values(token.UserID, token.TokenHash, token.JTI, token.DeviceID, token.ExpiresAt, token.CreatedAt)

	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		conn := r.tm.GetQueryEngine(txCtx)
		_, err := conn.Execx(txCtx, insertQuery)
		return err
	})
	if err != nil {
		return postgres.ConvertPGError(err)
	}

	return nil
}

// GetTokenByJTI получает refresh token по JTI
func (r *Repository) GetTokenByJTI(ctx context.Context, jti string) (*models.RefreshToken, error) {
	selectQuery := r.sb.Select("id", "user_id", "token_hash", "jti", "device_id", "expires_at", "used_at", "replaced_by_jti", "created_at").
		From("refresh_tokens").
		Where("jti = ?", jti)

	token := &models.RefreshToken{}

	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		conn := r.tm.GetQueryEngine(txCtx)
		return conn.Getx(txCtx, token, selectQuery)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, postgres.ConvertPGError(err)
	}

	return token, nil
}

// MarkAsUsed помечает токен как использованный
func (r *Repository) MarkAsUsed(ctx context.Context, jti, replacedByJTI string) error {
	updateQuery := r.sb.Update("refresh_tokens").
		Set("used_at", time.Now()).
		Set("replaced_by_jti", replacedByJTI).
		Where("jti = ?", jti)

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

// RevokeTokenByJTI отзывает токен по JTI
func (r *Repository) RevokeTokenByJTI(ctx context.Context, jti string) error {
	updateQuery := r.sb.Update("refresh_tokens").
		Set("used_at", time.Now()).
		Where("jti = ?", jti)

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

// CleanupExpiredTokens удаляет истекшие токены
func (r *Repository) CleanupExpiredTokens(ctx context.Context) error {
	deleteQuery := r.sb.Delete("refresh_tokens").
		Where("expires_at < ?", time.Now())

	err := r.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		conn := r.tm.GetQueryEngine(txCtx)
		_, err := conn.Execx(txCtx, deleteQuery)
		return err
	})
	if err != nil {
		return postgres.ConvertPGError(err)
	}

	return nil
}
