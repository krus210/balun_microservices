package main

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/logger"
	"github.com/sskorolev/balun_microservices/lib/secrets"

	"auth/internal/app/adapters"
	"auth/internal/app/crypto"
	deliveryGrpc "auth/internal/app/delivery/grpc"
	"auth/internal/app/keystore"
	"auth/internal/app/repository"
	"auth/internal/app/token"
	"auth/internal/app/usecase"
	"auth/internal/config"
	errorsMiddleware "auth/internal/middleware/errors"

	authPb "auth/pkg/api"
	usersPb "auth/pkg/users/api"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		logger.FatalKV(ctx, "failed to load config", "error", err.Error())
	}

	// Создаем приложение
	application, err := app.NewApp(ctx, cfg.StandardServiceConfig)
	if err != nil {
		logger.FatalKV(ctx, "failed to create app", "error", err.Error())
	}

	// Инициализируем logger
	if err := application.InitLogger(cfg.Logger, cfg.Service.Name, cfg.Service.Environment); err != nil {
		logger.FatalKV(ctx, "failed to initialize logger", "error", err.Error())
	}

	// Инициализируем tracer
	if err := application.InitTracer(cfg.Tracer); err != nil {
		logger.FatalKV(ctx, "failed to initialize tracer", "error", err.Error())
	}

	// Инициализируем metrics
	if err := application.InitMetrics(cfg.Metrics, cfg.Service.Name); err != nil {
		logger.FatalKV(ctx, "failed to initialize metrics", "error", err.Error())
	}

	// Инициализируем admin HTTP сервер (метрики и pprof)
	if cfg.Server.Admin != nil {
		if err := application.InitAdminServer(*cfg.Server.Admin); err != nil {
			logger.FatalKV(ctx, "failed to initialize admin server", "error", err.Error())
		}
	}

	logger.InfoKV(ctx, "starting auth service",
		"version", cfg.Service.Version,
		"environment", cfg.Service.Environment,
		"grpc_port", cfg.Server.GRPC.Port,
	)

	// Инициализируем PostgreSQL через lib/app
	if err := application.InitPostgres(ctx, cfg.Database); err != nil {
		logger.FatalKV(ctx, "failed to init postgres", "error", err)
	}

	logger.InfoKV(ctx, "database connection established",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.Name,
	)

	// Подключаемся к Users сервису
	if err := application.InitGRPCClient(ctx, "users", cfg.UsersService); err != nil {
		logger.FatalKV(ctx, "failed to connect to users service", "error", err.Error())
	}

	// Создаем зависимости

	// 1. Repository (единый)
	repo := repository.NewRepository(application.TransactionManager())

	// 2. Crypto
	passwordHasher := crypto.NewBcryptHasher(cfg.Crypto.Password.BcryptCost, cfg.Crypto.Password.MinLength)

	// 3. Keystore
	var keyStore keystore.KeyStore
	// Определяем какой secrets provider использовать в зависимости от окружения
	secretsProvider := cfg.Secrets.Dev
	if cfg.Service.Environment == "prod" || cfg.Service.Environment == "production" {
		secretsProvider = cfg.Secrets.Prod
	}

	if cfg.Keys.Storage == "vault" && secretsProvider.Vault.Enabled {
		// Создаем Secrets Provider для Vault
		vaultProvider, err := secrets.NewSecretsProvider(ctx, secrets.WithVault(secrets.VaultConfig{
			Address:    secretsProvider.Vault.Address,
			Token:      secretsProvider.Vault.Token,
			MountPath:  "secret",
			SecretPath: cfg.Keys.Vault.Path,
		}))
		if err != nil {
			logger.FatalKV(ctx, "failed to create vault provider", "error", err)
		}
		keyStore = keystore.NewVaultKeyStore(vaultProvider, cfg.Keys.Vault.Path)
		logger.InfoKV(ctx, "using Vault for RSA keys", "path", cfg.Keys.Vault.Path)
	} else {
		keyStore = keystore.NewDBKeyStore(application.TransactionManager())
		logger.InfoKV(ctx, "using database for RSA keys")

		// Авто-создание ключа при старте, если нужно
		if cfg.Keys.DB.AutoCreateOnStart {
			if err := ensureActiveKey(ctx, keyStore); err != nil {
				logger.FatalKV(ctx, "failed to ensure active key", "error", err)
			}
		}
	}

	// 4. Token Manager
	tokenManager := token.NewTokenManager(
		token.Config{
			Issuer:          cfg.Auth.Issuer,
			Audience:        cfg.Auth.Audience,
			AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
			RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
		},
		keyStore,
	)

	// 5. Adapters
	usersClient := adapters.NewUsersClient(usersPb.NewUsersServiceClient(application.GetGRPCClient("users")))

	// 6. Usecase
	authUsecase := usecase.NewUsecase(
		usersClient,
		repo, // единый репозиторий реализует UsersRepository
		repo, // и RefreshTokensRepository одновременно
		passwordHasher,
		tokenManager,
		keyStore,
		usecase.Config{
			AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
			RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
		},
	)

	// 7. Controller
	controller := deliveryGrpc.NewAuthController(authUsecase)

	// Инициализируем gRPC сервер
	application.InitGRPCServer(cfg.Server, errorsMiddleware.ErrorsUnaryInterceptor())

	// Регистрируем gRPC сервисы
	application.RegisterGRPC(func(s *grpc.Server) {
		authPb.RegisterAuthServiceServer(s, controller)
	})

	// Запускаем gRPC сервер через errgroup
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.InfoKV(gCtx, "starting auth service gRPC server", "grpc_port", cfg.Server.GRPC.Port)
		if err := application.Run(gCtx, *cfg.Server.GRPC); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	// Запускаем admin HTTP сервер
	if cfg.Server.Admin != nil {
		g.Go(func() error {
			logger.InfoKV(gCtx, "starting admin HTTP server", "admin_port", cfg.Server.Admin.Port)
			if err := application.ServeAdmin(gCtx); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		})
	}

	waitErr := app.WaitForShutdown(ctx, g.Wait, app.GracefulShutdownTimeout)
	switch {
	case waitErr == nil || errors.Is(waitErr, context.Canceled):
		logger.InfoKV(ctx, "auth service components stopped")
	case errors.Is(waitErr, context.DeadlineExceeded):
		logger.WarnKV(ctx, "graceful shutdown timeout exceeded, forcing cleanup")
	default:
		logger.FatalKV(ctx, "failed to serve", "error", waitErr.Error())
	}

	logger.InfoKV(ctx, "auth service shutdown complete")
}

// ensureActiveKey проверяет наличие активного ключа и создает его, если отсутствует
func ensureActiveKey(ctx context.Context, keyStore keystore.KeyStore) error {
	_, err := keyStore.GetActiveKey(ctx)
	if err == nil {
		logger.InfoKV(ctx, "active RSA key already exists")
		return nil
	}

	if !errors.Is(err, keystore.ErrNoActiveKey) {
		return fmt.Errorf("failed to check active key: %w", err)
	}

	logger.InfoKV(ctx, "no active RSA key found, creating new one")

	newKey, err := keyStore.CreateKey(ctx, keystore.KeyStatusActive)
	if err != nil {
		return fmt.Errorf("failed to create active key: %w", err)
	}

	logger.InfoKV(ctx, "created new active RSA key", "kid", newKey.KID)
	return nil
}
