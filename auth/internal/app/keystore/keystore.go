package keystore

import (
	"context"
	"errors"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrNoActiveKey = errors.New("no active key found")
)

// KeyStatus - статус RSA ключа
type KeyStatus string

const (
	KeyStatusActive  KeyStatus = "active"
	KeyStatusNext    KeyStatus = "next"
	KeyStatusExpired KeyStatus = "expired"
)

// RSAKey - структура RSA ключа
type RSAKey struct {
	KID           string
	PrivateKeyPEM string
	PublicKeyPEM  string
	Status        KeyStatus
}

// KeyStore - интерфейс для управления RSA ключами
type KeyStore interface {
	// GetActiveKey возвращает активный ключ для подписи
	GetActiveKey(ctx context.Context) (*RSAKey, error)

	// GetKeyByKID возвращает ключ по его ID
	GetKeyByKID(ctx context.Context, kid string) (*RSAKey, error)

	// GetAllActiveKeys возвращает все активные ключи (active + next)
	GetAllActiveKeys(ctx context.Context) ([]*RSAKey, error)

	// CreateKey создает новый RSA ключ
	CreateKey(ctx context.Context, status KeyStatus) (*RSAKey, error)

	// UpdateKeyStatus обновляет статус ключа
	UpdateKeyStatus(ctx context.Context, kid string, status KeyStatus) error
}
