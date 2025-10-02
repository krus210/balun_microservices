package usecase

import (
	"context"
	"fmt"

	"auth/internal/app/usecase/dto"

	"auth/internal/app/models"
)

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	user, err := s.usersRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("[AuthService][Register] userRepo GetUserByEmail error: %w", err)
	}
	if user != nil {
		return nil, models.ErrAlreadyExists
	}

	user, err = s.usersRepo.SaveUser(ctx, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("[AuthService][Register] userRepo SaveUser error: %w", err)
	}

	err = s.usersService.CreateUser(ctx, user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("[AuthService][Register] usersService CreateUser error: %w", err)
	}

	return user, nil
}
