package kafka_consume

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	sessionTimeoutMs = 6000
	readTimeout      = 2 * time.Second
)

type Handler interface {
	HandleMessage(message []byte) error
}

type Consumer struct {
	consumer *kafka.Consumer
	sender   Handler
	log      *slog.Logger
	topic    string
	group    string
}

func NewConsumer(sender Handler, address []string, topic, consumerGroup string, log *slog.Logger) (*Consumer, error) {
	const op = "kafka_produce.Consumer.New"
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(address, ","),
		"group.id":           consumerGroup,
		"session.timeout.ms": sessionTimeoutMs,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := c.Subscribe(topic, nil); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Consumer{consumer: c, sender: sender, log: log, topic: topic, group: consumerGroup}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	const op = "kafka_produce.Start"

	log := c.log.With(
		slog.String("op", op),
		slog.String("topic", c.topic),
		slog.String("group", c.group),
	)

	log.Info("consumer started")

	for {
		select {
		case <-ctx.Done():
			log.Info("consumer stopping", slog.Any("err", ctx.Err()))
			return
		default:
		}

		kafkaMsg, err := c.consumer.ReadMessage(readTimeout)
		if err != nil {
			if kerr, ok := err.(kafka.Error); ok {
				if kerr.Code() == kafka.ErrTimedOut {
					continue
				}
			}
			log.Error("read message failed", slog.Any("err", err))
			time.Sleep(1 * time.Second)
			continue
		}

		if kafkaMsg == nil {
			continue
		}

		if err := c.sender.HandleMessage(kafkaMsg.Value); err != nil {
			log.Error("handle message failed", slog.Any("err", err))
			continue
		}

		if _, err := c.consumer.CommitMessage(kafkaMsg); err != nil {
			log.Error("commit failed", slog.Any("err", err))
		}
	}
}

func (c *Consumer) Stop() error {
	c.log.Info("consumer stopping")
	return c.consumer.Close()
}
