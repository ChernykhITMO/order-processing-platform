package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain/events"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) GetNewEvent(ctx context.Context) (events.OrderCreated, int64, error) {
	const op = "storage.postgres.GetNewEvent"
	var (
		createdOrder events.OrderCreated
		payload      []byte
		eventID      int64
		found        bool
	)

	const query = `
		UPDATE events
		SET locked_at = $1
		WHERE id = (
			SELECT id
			FROM events
			WHERE sent_at IS NULL AND (locked_at IS NULL OR locked_at < now() - interval '1 minutes')
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING payload, id
	`

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, query, time.Now()).Scan(&payload, &eventID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("%s: %w", op, err)
		}
		found = true

		if err := json.Unmarshal(payload, &createdOrder); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
	if err != nil {
		return createdOrder, eventID, err
	}
	if !found {
		return events.OrderCreated{}, 0, nil
	}

	return createdOrder, eventID, nil
}
