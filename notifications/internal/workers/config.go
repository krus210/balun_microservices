package workers

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config содержит настройки воркеров
type Config struct {
	SaveEvents SaveEventsConfig `mapstructure:"save_events"`
	Delete     DeleteConfig     `mapstructure:"delete"`
}

// SaveEventsConfig содержит настройки воркера сохранения событий
type SaveEventsConfig struct {
	Interval  time.Duration `mapstructure:"interval"`
	BatchSize int           `mapstructure:"batch_size"`
}

// DeleteConfig содержит настройки воркера удаления
type DeleteConfig struct {
	Interval      time.Duration `mapstructure:"interval"`
	RetentionDays int           `mapstructure:"retention_days"`
}

// Load загружает конфигурацию workers из config.yaml
func Load() (*Config, error) {
	v := viper.New()

	// Настройка файла конфигурации
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Читаем файл конфигурации
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Настройка переменных окружения
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	var cfg Config

	// Извлекаем секцию workers из viper
	if err := v.UnmarshalKey("workers", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workers config: %w", err)
	}

	// Валидация
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("workers config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate проверяет корректность конфигурации workers
func (c *Config) Validate() error {
	if c.SaveEvents.BatchSize <= 0 {
		return fmt.Errorf("workers.save_events.batch_size must be positive")
	}
	if c.Delete.RetentionDays <= 0 {
		return fmt.Errorf("workers.delete.retention_days must be positive")
	}
	return nil
}
