package config

// StandardServiceConfig - стандартная конфигурация для микросервиса с БД
type StandardServiceConfig struct {
	Service  ServiceConfig  `mapstructure:"service"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Secrets  SecretsConfig  `mapstructure:"secrets"`

	// Опциональные поля для сервисов с дополнительными компонентами
	Kafka                *KafkaConfig                `mapstructure:"kafka,omitempty"`
	Outbox               *OutboxConfig               `mapstructure:"outbox,omitempty"`
	FriendRequestHandler *FriendRequestHandlerConfig `mapstructure:"friend_request_handler,omitempty"`
	UsersService         *TargetServiceConfig        `mapstructure:"users_service,omitempty"`
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
	if err := ValidateDatabaseConfig(c.Database); err != nil {
		return err
	}

	// Валидируем опциональные поля только если они заполнены
	if c.Kafka != nil {
		if err := ValidateKafkaConfig(*c.Kafka); err != nil {
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

	if c.UsersService != nil {
		if err := ValidateTargetServiceConfig(*c.UsersService, "users_service"); err != nil {
			return err
		}
	}

	return nil
}
