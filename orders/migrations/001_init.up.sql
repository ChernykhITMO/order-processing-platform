CREATE TABLE IF NOT EXISTS orders
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT,
    status     TEXT      NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items
(
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity   INT    NOT NULL CHECK (quantity >= 1),
    price      BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS events
(
    id         SERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload    JSONB NOT NULL,
    aggregate_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    locked_at TIMESTAMPTZ DEFAULT NULL,
    sent_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_events_unsent ON events (sent_at) WHERE sent_at IS NULL;
CREATE INDEX idx_events_unlocked ON events (locked_at) WHERE locked_at IS NULL;
