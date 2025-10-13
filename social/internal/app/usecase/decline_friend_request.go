package usecase

import (
	"context"
	"fmt"

	"social/internal/app/usecase/dto"

	"social/internal/app/models"
)

const (
	apiDeclineFriendRequest = "[SocialService][DeclineFriendRequest]"
)

func (s *SocialService) DeclineFriendRequest(ctx context.Context, req dto.ChangeFriendRequestDto) (*models.FriendRequest, error) {
	friendRequest, err := s.socialRepo.GetFriendRequest(ctx, req.RequestID)
	if err != nil {
		return nil, fmt.Errorf("%s: socialRepo GetFriendRequest error: %w", apiDeclineFriendRequest, err)
	}
	if friendRequest == nil {
		return nil, models.ErrNotFound
	}

	if req.UserID != friendRequest.ToUserID {
		return nil, models.ErrPermissionDenied
	}

	updatedFriendRequest, err := s.socialRepo.UpdateFriendRequest(ctx, req.RequestID, models.FriendRequestDeclined)
	if err != nil {
		return nil, fmt.Errorf("%s: socialRepo UpdateFriendRequest error: %w", apiDeclineFriendRequest, err)
	}

	return updatedFriendRequest, nil
}
