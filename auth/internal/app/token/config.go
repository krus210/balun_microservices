package token

import "time"

// Config - конфигурация для TokenManager
type Config struct {
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Audience        []string
}
