package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/postgres"
)

// App представляет основное приложение с инициализированной инфраструктурой
type App struct {
	config       config.Config
	pgConnection *postgres.Connection
	pgTxManager  postgres.TransactionManagerAPI
	grpcServer   *grpc.Server
	httpHandler  http.Handler
	adminServer  *http.Server
	grpcClients  map[string]*grpc.ClientConn
	cleanupFuncs []func()

	grpcRegistrar GRPCRegistrar
}

// ServerConfig определяет конфигурацию серверов для запуска
type ServerConfig struct {
	GRPC  *config.GRPCConfig
	HTTP  *config.HTTPConfig
	Admin *config.AdminConfig
}

// NewApp создает новое приложение с заданными опциями
func NewApp(ctx context.Context, cfg config.Config, opts ...Option) (*App, error) {
	app := &App{
		config:       cfg,
		grpcClients:  make(map[string]*grpc.ClientConn),
		cleanupFuncs: make([]func(), 0),
	}

	log.Printf("Starting %s service (version: %s, environment: %s)",
		cfg.GetService().Name,
		cfg.GetService().Version,
		cfg.GetService().Environment,
	)

	return app, nil
}

// InitPostgres инициализирует PostgreSQL connection pool
func (a *App) InitPostgres(ctx context.Context, dbCfg *config.DatabaseConfig) error {
	conn, cleanup, err := InitPostgres(ctx, dbCfg)
	if err != nil {
		return err
	}

	a.pgConnection = conn
	a.pgTxManager = InitTransactionManager(conn)
	a.cleanupFuncs = append(a.cleanupFuncs, cleanup)

	log.Printf("PostgreSQL connection established: %s:%d/%s", dbCfg.Host, dbCfg.Port, dbCfg.Name)

	return nil
}

// Postgres возвращает PostgreSQL connection
func (a *App) Postgres() *postgres.Connection {
	return a.pgConnection
}

// TransactionManager возвращает transaction manager
func (a *App) TransactionManager() postgres.TransactionManagerAPI {
	return a.pgTxManager
}

// InitGRPCServer инициализирует gRPC сервер
func (a *App) InitGRPCServer(cfg config.ServerConfig, customInterceptors ...grpc.UnaryServerInterceptor) {
	a.grpcServer = InitGRPCServer(cfg, customInterceptors...)
	log.Println("gRPC server initialized")
}

// RegisterGRPC регистрирует gRPC сервисы
func (a *App) RegisterGRPC(registrar GRPCRegistrar) {
	if a.grpcServer == nil {
		log.Fatal("gRPC server not initialized. Call InitGRPCServer() first")
	}
	registrar(a.grpcServer)
}

// GRPCServer возвращает gRPC сервер
func (a *App) GRPCServer() *grpc.Server {
	return a.grpcServer
}

// Run запускает только gRPC сервер
func (a *App) Run(ctx context.Context, grpcCfg config.GRPCConfig) error {
	return a.RunServers(ctx, ServerConfig{GRPC: &grpcCfg})
}

// Shutdown выполняет graceful shutdown и cleanup
func (a *App) Shutdown() {
	log.Println("Shutting down application...")

	// Выполняем cleanup функции в обратном порядке
	for i := len(a.cleanupFuncs) - 1; i >= 0; i-- {
		a.cleanupFuncs[i]()
	}

	log.Println("Application stopped")
}

// InitGRPCClient инициализирует gRPC клиент для подключения к другому сервису
func (a *App) InitGRPCClient(ctx context.Context, name string, targetCfg *config.TargetServiceConfig) error {
	conn, cleanup, err := InitGRPCClient(ctx, targetCfg)
	if err != nil {
		return fmt.Errorf("failed to init gRPC client '%s': %w", name, err)
	}

	a.grpcClients[name] = conn
	a.cleanupFuncs = append(a.cleanupFuncs, cleanup)

	log.Printf("gRPC client '%s' connected to %s", name, targetCfg.Address())
	return nil
}

// GetGRPCClient возвращает gRPC клиент по имени
func (a *App) GetGRPCClient(name string) *grpc.ClientConn {
	return a.grpcClients[name]
}

// InitHTTPServer инициализирует HTTP handler
func (a *App) InitHTTPServer(handler http.Handler) {
	a.httpHandler = handler
	log.Println("HTTP handler initialized")
}

// RunBoth запускает HTTP и gRPC серверы параллельно (для обратной совместимости)
func (a *App) RunBoth(ctx context.Context, grpcCfg config.GRPCConfig, httpCfg config.HTTPConfig) error {
	return a.RunServers(ctx, ServerConfig{
		GRPC: &grpcCfg,
		HTTP: &httpCfg,
	})
}

// RunServers запускает любую комбинацию серверов с graceful shutdown
func (a *App) RunServers(ctx context.Context, cfg ServerConfig) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 3) // до 3 серверов одновременно

	// Запуск gRPC сервера (если указан)
	if cfg.GRPC != nil {
		if a.grpcServer == nil {
			return fmt.Errorf("gRPC server not initialized")
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ServeGRPC(ctx, a.grpcServer, *cfg.GRPC); err != nil {
				if !errors.Is(err, context.Canceled) {
					errChan <- fmt.Errorf("gRPC server error: %w", err)
				}
			}
		}()
	}

	// Запуск HTTP сервера (если указан)
	if cfg.HTTP != nil {
		if a.httpHandler == nil {
			return fmt.Errorf("HTTP handler not initialized")
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ServeHTTP(ctx, a.httpHandler, *cfg.HTTP); err != nil {
				if !errors.Is(err, context.Canceled) {
					errChan <- fmt.Errorf("HTTP server error: %w", err)
				}
			}
		}()
	}

	// Запуск Admin сервера (если указан)
	if cfg.Admin != nil {
		if a.adminServer == nil {
			return fmt.Errorf("admin server not initialized")
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.ServeAdmin(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					errChan <- fmt.Errorf("admin server error: %w", err)
				}
			}
		}()
	}

	// Ожидание завершения всех серверов
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Собираем первую ошибку
	var firstErr error
	for err := range errChan {
		if err != nil && firstErr == nil {
			firstErr = err
			log.Printf("Server error: %v", err)
		}
	}

	// После завершения серверов выполняем cleanup
	a.Shutdown()

	return firstErr
}
