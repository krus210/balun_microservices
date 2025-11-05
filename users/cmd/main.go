package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	"users/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию с интеграцией secrets
	cfg, err := config.LoadWithSecrets(ctx)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Starting %s service (version: %s, environment: %s)",
		cfg.Service.Name, cfg.Service.Version, cfg.Service.Environment)

	server, cleanup, err := InitializeApp(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer cleanup()

	listenAddr := fmt.Sprintf(":%d", cfg.Server.GRPC.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
