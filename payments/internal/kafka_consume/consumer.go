package kafka_consume

import (
	"fmt"
	"log"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	sessionTimeout = 5000 // ms
	noTimeout      = -1
)

type Handler interface {
	HandleMessage(message []byte) error
}

type Consumer struct {
	consumer *kafka.Consumer
	handler  Handler
	stop     bool
}

func NewConsumer(handler Handler, address []string, topic, consumerGroup string) (*Consumer, error) {
	const op = "kafka.Consumer.New"
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(address, ","),
		"group.id":           consumerGroup,
		"session.timeout.ms": sessionTimeout,
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := c.Subscribe(topic, nil); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Consumer{consumer: c, handler: handler}, nil
}

func (c *Consumer) Start() {
	const op = "kafka.Start"
	for {
		if c.stop {
			break
		}
		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			log.Fatalf("%s: %w", op, err)
		}

		if kafkaMsg == nil {
			continue
		}

		if err := c.handler.HandleMessage(kafkaMsg.Value); err != nil {
			log.Fatalf("%s: %w", op, err)
		}

	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	return c.consumer.Close()
}
