package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type TxStorage struct {
	tx *sql.Tx
}

func (s *TxStorage) UpdatePaymentStatus(ctx context.Context, orderID int64, status string) error {
	const op = "storage.postgres.UpdatePaymentStatus"

	const query = `UPDATE payments SET status = $1 WHERE order_id = $2;`

	if _, err := s.tx.ExecContext(ctx, query, status, orderID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *TxStorage) UpsertPayment(ctx context.Context, orderID, userID, totalAmount int64, status string) error {
	const op = "storage.postgres.UpsertPayment"

	const query = `
		INSERT INTO payments (order_id, user_id, total_amount, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (order_id)
		DO UPDATE SET user_id = EXCLUDED.user_id,
		              total_amount = EXCLUDED.total_amount,
		              status = EXCLUDED.status;
	`

	if _, err := s.tx.ExecContext(ctx, query, orderID, userID, totalAmount, status); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
