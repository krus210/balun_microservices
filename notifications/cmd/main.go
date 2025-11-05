package main

import (
	"context"
	"lib/postgres"
	"log"
	"os/signal"
	"syscall"

	"notifications/internal/app/consumer"
	"notifications/internal/app/delivery"
	"notifications/internal/app/repository"
	"notifications/internal/app/worker"
	"notifications/internal/config"

	"golang.org/x/sync/errgroup"
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

	// Подключаемся к PostgreSQL
	conn, txMngr, err := postgres.New(ctx,
		postgres.WithHost(cfg.Database.Host),
		postgres.WithPort(cfg.Database.Port),
		postgres.WithDatabase(cfg.Database.Name),
		postgres.WithUser(cfg.Database.User),
		postgres.WithPassword(cfg.Database.Password),
		postgres.WithSSLMode(cfg.Database.SSLMode),
		postgres.WithMaxConnIdleTime(cfg.Database.MaxConnIdleTime),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	inboxRepo := repository.NewRepository(txMngr)

	handler := delivery.NewInboxHandler(inboxRepo)

	// Создаем Kafka consumer
	inboxConsumer, err := consumer.NewInboxConsumer(
		cfg.Kafka.GetBrokers(),
		cfg.Kafka.ConsumerGroupID,
		cfg.Kafka.ConsumerName,
		handler,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer inboxConsumer.Close()

	g, gCtx := errgroup.WithContext(ctx)

	// Запускаем воркер сохранения событий
	saveEventsWorker := worker.NewSaveEventsWorker(inboxRepo, txMngr)
	g.Go(func() error {
		saveEventsWorker.Start(gCtx)
		return nil
	})

	// Запускаем воркер удаления старых событий
	deleteWorker := worker.NewDeleteWorker(inboxRepo, txMngr)
	g.Go(func() error {
		go deleteWorker.Start(gCtx)
		return nil
	})

	// Запускаем Kafka consumer
	g.Go(func() error {
		return inboxConsumer.Run(gCtx, cfg.Kafka.Topics.FriendRequestEvents)
	})

	if err := g.Wait(); err != nil && ctx.Err() == nil {
		log.Println("error:", err)
	}

	log.Println("shutdown")
}
