package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config описывает конфигурацию gateway сервиса.
type Config struct {
	Service  ServiceConfig  `mapstructure:"service"`
	Server   ServerConfig   `mapstructure:"server"`
	Services ServicesConfig `mapstructure:"services"`
}

// ServiceConfig содержит общую информацию о сервисе.
type ServiceConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig описывает порты HTTP и gRPC серверов gateway.
type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// HTTPConfig описывает HTTP сервер.
type HTTPConfig struct {
	Port int `mapstructure:"port"`
}

// GRPCConfig описывает gRPC сервер.
type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

// ServicesConfig хранит адреса зависимых сервисов.
type ServicesConfig struct {
	Auth   TargetConfig `mapstructure:"auth"`
	Users  TargetConfig `mapstructure:"users"`
	Social TargetConfig `mapstructure:"social"`
	Chat   TargetConfig `mapstructure:"chat"`
}

// TargetConfig описывает хост и порт зависимого сервиса.
type TargetConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Load загружает конфигурацию из файла и переменных окружения (префикс APP_).
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
	v.SetDefault("service.name", "gateway")
	v.SetDefault("service.version", "1.0.0")
	v.SetDefault("service.environment", "dev")

	v.SetDefault("server.http.port", 8080)
	v.SetDefault("server.grpc.port", 8085)

	v.SetDefault("services.auth.host", "auth")
	v.SetDefault("services.auth.port", 8082)

	v.SetDefault("services.users.host", "users")
	v.SetDefault("services.users.port", 8082)

	v.SetDefault("services.social.host", "social")
	v.SetDefault("services.social.port", 8082)

	v.SetDefault("services.chat.host", "chat")
	v.SetDefault("services.chat.port", 8082)
}

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.Server.HTTP.Port <= 0 || c.Server.HTTP.Port > 65535 {
		return fmt.Errorf("server.http.port must be between 1 and 65535")
	}
	if c.Server.GRPC.Port <= 0 || c.Server.GRPC.Port > 65535 {
		return fmt.Errorf("server.grpc.port must be between 1 and 65535")
	}

	if err := c.Services.Auth.validate("services.auth"); err != nil {
		return err
	}
	if err := c.Services.Users.validate("services.users"); err != nil {
		return err
	}
	if err := c.Services.Social.validate("services.social"); err != nil {
		return err
	}
	if err := c.Services.Chat.validate("services.chat"); err != nil {
		return err
	}

	return nil
}

func (t TargetConfig) validate(prefix string) error {
	if t.Host == "" {
		return fmt.Errorf("%s.host is required", prefix)
	}
	if t.Port <= 0 || t.Port > 65535 {
		return fmt.Errorf("%s.port must be between 1 and 65535", prefix)
	}
	return nil
}
