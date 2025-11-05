package models

import "time"

type UserProfile struct {
	UserID    string
	Nickname  string
	Bio       *string
	AvatarURL *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
