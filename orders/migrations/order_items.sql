CREATE TABLE order_items
(
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity   INT    NOT NULL CHECK (quantity >= 1),
    price NUMERIC NOT NULL CHECK (price > 0)
);