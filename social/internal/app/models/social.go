package models

import "time"

type FriendRequestStatus int

const (
	FriendRequestPending  FriendRequestStatus = 0
	FriendRequestAccepted FriendRequestStatus = 1
	FriendRequestDeclined FriendRequestStatus = 2
)

type UserID int64

type FriendRequestID int64

type FriendRequest struct {
	ID         FriendRequestID
	FromUserID UserID
	ToUserID   UserID
	Status     FriendRequestStatus
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}
