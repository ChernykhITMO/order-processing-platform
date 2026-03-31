package services

import (
	"log/slog"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/storage/postgres"
)

type Order struct {
	log  *slog.Logger
	repo postgres.Repository
}

func New(log *slog.Logger, repo postgres.Repository) *Order {
	return &Order{
		log:  log,
		repo: repo,
	}
}
