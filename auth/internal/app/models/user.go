package models

import "time"

type User struct {
	ID        int64
	Email     string
	Password  string
	Token     *UserToken
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserToken struct {
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt time.Time
}
