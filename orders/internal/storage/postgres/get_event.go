package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain/events"
)

func (s *Storage) GetNewEvent(ctx context.Context) (events.OrderCreated, int64, error) {
	const op = "storage.postgres.GetNewEvent"
	var (
		createdOrder events.OrderCreated
		payload      []byte
		eventID      int64
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return createdOrder, eventID, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

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

	if err := tx.QueryRowContext(ctx, query, time.Now()).Scan(&payload, &eventID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return events.OrderCreated{}, 0, nil
		}
		return createdOrder, eventID, fmt.Errorf("%s: %w", op, err)
	}

	if err := json.Unmarshal(payload, &createdOrder); err != nil {
		return createdOrder, eventID, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return createdOrder, eventID, fmt.Errorf("%s: %w", op, err)
	}

	return createdOrder, eventID, nil
}
