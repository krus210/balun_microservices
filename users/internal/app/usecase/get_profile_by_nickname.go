package usecase

import (
	"context"
	"fmt"

	"users/internal/app/models"
)

func (s *UsersService) GetProfileByNickname(ctx context.Context, nickname string) (*models.UserProfile, error) {
	user, err := s.usersRepo.GetUserByNickname(ctx, nickname)
	if err != nil {
		return nil, fmt.Errorf("[UsersService][GetProfileByNickname] usersRepo GetUserByNickname error: %w", err)
	}
	if user == nil {
		return nil, models.ErrNotFound
	}

	return user, nil
}
