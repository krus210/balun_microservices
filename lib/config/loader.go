package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config определяет интерфейс для конфигурации сервиса
type Config interface {
	Validate() error
	GetService() ServiceConfig
	GetSecrets() *SecretsConfig
}

// LoadOptions определяет опции для загрузки конфигурации
type LoadOptions struct {
	ConfigName  string
	ConfigType  string
	ConfigPaths []string
	EnvPrefix   string

	// SetDefaults - функция для установки defaults (специфична для каждого сервиса)
	SetDefaults func(*viper.Viper)

	// Target - структура для десериализации
	Target interface{}
}

// Load загружает конфигурацию из файла и переменных окружения
func Load(opts LoadOptions) error {
	v := viper.New()

	// Установка defaults
	if opts.SetDefaults != nil {
		opts.SetDefaults(v)
	}

	// Настройка файла конфигурации
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)
	for _, path := range opts.ConfigPaths {
		v.AddConfigPath(path)
	}

	// Чтение файла конфигурации (если существует)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Файл не найден - это нормально, будем использовать defaults и env
	}

	// Настройка переменных окружения
	v.SetEnvPrefix(opts.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Явный биндинг для вложенных Vault переменных окружения
	// (AutomaticEnv не всегда корректно мапит глубоко вложенные структуры)
	bindVaultEnvVars(v)

	// Десериализация в структуру
	if err := v.Unmarshal(opts.Target); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// bindVaultEnvVars явно биндит переменные окружения для Vault
func bindVaultEnvVars(v *viper.Viper) {
	_ = v.BindEnv("secrets.prod.vault.token")
	_ = v.BindEnv("secrets.prod.vault.role_id")
	_ = v.BindEnv("secrets.prod.vault.secret_id")
	_ = v.BindEnv("secrets.dev.vault.token")
	_ = v.BindEnv("secrets.dev.vault.role_id")
	_ = v.BindEnv("secrets.dev.vault.secret_id")
}
