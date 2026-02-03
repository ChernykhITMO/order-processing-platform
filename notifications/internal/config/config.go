package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr           string        `yaml:"addr"`
	Password       string        `yaml:"password"`
	User           string        `yaml:"user"`
	DB             int           `yaml:"db"`
	MaxRetries     int           `yaml:"max_retries"`
	DialTimeout    time.Duration `yaml:"dial_timeout"`
	Timeout        time.Duration `yaml:"timeout"`
	TTL            time.Duration `yaml:"ttl"`
	KafkaBrokers   []string
	TopicStatus    string
	ConsumerGroup  string
	SessionTimeout time.Duration
	ReadTimeout    time.Duration
}

func Load() (Config, error) {
	redisAddr, err := getEnv("REDIS_ADDR")
	if err != nil {
		return Config{}, err
	}
	redisDB, err := getEnvInt("REDIS_DB")
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
	topicStatus, err := getEnv("KAFKA_TOPIC_STATUS")
	if err != nil {
		return Config{}, err
	}
	consumerGroup, err := getEnv("KAFKA_CONSUMER_GROUP")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Addr:           redisAddr,
		DB:             redisDB,
		TTL:            getEnvDuration("REDIS_TTL", 0),
		SessionTimeout: getEnvDuration("SESSION_TIMEOUT", 30*time.Second),
		KafkaBrokers:   kafkaBrokers,
		TopicStatus:    topicStatus,
		ConsumerGroup:  consumerGroup,
		ReadTimeout:    getEnvDuration("READ_TIMEOUT", time.Second),
	}, nil
}

func getEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", errors.New(key + " is empty")
	}
	return val, nil
}

func getEnvInt(key string) (int, error) {
	val, err := getEnv(key)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.New(key + " is invalid int: " + err.Error())
	}
	return parsed, nil
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return def
	}
	return parsed
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
