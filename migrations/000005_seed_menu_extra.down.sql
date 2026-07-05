DELETE FROM menu_items WHERE code IN ('MC003', 'DU002');
UPDATE menu_items SET code = '', unit = '', prep_time_minutes = 0, is_best_seller = false, is_new = false, is_promo = false, ingredients = '', allergy_info = '' WHERE code IN ('MC001', 'MC002', 'DU001');
