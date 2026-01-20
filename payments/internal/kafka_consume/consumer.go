package kafka_consume

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	sessionTimeout = 6000 // ms
)

type Handler interface {
	HandleMessage(message []byte) error
}

type Consumer struct {
	consumer *kafka.Consumer
	handler  Handler
	log      *slog.Logger
	topic    string
	group    string
}

func NewConsumer(handler Handler, address []string, topic, consumerGroup string, log *slog.Logger) (*Consumer, error) {
	const op = "kafka_produce.Consumer.New"
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(address, ","),
		"group.id":           consumerGroup,
		"session.timeout.ms": sessionTimeout,
		"auto.offset.reset":  "earliest",
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := c.Subscribe(topic, nil); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Consumer{consumer: c, handler: handler, log: log, topic: topic, group: consumerGroup}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	const op = "kafka_produce.Start"
	c.log.Info("consumer started", slog.String("op", op), slog.String("topic", c.topic), slog.String("group", c.group))

	for {
		select {
		case <-ctx.Done():
			c.log.Info("consumer stopping", slog.String("op", op), slog.Any("err", ctx.Err()))
			return
		default:
		}

		kafkaMsg, err := c.consumer.ReadMessage(sessionTimeout)
		if err != nil {
			if kerr, ok := err.(kafka.Error); ok {
				if kerr.Code() == kafka.ErrTimedOut {
					continue
				}
			}
			c.log.Error("read message failed", slog.String("op", op), slog.Any("err", err))
			break
		}
		if kafkaMsg == nil {
			continue
		}

		l := c.log.With(
			slog.String("topic", *kafkaMsg.TopicPartition.Topic),
			slog.Int("partition", int(kafkaMsg.TopicPartition.Partition)),
			slog.Int64("offset", int64(kafkaMsg.TopicPartition.Offset)),
		)

		if err := c.handler.HandleMessage(kafkaMsg.Value); err != nil {
			l.Error("handle message failed", slog.String("op", op), slog.Any("err", err))
			continue
		}

		l.Debug("message handled", slog.String("op", op), slog.Int("bytes", len(kafkaMsg.Value)))
	}
}

func (c *Consumer) Stop() error {
	c.log.Info("consumer stopping")
	return c.consumer.Close()
}
