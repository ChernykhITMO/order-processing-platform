package config

import "time"

type Config struct {
	Addr           string        `yaml:"addr"`
	Password       string        `yaml:"password"`
	User           string        `yaml:"user"`
	DB             int           `yaml:"db"`
	MaxRetries     int           `yaml:"max_retries"`
	DialTimeout    time.Duration `yaml:"dial_timeout"`
	Timeout        time.Duration `yaml:"timeout"`
	KafkaBrokers   []string
	TopicStatus    string
	ConsumerGroup  string
	SessionTimeout time.Duration
}
