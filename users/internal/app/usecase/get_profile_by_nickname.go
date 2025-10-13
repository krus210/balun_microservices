package usecase

import (
	"context"
	"fmt"

	"users/internal/app/models"
)

const (
	apiGetProfileByNickname = "[UsersService][GetProfileByNickname]"
)

func (s *UsersService) GetProfileByNickname(ctx context.Context, nickname string) (*models.UserProfile, error) {
	user, err := s.usersRepo.GetUserByNickname(ctx, nickname)
	if err != nil {
		return nil, fmt.Errorf("%s: usersRepo GetUserByNickname error: %w", apiGetProfileByNickname, err)
	}
	if user == nil {
		return nil, models.ErrNotFound
	}

	return user, nil
}
