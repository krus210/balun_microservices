package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config представляет конфигурацию auth сервиса.
type Config struct {
	Service      ServiceConfig      `mapstructure:"service"`
	Server       ServerConfig       `mapstructure:"server"`
	UsersService UsersServiceConfig `mapstructure:"users_service"`
}

// ServiceConfig содержит общую информацию о сервисе.
type ServiceConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig содержит настройки серверов.
type ServerConfig struct {
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// GRPCConfig содержит настройки gRPC сервера.
type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

// UsersServiceConfig содержит настройки подключения к Users сервису.
type UsersServiceConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Load загружает конфигурацию из файла и переменных окружения с префиксом APP_.
func Load() (*Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults устанавливает значения по умолчанию.
func setDefaults(v *viper.Viper) {
	v.SetDefault("service.name", "auth")
	v.SetDefault("service.version", "1.0.0")
	v.SetDefault("service.environment", "dev")

	v.SetDefault("server.grpc.port", 8082)

	v.SetDefault("users_service.host", "users")
	v.SetDefault("users_service.port", 8082)
}

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.Server.GRPC.Port <= 0 || c.Server.GRPC.Port > 65535 {
		return fmt.Errorf("server.grpc.port must be between 1 and 65535")
	}

	if c.UsersService.Host == "" {
		return fmt.Errorf("users_service.host is required")
	}
	if c.UsersService.Port <= 0 || c.UsersService.Port > 65535 {
		return fmt.Errorf("users_service.port must be between 1 and 65535")
	}

	return nil
}
