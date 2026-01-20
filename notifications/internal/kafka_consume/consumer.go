package kafka_consume

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Handler interface {
	HandleMessage(message []byte) error
}

type Consumer struct {
	consumer       *kafka.Consumer
	handler        Handler
	topic          string
	sessionTimeout time.Duration
	log            *slog.Logger
}

func NewConsumer(address []string, handler Handler, topic string, sessionTimeout time.Duration, log *slog.Logger) (*Consumer, error) {
	const op = "kafka_produce.Consumer.New"
	sessionTimeoutMs := int(sessionTimeout.Milliseconds())
	heartbeatIntervalMs := int((sessionTimeout / 3).Milliseconds())
	if sessionTimeoutMs < 1 {
		sessionTimeoutMs = 1
	}
	if heartbeatIntervalMs < 1 {
		heartbeatIntervalMs = 1
	}
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":     strings.Join(address, ","),
		"session.timeout.ms":    sessionTimeoutMs,
		"heartbeat.interval.ms": heartbeatIntervalMs,
		"group.id":              "my-group",
		"auto.offset.reset":     "earliest",
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := c.Subscribe(topic, nil); err != nil {
		return nil, fmt.Errorf("%s: subcribe to topic %w", op, err)
	}

	return &Consumer{consumer: c, topic: topic, handler: handler, sessionTimeout: sessionTimeout, log: log}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	const op = "kafka_produce.Consumer.Start"
	c.log.Info("consumer started", slog.String("op", op), slog.String("topic", c.topic))

	for {
		select {
		case <-ctx.Done():
			c.log.Info("consumer stopping", slog.String("op", op), slog.Any("err", ctx.Err()))
			return
		default:
		}

		kafkaMsg, err := c.consumer.ReadMessage(c.sessionTimeout)
		if err != nil {
			if kerr, ok := err.(kafka.Error); ok {
				if kerr.Code() == kafka.ErrTimedOut || kerr.Code() == kafka.ErrUnknownTopic {
					continue
				}
				c.log.Error("read message failed", slog.String("op", op), slog.Any("err", err))
				break
			}
		}

		l := c.log.With(
			slog.String("topic", *kafkaMsg.TopicPartition.Topic),
			slog.Int("partition", int(kafkaMsg.TopicPartition.Partition)),
			slog.Int64("offset", int64(kafkaMsg.TopicPartition.Offset)),
		)

		if err := c.handler.HandleMessage(kafkaMsg.Value); err != nil {
			c.log.Error("handle message", slog.String("op", op), slog.String("err", err.Error()))
			return
		}

		l.Debug("message handled", slog.String("op", op), slog.Int("bytes", len(kafkaMsg.Value)))
	}
}

func (c *Consumer) Stop() error {
	return c.consumer.Close()
}
