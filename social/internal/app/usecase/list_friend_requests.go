package usecase

import (
	"context"
	"fmt"

	"social/internal/app/models"
)

func (s *SocialService) ListFriendRequests(ctx context.Context, toUserId models.UserID) ([]*models.FriendRequest, error) {
	toUserExists, err := s.usersService.CheckUserExists(ctx, toUserId)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][ListFriendRequests] userService CheckUserExist error: %w", err)
	}
	if !toUserExists {
		return nil, models.ErrNotFound
	}

	friendRequests, _, err := s.socialRepo.GetFriendRequestsByToUserID(ctx, toUserId, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][ListFriendRequests] socialRepo GetFriendRequestsByToUserID error: %w", err)
	}

	return friendRequests, nil
}
