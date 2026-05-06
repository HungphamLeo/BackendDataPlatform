-- Tạo table accounts
CREATE TABLE IF NOT EXISTS accounts (
    id                  VARCHAR(36) PRIMARY KEY,
    user_id             VARCHAR(36) UNIQUE NOT NULL,
    balance             BIGINT NOT NULL DEFAULT 0,            -- tính theo cent
    reserved_balance    BIGINT NOT NULL DEFAULT 0,            -- balance bị giữ
    available_balance   BIGINT NOT NULL DEFAULT 0,            -- balance có thể dùng
    status              VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version             INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_status ON accounts(status);

-- Constraint: available = balance - reserved
-- (MySQL: CHECK (available_balance = balance - reserved_balance))