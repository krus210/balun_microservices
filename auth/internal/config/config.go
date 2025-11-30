package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	libconfig "github.com/sskorolev/balun_microservices/lib/config"
)

type Config struct {
	*libconfig.StandardServiceConfig
	Auth   AuthConfig
	Keys   KeysConfig
	Crypto CryptoConfig
}

type AuthConfig struct {
	Issuer          string
	Audience        []string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type KeysConfig struct {
	Storage string
	Vault   VaultKeysConfig
	DB      DBKeysConfig
}

type VaultKeysConfig struct {
	Path string
}

type DBKeysConfig struct {
	AutoCreateOnStart bool
}

type CryptoConfig struct {
	Password PasswordConfig
}

type PasswordConfig struct {
	BcryptCost int
	MinLength  int
}

func LoadConfig(ctx context.Context) (*Config, error) {
	// Инициализируем Viper для чтения config.yaml
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./auth")
	viper.AddConfigPath("/app")

	// Пытаемся прочитать config.yaml (игнорируем ошибку если файл не найден)
	_ = viper.ReadInConfig()

	// Настраиваем переменные окружения с префиксом APP_
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Загружаем базовую конфигурацию через lib/config
	// Для Vault в production окружении нужно добавить secrets конфигурацию
	serviceCfg, err := libconfig.LoadServiceConfig(ctx, "auth",
		libconfig.WithUsersService("users", 8082),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load service config: %w", err)
	}

	cfg := &Config{
		StandardServiceConfig: serviceCfg,
	}

	// Загружаем auth-специфичные настройки
	if err := loadAuthConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to load auth config: %w", err)
	}

	return cfg, nil
}

func loadAuthConfig(cfg *Config) error {
	// Auth
	cfg.Auth = AuthConfig{
		Issuer:          viper.GetString("auth.issuer"),
		Audience:        viper.GetStringSlice("auth.audience"),
		AccessTokenTTL:  viper.GetDuration("auth.access_token_ttl"),
		RefreshTokenTTL: viper.GetDuration("auth.refresh_token_ttl"),
	}

	// Валидация auth
	if cfg.Auth.Issuer == "" {
		return fmt.Errorf("auth.issuer is required")
	}
	if len(cfg.Auth.Audience) == 0 {
		return fmt.Errorf("auth.audience is required")
	}
	if cfg.Auth.AccessTokenTTL == 0 {
		cfg.Auth.AccessTokenTTL = 15 * time.Minute // default
	}
	if cfg.Auth.RefreshTokenTTL == 0 {
		cfg.Auth.RefreshTokenTTL = 720 * time.Hour // default 30 days
	}

	// Keys
	cfg.Keys = KeysConfig{
		Storage: viper.GetString("keys.storage"),
		Vault: VaultKeysConfig{
			Path: viper.GetString("keys.vault.path"),
		},
		DB: DBKeysConfig{
			AutoCreateOnStart: viper.GetBool("keys.db.auto_create_on_start"),
		},
	}

	// Валидация keys
	if cfg.Keys.Storage != "vault" && cfg.Keys.Storage != "db" {
		cfg.Keys.Storage = "vault" // default
	}

	if cfg.Keys.Storage == "vault" && cfg.Keys.Vault.Path == "" {
		return fmt.Errorf("keys.vault.path is required when storage is vault")
	}

	// Crypto
	cfg.Crypto = CryptoConfig{
		Password: PasswordConfig{
			BcryptCost: viper.GetInt("crypto.password.bcrypt_cost"),
			MinLength:  viper.GetInt("crypto.password.min_length"),
		},
	}

	// Валидация crypto
	if cfg.Crypto.Password.BcryptCost == 0 {
		cfg.Crypto.Password.BcryptCost = 12 // default
	}
	if cfg.Crypto.Password.MinLength == 0 {
		cfg.Crypto.Password.MinLength = 6 // default
	}

	return nil
}
