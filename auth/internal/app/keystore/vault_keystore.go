package keystore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sskorolev/balun_microservices/lib/secrets"
)

// VaultKeyStore - реализация KeyStore через HashiCorp Vault
type VaultKeyStore struct {
	provider  secrets.SecretsProvider
	basePath  string
	generator *RSAKeyGenerator
}

func NewVaultKeyStore(provider secrets.SecretsProvider, basePath string) *VaultKeyStore {
	if basePath == "" {
		basePath = "rsa-keys"
	}
	return &VaultKeyStore{
		provider:  provider,
		basePath:  basePath,
		generator: NewRSAKeyGenerator(),
	}
}

// GetActiveKey возвращает активный ключ
func (s *VaultKeyStore) GetActiveKey(ctx context.Context) (*RSAKey, error) {
	// Получаем метаданные о ключах (список активных)
	metadata, err := s.getKeysMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Ищем активный ключ
	for _, keyMeta := range metadata {
		if keyMeta.Status == KeyStatusActive {
			return s.GetKeyByKID(ctx, keyMeta.KID)
		}
	}

	return nil, ErrNoActiveKey
}

// GetKeyByKID возвращает ключ по KID
func (s *VaultKeyStore) GetKeyByKID(ctx context.Context, kid string) (*RSAKey, error) {
	// Путь к ключу в Vault: {basePath}/{kid}
	keyPath := fmt.Sprintf("%s/%s", s.basePath, kid)

	// Получаем данные ключа как JSON
	keyData, err := s.provider.Get(ctx, keyPath)
	if err != nil {
		if err == secrets.ErrSecretNotFound {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get key from vault: %w", err)
	}

	// Парсим JSON
	var key RSAKey
	if err := json.Unmarshal([]byte(keyData), &key); err != nil {
		return nil, fmt.Errorf("failed to unmarshal key data: %w", err)
	}

	return &key, nil
}

// GetAllActiveKeys возвращает все активные ключи (active + next)
func (s *VaultKeyStore) GetAllActiveKeys(ctx context.Context) ([]*RSAKey, error) {
	metadata, err := s.getKeysMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var keys []*RSAKey
	for _, keyMeta := range metadata {
		if keyMeta.Status == KeyStatusActive || keyMeta.Status == KeyStatusNext {
			key, err := s.GetKeyByKID(ctx, keyMeta.KID)
			if err != nil {
				return nil, err
			}
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// CreateKey создает новый RSA ключ
func (s *VaultKeyStore) CreateKey(ctx context.Context, status KeyStatus) (*RSAKey, error) {
	// Генерируем RSA ключ
	privateKey, err := s.generator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	privateKeyPEM, err := EncodePrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	publicKeyPEM, err := EncodePublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	kid := uuid.New().String()
	key := &RSAKey{
		KID:           kid,
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
		Status:        status,
	}

	// Сохраняем ключ в Vault
	_ = fmt.Sprintf("%s/%s", s.basePath, kid) // keyPath для будущей реализации
	_, err = json.Marshal(key)                // keyData для будущей реализации
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	// Note: lib/secrets не имеет метода Set, поэтому нужно использовать Vault API напрямую
	// Для упрощения, пока вернем ошибку
	return nil, fmt.Errorf("vault key creation not implemented yet - use Vault CLI or API")
}

// UpdateKeyStatus обновляет статус ключа
func (s *VaultKeyStore) UpdateKeyStatus(ctx context.Context, kid string, status KeyStatus) error {
	// Получаем текущий ключ
	key, err := s.GetKeyByKID(ctx, kid)
	if err != nil {
		return err
	}

	// Обновляем статус
	key.Status = status

	// Note: lib/secrets не имеет метода Set, поэтому нужно использовать Vault API напрямую
	return fmt.Errorf("vault key update not implemented yet - use Vault CLI or API")
}

// getKeysMetadata получает метаданные всех ключей
func (s *VaultKeyStore) getKeysMetadata(ctx context.Context) ([]RSAKey, error) {
	// Получаем список ключей из Vault
	// Для упрощения, будем хранить метаданные в отдельном ключе
	metadataPath := fmt.Sprintf("%s/_metadata", s.basePath)

	metadataJSON, err := s.provider.Get(ctx, metadataPath)
	if err != nil {
		if errors.Is(err, secrets.ErrSecretNotFound) {
			// Если метаданных нет, возвращаем пустой список
			return []RSAKey{}, nil
		}
		return nil, fmt.Errorf("failed to get keys metadata: %w", err)
	}

	var metadata []RSAKey
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}
