UPDATE menu_items SET code = 'MC001', unit = 'Tô', prep_time_minutes = 15, is_best_seller = true, ingredients = 'Bánh phở, thịt bò, hành, rau thơm', allergy_info = 'Có thể chứa gluten' WHERE name = 'Phở bò';
UPDATE menu_items SET code = 'MC002', unit = 'Đĩa', prep_time_minutes = 12, is_new = true, ingredients = 'Cơm, gà, dưa leo, cà chua', allergy_info = '' WHERE name = 'Cơm gà';
UPDATE menu_items SET code = 'DU001', unit = 'Ly', prep_time_minutes = 2, is_promo = true, ingredients = 'Trà, đá', allergy_info = '' WHERE name = 'Trà đá';

INSERT INTO menu_items (category_id, code, name, price, status, unit, prep_time_minutes, is_promo, is_best_seller, is_new, description, ingredients, allergy_info) VALUES
    ('44444444-4444-4444-4444-444444444441', 'MC003', 'Bún chả', 65000, 'available', 'Phần', 18, false, true, false, 'Bún chả Hà Nội', 'Bún, chả nướng, nước mắm chua ngọt', 'Có thể chứa đậu phộng'),
    ('44444444-4444-4444-4444-444444444442', 'DU002', 'Nước cam ép', 25000, 'low_stock', 'Ly', 5, true, false, true, 'Cam tươi ép nguyên chất', 'Cam', '');
