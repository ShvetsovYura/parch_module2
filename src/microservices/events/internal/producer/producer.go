package producer

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"oyevents/internal/types"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ClientProducer struct {
	producer *kafka.Producer
}

func NewEventsProducer(cfg types.ProducerConfig) *ClientProducer {
	cfgMap := kafka.ConfigMap{
		"bootstrap.servers": cfg.BootstrapServers,
		"acks":              cfg.Acks,
		"client.id":         cfg.ClientID,
		// "security.protocol": "PLAINTEXT",
	}

	p, err := kafka.NewProducer(&cfgMap)
	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}
	return &ClientProducer{
		producer: p,
	}
}

func (p *ClientProducer) Run(ctx context.Context, wg *sync.WaitGroup, reqCh <-chan types.EventMessage) {

	deliveryChan := make(chan kafka.Event)

	defer func() {
		p.producer.Close()
		close(deliveryChan)
	}()

	slog.Info("Producer running", slog.Any("producer", p))

	for {
		select {
		case <-ctx.Done():
			slog.Info("Получен сигнал выхода, остановка продьюсера...")
			wg.Done()
		case value := <-reqCh:
			var payload []byte
			val, ok := value.Msg.([]byte)
			if ok {
				payload = val
			} else {
				var err error
				payload, err = json.Marshal(value.Msg)
				if err != nil {
					slog.Warn("Failed to serialize payload", slog.Any("error", err))
					continue
				}
			}

			err := p.producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &value.Topic, Partition: kafka.PartitionAny},
				Value:          payload,
			}, deliveryChan)

			if err != nil {
				slog.Warn("Produce failed", slog.Any("error", err))
			}

			e := <-deliveryChan
			m := e.(*kafka.Message)

			if m.TopicPartition.Error != nil {
				slog.Info("Delivery failed", slog.Any("error", m.TopicPartition.Error))
			} else {
				slog.Info("Delivered message ",
					slog.String("topic", *m.TopicPartition.Topic),
					slog.Int("partition", int(m.TopicPartition.Partition)),
					slog.Any("offset", m.TopicPartition.Offset))
			}
		}

	}

}
