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

	updatedFriendRequest := &models.FriendRequest{}
	err = s.transactionalManager.RunReadCommitted(ctx,
		func(txCtx context.Context) error { // TRANSANCTION SCOPE

			updatedFriendRequest, err = s.socialRepo.UpdateFriendRequest(ctx, req.RequestID, models.FriendRequestDeclined)
			if err != nil {
				return fmt.Errorf("%s: socialRepo UpdateFriendRequest error: %w", apiAcceptFriendRequest, err)
			}

			err := s.outboxRepository.SaveFriendRequestUpdatedID(ctx, updatedFriendRequest.ID, models.FriendRequestDeclined)
			if err != nil {
				return fmt.Errorf("%s: outboxRepository SaveFriendRequestUpdatedID error: %w", apiSendFriendRequest, err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return updatedFriendRequest, nil
}
