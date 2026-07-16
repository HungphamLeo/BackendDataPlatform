package models

import (
	"time"
)

// MarketTicker represents a real-time ticker snapshot.
type MarketTicker struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Symbol    string    `gorm:"index:idx_ticker_sym_exch_ts;not null"`
	Exchange  string    `gorm:"index:idx_ticker_sym_exch_ts;not null"`
	Price     float64   `gorm:"not null"`
	Volume    float64
	Bid       float64
	Ask       float64
	Timestamp time.Time `gorm:"index:idx_ticker_sym_exch_ts;not null"`
	CreatedAt time.Time
}

func (MarketTicker) TableName() string { return "market_ticker" }

// MarketOrderBook persists a depth snapshot as JSON blobs.
type MarketOrderBook struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Symbol    string    `gorm:"index;not null"`
	Exchange  string    `gorm:"index;not null"`
	BidsJSON  string    `gorm:"type:text"`
	AsksJSON  string    `gorm:"type:text"`
	Timestamp time.Time `gorm:"index;not null"`
	CreatedAt time.Time
}

func (MarketOrderBook) TableName() string { return "market_orderbook" }

// MarketKline represents a completed candlestick bar.
type MarketKline struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Symbol    string    `gorm:"index:idx_kline;not null"`
	Exchange  string    `gorm:"index:idx_kline;not null"`
	Interval  string    `gorm:"index:idx_kline;not null"`
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Timestamp time.Time `gorm:"index:idx_kline;not null"`
	CreatedAt time.Time
}

func (MarketKline) TableName() string { return "market_kline" }
