CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255) NOT NULL,
    entry VARCHAR(50) NOT NULL,
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255) NOT NULL,
    delivery_service VARCHAR(100) NOT NULL,
    shardkey VARCHAR(50) NOT NULL,
    sm_id INTEGER NOT NULL CHECK (sm_id >= 0),
    date_created TIMESTAMPTZ NOT NULL,
    oof_shard VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS deliveries (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    zip VARCHAR(50) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    region VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS payments (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction VARCHAR(255) NOT NULL,
    request_id VARCHAR(255),
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    provider VARCHAR(100) NOT NULL,
    amount INTEGER NOT NULL CHECK (amount >= 0),
    payment_dt BIGINT NOT NULL CHECK (payment_dt >= 0),
    bank VARCHAR(100) NOT NULL,
    delivery_cost INTEGER NOT NULL CHECK (delivery_cost >= 0),
    goods_total INTEGER NOT NULL CHECK (goods_total >= 0),
    custom_fee INTEGER NOT NULL DEFAULT 0 CHECK (custom_fee >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id INTEGER NOT NULL CHECK (chrt_id >= 0),
    track_number VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    rid VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sale INTEGER NOT NULL DEFAULT 0 CHECK (sale >= 0),
    size VARCHAR(50) NOT NULL,
    total_price INTEGER NOT NULL CHECK (total_price >= 0),
    nm_id INTEGER NOT NULL CHECK (nm_id >= 0),
    brand VARCHAR(255) NOT NULL,
    status INTEGER NOT NULL CHECK (status >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_orders_track_number ON orders(track_number);
CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders(date_created);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_phone ON deliveries(phone);
CREATE INDEX IF NOT EXISTS idx_deliveries_email ON deliveries(email);
CREATE INDEX IF NOT EXISTS idx_payments_transaction ON payments(transaction);
CREATE INDEX IF NOT EXISTS idx_items_chrt_id ON items(chrt_id);
CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid);
CREATE INDEX IF NOT EXISTS idx_items_nm_id ON items(nm_id);

