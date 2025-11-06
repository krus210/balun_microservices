package app

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/sskorolev/balun_microservices/lib/config"
)

// GRPCRegistrar определяет функцию для регистрации gRPC сервисов
type GRPCRegistrar func(*grpc.Server)

// InitGRPCServer создает новый gRPC сервер с настройками по умолчанию
func InitGRPCServer(interceptors ...grpc.UnaryServerInterceptor) *grpc.Server {
	opts := []grpc.ServerOption{}

	if len(interceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(interceptors...))
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
		server.GracefulStop()
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
