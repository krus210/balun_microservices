package config

import (
	"fmt"
	"time"
)

// ServiceConfig содержит общую информацию о сервисе
type ServiceConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig содержит настройки серверов
type ServerConfig struct {
	HTTP *HTTPConfig `mapstructure:"http,omitempty"`
	GRPC *GRPCConfig `mapstructure:"grpc,omitempty"`
}

// HTTPConfig содержит настройки HTTP сервера
type HTTPConfig struct {
	Port int `mapstructure:"port"`
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

// TargetServiceConfig содержит настройки подключения к зависимому сервису
type TargetServiceConfig struct {
	Host       string            `mapstructure:"host"`
	Port       int               `mapstructure:"port"`
	GRPCClient *GRPCClientConfig `mapstructure:"grpc_client,omitempty"`
}

// Address возвращает полный адрес сервиса
func (t TargetServiceConfig) Address() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

// GRPCClientConfig содержит настройки gRPC клиента
type GRPCClientConfig struct {
	Timeout        time.Duration         `mapstructure:"timeout"`
	Retry          *RetryConfig          `mapstructure:"retry,omitempty"`
	CircuitBreaker *CircuitBreakerConfig `mapstructure:"circuit_breaker,omitempty"`
}

// RetryConfig содержит настройки retry логики
type RetryConfig struct {
	MaxAttempts    int                `mapstructure:"max_attempts"`
	Backoff        RetryBackoffConfig `mapstructure:"backoff"`
	RetryableCodes []string           `mapstructure:"retryable_codes"`
}

// RetryBackoffConfig содержит настройки exponential backoff
type RetryBackoffConfig struct {
	Base   time.Duration `mapstructure:"base"`
	Max    time.Duration `mapstructure:"max"`
	Jitter bool          `mapstructure:"jitter"`
}

// CircuitBreakerConfig содержит настройки circuit breaker
type CircuitBreakerConfig struct {
	FailuresForOpen  int           `mapstructure:"failures_for_open"`
	Window           time.Duration `mapstructure:"window"`
	HalfOpenMaxCalls int           `mapstructure:"half_open_max_calls"`
	OpenStateFor     time.Duration `mapstructure:"open_state_for"`
}

// KafkaConfig содержит настройки подключения к Apache Kafka
type KafkaConfig struct {
	Brokers  string      `mapstructure:"brokers"`
	ClientID string      `mapstructure:"client_id"`
	Topics   KafkaTopics `mapstructure:"topics"`
}

// KafkaTopics содержит названия топиков Kafka
type KafkaTopics struct {
	FriendRequestEvents string `mapstructure:"friend_request_events"`
}

// GetBrokers возвращает строку с адресами брокеров Kafka
func (c KafkaConfig) GetBrokers() string {
	return c.Brokers
}

// OutboxConfig содержит настройки Transactional Outbox процессора
type OutboxConfig struct {
	Processor OutboxProcessorConfig `mapstructure:"processor"`
}

// OutboxProcessorConfig содержит параметры работы outbox процессора
type OutboxProcessorConfig struct {
	BatchSize     int           `mapstructure:"batch_size"`
	MaxRetry      int           `mapstructure:"max_retry"`
	RetryInterval time.Duration `mapstructure:"retry_interval"`
	Window        time.Duration `mapstructure:"window"`
}

// FriendRequestHandlerConfig содержит настройки обработчика заявок в друзья
type FriendRequestHandlerConfig struct {
	BatchSize int `mapstructure:"batch_size"`
}

// KafkaConsumerConfig содержит настройки Kafka consumer
type KafkaConsumerConfig struct {
	Brokers         string      `mapstructure:"brokers"`
	ConsumerGroupID string      `mapstructure:"consumer_group_id"`
	ConsumerName    string      `mapstructure:"consumer_name"`
	Topics          KafkaTopics `mapstructure:"topics"`
}

// GetBrokers возвращает строку с адресами брокеров Kafka
func (c KafkaConsumerConfig) GetBrokers() string {
	return c.Brokers
}
