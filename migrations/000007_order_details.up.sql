ALTER TABLE orders
    ADD COLUMN client_request_id TEXT UNIQUE,
    ADD COLUMN subtotal NUMERIC(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN vat_amount NUMERIC(12,2) NOT NULL DEFAULT 0;

ALTER TABLE order_items
    ADD COLUMN unit_price NUMERIC(12,2) NOT NULL DEFAULT 0;

CREATE TABLE order_item_options (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_item_id  UUID NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    option_id      UUID NOT NULL REFERENCES menu_item_options(id),
    name           TEXT NOT NULL,
    price_delta    NUMERIC(12,2) NOT NULL
);
CREATE INDEX idx_order_item_options_order_item ON order_item_options(order_item_id);
