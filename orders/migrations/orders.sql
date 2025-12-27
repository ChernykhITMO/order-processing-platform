CREATE TABLE orders
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT,
    status     TEXT      NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);