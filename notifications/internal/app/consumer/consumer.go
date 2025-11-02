package consumer

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/IBM/sarama"
)

var topicRE = regexp.MustCompile(`^[A-Za-z0-9._-]{1,249}$`)

type handler interface {
	SaveInboxMessage(ctx context.Context, message *sarama.ConsumerMessage) (needMark bool, err error)
}

type InboxConsumer struct {
	group        sarama.ConsumerGroup
	batchSize    int
	batchTimeout time.Duration
	consumerName string
	handler      handler
}

func NewInboxConsumer(brokers []string, groupID string, consumerName string, handler handler) (*InboxConsumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_5_0_0
	cfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Return.Errors = true // ОБЯЗАТЕЛЬНО ЧИТАЕМ cg.group.Errors()

	cg, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}

	return &InboxConsumer{
		group:        cg,
		batchSize:    128,
		batchTimeout: 300 * time.Millisecond,
		consumerName: consumerName,
		handler:      handler,
	}, nil
}

func (c *InboxConsumer) Close() error { return c.group.Close() }

func (c *InboxConsumer) Run(ctx context.Context, topics ...string) error {
	for _, t := range topics {
		if !topicRE.MatchString(t) {
			return sarama.ConfigurationError("invalid topic: " + t)
		}
	}

	// отдельная горутина для ошибок Сonsumer Group (полезно для диагностики)
	go func() {
		for err := range c.group.Errors() {
			log.Printf("[consumer-group] error: %v", err)
		}
	}()

	handler := &consumerGroupHandler{c: c}
	for {
		if err := c.group.Consume(ctx, topics, handler); err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

type consumerGroupHandler struct{ c *InboxConsumer }

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim вызывается отдельно на КАЖДУЮ партицию (важно для порядка сообщений)
func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	batch := make([]*sarama.ConsumerMessage, 0, h.c.batchSize)

	timer := time.NewTimer(h.c.batchTimeout)
	defer timer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		for _, msg := range batch {
			needMark, err := h.c.handler.SaveInboxMessage(sess.Context(), msg)
			if err != nil {
				log.Printf(err.Error())
			}
			if needMark {
				sess.MarkMessage(msg, "")
			}
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-sess.Context().Done():
			return nil

		case m, ok := <-claim.Messages():
			if !ok {
				flush()
				return nil
			}
			batch = append(batch, m)
			if len(batch) >= h.c.batchSize {
				flush()
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(h.c.batchTimeout)
			}

		case <-timer.C:
			flush()
			timer.Reset(h.c.batchTimeout)
		}
	}
}
