package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/config"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db        *pgxpool.Pool
	txManager *txmanager.Manager
}

func New(cfg config.DBConfig) (*Storage, error) {
	const op = "storage.postgres.New"

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns

	db, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db:        db,
		txManager: txmanager.New(db),
	}, nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
