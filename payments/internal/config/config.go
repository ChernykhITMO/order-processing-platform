package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DB            DBConfig
	HealthAddr    string
	KafkaBrokers  []string
	TopicOrder    string
	TopicStatus   string
	EventType     string
	ConsumerGroup string
	SenderPeriod  time.Duration
}

type DBConfig struct {
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
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
	maxConns, err := getEnvInt32WithDefault("PAYMENTS_PG_MAX_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	minConns, err := getEnvInt32WithDefault("PAYMENTS_PG_MIN_CONNS", 2)
	if err != nil {
		return Config{}, err
	}
	maxConnIdleTime, err := getEnvDurationWithDefault("PAYMENTS_PG_MAX_CONN_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	healthCheckPeriod, err := getEnvDurationWithDefault("PAYMENTS_PG_HEALTH_CHECK_PERIOD", 30*time.Second)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DB: DBConfig{
			DSN:               dsn,
			MaxConns:          maxConns,
			MinConns:          minConns,
			MaxConnIdleTime:   maxConnIdleTime,
			HealthCheckPeriod: healthCheckPeriod,
		},
		HealthAddr:    getEnvOrDefault("PAYMENTS_HEALTH_ADDR", ":8082"),
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

func getEnvOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
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

func getEnvInt32WithDefault(key string, def int32) (int32, error) {
	val := os.Getenv(key)
	if val == "" {
		return def, nil
	}

	parsed, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, errors.New(key + " is invalid int: " + err.Error())
	}

	return int32(parsed), nil
}

func getEnvDurationWithDefault(key string, def time.Duration) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		return def, nil
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
