CREATE TABLE IF NOT EXISTS market_ticker (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol      VARCHAR(20)  NOT NULL,
    exchange    VARCHAR(20)  NOT NULL,
    price       DOUBLE       NOT NULL,
    volume      DOUBLE       NOT NULL DEFAULT 0,
    bid         DOUBLE       NOT NULL DEFAULT 0,
    ask         DOUBLE       NOT NULL DEFAULT 0,
    timestamp   DATETIME(3)  NOT NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ticker_sym_exch_ts (symbol, exchange, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
