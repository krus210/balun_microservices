package config

// StandardServiceConfig - стандартная конфигурация для микросервиса
type StandardServiceConfig struct {
	Service  ServiceConfig   `mapstructure:"service"`
	Server   ServerConfig    `mapstructure:"server"`
	Database *DatabaseConfig `mapstructure:"database,omitempty"`
	Secrets  SecretsConfig   `mapstructure:"secrets"`
	Logger   LoggerConfig    `mapstructure:"logger"`
	Tracer   TracerConfig    `mapstructure:"tracer"`
	Metrics  MetricsConfig   `mapstructure:"metrics"`

	// Опциональные поля для сервисов с дополнительными компонентами
	Kafka                *KafkaConfig                `mapstructure:"kafka,omitempty"`
	KafkaConsumer        *KafkaConsumerConfig        `mapstructure:"kafka_consumer,omitempty"`
	Outbox               *OutboxConfig               `mapstructure:"outbox,omitempty"`
	FriendRequestHandler *FriendRequestHandlerConfig `mapstructure:"friend_request_handler,omitempty"`

	// Подключения к другим сервисам
	AuthService   *TargetServiceConfig `mapstructure:"auth_service,omitempty"`
	UsersService  *TargetServiceConfig `mapstructure:"users_service,omitempty"`
	SocialService *TargetServiceConfig `mapstructure:"social_service,omitempty"`
	ChatService   *TargetServiceConfig `mapstructure:"chat_service,omitempty"`
}

// GetService реализует интерфейс Config
func (c *StandardServiceConfig) GetService() ServiceConfig {
	return c.Service
}

// GetSecrets реализует интерфейс Config
func (c *StandardServiceConfig) GetSecrets() *SecretsConfig {
	return &c.Secrets
}

// Validate проверяет корректность конфигурации
func (c *StandardServiceConfig) Validate() error {
	// Валидируем обязательные поля
	if err := ValidateServiceConfig(c.Service); err != nil {
		return err
	}
	if err := ValidateServerConfig(c.Server); err != nil {
		return err
	}
	if err := ValidateLoggerConfig(c.Logger); err != nil {
		return err
	}
	if err := ValidateTracerConfig(c.Tracer); err != nil {
		return err
	}
	if err := ValidateMetricsConfig(c.Metrics); err != nil {
		return err
	}

	// Валидируем опциональные поля только если они заполнены
	if c.Database != nil {
		if err := ValidateDatabaseConfig(*c.Database); err != nil {
			return err
		}
	}

	if c.Kafka != nil {
		if err := ValidateKafkaConfig(*c.Kafka); err != nil {
			return err
		}
	}

	if c.KafkaConsumer != nil {
		if err := ValidateKafkaConsumerConfig(*c.KafkaConsumer); err != nil {
			return err
		}
	}

	if c.Outbox != nil {
		if err := ValidateOutboxConfig(*c.Outbox); err != nil {
			return err
		}
	}

	if c.FriendRequestHandler != nil {
		if err := ValidateFriendRequestHandlerConfig(*c.FriendRequestHandler); err != nil {
			return err
		}
	}

	if c.AuthService != nil {
		if err := ValidateTargetServiceConfig(*c.AuthService, "auth_service"); err != nil {
			return err
		}
	}

	if c.UsersService != nil {
		if err := ValidateTargetServiceConfig(*c.UsersService, "users_service"); err != nil {
			return err
		}
	}

	if c.SocialService != nil {
		if err := ValidateTargetServiceConfig(*c.SocialService, "social_service"); err != nil {
			return err
		}
	}

	if c.ChatService != nil {
		if err := ValidateTargetServiceConfig(*c.ChatService, "chat_service"); err != nil {
			return err
		}
	}

	return nil
}
