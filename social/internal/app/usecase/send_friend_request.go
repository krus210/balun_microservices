package usecase

import (
	"context"
	"fmt"

	"social/internal/app/models"
	"social/internal/app/usecase/dto"
)

const (
	apiSendFriendRequest = "[SocialService][SendFriendRequest]"
)

func (s *SocialService) SendFriendRequest(ctx context.Context, req dto.FriendRequestDto) (*models.FriendRequest, error) {
	toUserExists, err := s.usersService.CheckUserExists(ctx, req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("%s: userService CheckUserExist error: %w", apiSendFriendRequest, err)
	}
	if !toUserExists {
		return nil, models.ErrNotFound
	}

	friendRequest, err := s.socialRepo.GetFriendRequestByUserIDs(ctx, req.FromUserID, req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("%s: socialRepo GetFriendRequestByUserIDs error: %w", apiSendFriendRequest, err)
	}
	if friendRequest != nil {
		return nil, models.ErrAlreadyExists
	}

	friendRequest, err = s.socialRepo.GetFriendRequestByUserIDs(ctx, req.ToUserID, req.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("%s: socialRepo GetFriendRequestByUserIDs error: %w", apiSendFriendRequest, err)
	}

	if friendRequest != nil {
		return nil, models.ErrAlreadyExists
	}

	friendRequest = &models.FriendRequest{
		FromUserID: req.FromUserID,
		ToUserID:   req.ToUserID,
		Status:     models.FriendRequestPending,
	}

	savedFriendRequest := &models.FriendRequest{}
	err = s.transactionalManager.RunReadCommitted(ctx,
		func(txCtx context.Context) error { // TRANSANCTION SCOPE

			savedFriendRequest, err = s.socialRepo.SaveFriendRequest(txCtx, friendRequest)
			if err != nil {
				return fmt.Errorf("%s: socialRepo SaveFriendRequest error: %w", apiSendFriendRequest, err)
			}

			err = s.outboxRepository.SaveFriendRequestCreatedID(txCtx, savedFriendRequest.ID)
			if err != nil {
				return fmt.Errorf("%s: outboxRepository SaveFriendRequestCreatedID error: %w", apiSendFriendRequest, err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return savedFriendRequest, nil
}
