package secrets

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// fileProvider читает секреты из YAML файла
// Приватный тип, используйте NewSecretsProvider с WithFile для создания
type fileProvider struct {
	filePath string
	secrets  map[string]interface{}
	mu       sync.RWMutex
	loaded   bool
}

// newFileProvider создает новый fileProvider для чтения секретов из YAML файла
// filePath должен быть абсолютным путем к файлу
// Для внутреннего использования библиотекой, используйте NewSecretsProvider с WithFile
func newFileProvider(filePath string) *fileProvider {
	return &fileProvider{
		filePath: filePath,
		secrets:  make(map[string]interface{}),
	}
}

// load загружает секреты из YAML файла
func (p *fileProvider) load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.loaded {
		return nil
	}

	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return fmt.Errorf("failed to read secrets file %s: %w", p.filePath, err)
	}

	if err := yaml.Unmarshal(data, &p.secrets); err != nil {
		return fmt.Errorf("failed to parse secrets file %s: %w", p.filePath, err)
	}

	p.loaded = true
	return nil
}

// Get получает секрет из файла в виде строки
func (p *fileProvider) Get(ctx context.Context, key string) (string, error) {
	if err := p.load(); err != nil {
		return "", err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	value, exists := p.secrets[key]
	if !exists {
		return "", fmt.Errorf("%w: %s in file %s", ErrSecretNotFound, key, p.filePath)
	}

	// Преобразуем значение в строку
	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("secret %s in file %s is not a string", key, p.filePath)
	}

	return strValue, nil
}

// GetBytes получает секрет из файла в виде байтов
// Если значение закодировано в base64, автоматически декодирует его
func (p *fileProvider) GetBytes(ctx context.Context, key string) ([]byte, error) {
	value, err := p.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Пробуем декодировать base64
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// Если не получилось декодировать - возвращаем как есть
		return []byte(value), nil
	}

	return decoded, nil
}
