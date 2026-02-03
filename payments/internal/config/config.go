package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	DBDSN         string
	KafkaBrokers  []string
	TopicOrder    string
	TopicStatus   string
	EventType     string
	ConsumerGroup string
	SenderPeriod  time.Duration
}

func Load() (Config, error) {
	dsn, err := getEnv("PAYMENTS_PG_DSN")
	if err != nil {
		return Config{}, err
	}
	brokersRaw, err := getEnv("KAFKA_BROKERS")
	if err != nil {
		return Config{}, err
	}
	kafkaBrokers := parseKafkaBrokers(brokersRaw)
	if len(kafkaBrokers) == 0 {
		return Config{}, errors.New("KAFKA_BROKERS is empty")
	}
	topicOrder, err := getEnv("KAFKA_TOPIC_ORDER")
	if err != nil {
		return Config{}, err
	}
	topicStatus, err := getEnv("KAFKA_TOPIC_STATUS")
	if err != nil {
		return Config{}, err
	}
	eventType, err := getEnv("KAFKA_EVENT_TYPE")
	if err != nil {
		return Config{}, err
	}
	consumerGroup, err := getEnv("KAFKA_CONSUMER_GROUP")
	if err != nil {
		return Config{}, err
	}
	senderPeriod, err := getEnvDuration("KAFKA_SENDER_PERIOD")
	if err != nil {
		return Config{}, err
	}

	return Config{
		DBDSN:         dsn,
		KafkaBrokers:  kafkaBrokers,
		TopicOrder:    topicOrder,
		TopicStatus:   topicStatus,
		EventType:     eventType,
		ConsumerGroup: consumerGroup,
		SenderPeriod:  senderPeriod,
	}, nil
}

func getEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", errors.New(key + " is empty")
	}
	return val, nil
}

func getEnvDuration(key string) (time.Duration, error) {
	val, err := getEnv(key)
	if err != nil {
		return 0, err
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return 0, errors.New(key + " is invalid duration: " + err.Error())
	}
	return parsed, nil
}

func parseKafkaBrokers(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
