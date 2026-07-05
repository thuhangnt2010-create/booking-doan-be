DELETE FROM menu_item_options WHERE item_id IN (SELECT id FROM menu_items WHERE code = 'MC001');
