package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(address []string) (*Producer, error) {
	const op = "kafka.NewProducer"
	conf := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(address, ","),
	}

	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Producer{producer: p}, nil
}

func (p *Producer) Produce(ctx context.Context, message []byte, topic string) error {
	const op = "kafka.Produce"
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny},
		Value: message,
	}

	kafkaChan := make(chan kafka.Event)
	if err := p.producer.Produce(kafkaMsg, kafkaChan); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	select {
	case ans := <-kafkaChan:
		switch ev := ans.(type) {
		case *kafka.Message:
			return nil
		case kafka.Error:
			return ev
		default:
			return fmt.Errorf("%s: unknown event type", op)
		}
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}
}

func (p *Producer) Close() {
	p.producer.Flush(int((5 * time.Millisecond) / time.Millisecond))
	p.producer.Close()
}
