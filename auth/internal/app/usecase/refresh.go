package usecase

import (
	"context"
	"fmt"
	"time"

	"auth/internal/app/usecase/dto"

	"auth/internal/app/models"

	"github.com/google/uuid"
)

func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*models.User, error) {
	user, err := s.usersRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[AuthService][Refresh] userRepo GetUserByID error: %w", err)
	}
	if user == nil {
		return nil, models.ErrNotFound
	}
	if user.Token.RefreshToken != req.RefreshToken {
		return nil, ErrWrongToken
	}

	user.Token = &models.UserToken{
		AccessToken:    uuid.New().String(),
		RefreshToken:   uuid.New().String(),
		TokenExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	err = s.usersRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("[AuthService][Refresh] userRepo UpdateUser error: %w", err)
	}

	return user, nil
}
