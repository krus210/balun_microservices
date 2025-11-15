package config

import (
	"fmt"
)

// ValidatePort проверяет корректность порта
func ValidatePort(port int, fieldName string) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", fieldName)
	}
	return nil
}

// ValidateRequired проверяет обязательное строковое поле
func ValidateRequired(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidatePositive проверяет положительное целое число
func ValidatePositive(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// ValidateNonNegative проверяет неотрицательное целое число
func ValidateNonNegative(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative", fieldName)
	}
	return nil
}

// ValidateServiceConfig валидирует ServiceConfig
func ValidateServiceConfig(cfg ServiceConfig) error {
	if err := ValidateRequired(cfg.Name, "service.name"); err != nil {
		return err
	}
	return nil
}

// ValidateServerConfig валидирует ServerConfig
func ValidateServerConfig(cfg ServerConfig) error {
	if cfg.HTTP != nil {
		if err := ValidatePort(cfg.HTTP.Port, "server.http.port"); err != nil {
			return err
		}
	}

	if cfg.GRPC != nil {
		if err := ValidatePort(cfg.GRPC.Port, "server.grpc.port"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDatabaseConfig валидирует DatabaseConfig
func ValidateDatabaseConfig(cfg DatabaseConfig) error {
	if err := ValidateRequired(cfg.Host, "database.host"); err != nil {
		return err
	}
	if err := ValidateRequired(cfg.Name, "database.name"); err != nil {
		return err
	}
	if err := ValidatePort(cfg.Port, "database.port"); err != nil {
		return err
	}
	return nil
}

// ValidateTargetServiceConfig валидирует TargetServiceConfig
func ValidateTargetServiceConfig(cfg TargetServiceConfig, prefix string) error {
	if err := ValidateRequired(cfg.Host, prefix+".host"); err != nil {
		return err
	}
	if err := ValidatePort(cfg.Port, prefix+".port"); err != nil {
		return err
	}
	return nil
}

// ValidateKafkaConfig валидирует KafkaConfig
func ValidateKafkaConfig(cfg KafkaConfig) error {
	if err := ValidateRequired(cfg.GetBrokers(), "kafka.brokers"); err != nil {
		return err
	}
	if err := ValidateRequired(cfg.Topics.FriendRequestEvents, "kafka.topics.friend_request_events"); err != nil {
		return err
	}
	return nil
}

// ValidateOutboxConfig валидирует OutboxConfig
func ValidateOutboxConfig(cfg OutboxConfig) error {
	if err := ValidatePositive(cfg.Processor.BatchSize, "outbox.processor.batch_size"); err != nil {
		return err
	}
	if err := ValidateNonNegative(cfg.Processor.MaxRetry, "outbox.processor.max_retry"); err != nil {
		return err
	}
	return nil
}

// ValidateFriendRequestHandlerConfig валидирует FriendRequestHandlerConfig
func ValidateFriendRequestHandlerConfig(cfg FriendRequestHandlerConfig) error {
	if err := ValidatePositive(cfg.BatchSize, "friend_request_handler.batch_size"); err != nil {
		return err
	}
	return nil
}

// ValidateKafkaConsumerConfig валидирует KafkaConsumerConfig
func ValidateKafkaConsumerConfig(cfg KafkaConsumerConfig) error {
	if err := ValidateRequired(cfg.GetBrokers(), "kafka_consumer.brokers"); err != nil {
		return err
	}
	if err := ValidateRequired(cfg.ConsumerGroupID, "kafka_consumer.consumer_group_id"); err != nil {
		return err
	}
	if err := ValidateRequired(cfg.ConsumerName, "kafka_consumer.consumer_name"); err != nil {
		return err
	}
	if err := ValidateRequired(cfg.Topics.FriendRequestEvents, "kafka_consumer.topics.friend_request_events"); err != nil {
		return err
	}
	return nil
}

// ValidateLoggerConfig валидирует LoggerConfig
func ValidateLoggerConfig(cfg LoggerConfig) error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}
	if !validLevels[cfg.Level] {
		return fmt.Errorf("logger.level must be one of: debug, info, warn, error, fatal, panic")
	}
	return nil
}

// ValidateTracerConfig валидирует TracerConfig
func ValidateTracerConfig(cfg TracerConfig) error {
	// Валидация только если трейсинг включен
	if !cfg.Enabled {
		return nil
	}

	if err := ValidateRequired(cfg.ServiceName, "tracer.service_name"); err != nil {
		return err
	}

	// Проверяем наличие хотя бы одной конфигурации агента
	if cfg.JaegerHost == "" && cfg.JaegerAgentHost == "" {
		return fmt.Errorf("tracer.jaeger_host or tracer.jaeger_agent_host must be specified when tracer is enabled")
	}

	if cfg.JaegerAgentHost != "" && cfg.JaegerAgentPort <= 0 {
		return fmt.Errorf("tracer.jaeger_agent_port must be positive when jaeger_agent_host is specified")
	}

	return nil
}
