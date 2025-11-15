package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/grpc/server/interceptors"
)

// GRPCRegistrar определяет функцию для регистрации gRPC сервисов
type GRPCRegistrar func(*grpc.Server)

// InitGRPCServer создает новый gRPC сервер с настройками по умолчанию и встроенными интерсепторами
//
// Порядок interceptors (важен!):
// 1. Panic recovery - перехват паник
// 2. Rate limit (если enabled) - ограничение запросов
// 3. Timeout (если enabled) - таймауты для запросов
// 4. Custom interceptors - пользовательские интерсепторы (например, errors middleware)
//
// OpenTelemetry tracing настраивается через stats handler (не через interceptor)
func InitGRPCServer(cfg config.ServerConfig, customInterceptors ...grpc.UnaryServerInterceptor) *grpc.Server {
	var interceptorChain []grpc.UnaryServerInterceptor

	// 1. Panic recovery (всегда первый для перехвата любых паник)
	interceptorChain = append(interceptorChain, interceptors.PanicRecoveryUnaryInterceptor())

	interceptorChain = append(interceptorChain, interceptors.DebugOpenTelemetryUnaryServerInterceptor(true, true))

	interceptorChain = append(interceptorChain, interceptors.LogErrorUnaryInterceptor())

	// 2. Rate limit (если enabled)
	if cfg.RateLimit != nil && cfg.RateLimit.Enabled {
		interceptorChain = append(interceptorChain, interceptors.RateLimitUnaryInterceptor(*cfg.RateLimit))
	}

	// 3. Timeout (если enabled)
	if cfg.Timeout != nil && cfg.Timeout.Enabled {
		interceptorChain = append(interceptorChain, interceptors.TimeoutUnaryInterceptor(*cfg.Timeout))
	}

	// 4. Custom interceptors (например, ErrorsUnaryInterceptor)
	interceptorChain = append(interceptorChain, customInterceptors...)

	opts := []grpc.ServerOption{
		// OpenTelemetry tracing через stats handler (современный подход, заменяет deprecated interceptors)
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(interceptorChain...),
	}

	server := grpc.NewServer(opts...)

	// Включаем reflection для удобной разработки (grpc_cli)
	reflection.Register(server)

	return server
}

// ServeGRPC запускает gRPC сервер на указанном порту
func ServeGRPC(ctx context.Context, server *grpc.Server, grpcCfg config.GRPCConfig) error {
	listenAddr := fmt.Sprintf(":%d", grpcCfg.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	errChan := make(chan error, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.Serve(lis); err != nil {
			errChan <- fmt.Errorf("grpc server failed: %w", err)
		}
	}()

	// Ждем сигнала отмены или ошибки
	select {
	case <-ctx.Done():
		gracefulDone := make(chan struct{})

		go func() {
			server.GracefulStop()
			close(gracefulDone)
		}()

		select {
		case <-gracefulDone:
			return ctx.Err()
		case <-time.After(GracefulShutdownTimeout):
			log.Printf("gRPC graceful shutdown exceeded %s, forcing stop", GracefulShutdownTimeout)
			server.Stop()
			return ctx.Err()
		}
	case err := <-errChan:
		return err
	}
}
