package models

import (
	"time"
)

type User struct {
	ID           string     `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Token        *UserToken `db:"-"` // не мапится на БД
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

type UserToken struct {
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt time.Time
}

// RefreshToken - модель refresh токена для БД
type RefreshToken struct {
	ID            string     `db:"id"`
	UserID        string     `db:"user_id"`
	TokenHash     string     `db:"token_hash"` // SHA-256 хеш токена
	JTI           string     `db:"jti"`        // JWT ID
	DeviceID      *string    `db:"device_id"`
	ExpiresAt     time.Time  `db:"expires_at"`
	UsedAt        *time.Time `db:"used_at"`
	ReplacedByJTI *string    `db:"replaced_by_jti"`
	CreatedAt     time.Time  `db:"created_at"`
}
