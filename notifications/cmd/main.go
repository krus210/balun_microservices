package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"notifications/internal/app/consumer"
	"notifications/internal/app/delivery"
	"notifications/internal/app/repository"
	"notifications/internal/app/worker"
	"notifications/pkg/postgres"
	"notifications/pkg/postgres/transaction_manager"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conn, err := postgres.NewConnectionPool(ctx, DSN(),
		postgres.WithMaxConnIdleTime(time.Minute),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	txMngr := transaction_manager.New(conn)

	inboxRepo := repository.NewRepository(txMngr)

	handler := delivery.NewInboxHandler(inboxRepo)

	inboxConsumer, err := consumer.NewInboxConsumer(
		[]string{kafkaBrokers},
		kafkaConsumerGroupID,
		kafkaConsumerName,
		handler,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer inboxConsumer.Close()

	g, gCtx := errgroup.WithContext(ctx)

	saveEventsWorker := worker.NewSaveEventsWorker(inboxRepo, txMngr)
	g.Go(func() error {
		saveEventsWorker.Start(gCtx)
		return nil
	})

	deleteWorker := worker.NewDeleteWorker(inboxRepo, txMngr)
	g.Go(func() error {
		go deleteWorker.Start(gCtx)
		return nil
	})

	g.Go(func() error {
		return inboxConsumer.Run(gCtx, kafkaFriendRequestEventTopicName)
	})

	if err := g.Wait(); err != nil && ctx.Err() == nil {
		log.Println("error:", err)
	}

	log.Println("shutdown")
}
