package usecase

import (
	"context"

	"auth/internal/app/models"
	"auth/internal/usecase/dto"
)

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	user, err := s.usersRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, models.ErrAlreadyExists
	}

	user, err = s.usersRepo.SaveUser(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	err = s.usersService.CreateUser(ctx, user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
