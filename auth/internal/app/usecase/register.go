package usecase

import (
	"context"
	"fmt"

	"auth/internal/app/usecase/dto"

	"auth/internal/app/models"
)

const (
	apiRegister = "[AuthService][Register]"
)

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	user, err := s.usersRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: userRepo GetUserByEmail error: %w", apiRegister, err)
	}
	if user != nil {
		return nil, models.ErrAlreadyExists
	}

	user, err = s.usersRepo.SaveUser(ctx, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: userRepo SaveUser error: %w", apiRegister, err)
	}

	err = s.usersService.CreateUser(ctx, user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: usersService CreateUser error: %w", apiRegister, err)
	}

	return user, nil
}
