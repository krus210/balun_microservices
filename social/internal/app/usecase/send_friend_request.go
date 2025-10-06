package usecase

import (
	"context"
	"fmt"

	"social/internal/app/models"
	"social/internal/app/usecase/dto"
)

func (s *SocialService) SendFriendRequest(ctx context.Context, req dto.FriendRequestDto) (*models.FriendRequest, error) {
	toUserExists, err := s.usersService.CheckUserExists(ctx, req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][SendFriendRequest] userService CheckUserExist error: %w", err)
	}
	if !toUserExists {
		return nil, models.ErrNotFound
	}

	friendRequest, err := s.socialRepo.GetFriendRequestByUserIDs(ctx, req.FromUserID, req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][SendFriendRequest] sociaRepo GetFriendRequests error: %w", err)
	}
	if friendRequest != nil {
		return nil, models.ErrAlreadyExists
	}

	friendRequest, err = s.socialRepo.GetFriendRequestByUserIDs(ctx, req.ToUserID, req.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][SendFriendRequest] sociaRepo GetFriendRequests error: %w", err)
	}

	if friendRequest != nil {
		return nil, models.ErrAlreadyExists
	}

	friendRequest = &models.FriendRequest{
		FromUserID: req.FromUserID,
		ToUserID:   req.ToUserID,
		Status:     models.FriendRequestRequested,
	}

	savedFriendRequest, err := s.socialRepo.SaveFriendRequest(ctx, friendRequest)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][SendFriendRequest] sociaRepo SaveFriendRequest error: %w", err)
	}

	return savedFriendRequest, nil
}
