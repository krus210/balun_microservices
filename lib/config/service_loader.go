package config

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/sskorolev/balun_microservices/lib/secrets"
)

// ServiceOption определяет опцию для загрузки конфигурации сервиса
type ServiceOption func(*serviceOptions)

// serviceOptions содержит параметры для загрузки конфигурации
type serviceOptions struct {
	serviceName       string
	databaseName      string
	grpcPort          int
	secretsPathSuffix string
	customDefaults    func(*viper.Viper)
}

// WithDatabaseName устанавливает имя базы данных
func WithDatabaseName(name string) ServiceOption {
	return func(opts *serviceOptions) {
		opts.databaseName = name
	}
}

// WithGRPCPort устанавливает порт gRPC сервера
func WithGRPCPort(port int) ServiceOption {
	return func(opts *serviceOptions) {
		opts.grpcPort = port
	}
}

// WithSecretsPathSuffix устанавливает суффикс пути для Vault secrets
func WithSecretsPathSuffix(suffix string) ServiceOption {
	return func(opts *serviceOptions) {
		opts.secretsPathSuffix = suffix
	}
}

// WithCustomDefaults устанавливает кастомную функцию для установки defaults
func WithCustomDefaults(fn func(*viper.Viper)) ServiceOption {
	return func(opts *serviceOptions) {
		opts.customDefaults = fn
	}
}

// LoadServiceConfig загружает стандартную конфигурацию сервиса с секретами
func LoadServiceConfig(ctx context.Context, serviceName string, opts ...ServiceOption) (*StandardServiceConfig, error) {
	// Применяем опции
	options := &serviceOptions{
		serviceName:  serviceName,
		databaseName: serviceName, // по умолчанию совпадает с именем сервиса
		grpcPort:     8082,        // стандартный порт для всех сервисов
	}
	for _, opt := range opts {
		opt(options)
	}

	// Создаём функцию setDefaults
	setDefaults := func(v *viper.Viper) {
		// Service defaults
		v.SetDefault("service.name", options.serviceName)
		v.SetDefault("service.version", "1.0.0")
		v.SetDefault("service.environment", "dev")

		// Server defaults
		v.SetDefault("server.grpc.port", options.grpcPort)

		// Database defaults
		v.SetDefault("database.host", "localhost")
		v.SetDefault("database.port", 5432)
		v.SetDefault("database.name", options.databaseName)
		v.SetDefault("database.sslmode", "disable")
		v.SetDefault("database.max_conn_idle_time", time.Minute)

		// Secrets defaults для dev
		v.SetDefault("secrets.dev.env_prefix", "APP_")
		v.SetDefault("secrets.dev.file_path", "./secrets.yaml")
		v.SetDefault("secrets.dev.vault.enabled", false)

		// Secrets defaults для prod
		v.SetDefault("secrets.prod.env_prefix", "APP_")
		v.SetDefault("secrets.prod.file_path", "/etc/secrets/secrets.yaml")
		v.SetDefault("secrets.prod.vault.enabled", true)
		v.SetDefault("secrets.prod.vault.address", "http://vault:8200")
		v.SetDefault("secrets.prod.vault.mount_path", "secret")

		// Если указан суффикс пути для secrets, используем его
		if options.secretsPathSuffix != "" {
			v.SetDefault("secrets.prod.vault.secret_path", options.secretsPathSuffix)
		} else {
			v.SetDefault("secrets.prod.vault.secret_path", fmt.Sprintf("%s/production", options.serviceName))
		}

		// Вызываем кастомную функцию если она есть
		if options.customDefaults != nil {
			options.customDefaults(v)
		}
	}

	// Загружаем конфигурацию
	cfg := &StandardServiceConfig{}
	err := Load(LoadOptions{
		ConfigName:  "config",
		ConfigType:  "yaml",
		ConfigPaths: []string{".", "./config"},
		EnvPrefix:   "APP",
		SetDefaults: setDefaults,
		Target:      cfg,
	})
	if err != nil {
		return nil, err
	}

	// Валидируем
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Загружаем секреты
	if err := LoadConfigWithSecrets(ctx, cfg, &databaseSecretsLoader{db: &cfg.Database}); err != nil {
		return nil, err
	}

	return cfg, nil
}

// databaseSecretsLoader реализует SecretsLoader для загрузки секретов БД
type databaseSecretsLoader struct {
	db *DatabaseConfig
}

func (l *databaseSecretsLoader) LoadSecrets(ctx context.Context, provider secrets.SecretsProvider) error {
	return LoadDatabaseSecrets(ctx, provider, l.db)
}
