package models

import "time"

type FriendRequestStatus int

const (
	FriendRequestPending  FriendRequestStatus = 0
	FriendRequestAccepted FriendRequestStatus = 1
	FriendRequestDeclined FriendRequestStatus = 2
)

type FriendRequest struct {
	ID         int64
	FromUserID int64
	ToUserID   int64
	Status     FriendRequestStatus
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}
