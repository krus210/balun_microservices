package usecase

import (
	"context"
	"fmt"

	"users/internal/app/models"
	"users/internal/app/usecase/dto"
)

func (s *UsersService) CreateProfile(ctx context.Context, req dto.CreateProfileRequest) (*models.UserProfile, error) {
	user, err := s.usersRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[UsersService][CreateProfile] usersRepo GetUserByID error: %w", err)
	}
	if user != nil {
		return nil, models.ErrAlreadyExists
	}

	user = &models.UserProfile{
		UserID:    req.UserID,
		Nickname:  req.Nickname,
		Bio:       req.Bio,
		AvatarURL: req.AvatarURL,
	}

	saveUser, err := s.usersRepo.SaveUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("[UsersService][CreateProfile] usersRepo SaveUser error: %w", err)
	}

	return saveUser, nil
}
