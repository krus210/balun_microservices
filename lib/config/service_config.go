package config

// StandardServiceConfig - стандартная конфигурация для микросервиса с БД
type StandardServiceConfig struct {
	Service  ServiceConfig  `mapstructure:"service"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Secrets  SecretsConfig  `mapstructure:"secrets"`
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
	if err := ValidateServiceConfig(c.Service); err != nil {
		return err
	}
	if err := ValidateServerConfig(c.Server); err != nil {
		return err
	}
	if err := ValidateDatabaseConfig(c.Database); err != nil {
		return err
	}
	return nil
}
