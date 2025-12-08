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
	apiLogin = "[AuthService][Login]"
)

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*models.User, error) {
	// Получаем пользователя по email
	user, err := s.usersRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: userRepo GetUserByEmail error: %w", apiLogin, err)
	}
	if user == nil {
		return nil, models.ErrNotFound
	}

	// Проверяем пароль
	if err := s.passwordHasher.Verify(user.PasswordHash, req.Password); err != nil {
		return nil, ErrWrongPassword
	}

	// Создаем access token
	accessToken, err := s.tokenManager.CreateAccessToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create access token: %w", apiLogin, err)
	}

	// Создаем refresh token
	refreshToken, jti, err := s.tokenManager.CreateRefreshToken(ctx, user.ID, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create refresh token: %w", apiLogin, err)
	}

	// Хешируем refresh token для хранения в БД
	tokenHash := crypto.HashToken(refreshToken)

	// Сохраняем refresh token в БД
	var deviceIDPtr *string
	if req.DeviceID != "" {
		deviceIDPtr = &req.DeviceID
	}

	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		JTI:       jti,
		DeviceID:  deviceIDPtr,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokensRepo.CreateToken(ctx, refreshTokenModel); err != nil {
		return nil, fmt.Errorf("%s: failed to save refresh token: %w", apiLogin, err)
	}

	// Заполняем токены в пользователя
	user.Token = &models.UserToken{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		TokenExpiresAt: time.Now().Add(s.cfg.AccessTokenTTL),
	}

	return user, nil
}
