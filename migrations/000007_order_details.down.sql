DROP TABLE IF EXISTS order_item_options;
ALTER TABLE order_items DROP COLUMN unit_price;
ALTER TABLE orders
    DROP COLUMN client_request_id,
    DROP COLUMN subtotal,
    DROP COLUMN vat_amount;
