ALTER TABLE menu_items
    ADD COLUMN code TEXT NOT NULL DEFAULT '',
    ADD COLUMN unit TEXT NOT NULL DEFAULT '',
    ADD COLUMN prep_time_minutes INT NOT NULL DEFAULT 0,
    ADD COLUMN is_promo BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN is_best_seller BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN is_new BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN ingredients TEXT NOT NULL DEFAULT '',
    ADD COLUMN allergy_info TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_menu_items_code ON menu_items(code);
