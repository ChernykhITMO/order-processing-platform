package postgres

import "context"

func (s *TxStorage) SaveEvent(ctx context.Context, eventType string, payload []byte, aggregateID int64) error {
	const query = `
          INSERT INTO events (event_type, payload, aggregate_id)
          VALUES ($1, $2, $3)
      `
	_, err := s.tx.ExecContext(ctx, query, eventType, payload, aggregateID)
	return err
}
