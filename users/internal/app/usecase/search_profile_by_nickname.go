package usecase

import (
	"context"
	"fmt"

	"users/internal/app/models"
	"users/internal/app/usecase/dto"
)

const (
	apiSearchByNickname = "[UsersService][SearchByNickname]"
)

func (s *UsersService) SearchByNickname(ctx context.Context, req dto.SearchByNicknameRequest) ([]*models.UserProfile, error) {
	users, err := s.usersRepo.SearchByNickname(ctx, req.Query, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("%s: usersRepo SearchByNickname error: %w", apiSearchByNickname, err)
	}

	return users, nil
}
