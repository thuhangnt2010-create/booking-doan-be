CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE restaurants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE branches (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id  UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tables (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id  UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    area       TEXT NOT NULL DEFAULT '',
    code       TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'ready' CHECK (status IN ('ready', 'serving', 'locked', 'maintenance', 'paying')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (branch_id, code)
);

CREATE TABLE qr_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id   UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_qr_codes_table_active ON qr_codes(table_id) WHERE active = true;

CREATE TABLE sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id   UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'payment_requested', 'closed')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at   TIMESTAMPTZ
);
CREATE INDEX idx_sessions_table_status ON sessions(table_id, status);

CREATE TABLE menu_categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id  UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    position   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE menu_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES menu_categories(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    price       NUMERIC(12,2) NOT NULL,
    status      TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'low_stock', 'out_of_stock', 'suspended')),
    image_key   TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_menu_items_category ON menu_items(category_id);

CREATE TABLE menu_item_options (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id     UUID NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    name        TEXT NOT NULL,
    price_delta NUMERIC(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE orders (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    code       TEXT NOT NULL UNIQUE,
    status     TEXT NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'received', 'cooking', 'done', 'served', 'cancelled')),
    total      NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_orders_session ON orders(session_id);

CREATE TABLE order_items (
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id  UUID NOT NULL REFERENCES menu_items(id),
    qty      INT NOT NULL CHECK (qty > 0),
    note     TEXT NOT NULL DEFAULT '',
    status   TEXT NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'received', 'cooking', 'done', 'served', 'cancelled'))
);
CREATE INDEX idx_order_items_order ON order_items(order_id);

CREATE TABLE staff_call_requests (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    type       TEXT NOT NULL DEFAULT 'other',
    status     TEXT NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'received', 'processing', 'done')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_staff_calls_session ON staff_call_requests(session_id);

CREATE TABLE payment_requests (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id    UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    status        TEXT NOT NULL DEFAULT 'requested' CHECK (status IN ('requested', 'confirmed', 'cancelled')),
    requested_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at  TIMESTAMPTZ
);
CREATE INDEX idx_payment_requests_session ON payment_requests(session_id);

CREATE TABLE admin_users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'staff' CHECK (role IN ('admin', 'staff', 'kitchen')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
