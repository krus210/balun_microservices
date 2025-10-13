package usecase

import (
	"context"
	"fmt"

	"users/internal/app/models"
	"users/internal/app/usecase/dto"
)

const (
	apiUpdateProfile = "[UsersService][UpdateProfile]"
)

func (s *UsersService) UpdateProfile(ctx context.Context, req dto.UpdateProfileRequest) (*models.UserProfile, error) {
	user, err := s.usersRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: usersRepo GetUserByID error: %w", apiUpdateProfile, err)
	}
	if user == nil {
		return nil, models.ErrNotFound
	}

	if req.Nickname != nil && *req.Nickname != "" {
		user.Nickname = *req.Nickname
	}
	if req.Bio != nil && *req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.AvatarURL != nil && *req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	updatedUser, err := s.usersRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("%s: usersRepo UpdateUser error: %w", apiUpdateProfile, err)
	}

	return updatedUser, nil
}
