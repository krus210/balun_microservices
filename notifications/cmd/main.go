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

	// Создаем приложение
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	// Инициализируем PostgreSQL
	if err := application.InitPostgres(ctx, cfg.Database); err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}

	// Загружаем конфигурацию workers
	workersCfg, err := workersConfig.Load()
	if err != nil {
		log.Fatalf("failed to load workers config: %v", err)
	}

	inboxRepo := repository.NewRepository(application.TransactionManager())

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
	saveEventsWorker := worker.NewSaveEventsWorker(inboxRepo, application.TransactionManager()).
		WithTickInterval(workersCfg.SaveEvents.Interval).
		WithBatchSize(workersCfg.SaveEvents.BatchSize)
	g.Go(func() error {
		saveEventsWorker.Start(gCtx)
		return nil
	})

	// Запускаем воркер удаления старых событий с настройками из конфигурации
	deleteWorker := worker.NewDeleteWorker(inboxRepo, application.TransactionManager()).
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

	// Выполняем graceful shutdown
	application.Shutdown()

	log.Println("shutdown")
}
