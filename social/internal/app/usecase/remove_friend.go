package usecase

import (
	"context"
	"fmt"

	"social/internal/app/models"
	"social/internal/app/usecase/dto"
)

const (
	apiRemoveFriend = "[SocialService][RemoveFriend]"
)

func (s *SocialService) RemoveFriend(ctx context.Context, req dto.FriendRequestDto) error {
	friendRequest, err := s.getAcceptedFriendRequest(ctx, req.FromUserID, req.ToUserID)
	if err != nil {
		return fmt.Errorf("%s: getAcceptedFriendRequest error: %w", apiRemoveFriend, err)
	}

	err = s.socialRepo.DeleteFriendRequest(ctx, friendRequest.ID)
	if err != nil {
		return fmt.Errorf("%s: socialRepo DeleteFriendRequest error: %w", apiRemoveFriend, err)
	}

	return nil
}

func (s *SocialService) getAcceptedFriendRequest(ctx context.Context, firstUserID models.UserID, secondUserID models.UserID) (*models.FriendRequest, error) {
	friendRequest, err := s.socialRepo.GetFriendRequestByUserIDs(ctx, firstUserID, secondUserID)
	if err != nil {
		return nil, err
	}

	if friendRequest != nil && friendRequest.Status == models.FriendRequestAccepted {
		return friendRequest, nil
	}

	friendRequest, err = s.socialRepo.GetFriendRequestByUserIDs(ctx, secondUserID, firstUserID)
	if err != nil {
		return nil, err
	}

	if friendRequest == nil || friendRequest.Status != models.FriendRequestAccepted {
		return nil, models.ErrNotFound
	}

	return friendRequest, nil
}
