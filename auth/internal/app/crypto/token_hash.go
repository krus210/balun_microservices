package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken - создает SHA-256 хеш токена для хранения в БД
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
