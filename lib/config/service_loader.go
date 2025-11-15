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
	databaseEnabled   bool
	grpcPort          int
	secretsPathSuffix string
	customDefaults    func(*viper.Viper)

	// Опциональные компоненты
	kafka                *KafkaConfig
	kafkaConsumer        *KafkaConsumerConfig
	outbox               *OutboxConfig
	friendRequestHandler *FriendRequestHandlerConfig

	// Подключения к другим сервисам
	authService   *TargetServiceConfig
	usersService  *TargetServiceConfig
	socialService *TargetServiceConfig
	chatService   *TargetServiceConfig
}

// WithDatabaseName устанавливает имя базы данных
func WithDatabaseName(name string) ServiceOption {
	return func(opts *serviceOptions) {
		opts.databaseName = name
		opts.databaseEnabled = true
	}
}

// WithoutDatabase отключает блок database и загрузку связанных секретов
func WithoutDatabase() ServiceOption {
	return func(opts *serviceOptions) {
		opts.databaseEnabled = false
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

// WithKafka включает Kafka конфигурацию
func WithKafka(brokers, clientID, friendRequestTopic string) ServiceOption {
	return func(opts *serviceOptions) {
		opts.kafka = &KafkaConfig{
			Brokers:  brokers,
			ClientID: clientID,
			Topics: KafkaTopics{
				FriendRequestEvents: friendRequestTopic,
			},
		}
	}
}

// WithKafkaConsumer включает Kafka consumer конфигурацию
func WithKafkaConsumer(brokers, consumerGroupID, consumerName, friendRequestTopic string) ServiceOption {
	return func(opts *serviceOptions) {
		opts.kafkaConsumer = &KafkaConsumerConfig{
			Brokers:         brokers,
			ConsumerGroupID: consumerGroupID,
			ConsumerName:    consumerName,
			Topics: KafkaTopics{
				FriendRequestEvents: friendRequestTopic,
			},
		}
	}
}

// WithOutbox включает Outbox процессор конфигурацию
func WithOutbox(batchSize, maxRetry int, retryInterval, window time.Duration) ServiceOption {
	return func(opts *serviceOptions) {
		opts.outbox = &OutboxConfig{
			Processor: OutboxProcessorConfig{
				BatchSize:     batchSize,
				MaxRetry:      maxRetry,
				RetryInterval: retryInterval,
				Window:        window,
			},
		}
	}
}

// WithFriendRequestHandler включает конфигурацию обработчика заявок в друзья
func WithFriendRequestHandler(batchSize int) ServiceOption {
	return func(opts *serviceOptions) {
		opts.friendRequestHandler = &FriendRequestHandlerConfig{
			BatchSize: batchSize,
		}
	}
}

// WithAuthService включает конфигурацию подключения к Auth сервису
func WithAuthService(host string, port int) ServiceOption {
	return func(opts *serviceOptions) {
		opts.authService = &TargetServiceConfig{
			Host: host,
			Port: port,
		}
	}
}

// WithUsersService включает конфигурацию подключения к Users сервису
func WithUsersService(host string, port int) ServiceOption {
	return func(opts *serviceOptions) {
		opts.usersService = &TargetServiceConfig{
			Host: host,
			Port: port,
		}
	}
}

// WithSocialService включает конфигурацию подключения к Social сервису
func WithSocialService(host string, port int) ServiceOption {
	return func(opts *serviceOptions) {
		opts.socialService = &TargetServiceConfig{
			Host: host,
			Port: port,
		}
	}
}

// WithChatService включает конфигурацию подключения к Chat сервису
func WithChatService(host string, port int) ServiceOption {
	return func(opts *serviceOptions) {
		opts.chatService = &TargetServiceConfig{
			Host: host,
			Port: port,
		}
	}
}

// setGRPCClientDefaults устанавливает defaults для gRPC клиента target сервиса
func setGRPCClientDefaults(v *viper.Viper, servicePrefix string) {
	// Timeout
	v.SetDefault(servicePrefix+".grpc_client.timeout", 5*time.Second)

	// Retry configuration
	v.SetDefault(servicePrefix+".grpc_client.retry.max_attempts", 3)
	v.SetDefault(servicePrefix+".grpc_client.retry.backoff.base", 100*time.Millisecond)
	v.SetDefault(servicePrefix+".grpc_client.retry.backoff.max", 2*time.Second)
	v.SetDefault(servicePrefix+".grpc_client.retry.backoff.jitter", true)
	v.SetDefault(servicePrefix+".grpc_client.retry.retryable_codes", []string{
		"UNAVAILABLE",
		"DEADLINE_EXCEEDED",
		"RESOURCE_EXHAUSTED",
		"ABORTED",
	})

	// Circuit Breaker configuration
	v.SetDefault(servicePrefix+".grpc_client.circuit_breaker.failures_for_open", 5)
	v.SetDefault(servicePrefix+".grpc_client.circuit_breaker.window", 30*time.Second)
	v.SetDefault(servicePrefix+".grpc_client.circuit_breaker.half_open_max_calls", 5)
	v.SetDefault(servicePrefix+".grpc_client.circuit_breaker.open_state_for", 60*time.Second)
}

// LoadServiceConfig загружает стандартную конфигурацию сервиса с секретами
func LoadServiceConfig(ctx context.Context, serviceName string, opts ...ServiceOption) (*StandardServiceConfig, error) {
	// Применяем опции
	options := &serviceOptions{
		serviceName:     serviceName,
		databaseName:    serviceName, // по умолчанию совпадает с именем сервиса
		databaseEnabled: true,
		grpcPort:        8082, // стандартный порт для всех сервисов
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

		// Logger defaults
		v.SetDefault("logger.level", "info")

		// Tracer defaults
		v.SetDefault("tracer.enabled", true)
		v.SetDefault("tracer.service_name", options.serviceName)
		v.SetDefault("tracer.jaeger_agent_host", "jaeger-agent")
		v.SetDefault("tracer.jaeger_agent_port", 4317)
		v.SetDefault("tracer.sampler_type", "const")
		v.SetDefault("tracer.sampler_param", 1)

		// Database defaults
		if options.databaseEnabled {
			v.SetDefault("database.host", "localhost")
			v.SetDefault("database.port", 5432)
			v.SetDefault("database.name", options.databaseName)
			v.SetDefault("database.sslmode", "disable")
			v.SetDefault("database.max_conn_idle_time", time.Minute)
		}

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

		// Опциональные компоненты - defaults только если указаны через опции
		if options.kafka != nil {
			v.SetDefault("kafka.brokers", options.kafka.Brokers)
			v.SetDefault("kafka.client_id", options.kafka.ClientID)
			v.SetDefault("kafka.topics.friend_request_events", options.kafka.Topics.FriendRequestEvents)
		}

		if options.kafkaConsumer != nil {
			v.SetDefault("kafka_consumer.brokers", options.kafkaConsumer.Brokers)
			v.SetDefault("kafka_consumer.consumer_group_id", options.kafkaConsumer.ConsumerGroupID)
			v.SetDefault("kafka_consumer.consumer_name", options.kafkaConsumer.ConsumerName)
			v.SetDefault("kafka_consumer.topics.friend_request_events", options.kafkaConsumer.Topics.FriendRequestEvents)
		}

		if options.outbox != nil {
			v.SetDefault("outbox.processor.batch_size", options.outbox.Processor.BatchSize)
			v.SetDefault("outbox.processor.max_retry", options.outbox.Processor.MaxRetry)
			v.SetDefault("outbox.processor.retry_interval", options.outbox.Processor.RetryInterval)
			v.SetDefault("outbox.processor.window", options.outbox.Processor.Window)
		}

		if options.friendRequestHandler != nil {
			v.SetDefault("friend_request_handler.batch_size", options.friendRequestHandler.BatchSize)
		}

		if options.authService != nil {
			v.SetDefault("auth_service.host", options.authService.Host)
			v.SetDefault("auth_service.port", options.authService.Port)
			setGRPCClientDefaults(v, "auth_service")
		}

		if options.usersService != nil {
			v.SetDefault("users_service.host", options.usersService.Host)
			v.SetDefault("users_service.port", options.usersService.Port)
			setGRPCClientDefaults(v, "users_service")
		}

		if options.socialService != nil {
			v.SetDefault("social_service.host", options.socialService.Host)
			v.SetDefault("social_service.port", options.socialService.Port)
			setGRPCClientDefaults(v, "social_service")
		}

		if options.chatService != nil {
			v.SetDefault("chat_service.host", options.chatService.Host)
			v.SetDefault("chat_service.port", options.chatService.Port)
			setGRPCClientDefaults(v, "chat_service")
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

	// Если база данных отключена опциями, полностью игнорируем конфигурацию
	if !options.databaseEnabled {
		cfg.Database = nil
	}

	// Загружаем секреты БД только если сервис использует базу данных
	if options.databaseEnabled && cfg.Database != nil {
		if err := LoadConfigWithSecrets(ctx, cfg, &databaseSecretsLoader{db: cfg.Database}); err != nil {
			return nil, err
		}
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
