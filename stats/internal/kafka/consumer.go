package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	logger "stats/internal/logger/slog"
	"stats/internal/storage"
	"time"

	"github.com/IBM/sarama"
)

const (
	consumerGroup  = "wallet-consumers"
	sessionTimeout = 5000 //ms
)

// Consumer представляет Kafka consumer service
type Consumer struct {
	Consumer sarama.ConsumerGroup
	Topic    string
	Storage  storage.Storage
	Logger   *slog.Logger
	Done     chan struct{}
}

// NewConsumer создает новый экземпляр Consumer
func NewConsumer(brokers []string, topic string, storage storage.Storage, logger *slog.Logger) (*Consumer, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Consumer.Group.Heartbeat.Interval = 3000
	kafkaConfig.Consumer.Group.Session.Timeout = sessionTimeout
	kafkaConfig.Consumer.Offsets.AutoCommit.Enable = false

	consumerGroup, err := sarama.NewConsumerGroup(
		brokers,
		consumerGroup,
		kafkaConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &Consumer{
		Consumer: consumerGroup,
		Topic:    topic,
		Storage:  storage,
		Logger:   logger,
		Done:     make(chan struct{}),
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	defer close(c.Done)

	handler := &consumerHandler{
		logger:  c.Logger,
		storage: c.Storage,
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c.Logger.Info("Starting Kafka consumer",
		slog.String("topic", c.Topic),
		slog.String("group", consumerGroup),
	)

	for {
		select {
		case <-ctx.Done():
			c.Logger.Info("Shutting down consumer")
			return nil
		default:
			if err := c.Consumer.Consume(ctx, []string{c.Topic}, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return nil
				}
				c.Logger.Error("Consume error", logger.Err(err))
				return fmt.Errorf("consume error: %w", err)
			}
		}
	}
}

// Close останавливает consumer
func (c *Consumer) Close() error {
	if err := c.Consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	<-c.Done // Ждем завершения Run()
	return nil
}

// consumerHandler реализует sarama.ConsumerGroupHandler
type consumerHandler struct {
	logger  *slog.Logger
	storage storage.Storage
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Debug("Consumer group session setup")
	return nil
}

func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Debug("Consumer group session cleanup")
	return nil
}

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		startTime := time.Now()

		var event Event
		if err := json.Unmarshal(message.Value, &event); err != nil {
			slog.Error("Message unmarshal failed",
				logger.Err(err),
				slog.String("topic", message.Topic),
				slog.Int64("partition", int64(message.Partition)),
				slog.Int64("offset", message.Offset),
			)
			continue
		}

		tx, err := h.storage.BeginTx(context.Background())
		if err != nil {
			slog.Error("BeginTx error",
				logger.Err(err),
				slog.String("topic", message.Topic),
				slog.Int64("partition", int64(message.Partition)),
				slog.Int64("offset", message.Offset),
			)

			continue
		}

		if err := h.handleEvent(event, tx); err != nil {
			slog.Error("Message processing failed",
				logger.Err(err),
				slog.String("topic", message.Topic),
				slog.Int64("partition", int64(message.Partition)),
				slog.Int64("offset", message.Offset),
			)

			tx.Rollback()
			continue
		}

		if err := tx.Commit(); err != nil {
			continue
		}

		session.MarkMessage(message, "")
		session.Commit()

		h.logger.Debug("Message processed",
			slog.String("topic", message.Topic),
			slog.Int64("partition", int64(message.Partition)),
			slog.Int64("offset", message.Offset),
			slog.Duration("duration", time.Since(startTime)),
		)
	}
	return nil
}

func (h *consumerHandler) handleEvent(event Event, tx storage.Transaction) error {
	switch event.Type {
	case EventWalletCreated:
		var payload WalletCreatedPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		log.Printf("Wallet created: ID=%s", payload.ID)
		return tx.UpdateStats(context.Background(), storage.OpCreate)

	case EventWalletDeposited:
		var payload WalletDepositedPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		log.Printf("Deposit: ID=%s, Amount=%.2f", payload.ID, payload.Amount)
		return tx.UpdateStats(context.Background(), storage.OpDeposit, payload.Amount)

	case EventWalletWithdrawn:
		var payload WalletWithdrawnPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		log.Printf("Withdrawal: ID=%s, Amount=%.2f", payload.ID, payload.Amount)
		return tx.UpdateStats(context.Background(), storage.OpWithdraw, payload.Amount)

	case EventWalletTransferred:
		var payload WalletTransferredPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		log.Printf("Transfer: From=%s, To=%s, Amount=%.2f", payload.ID, payload.TransferTo, payload.Amount)
		return tx.UpdateStats(context.Background(), storage.OpTransfer, payload.Amount)

	case EventWalletDeleted:
		var payload WalletDeletedPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		log.Printf("Wallet deleted: ID=%s", payload.ID)
		return tx.UpdateStats(context.Background(), storage.OpDelete)

	default:
		log.Printf("Unknown event type: %s", event.Type)
		return nil
	}
}
