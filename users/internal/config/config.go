package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config представляет полную конфигурацию users сервиса
type Config struct {
	Service  ServiceConfig  `mapstructure:"service"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Secrets  SecretsConfig  `mapstructure:"secrets"`
}

// ServiceConfig содержит общую информацию о сервисе
type ServiceConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig содержит настройки серверов
type ServerConfig struct {
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// GRPCConfig содержит настройки gRPC сервера
type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig содержит настройки подключения к базе данных
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time"`
}

// DSN возвращает строку подключения к PostgreSQL
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

// SecretsConfig содержит настройки провайдера секретов для разных окружений
type SecretsConfig struct {
	Dev  SecretsProviderConfig `mapstructure:"dev"`
	Prod SecretsProviderConfig `mapstructure:"prod"`
}

// SecretsProviderConfig содержит настройки провайдера секретов
type SecretsProviderConfig struct {
	EnvPrefix string             `mapstructure:"env_prefix"`
	FilePath  string             `mapstructure:"file_path"`
	Vault     VaultSecretsConfig `mapstructure:"vault"`
}

// VaultSecretsConfig содержит настройки HashiCorp Vault
type VaultSecretsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Address    string `mapstructure:"address"`
	Token      string `mapstructure:"token"`
	RoleID     string `mapstructure:"role_id"`
	SecretID   string `mapstructure:"secret_id"`
	MountPath  string `mapstructure:"mount_path"`
	SecretPath string `mapstructure:"secret_path"`
}

// Load загружает конфигурацию из файла и переменных окружения с префиксом APP_
func Load() (*Config, error) {
	v := viper.New()

	// Устанавливаем defaults
	setDefaults(v)

	// Настройка файла конфигурации
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Читаем файл конфигурации (если существует)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Файл не найден - это нормально, будем использовать defaults и env
	}

	// Настройка переменных окружения
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Явный биндинг для вложенных Vault переменных окружения
	// (AutomaticEnv не всегда корректно мапит глубоко вложенные структуры)
	_ = v.BindEnv("secrets.prod.vault.token")
	_ = v.BindEnv("secrets.prod.vault.role_id")
	_ = v.BindEnv("secrets.prod.vault.secret_id")
	_ = v.BindEnv("secrets.dev.vault.token")
	_ = v.BindEnv("secrets.dev.vault.role_id")
	_ = v.BindEnv("secrets.dev.vault.secret_id")

	// Десериализация в структуру
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Валидация конфигурации
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults устанавливает значения по умолчанию
func setDefaults(v *viper.Viper) {
	// Service defaults
	v.SetDefault("service.name", "users")
	v.SetDefault("service.version", "1.0.0")
	v.SetDefault("service.environment", "dev")

	// Server defaults
	v.SetDefault("server.grpc.port", 8082)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "users")
	// user и password НЕ имеют defaults - должны передаваться через ENV или secrets
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
	v.SetDefault("secrets.prod.vault.secret_path", "users/production")
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	// Проверка обязательных полей базы данных
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}

	// Проверка портов
	if c.Server.GRPC.Port <= 0 || c.Server.GRPC.Port > 65535 {
		return fmt.Errorf("server.grpc.port must be between 1 and 65535")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("database.port must be between 1 and 65535")
	}

	return nil
}
