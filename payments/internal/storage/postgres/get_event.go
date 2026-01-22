package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/domain/events"
)

func (s *Storage) GetNewEvent(ctx context.Context) (events.PaymentStatus, error) {
	const op = "storage.postgres.GetNewEvent"

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

	var (
		payment events.PaymentStatus
		payload []byte
		id      int64
	)

	if err := s.db.QueryRowContext(ctx, query, time.Now()).Scan(&payload, &id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return events.PaymentStatus{}, nil
		}
		return events.PaymentStatus{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := json.Unmarshal(payload, &payment); err != nil {
		return payment, fmt.Errorf("%s: %w", op, err)
	}

	payment.EventID = id

	return payment, nil
}
