CREATE TABLE IF NOT EXISTS market_kline (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol      VARCHAR(20)  NOT NULL,
    exchange    VARCHAR(20)  NOT NULL,
    `interval`  VARCHAR(10)  NOT NULL,
    open        DOUBLE       NOT NULL,
    high        DOUBLE       NOT NULL,
    low         DOUBLE       NOT NULL,
    close       DOUBLE       NOT NULL,
    volume      DOUBLE       NOT NULL,
    timestamp   DATETIME(3)  NOT NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_kline (symbol, exchange, `interval`, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
