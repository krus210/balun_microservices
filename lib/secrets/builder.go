package secrets

import (
	"context"
	"fmt"
)

// Option определяет функциональную опцию для конфигурации SecretsProvider
type Option func(*config)

// config содержит настройки для создания SecretsProvider
type config struct {
	providers []providerConfig
}

// providerConfig описывает конфигурацию отдельного провайдера
type providerConfig struct {
	providerType string
	envPrefix    string
	filePath     string
	vaultConfig  VaultConfig
}

// NewSecretsProvider создает новый SecretsProvider с указанными опциями
//
// По умолчанию (без опций) создается EnvProvider с префиксом "APP_"
//
// Если указано несколько провайдеров, они объединяются в CompositeProvider
// в том порядке, в котором были добавлены опции (рекомендуется: Env → File → Vault)
//
// Примеры:
//
//	// По умолчанию - только ENV с префиксом APP_
//	provider, err := secrets.NewSecretsProvider(ctx)
//
//	// ENV + File
//	provider, err := secrets.NewSecretsProvider(ctx,
//	    secrets.WithEnv("APP_"),
//	    secrets.WithFile("./secrets.yaml"),
//	)
//
//	// Полная цепочка: ENV + File + Vault
//	provider, err := secrets.NewSecretsProvider(ctx,
//	    secrets.WithEnv("APP_"),
//	    secrets.WithFile("/etc/secrets.yaml"),
//	    secrets.WithVault(secrets.VaultConfig{
//	        Address:    "http://vault:8200",
//	        Token:      "token",
//	        SecretPath: "myapp/prod",
//	    }),
//	)
func NewSecretsProvider(ctx context.Context, opts ...Option) (SecretsProvider, error) {
	cfg := &config{
		providers: make([]providerConfig, 0),
	}

	// Применяем опции
	for _, opt := range opts {
		opt(cfg)
	}

	// Если не указано ни одной опции, используем EnvProvider по умолчанию
	if len(cfg.providers) == 0 {
		return newEnvProvider(withPrefix("APP_")), nil
	}

	// Создаем провайдеры
	providers := make([]SecretsProvider, 0, len(cfg.providers))
	for _, pc := range cfg.providers {
		var provider SecretsProvider
		var err error

		switch pc.providerType {
		case "env":
			provider = newEnvProvider(withPrefix(pc.envPrefix))

		case "file":
			provider = newFileProvider(pc.filePath)

		case "vault":
			provider, err = newVaultProvider(ctx, pc.vaultConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create vault provider: %w", err)
			}

		default:
			return nil, fmt.Errorf("unknown provider type: %s", pc.providerType)
		}

		providers = append(providers, provider)
	}

	// Если только один провайдер - возвращаем его напрямую
	if len(providers) == 1 {
		return providers[0], nil
	}

	// Если несколько провайдеров - создаем композитный
	return newCompositeProvider(providers...), nil
}

// WithEnv добавляет EnvProvider с указанным префиксом
//
// Префикс автоматически добавляется к имени переменной окружения.
// Например, с префиксом "APP_" ключ "database.password" будет искаться
// в переменной окружения "APP_DATABASE_PASSWORD"
//
// Пример:
//
//	secrets.NewSecretsProvider(ctx, secrets.WithEnv("APP_"))
func WithEnv(prefix string) Option {
	return func(c *config) {
		c.providers = append(c.providers, providerConfig{
			providerType: "env",
			envPrefix:    prefix,
		})
	}
}

// WithEnvPrefix - алиас для WithEnv для совместимости с распространенным именованием
//
// Пример:
//
//	secrets.NewSecretsProvider(ctx, secrets.WithEnvPrefix("APP_"))
func WithEnvPrefix(prefix string) Option {
	return WithEnv(prefix)
}

// WithFile добавляет FileProvider для чтения секретов из YAML файла
//
// filePath должен быть абсолютным путем к YAML файлу с секретами.
// Файл должен содержать плоскую структуру ключ-значение.
//
// Формат файла:
//
//	database.password: "my-secret"
//	api.token: "abc-xyz"
//	tls.cert: "LS0tLS1CRUdJTi..."  # base64 для бинарных данных
//
// Пример:
//
//	secrets.NewSecretsProvider(ctx, secrets.WithFile("./secrets.yaml"))
func WithFile(filePath string) Option {
	return func(c *config) {
		c.providers = append(c.providers, providerConfig{
			providerType: "file",
			filePath:     filePath,
		})
	}
}

// WithVault добавляет VaultProvider для чтения секретов из HashiCorp Vault
//
// Поддерживаются два метода аутентификации:
//   - Token: указать cfg.Token
//   - AppRole: указать cfg.RoleID и cfg.SecretID
//
// Пример с Token:
//
//	secrets.NewSecretsProvider(ctx, secrets.WithVault(secrets.VaultConfig{
//	    Address:    "http://localhost:8200",
//	    Token:      "dev-root-token",
//	    MountPath:  "secret",
//	    SecretPath: "myapp/production",
//	}))
//
// Пример с AppRole:
//
//	secrets.NewSecretsProvider(ctx, secrets.WithVault(secrets.VaultConfig{
//	    Address:    "http://localhost:8200",
//	    RoleID:     "your-role-id",
//	    SecretID:   "your-secret-id",
//	    SecretPath: "myapp/production",
//	}))
func WithVault(cfg VaultConfig) Option {
	return func(c *config) {
		c.providers = append(c.providers, providerConfig{
			providerType: "vault",
			vaultConfig:  cfg,
		})
	}
}
