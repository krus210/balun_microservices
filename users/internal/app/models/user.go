package models

import "time"

type UserProfile struct {
	UserID    int64
	Nickname  string
	Bio       *string
	AvatarURL *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
