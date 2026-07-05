INSERT INTO menu_item_options (item_id, type, name, price_delta)
SELECT id, 'size', 'Nhỏ', 0 FROM menu_items WHERE code = 'MC001'
UNION ALL
SELECT id, 'size', 'Lớn', 10000 FROM menu_items WHERE code = 'MC001'
UNION ALL
SELECT id, 'spice', 'Cay vừa', 0 FROM menu_items WHERE code = 'MC001'
UNION ALL
SELECT id, 'spice', 'Cay nhiều', 0 FROM menu_items WHERE code = 'MC001';
