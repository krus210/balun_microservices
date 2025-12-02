package usecase

import (
	"context"
	"fmt"
	"time"

	"auth/internal/app/crypto"
	"auth/internal/app/models"
	"auth/internal/app/usecase/dto"
)

const (
	apiRefresh = "[AuthService][Refresh]"
)

func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*models.User, error) {
	// 1. Валидация refresh токена и извлечение claims
	claims, err := s.tokenManager.VerifyRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid refresh token: %w", apiRefresh, ErrInvalidToken)
	}

	// 2. Получаем токен из БД по JTI
	tokenHash := crypto.HashToken(req.RefreshToken)
	storedToken, err := s.refreshTokensRepo.GetTokenByJTI(ctx, claims.JWTID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get token from DB: %w", apiRefresh, err)
	}

	// 3. Проверка: токен уже использован (anti-reuse)
	if storedToken.UsedAt != nil {
		return nil, ErrTokenUsed
	}

	// 4. Проверка: токен истек
	if storedToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// 5. Проверка: хеш совпадает
	if storedToken.TokenHash != tokenHash {
		return nil, ErrInvalidToken
	}

	// 6. Получаем пользователя
	user, err := s.usersRepo.GetUserByID(ctx, claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get user: %w", apiRefresh, err)
	}

	// 7. Создаем новый access token
	newAccessToken, err := s.tokenManager.CreateAccessToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create access token: %w", apiRefresh, err)
	}

	// 8. Создаем новый refresh token
	newRefreshToken, newJTI, err := s.tokenManager.CreateRefreshToken(ctx, user.ID, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create refresh token: %w", apiRefresh, err)
	}

	// 9. Помечаем старый токен как использованный
	if err := s.refreshTokensRepo.MarkAsUsed(ctx, claims.JWTID, newJTI); err != nil {
		return nil, fmt.Errorf("%s: failed to mark token as used: %w", apiRefresh, err)
	}

	// 10. Сохраняем новый refresh token
	var deviceIDPtr *string
	if req.DeviceID != "" {
		deviceIDPtr = &req.DeviceID
	}

	newTokenHash := crypto.HashToken(newRefreshToken)
	newRefreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		JTI:       newJTI,
		DeviceID:  deviceIDPtr,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokensRepo.CreateToken(ctx, newRefreshTokenModel); err != nil {
		return nil, fmt.Errorf("%s: failed to save new refresh token: %w", apiRefresh, err)
	}

	// 11. Заполняем токены в пользователя
	user.Token = &models.UserToken{
		AccessToken:    newAccessToken,
		RefreshToken:   newRefreshToken,
		TokenExpiresAt: time.Now().Add(s.cfg.AccessTokenTTL),
	}

	return user, nil
}
