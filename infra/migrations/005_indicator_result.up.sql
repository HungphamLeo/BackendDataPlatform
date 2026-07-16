CREATE TABLE IF NOT EXISTS indicator_result (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    symbol          VARCHAR(20)  NOT NULL,
    exchange        VARCHAR(20)  NOT NULL,
    `interval`      VARCHAR(10)  NOT NULL,
    indicator_name  VARCHAR(20)  NOT NULL,
    value           DOUBLE       NOT NULL,
    meta_json       VARCHAR(512) DEFAULT NULL,
    timestamp       DATETIME     NOT NULL,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ind_sym_exch_intv_name_ts (symbol, exchange, `interval`, indicator_name, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
