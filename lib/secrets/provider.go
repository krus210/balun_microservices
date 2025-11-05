package secrets

import (
	"context"
	"errors"
)

// ErrSecretNotFound возвращается когда секрет не найден ни в одном источнике
var ErrSecretNotFound = errors.New("secret not found")

// SecretsProvider определяет интерфейс для получения секретов из различных источников
type SecretsProvider interface {
	// Get получает секрет в виде строки по ключу
	Get(ctx context.Context, key string) (string, error)

	// GetBytes получает секрет в виде байтов (для бинарных данных, например TLS ключей)
	GetBytes(ctx context.Context, key string) ([]byte, error)
}
