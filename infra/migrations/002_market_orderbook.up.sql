CREATE TABLE IF NOT EXISTS market_orderbook (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol      VARCHAR(20)  NOT NULL,
    exchange    VARCHAR(20)  NOT NULL,
    bids_json   MEDIUMTEXT,
    asks_json   MEDIUMTEXT,
    timestamp   DATETIME(3)  NOT NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ob_sym_exch_ts (symbol, exchange, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
