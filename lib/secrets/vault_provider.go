package secrets

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
)

// vaultProvider читает секреты из HashiCorp Vault
// Приватный тип, используйте NewSecretsProvider с WithVault для создания
type vaultProvider struct {
	client     *api.Client
	mountPath  string // путь к KV engine (по умолчанию "secret")
	secretPath string // базовый путь к секретам внутри mount
}

// VaultConfig содержит конфигурацию для подключения к Vault
type VaultConfig struct {
	// Address - адрес Vault сервера (например, "https://vault.example.com:8200")
	Address string

	// Token - токен для аутентификации (если используется)
	Token string

	// AppRole аутентификация (опционально)
	RoleID   string
	SecretID string

	// MountPath - путь к KV v2 engine (по умолчанию "secret")
	MountPath string

	// SecretPath - базовый путь к секретам (например, "myapp/production")
	SecretPath string
}

// newVaultProvider создает новый vaultProvider с указанной конфигурацией
// Для внутреннего использования библиотекой, используйте NewSecretsProvider с WithVault
func newVaultProvider(ctx context.Context, cfg VaultConfig) (*vaultProvider, error) {
	// Создаем конфигурацию клиента
	config := api.DefaultConfig()
	config.Address = cfg.Address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Аутентификация
	if cfg.Token != "" {
		// Используем Token аутентификацию
		client.SetToken(cfg.Token)
	} else if cfg.RoleID != "" && cfg.SecretID != "" {
		// Используем AppRole аутентификацию
		if err := authenticateWithAppRole(ctx, client, cfg.RoleID, cfg.SecretID); err != nil {
			return nil, fmt.Errorf("failed to authenticate with approle: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either Token or AppRole credentials must be provided")
	}

	// Устанавливаем mount path по умолчанию
	mountPath := cfg.MountPath
	if mountPath == "" {
		mountPath = "secret"
	}

	return &vaultProvider{
		client:     client,
		mountPath:  mountPath,
		secretPath: cfg.SecretPath,
	}, nil
}

// authenticateWithAppRole выполняет аутентификацию через AppRole
func authenticateWithAppRole(ctx context.Context, client *api.Client, roleID, secretID string) error {
	appRoleAuth, err := approle.NewAppRoleAuth(
		roleID,
		&approle.SecretID{FromString: secretID},
	)
	if err != nil {
		return fmt.Errorf("unable to initialize AppRole auth: %w", err)
	}

	authInfo, err := client.Auth().Login(ctx, appRoleAuth)
	if err != nil {
		return fmt.Errorf("unable to login with AppRole: %w", err)
	}

	if authInfo == nil {
		return fmt.Errorf("no auth info was returned")
	}

	return nil
}

// Get получает секрет из Vault в виде строки
// Ключ может быть в формате "path/to/secret" или "secret.key"
func (p *vaultProvider) Get(ctx context.Context, key string) (string, error) {
	// Формируем полный путь к секрету
	// Для KV v2: /mountPath/data/secretPath
	fullPath := fmt.Sprintf("%s/data/%s", p.mountPath, p.buildSecretPath(key))

	secret, err := p.client.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read secret from vault: %w", err)
	}

	if secret == nil {
		return "", fmt.Errorf("%w: %s at path %s", ErrSecretNotFound, key, fullPath)
	}

	// KV v2 возвращает данные в secret.Data["data"]
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected data format from vault for key %s", key)
	}

	// Ищем значение по ключу
	value, exists := data[key]
	if !exists {
		return "", fmt.Errorf("%w: %s in vault path %s", ErrSecretNotFound, key, fullPath)
	}

	// Преобразуем в строку
	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("secret %s is not a string", key)
	}

	return strValue, nil
}

// GetBytes получает секрет из Vault в виде байтов
// Если значение закодировано в base64, автоматически декодирует его
func (p *vaultProvider) GetBytes(ctx context.Context, key string) ([]byte, error) {
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

// buildSecretPath формирует полный путь к секрету
func (p *vaultProvider) buildSecretPath(key string) string {
	if p.secretPath != "" {
		path := strings.Trim(p.secretPath, "/")
		if path == "" {
			return ""
		}
		return strings.TrimPrefix(path, p.mountPath+"/")
	}
	return strings.TrimPrefix(strings.Trim(key, "/"), p.mountPath+"/")
}
