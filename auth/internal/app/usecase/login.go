package usecase

import (
	"context"
	"fmt"
	"time"

	"auth/internal/app/usecase/dto"

	"auth/internal/app/models"

	"github.com/google/uuid"
)

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*models.User, error) {
	user, err := s.usersRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("[AuthService][Login] userRepo GetUserByEmail error: %w", err)
	}
	if user == nil {
		return nil, models.ErrNotFound
	}
	if user.Password != req.Password {
		return nil, ErrWrongPassword
	}

	user.Token = &models.UserToken{
		AccessToken:    uuid.New().String(),
		RefreshToken:   uuid.New().String(),
		TokenExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	err = s.usersRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("[AuthService][Login] userRepo UpdateUser error: %w", err)
	}

	return user, nil
}
