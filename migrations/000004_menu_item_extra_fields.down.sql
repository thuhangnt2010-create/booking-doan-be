DROP INDEX IF EXISTS idx_menu_items_code;
ALTER TABLE menu_items
    DROP COLUMN code,
    DROP COLUMN unit,
    DROP COLUMN prep_time_minutes,
    DROP COLUMN is_promo,
    DROP COLUMN is_best_seller,
    DROP COLUMN is_new,
    DROP COLUMN ingredients,
    DROP COLUMN allergy_info;
