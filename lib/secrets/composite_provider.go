package secrets

import (
	"context"
	"errors"
	"fmt"
)

// compositeProvider объединяет несколько провайдеров секретов в единую цепочку поиска
// Провайдеры проверяются по порядку, поиск останавливается при первом найденном значении
// Приватный тип, используйте NewSecretsProvider с несколькими опциями для создания
type compositeProvider struct {
	providers []SecretsProvider
}

// newCompositeProvider создает новый compositeProvider с указанными провайдерами
// Провайдеры будут опрашиваться в том порядке, в котором они переданы
//
// Рекомендуемый порядок: Env → File → Vault
// Это позволяет переопределить секреты через переменные окружения для локальной разработки
//
// Для внутреннего использования библиотекой, используйте NewSecretsProvider с несколькими опциями
func newCompositeProvider(providers ...SecretsProvider) *compositeProvider {
	return &compositeProvider{
		providers: providers,
	}
}

// Get получает секрет в виде строки, последовательно опрашивая все провайдеры
// Возвращает значение из первого провайдера, который нашел секрет
// Если ни один провайдер не нашел секрет, возвращает ErrSecretNotFound
func (p *compositeProvider) Get(ctx context.Context, key string) (string, error) {
	var errs []error

	for _, provider := range p.providers {
		value, err := provider.Get(ctx, key)
		if err == nil {
			return value, nil
		}

		// Если секрет не найден - продолжаем поиск в следующем провайдере
		if errors.Is(err, ErrSecretNotFound) {
			errs = append(errs, err)
			continue
		}

		// Для других ошибок (например, проблемы с подключением к Vault) возвращаем ошибку
		return "", fmt.Errorf("error from provider: %w", err)
	}

	// Секрет не найден ни в одном провайдере
	return "", fmt.Errorf("%w: %s (checked %d providers)", ErrSecretNotFound, key, len(p.providers))
}

// GetBytes получает секрет в виде байтов, последовательно опрашивая все провайдеры
// Возвращает значение из первого провайдера, который нашел секрет
// Если ни один провайдер не нашел секрет, возвращает ErrSecretNotFound
func (p *compositeProvider) GetBytes(ctx context.Context, key string) ([]byte, error) {
	var errs []error

	for _, provider := range p.providers {
		value, err := provider.GetBytes(ctx, key)
		if err == nil {
			return value, nil
		}

		// Если секрет не найден - продолжаем поиск в следующем провайдере
		if errors.Is(err, ErrSecretNotFound) {
			errs = append(errs, err)
			continue
		}

		// Для других ошибок (например, проблемы с подключением к Vault) возвращаем ошибку
		return nil, fmt.Errorf("error from provider: %w", err)
	}

	// Секрет не найден ни в одном провайдере
	return nil, fmt.Errorf("%w: %s (checked %d providers)", ErrSecretNotFound, key, len(p.providers))
}
