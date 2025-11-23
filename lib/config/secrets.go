package config

import (
	"context"
	"fmt"

	"github.com/sskorolev/balun_microservices/lib/secrets"
)

// SecretsLoader определяет интерфейс для загрузки секретов
type SecretsLoader interface {
	LoadSecrets(ctx context.Context, provider secrets.SecretsProvider) error
}

// LoadConfigWithSecrets загружает конфигурацию с поддержкой секретов
func LoadConfigWithSecrets(ctx context.Context, cfg Config, secretsLoader SecretsLoader) error {
	// Создаем SecretsProvider
	provider, err := NewSecretsProviderFromConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create secrets provider: %w", err)
	}

	// Загружаем секреты
	if err := secretsLoader.LoadSecrets(ctx, provider); err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	return nil
}

// NewSecretsProviderFromConfig создает SecretsProvider на основе конфигурации
func NewSecretsProviderFromConfig(ctx context.Context, cfg Config) (secrets.SecretsProvider, error) {
	secretsCfg := cfg.GetSecrets()
	if secretsCfg == nil {
		return nil, fmt.Errorf("secrets configuration not found")
	}

	// Выбираем конфигурацию в зависимости от окружения
	var providerCfg SecretsProviderConfig
	switch cfg.GetService().Environment {
	case "prod", "production":
		providerCfg = secretsCfg.Prod
	default:
		providerCfg = secretsCfg.Dev
	}

	// Формируем список опций
	opts := make([]secrets.Option, 0, 3)

	// Всегда добавляем ENV провайдер
	opts = append(opts, secrets.WithEnv(providerCfg.EnvPrefix))

	// Всегда добавляем File провайдер
	opts = append(opts, secrets.WithFile(providerCfg.FilePath))

	// Добавляем Vault провайдер только если он включен
	if providerCfg.Vault.Enabled {
		vaultCfg := secrets.VaultConfig{
			Address:    providerCfg.Vault.Address,
			Token:      providerCfg.Vault.Token,
			RoleID:     providerCfg.Vault.RoleID,
			SecretID:   providerCfg.Vault.SecretID,
			MountPath:  providerCfg.Vault.MountPath,
			SecretPath: providerCfg.Vault.SecretPath,
		}

		opts = append(opts, secrets.WithVault(vaultCfg))
	}

	return secrets.NewSecretsProvider(ctx, opts...)
}

// LoadDatabaseSecrets - helper для загрузки секретов БД
func LoadDatabaseSecrets(ctx context.Context, provider secrets.SecretsProvider, db *DatabaseConfig) error {
	// Пытаемся загрузить user из secrets
	user, err := provider.Get(ctx, "database.user")
	if err != nil {
		// Если не удалось получить из secrets provider, используем значение из конфига
		if db.User == "" {
			return fmt.Errorf("database user not found: %w", err)
		}
	} else {
		db.User = user
	}

	// Пытаемся загрузить password из secrets
	password, err := provider.Get(ctx, "database.password")
	if err != nil {
		// Если не удалось получить из secrets provider, используем значение из конфига
		if db.Password == "" {
			return fmt.Errorf("database password not found: %w", err)
		}
	} else {
		db.Password = password
	}

	// Финальная проверка
	if db.User == "" || db.Password == "" {
		return fmt.Errorf("database credentials are required")
	}

	return nil
}
