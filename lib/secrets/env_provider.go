package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// envProvider читает секреты из переменных окружения
// Приватный тип, используйте NewSecretsProvider с WithEnv/WithEnvPrefix для создания
type envProvider struct {
	prefix string
}

// envProviderOption определяет опции для конфигурации envProvider
type envProviderOption func(*envProvider)

// withPrefix устанавливает префикс для переменных окружения (например, "APP_")
func withPrefix(prefix string) envProviderOption {
	return func(p *envProvider) {
		p.prefix = prefix
	}
}

// newEnvProvider создает новый envProvider с указанными опциями
// По умолчанию использует префикс "APP_"
// Для внутреннего использования библиотекой, используйте NewSecretsProvider с WithEnv/WithEnvPrefix
func newEnvProvider(opts ...envProviderOption) *envProvider {
	p := &envProvider{
		prefix: "APP_", // значение по умолчанию
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Get получает секрет из переменной окружения в виде строки
// Ключ автоматически преобразуется: точки заменяются на подчеркивания и добавляется префикс
// Например: "database.password" -> "APP_DATABASE_PASSWORD"
func (p *envProvider) Get(ctx context.Context, key string) (string, error) {
	envKey := p.buildEnvKey(key)
	value := os.Getenv(envKey)

	if value == "" {
		return "", fmt.Errorf("%w: %s (env key: %s)", ErrSecretNotFound, key, envKey)
	}

	return value, nil
}

// GetBytes получает секрет из переменной окружения в виде байтов
func (p *envProvider) GetBytes(ctx context.Context, key string) ([]byte, error) {
	value, err := p.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	return []byte(value), nil
}

// buildEnvKey преобразует ключ в имя переменной окружения
// Заменяет точки и дефисы на подчеркивания, приводит к верхнему регистру и добавляет префикс
func (p *envProvider) buildEnvKey(key string) string {
	// Заменяем точки и дефисы на подчеркивания
	envKey := strings.ReplaceAll(key, ".", "_")
	envKey = strings.ReplaceAll(envKey, "-", "_")

	// Приводим к верхнему регистру
	envKey = strings.ToUpper(envKey)

	// Добавляем префикс
	if p.prefix != "" {
		envKey = p.prefix + envKey
	}

	return envKey
}
