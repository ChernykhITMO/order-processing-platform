package kafka_produce

import (
	"context"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const flushTimeoutMs = 5000

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(address []string) (*Producer, error) {
	const op = "kafka_produce.NewProducer"
	conf := &kafka.ConfigMap{
		"bootstrap.servers":   strings.Join(address, ","),
		"acks":                "all",
		"enable.idempotence":  true,
		"retries":             10,
		"request.timeout.ms":  15000,
		"delivery.timeout.ms": 60000,
	}

	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Producer{producer: p}, nil
}

func (p *Producer) Produce(ctx context.Context, message []byte, topic string) error {
	const op = "kafka_produce.Produce"
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny},
		Value: message,
		Key:   nil,
	}

	kafkaChan := make(chan kafka.Event, 1)
	if err := p.producer.Produce(kafkaMsg, kafkaChan); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", op, ctx.Err())
	case ans := <-kafkaChan:
		switch ev := ans.(type) {
		case *kafka.Message:
			if err := ev.TopicPartition.Error; err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil

		default:
			return fmt.Errorf("%s: unexpected event type: %T", op, ans)
		}
	}
}

func (p *Producer) Close() error {
	remaining := p.producer.Flush(flushTimeoutMs)
	p.producer.Close()
	if remaining > 0 {
		return fmt.Errorf("kafka producer: %d messages not delivered on close", remaining)
	}
	return nil
}
