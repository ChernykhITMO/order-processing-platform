CREATE TABLE IF NOT EXISTS payments
(
    order_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total_amount BIGINT NOT NULL,
    status TEXT DEFAULT 'pending'
        check (status in ('pending', 'succeeded', 'failed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
)