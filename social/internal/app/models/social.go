package models

import "time"

type FriendRequestStatus string

const (
	FriendRequestAccepted  FriendRequestStatus = "PENDING"
	FriendRequestRequested FriendRequestStatus = "REQUESTED"
	FriendRequestDeclined  FriendRequestStatus = "DECLINED"
)

type FriendRequest struct {
	ID         int64
	FromUserID int64
	ToUserID   int64
	Status     FriendRequestStatus
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}
