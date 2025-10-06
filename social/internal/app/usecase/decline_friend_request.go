package usecase

import (
	"context"
	"fmt"

	"social/internal/app/usecase/dto"

	"social/internal/app/models"
)

func (s *SocialService) DeclineFriendRequest(ctx context.Context, req dto.ChangeFriendRequestDto) (*models.FriendRequest, error) {
	friendRequest, err := s.socialRepo.GetFriendRequest(ctx, req.RequestID)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][DeclineFriendRequest] sociaRepo GetFriendRequest error: %w", err)
	}
	if friendRequest == nil {
		return nil, models.ErrNotFound
	}

	if req.UserID != friendRequest.ToUserID {
		return nil, models.ErrPermissionDenied
	}

	updatedFriendRequest, err := s.socialRepo.UpdateFriendRequest(ctx, req.RequestID, models.FriendRequestDeclined)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][DeclineFriendRequest] sociaRepo UpdateFriendRequest error: %w", err)
	}

	return updatedFriendRequest, nil
}
