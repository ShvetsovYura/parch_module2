package consumer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"oyevents/internal/types"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type EventsConsumer struct {
	consumer *kafka.Consumer
	topics   []string
}

func NewEventsConsumer(topics []string, config types.ConsumerConfig) (*EventsConsumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  config.BootstrapServers,
		"group.id":           config.GroupID,
		"auto.offset.reset":  config.AutoOffsetReset,
		"enable.auto.commit": config.EnableAutoCommit,
		"session.timeout.ms": config.SessionTimeoutMs,
		// "security.protocol":  "PLAINTEXT",
		"client.id": config.ClientID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %s", err)
	}
	return &EventsConsumer{
		topics:   topics,
		consumer: consumer,
	}, nil
}

func (c *EventsConsumer) Run(ctx context.Context, wg *sync.WaitGroup) error {

	err := c.consumer.SubscribeTopics(c.topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	defer c.consumer.Close()

	for {
		msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
		if err == nil {
			value := string(msg.Value)
			slog.Info("Received message", slog.String("content", value), slog.String("topic", *msg.TopicPartition.Topic))

		} else {
			// Handle errors
			var kafkaErr kafka.Error
			if errors.As(err, &kafkaErr) && kafkaErr.IsFatal() {
				log.Fatalf("Fatal error: %s", kafkaErr)
			}
		}
	}
}
