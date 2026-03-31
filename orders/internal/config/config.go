package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env    string
	GRPC   GRPCConfig
	Health HealthConfig
	DB     DBConfig
	Kafka  KafkaConfig
}

type GRPCConfig struct {
	Port int
}

type HealthConfig struct {
	Addr string
}

type DBConfig struct {
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	Period  time.Duration
}

func Load(envKey, grpcPortKey, healthAddrKey, pgDSNKey, kafkaBrokersKey, kafkaTopicKey, kafkaPeriodKey string) (*Config, error) {
	env := getEnv(envKey)
	if env == "" {
		return nil, fmt.Errorf("env %s is empty", envKey)
	}

	grpcPortStr := getEnv(grpcPortKey)
	if grpcPortStr == "" {
		return nil, fmt.Errorf("env %s is empty", grpcPortKey)
	}
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", grpcPortKey, err)
	}

	pgDSN := getEnv(pgDSNKey)
	if pgDSN == "" {
		return nil, fmt.Errorf("env %s is empty", pgDSNKey)
	}
	healthAddr := getEnvWithDefault(healthAddrKey, ":8081")
	maxConns, err := getEnvInt32WithDefault("ORDERS_PG_MAX_CONNS", 10)
	if err != nil {
		return nil, err
	}
	minConns, err := getEnvInt32WithDefault("ORDERS_PG_MIN_CONNS", 2)
	if err != nil {
		return nil, err
	}
	maxConnIdleTime, err := getEnvDurationWithDefault("ORDERS_PG_MAX_CONN_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return nil, err
	}
	healthCheckPeriod, err := getEnvDurationWithDefault("ORDERS_PG_HEALTH_CHECK_PERIOD", 30*time.Second)
	if err != nil {
		return nil, err
	}

	kafkaBrokers := parseCSV(getEnv(kafkaBrokersKey))
	kafkaTopic := getEnv(kafkaTopicKey)
	kafkaPeriod, err := parseDuration(getEnv(kafkaPeriodKey))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", kafkaPeriodKey, err)
	}

	return &Config{
		Env: env,
		GRPC: GRPCConfig{
			Port: grpcPort,
		},
		Health: HealthConfig{
			Addr: healthAddr,
		},
		DB: DBConfig{
			DSN:               pgDSN,
			MaxConns:          maxConns,
			MinConns:          minConns,
			MaxConnIdleTime:   maxConnIdleTime,
			HealthCheckPeriod: healthCheckPeriod,
		},
		Kafka: KafkaConfig{
			Brokers: kafkaBrokers,
			Topic:   kafkaTopic,
			Period:  kafkaPeriod,
		},
	}, nil
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func getEnvWithDefault(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return def
}

func getEnvInt32WithDefault(key string, def int32) (int32, error) {
	value := os.Getenv(key)
	if value == "" {
		return def, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return int32(parsed), nil
}

func getEnvDurationWithDefault(key string, def time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return def, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return parsed, nil
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
