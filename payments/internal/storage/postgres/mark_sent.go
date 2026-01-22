package postgres

import (
	"context"
	"fmt"
	"time"
)

func (s *Storage) MarkSent(ctx context.Context, id int64) error {
	const op = "storage.postgres.MarkSent"

	const query = `UPDATE events SET sent_at = $1, locked_at = NULL WHERE id = $2`

	if _, err := s.db.ExecContext(ctx, query, time.Now(), id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
