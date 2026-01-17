package redis_storage

import (
	"context"
	"encoding/json"

	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain"
	"github.com/ChernykhITMO/order-processing-platform/notifications/internal/domain/events"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
}

func New(cfg config.Config) *Storage {
	client := redis.NewClient(&redis.Options{
		Addr:        cfg.Addr,
		Password:    cfg.Password,
		Username:    cfg.User,
		DB:          cfg.DB,
		MaxRetries:  cfg.MaxRetries,
		DialTimeout: cfg.DialTimeout,
	})
	return &Storage{client: client}
}
func (s *Storage) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *Storage) SaveNotification(ctx context.Context, key string, value events.Payment) error {
	payment, err := json.Marshal(&value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, payment, 0).Err()
}
func (s *Storage) GetNotification(ctx context.Context, key string) (events.Payment, error) {
	var payment events.Payment
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return payment, domain.ErrNotFound
		}
		return payment, err
	}

	if err := json.Unmarshal([]byte(val), &payment); err != nil {
		return payment, err
	}
	return payment, nil
}

func (s *Storage) Close() error {
	return s.client.Close()
}
