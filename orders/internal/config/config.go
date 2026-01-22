package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env   string
	GRPC  GRPCConfig
	DB    DBConfig
	Kafka KafkaConfig
}

type GRPCConfig struct {
	Port int
}

type DBConfig struct {
	DSN string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	Period  time.Duration
}

func MustLoad(envKey, grpcPortKey, pgDSNKey, kafkaBrokersKey, kafkaTopicKey, kafkaPeriodKey string) *Config {
	env := getEnv(envKey)
	if env == "" {
		return nil
	}

	grpcPortStr := getEnv(grpcPortKey)
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		return nil
	}

	pgDSN := getEnv(pgDSNKey)
	if pgDSN == "" {
		return nil
	}

	kafkaBrokers := parseCSV(getEnv(kafkaBrokersKey))
	kafkaTopic := getEnv(kafkaTopicKey)
	kafkaPeriod, err := parseDuration(getEnv(kafkaPeriodKey))
	if err != nil {
		return nil
	}

	return &Config{
		Env: env,
		GRPC: GRPCConfig{
			Port: grpcPort,
		},
		DB: DBConfig{
			DSN: pgDSN,
		},
		Kafka: KafkaConfig{
			Brokers: kafkaBrokers,
			Topic:   kafkaTopic,
			Period:  kafkaPeriod,
		},
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func parseDuration(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	return time.ParseDuration(value)
}

func parseCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		item := strings.TrimSpace(p)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}
