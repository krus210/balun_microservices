package dto

import "social/internal/app/models"

type FriendRequestDto struct {
	FromUserID int64
	ToUserID   int64
}

type ListFriendsDto struct {
	UserID int64
	Limit  int64
	Cursor *string
}

type ListFriendsResponse struct {
	Friends    []*models.FriendRequest
	NextCursor *string
}
