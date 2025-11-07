package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/postgres"

	"notifications/internal/app/consumer"
	"notifications/internal/app/delivery"
	"notifications/internal/app/repository"
	"notifications/internal/app/worker"
	workersConfig "notifications/internal/workers"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию через lib/config
	cfg, err := config.LoadServiceConfig(ctx, "notifications")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Starting %s service (version: %s, environment: %s)",
		cfg.Service.Name, cfg.Service.Version, cfg.Service.Environment)

	// Загружаем конфигурацию workers
	workersCfg, err := workersConfig.Load()
	if err != nil {
		log.Fatalf("failed to load workers config: %v", err)
	}

	// Инициализируем PostgreSQL через lib/app
	conn, cleanup, err := app.InitPostgres(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	defer cleanup()

	// Создаем Transaction Manager
	txMngr := postgres.NewTransactionManager(conn)

	inboxRepo := repository.NewRepository(txMngr)

	handler := delivery.NewInboxHandler(inboxRepo)

	// Создаем Kafka consumer
	// Конвертируем строку с брокерами в слайс
	brokers := strings.Split(cfg.KafkaConsumer.GetBrokers(), ",")
	inboxConsumer, err := consumer.NewInboxConsumer(
		brokers,
		cfg.KafkaConsumer.ConsumerGroupID,
		cfg.KafkaConsumer.ConsumerName,
		handler,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer inboxConsumer.Close()

	g, gCtx := errgroup.WithContext(ctx)

	// Запускаем воркер сохранения событий с настройками из конфигурации
	saveEventsWorker := worker.NewSaveEventsWorker(inboxRepo, txMngr).
		WithTickInterval(workersCfg.SaveEvents.Interval).
		WithBatchSize(workersCfg.SaveEvents.BatchSize)
	g.Go(func() error {
		saveEventsWorker.Start(gCtx)
		return nil
	})

	// Запускаем воркер удаления старых событий с настройками из конфигурации
	deleteWorker := worker.NewDeleteWorker(inboxRepo, txMngr).
		WithTickInterval(workersCfg.Delete.Interval).
		WithRetentionPeriod(time.Duration(workersCfg.Delete.RetentionDays) * 24 * time.Hour)
	g.Go(func() error {
		deleteWorker.Start(gCtx)
		return nil
	})

	// Запускаем Kafka consumer
	g.Go(func() error {
		return inboxConsumer.Run(gCtx, cfg.KafkaConsumer.Topics.FriendRequestEvents)
	})

	if err := g.Wait(); err != nil && ctx.Err() == nil {
		log.Println("error:", err)
	}

	log.Println("shutdown")
}
