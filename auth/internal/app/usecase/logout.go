package usecase

import (
	"context"
	"fmt"

	"auth/internal/app/usecase/dto"
)

const (
	apiLogout = "[AuthService][Logout]"
)

func (s *AuthService) Logout(ctx context.Context, req dto.LogoutRequest) error {
	// Валидация refresh токена
	claims, err := s.tokenManager.VerifyRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return fmt.Errorf("%s: invalid refresh token: %w", apiLogout, ErrInvalidToken)
	}

	// Отзываем токен по JTI
	if err := s.refreshTokensRepo.RevokeTokenByJTI(ctx, claims.JWTID); err != nil {
		return fmt.Errorf("%s: failed to revoke token: %w", apiLogout, err)
	}

	return nil
}
