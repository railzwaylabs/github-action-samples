BEGIN;

CREATE TABLE orders (
    id          UUID PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL,
    deleted_at  TIMESTAMPTZ,
    customer_id VARCHAR(255) NOT NULL,
    order_id    UUID NOT NULL UNIQUE,
    status      SMALLINT NOT NULL DEFAULT 1
);

CREATE INDEX idx_orders_customer_id ON orders (customer_id);
CREATE INDEX idx_orders_deleted_at ON orders (deleted_at);

CREATE TABLE order_items (
    id          UUID PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL,
    deleted_at  TIMESTAMPTZ,
    order_id    UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id  VARCHAR(255) NOT NULL,
    quantity    BIGINT NOT NULL CHECK (quantity > 0),
    price       BIGINT NOT NULL CHECK (price >= 0),
    subtotal    BIGINT NOT NULL CHECK (subtotal >= 0),
    CONSTRAINT order_items_subtotal_consistent CHECK (subtotal = quantity * price)
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);
CREATE INDEX idx_order_items_deleted_at ON order_items (deleted_at);

COMMIT;
