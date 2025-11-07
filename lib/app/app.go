package app

import (
	"context"
	"fmt"
	"log"

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
	cleanupFuncs []func()

	grpcRegistrar GRPCRegistrar
}

// NewApp создает новое приложение с заданными опциями
func NewApp(ctx context.Context, cfg config.Config, opts ...Option) (*App, error) {
	app := &App{
		config:       cfg,
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
func (a *App) InitGRPCServer(interceptors ...grpc.UnaryServerInterceptor) {
	a.grpcServer = InitGRPCServer(interceptors...)
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

// Run запускает приложение
func (a *App) Run(ctx context.Context, grpcCfg config.GRPCConfig) error {
	if a.grpcServer == nil {
		return fmt.Errorf("gRPC server not initialized")
	}

	log.Printf("gRPC server starting on port %d", grpcCfg.Port)

	err := ServeGRPC(ctx, a.grpcServer, grpcCfg)

	// После завершения сервера выполняем cleanup
	a.Shutdown()

	return err
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
