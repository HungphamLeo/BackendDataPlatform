CREATE TABLE IF NOT EXISTS historical_ohlcv (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol      VARCHAR(20)  NOT NULL,
    exchange    VARCHAR(20)  NOT NULL,
    `interval`  VARCHAR(10)  NOT NULL,
    open        DOUBLE       NOT NULL,
    high        DOUBLE       NOT NULL,
    low         DOUBLE       NOT NULL,
    close       DOUBLE       NOT NULL,
    volume      DOUBLE       NOT NULL,
    timestamp   DATETIME     NOT NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_ohlcv_sym_exch_intv_ts (symbol, exchange, `interval`, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
