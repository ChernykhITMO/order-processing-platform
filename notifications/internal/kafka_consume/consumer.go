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
	readTimeout    time.Duration
	log            *slog.Logger
}

func NewConsumer(address []string, handler Handler, topic, consumerGroup string, sessionTimeout, readTimeout time.Duration, log *slog.Logger) (*Consumer, error) {
	const op = "kafka_consume.NewConsumer"

	cfg := &kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(address, ","),
		"session.timeout.ms": int(sessionTimeout.Milliseconds()),
		"group.id":           consumerGroup,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := c.Subscribe(topic, nil); err != nil {
		return nil, fmt.Errorf("%s: subcribe to topic %w", op, err)
	}

	return &Consumer{
		consumer:       c,
		topic:          topic,
		handler:        handler,
		sessionTimeout: sessionTimeout,
		readTimeout:    readTimeout,
		log:            log}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	const op = "kafka_consume.Consumer.Start"

	log := c.log.With(
		slog.String("op", op),
		slog.String("topic", c.topic),
	)

	log.Debug("consumer started")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		kafkaMsg, err := c.consumer.ReadMessage(c.readTimeout)
		if err != nil {
			if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
				continue
			}
			if ctx.Err() != nil {
				return nil
			}
			log.Error("read message failed", slog.Any("err", err))
			continue
		}

		if kafkaMsg == nil {
			continue
		}

		if err := c.handler.HandleMessage(kafkaMsg.Value); err != nil {
			log.Error("handle message failed", slog.Any("err", err))
			continue
		}

		if _, err := c.consumer.CommitMessage(kafkaMsg); err != nil {
			log.Error("commit failed", slog.Any("err", err))
		}
	}
}

func (c *Consumer) Stop() error {
	const op = "app.Stop"
	log := c.log.With(slog.String("op", op))
	log.Debug("stopping consumer")
	return c.consumer.Close()
}
