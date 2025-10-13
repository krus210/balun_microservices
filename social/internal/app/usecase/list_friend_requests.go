package usecase

import (
	"context"
	"fmt"

	"social/internal/app/models"
)

const (
	apiListFriendRequests = "[SocialService][ListFriendRequests]"
)

func (s *SocialService) ListFriendRequests(ctx context.Context, toUserId models.UserID) ([]*models.FriendRequest, error) {
	toUserExists, err := s.usersService.CheckUserExists(ctx, toUserId)
	if err != nil {
		return nil, fmt.Errorf("%s: userService CheckUserExist error: %w", apiListFriendRequests, err)
	}
	if !toUserExists {
		return nil, models.ErrNotFound
	}

	friendRequests, _, err := s.socialRepo.GetFriendRequestsByToUserID(ctx, toUserId, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: socialRepo GetFriendRequestsByToUserID error: %w", apiListFriendRequests, err)
	}

	return friendRequests, nil
}
