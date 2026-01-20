package postgres

import (
	"context"
	"fmt"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/ports"
)

const rowsInserted = 1

func (s *TxStorage) TryMarkProcessed(ctx context.Context, eventId int64) (bool, error) {
	const op = "storage.postgres.TryMarkProcessed"

	const query = `
		INSERT INTO processed_events (event_id, processed_at)
		VALUES ($1, NOW())
		ON CONFLICT (event_id) DO NOTHING;
	`

	res, err := s.tx.ExecContext(ctx, query, eventId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return n == rowsInserted, nil
}

func (s *Storage) RunInTx(ctx context.Context, fn func(tx ports.StorageTx) error) error {
	const op = "storage.postgres.RunInTx"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	txStorage := &TxStorage{tx: tx}

	if err := fn(txStorage); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
