package usecase

import (
	"context"
	"fmt"

	"users/internal/app/models"
)

const (
	apiGetProfileByID = "[UsersService][GetProfileByID]"
)

func (s *UsersService) GetProfileByID(ctx context.Context, id int64) (*models.UserProfile, error) {
	user, err := s.usersRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: usersRepo GetUserByID error: %w", apiGetProfileByID, err)
	}
	if user == nil {
		return nil, models.ErrNotFound
	}

	return user, nil
}
