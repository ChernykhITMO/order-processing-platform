CREATE TABLE order_items
(
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity   INT    NOT NULL CHECK (quantity >= 1),
    price      BIGINT
);

INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (1, 1, 1, 1);