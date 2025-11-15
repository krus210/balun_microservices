package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/logger"

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

	// Инициализируем PostgreSQL
	if err := application.InitPostgres(ctx, cfg.Database); err != nil {
		logger.FatalKV(ctx, "failed to init postgres", "error", err.Error())
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

	g, gCtx := errgroup.WithContext(ctx)

	// Запускаем воркер сохранения событий с настройками из конфигурации
	saveEventsWorker := worker.NewSaveEventsWorker(inboxRepo, application.TransactionManager()).
		WithTickInterval(workersCfg.SaveEvents.Interval).
		WithBatchSize(workersCfg.SaveEvents.BatchSize)
	g.Go(func() error {
		logger.InfoKV(gCtx, "starting save events worker")
		saveEventsWorker.Start(gCtx)
		logger.InfoKV(gCtx, "save events worker stopped")
		return nil
	})

	// Запускаем воркер удаления старых событий с настройками из конфигурации
	deleteWorker := worker.NewDeleteWorker(inboxRepo, application.TransactionManager()).
		WithTickInterval(workersCfg.Delete.Interval).
		WithRetentionPeriod(time.Duration(workersCfg.Delete.RetentionDays) * 24 * time.Hour)
	g.Go(func() error {
		logger.InfoKV(gCtx, "starting delete events worker")
		deleteWorker.Start(gCtx)
		logger.InfoKV(gCtx, "delete events worker stopped")
		return nil
	})

	// Запускаем Kafka consumer
	g.Go(func() error {
		logger.InfoKV(gCtx, "starting kafka consumer")
		return inboxConsumer.Run(gCtx, cfg.KafkaConsumer.Topics.FriendRequestEvents)
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

	logger.InfoKV(ctx, "starting notifications service",
		"version", cfg.Service.Version,
		"environment", cfg.Service.Environment,
		"brokers", cfg.KafkaConsumer.GetBrokers(),
	)

	waitErr := app.WaitForShutdown(ctx, g.Wait, app.GracefulShutdownTimeout)
	switch {
	case waitErr == nil || errors.Is(waitErr, context.Canceled):
		logger.InfoKV(ctx, "notifications service components stopped")
	case errors.Is(waitErr, context.DeadlineExceeded):
		logger.WarnKV(ctx, "graceful shutdown timeout exceeded, forcing cleanup")
	default:
		logger.FatalKV(ctx, "failed to serve", "error", waitErr.Error())
	}

	// Закрываем Kafka consumer (коммит оффсетов)
	logger.InfoKV(ctx, "closing kafka consumer")
	if err := inboxConsumer.Close(); err != nil {
		logger.ErrorKV(ctx, "failed to close kafka consumer", "error", err.Error())
	}

	// Закрываем PostgreSQL
	logger.InfoKV(ctx, "closing database connections")
	application.Shutdown()

	logger.InfoKV(ctx, "notifications service shutdown complete")
}
