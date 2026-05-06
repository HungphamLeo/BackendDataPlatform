-- Tạo table orders
CREATE TABLE IF NOT EXISTS orders (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL,
    symbol          VARCHAR(10) NOT NULL,
    side            VARCHAR(10) NOT NULL,  -- BUY, SELL
    price           BIGINT NOT NULL,       -- tính theo cent
    quantity        DOUBLE PRECISION NOT NULL,
    status          VARCHAR(20) NOT NULL,  -- PENDING, FILLED, ...
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version         INTEGER NOT NULL DEFAULT 1,  -- optimistic locking
    
    -- Indexes để tăng tốc độ query
    CONSTRAINT orders_user_symbol_check CHECK (symbol != '')
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

-- Tạo table outbox (cho event publishing)
CREATE TABLE IF NOT EXISTS outbox (
    id              VARCHAR(36) PRIMARY KEY,
    topic           VARCHAR(100) NOT NULL,
    payload         BYTEA NOT NULL,         -- protobuf bytes
    event_type      VARCHAR(100) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at    TIMESTAMP,
    failed_at       TIMESTAMP,
    error           TEXT,
    retries         INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT outbox_topic_check CHECK (topic != '')
);

CREATE INDEX idx_outbox_published_at ON outbox(published_at) WHERE published_at IS NULL;
CREATE INDEX idx_outbox_created_at ON outbox(created_at);