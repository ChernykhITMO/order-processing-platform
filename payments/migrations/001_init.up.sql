CREATE TABLE IF NOT EXISTS payments
(
    order_id     BIGINT PRIMARY KEY,
    user_id      BIGINT      NOT NULL,
    total_amount BIGINT      NOT NULL,
    status       TEXT                 DEFAULT 'pending'
        check (status in ('pending', 'succeeded', 'failed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS events
(
    id           SERIAL PRIMARY KEY,
    event_type   TEXT   NOT NULL,
    payload      JSONB  NOT NULL,
    aggregate_id BIGINT NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    locked_at    TIMESTAMPTZ DEFAULT NULL,
    sent_at      TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_events_unsent ON events (sent_at) WHERE sent_at IS NULL;
CREATE INDEX idx_events_unlocked ON events (locked_at) WHERE locked_at IS NULL;