package usecase

import (
	"context"
	"fmt"

	"social/internal/app/models"
)

func (s *SocialService) AcceptFriendRequest(ctx context.Context, requestID int64) (*models.FriendRequest, error) {
	friendRequest, err := s.socialRepo.GetFriendRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][AcceptFriendRequest] sociaRepo GetFriendRequest error: %w", err)
	}
	if friendRequest == nil {
		return nil, models.ErrNotFound
	}

	updatedFriendRequest, err := s.socialRepo.UpdateFriendRequest(ctx, requestID, models.FriendRequestAccepted)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][AcceptFriendRequest] sociaRepo UpdateFriendRequest error: %w", err)
	}

	return updatedFriendRequest, nil
}
