package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config представляет полную конфигурацию social сервиса
type Config struct {
	Service              ServiceConfig              `mapstructure:"service"`
	Server               ServerConfig               `mapstructure:"server"`
	Database             DatabaseConfig             `mapstructure:"database"`
	Kafka                KafkaConfig                `mapstructure:"kafka"`
	Outbox               OutboxConfig               `mapstructure:"outbox"`
	FriendRequestHandler FriendRequestHandlerConfig `mapstructure:"friend_request_handler"`
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

// KafkaConfig содержит настройки Kafka
type KafkaConfig struct {
	Brokers []string     `mapstructure:"brokers"`
	Topics  TopicsConfig `mapstructure:"topics"`
}

// BrokersList возвращает список брокеров в виде строки через запятую
func (c KafkaConfig) BrokersList() []string {
	return c.Brokers
}

// TopicsConfig содержит названия топиков
type TopicsConfig struct {
	FriendRequestEvents string `mapstructure:"friend_request_events"`
}

// OutboxConfig содержит настройки outbox процессора
type OutboxConfig struct {
	Processor ProcessorConfig `mapstructure:"processor"`
}

// ProcessorConfig содержит настройки процессора событий
type ProcessorConfig struct {
	BatchSize     int           `mapstructure:"batch_size"`
	MaxRetry      int           `mapstructure:"max_retry"`
	RetryInterval time.Duration `mapstructure:"retry_interval"`
	Window        time.Duration `mapstructure:"window"`
}

// FriendRequestHandlerConfig содержит настройки обработчика заявок в друзья
type FriendRequestHandlerConfig struct {
	BatchSize int `mapstructure:"batch_size"`
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
	v.SetDefault("service.name", "social")
	v.SetDefault("service.version", "1.0.0")
	v.SetDefault("service.environment", "dev")

	// Server defaults
	v.SetDefault("server.grpc.port", 8082)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "social")
	// user и password НЕ имеют defaults - должны передаваться через ENV или secrets
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_conn_idle_time", time.Minute)

	// Kafka defaults
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.topics.friend_request_events", "friend-request-events")

	// Outbox defaults
	v.SetDefault("outbox.processor.batch_size", 10)
	v.SetDefault("outbox.processor.max_retry", 10)
	v.SetDefault("outbox.processor.retry_interval", 30*time.Second)
	v.SetDefault("outbox.processor.window", time.Hour)

	// Friend request handler defaults
	v.SetDefault("friend_request_handler.batch_size", 100)
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
	if c.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database.password is required")
	}

	// Проверка Kafka
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers is required")
	}
	if c.Kafka.Topics.FriendRequestEvents == "" {
		return fmt.Errorf("kafka.topics.friend_request_events is required")
	}

	// Проверка портов
	if c.Server.GRPC.Port <= 0 || c.Server.GRPC.Port > 65535 {
		return fmt.Errorf("server.grpc.port must be between 1 and 65535")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("database.port must be between 1 and 65535")
	}

	// Проверка outbox настроек
	if c.Outbox.Processor.BatchSize <= 0 {
		return fmt.Errorf("outbox.processor.batch_size must be positive")
	}
	if c.Outbox.Processor.MaxRetry < 0 {
		return fmt.Errorf("outbox.processor.max_retry must be non-negative")
	}

	// Проверка friend request handler
	if c.FriendRequestHandler.BatchSize <= 0 {
		return fmt.Errorf("friend_request_handler.batch_size must be positive")
	}

	return nil
}
