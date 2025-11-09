package app

import (
	"github.com/spf13/viper"
	"github.com/sskorolev/balun_microservices/lib/config"
)

// Option определяет функциональную опцию для настройки App
type Option func(*appOptions) error

// appOptions содержит опции для создания App
type appOptions struct {
	configFile   string
	envPrefix    string
	setDefaults  func(*viper.Viper)
	configTarget interface{}
}

// WithConfigFile устанавливает путь к конфигурационному файлу
func WithConfigFile(path string) Option {
	return func(opts *appOptions) error {
		opts.configFile = path
		return nil
	}
}

// WithEnvPrefix устанавливает префикс для переменных окружения
func WithEnvPrefix(prefix string) Option {
	return func(opts *appOptions) error {
		opts.envPrefix = prefix
		return nil
	}
}

// WithSetDefaults устанавливает функцию для установки defaults в viper
func WithSetDefaults(setDefaults func(*viper.Viper)) Option {
	return func(opts *appOptions) error {
		opts.setDefaults = setDefaults
		return nil
	}
}

// WithConfigTarget устанавливает структуру для десериализации конфигурации
func WithConfigTarget(target interface{}) Option {
	return func(opts *appOptions) error {
		opts.configTarget = target
		return nil
	}
}

// defaultOptions возвращает опции по умолчанию
func defaultOptions() *appOptions {
	return &appOptions{
		configFile: "config",
		envPrefix:  "APP",
	}
}

// applyOptions применяет опции к appOptions
func applyOptions(opts []Option) (*appOptions, error) {
	options := defaultOptions()
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}
	return options, nil
}

// loadOptionsToConfigOptions конвертирует appOptions в config.LoadOptions
func loadOptionsToConfigOptions(opts *appOptions) config.LoadOptions {
	return config.LoadOptions{
		ConfigName:  opts.configFile,
		ConfigType:  "yaml",
		ConfigPaths: []string{".", "./config"},
		EnvPrefix:   opts.envPrefix,
		SetDefaults: opts.setDefaults,
		Target:      opts.configTarget,
	}
}
