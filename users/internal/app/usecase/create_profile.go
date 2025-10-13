package usecase

import (
	"context"
	"fmt"

	"users/internal/app/models"
	"users/internal/app/usecase/dto"
)

const (
	apiCreateProfile = "[UsersService][CreateProfile]"
)

func (s *UsersService) CreateProfile(ctx context.Context, req dto.CreateProfileRequest) (*models.UserProfile, error) {
	user, err := s.usersRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: usersRepo GetUserByID error: %w", apiCreateProfile, err)
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
		return nil, fmt.Errorf("%s: usersRepo SaveUser error: %w", apiCreateProfile, err)
	}

	return saveUser, nil
}
