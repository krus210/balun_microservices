package usecase

import (
	"context"
	"fmt"

	"users/internal/app/models"
	"users/internal/app/usecase/dto"
)

func (s *UsersService) SearchByNickname(ctx context.Context, req dto.SearchByNicknameRequest) ([]*models.UserProfile, error) {
	users, err := s.usersRepo.SearchByNickname(ctx, req.Query, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("[UsersService][SearchByNickname] usersRepo SearchByNickname error: %w", err)
	}

	return users, nil
}
