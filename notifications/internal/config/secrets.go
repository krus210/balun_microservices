package config

import (
	"context"
	"fmt"

	"github.com/sskorolev/balun_microservices/lib/secrets"
)

// LoadWithSecrets загружает конфигурацию и инициализирует SecretsProvider
// для безопасного хранения паролей и других чувствительных данных
func LoadWithSecrets(ctx context.Context) (*Config, error) {
	// Загружаем обычную конфигурацию через Viper
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Создаем SecretsProvider на основе конфигурации
	provider, err := newSecretsProviderFromConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets provider: %w", err)
	}

	// Загружаем чувствительные данные из SecretsProvider
	// database.user
	dbUser, err := provider.Get(ctx, "database.user")
	if err != nil {
		// Если не удалось получить из secrets provider, используем значение из viper
		// (для обратной совместимости)
		if cfg.Database.User == "" {
			return nil, fmt.Errorf("database user not found in secrets or config: %w", err)
		}
	} else {
		cfg.Database.User = dbUser
	}

	// database.password
	dbPassword, err := provider.Get(ctx, "database.password")
	if err != nil {
		// Если не удалось получить из secrets provider, используем значение из viper
		// (для обратной совместимости)
		if cfg.Database.Password == "" {
			return nil, fmt.Errorf("database password not found in secrets or config: %w", err)
		}
	} else {
		cfg.Database.Password = dbPassword
	}

	// Финальная валидация database credentials после всех попыток загрузки
	if cfg.Database.User == "" {
		return nil, fmt.Errorf("database.user is required")
	}
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("database.password is required")
	}

	return cfg, nil
}

// newSecretsProviderFromConfig создает SecretsProvider на основе конфигурации
func newSecretsProviderFromConfig(ctx context.Context, cfg *Config) (secrets.SecretsProvider, error) {
	// Выбираем конфигурацию в зависимости от окружения
	var providerCfg SecretsProviderConfig
	switch cfg.Service.Environment {
	case "prod", "production":
		providerCfg = cfg.Secrets.Prod
	default:
		providerCfg = cfg.Secrets.Dev
	}

	// Формируем список опций для создания провайдера
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
