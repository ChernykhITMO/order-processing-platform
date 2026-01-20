package postgres

import (
	"context"
	"fmt"
	"time"
)

func (s *Storage) MarkSent(ctx context.Context, eventID int64) error {
	const op = "storage.postgres.MarkSent"
	const query = `UPDATE events SET sent_at = $1, locked_at = NULL WHERE id = $2`

	currentTime := time.Now()
	if _, err := s.db.ExecContext(ctx, query, currentTime, eventID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
