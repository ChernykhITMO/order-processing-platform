package config

import "time"

type Config struct {
	DBDSN         string
	KafkaBrokers  []string
	TopicOrder    string
	TopicStatus   string
	EventType     string
	ConsumerGroup string
	SenderPeriod  time.Duration
}
