package dto

import "social/internal/app/models"

type FriendRequestDto struct {
	FromUserID models.UserID
	ToUserID   models.UserID
}

type ChangeFriendRequestDto struct {
	UserID    models.UserID
	RequestID models.FriendRequestID
}

type ListFriendsDto struct {
	UserID models.UserID
	Limit  int64
	Cursor *string
}

type ListFriendsResponse struct {
	Friends    []*models.FriendRequest
	NextCursor *string
}
