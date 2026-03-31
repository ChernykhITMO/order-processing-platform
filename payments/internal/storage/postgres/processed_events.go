package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const rowsInserted = 1

func (s *TxStorage) TryMarkProcessed(ctx context.Context, eventId int64) (bool, error) {
	const op = "storage.postgres.TryMarkProcessed"

	const query = `
		INSERT INTO processed_events (event_id, processed_at)
		VALUES ($1, NOW())
		ON CONFLICT (event_id) DO NOTHING;
	`

	res, err := s.tx.Exec(ctx, query, eventId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return res.RowsAffected() == rowsInserted, nil
}

func (s *Storage) RunInTx(ctx context.Context, fn func(tx TxRepository) error) error {
	const op = "storage.postgres.RunInTx"

	return s.txManager.WithinTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		txStorage := &TxStorage{tx: tx}
		if err := fn(txStorage); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
}
