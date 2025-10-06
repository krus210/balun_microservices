package usecase

import (
	"context"
	"fmt"

	"social/internal/app/models"
	"social/internal/app/usecase/dto"
)

func (s *SocialService) ListFriends(ctx context.Context, req dto.ListFriendsDto) (*dto.ListFriendsResponse, error) {
	toUserExists, err := s.usersService.CheckUserExists(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][ListFriends] userService CheckUserExist error: %w", err)
	}
	if !toUserExists {
		return nil, models.ErrNotFound
	}

	friends := make([]*models.FriendRequest, 0)

	toUserIDFriendRequests, nextCursor, err := s.socialRepo.GetFriendRequestsByToUserID(ctx, req.UserID, &req.Limit, req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][ListFriends] socialRepo GetFriendRequestsByToUserID error: %w", err)
	}

	for _, friendRequest := range toUserIDFriendRequests {
		if friendRequest != nil && friendRequest.Status == models.FriendRequestAccepted {
			friends = append(friends, friendRequest)
		}
	}

	if int64(len(friends)) >= req.Limit {
		return &dto.ListFriendsResponse{
			Friends:    friends,
			NextCursor: nextCursor,
		}, nil
	}

	newLimit := req.Limit - int64(len(friends))
	fromUserIDFriendRequests, nextCursor, err := s.socialRepo.GetFriendRequestsByFromUserID(ctx, req.UserID, &newLimit, nextCursor)
	if err != nil {
		return nil, fmt.Errorf("[SocialService][ListFriends] socialRepo GetFriendRequestsByToUserID error: %w", err)
	}

	for _, friendRequest := range fromUserIDFriendRequests {
		if friendRequest != nil && friendRequest.Status == models.FriendRequestAccepted {
			friends = append(friends, friendRequest)
		}
	}

	return &dto.ListFriendsResponse{
		Friends:    friends,
		NextCursor: nextCursor,
	}, nil
}
