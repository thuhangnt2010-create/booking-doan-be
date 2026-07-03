INSERT INTO restaurants (id, name) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Nhà hàng Demo');

INSERT INTO branches (id, restaurant_id, name) VALUES
    ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Chi nhánh Quận 1');

INSERT INTO tables (id, branch_id, area, code, status) VALUES
    ('33333333-3333-3333-3333-333333333331', '22222222-2222-2222-2222-222222222222', 'Tầng 1', 'B01', 'ready'),
    ('33333333-3333-3333-3333-333333333332', '22222222-2222-2222-2222-222222222222', 'Tầng 1', 'B02', 'ready');

INSERT INTO qr_codes (table_id, token, active) VALUES
    ('33333333-3333-3333-3333-333333333331', 'demo-qr-token-b01', true),
    ('33333333-3333-3333-3333-333333333332', 'demo-qr-token-b02', true);

INSERT INTO menu_categories (id, branch_id, name, position) VALUES
    ('44444444-4444-4444-4444-444444444441', '22222222-2222-2222-2222-222222222222', 'Món chính', 1),
    ('44444444-4444-4444-4444-444444444442', '22222222-2222-2222-2222-222222222222', 'Đồ uống', 2);

INSERT INTO menu_items (category_id, name, price, status, description) VALUES
    ('44444444-4444-4444-4444-444444444441', 'Phở bò', 55000, 'available', 'Phở bò truyền thống'),
    ('44444444-4444-4444-4444-444444444441', 'Cơm gà', 45000, 'available', 'Cơm gà xối mỡ'),
    ('44444444-4444-4444-4444-444444444442', 'Trà đá', 5000, 'available', 'Trà đá miễn phí');
