package dto

import (
	"time"
)

// TickerDTO đại diện cho dữ liệu ticker được trả về cho client
type TickerDTO struct {
	Symbol      string    `json:"symbol"`      // Symbol (ví dụ: BTC/USDT)
	Exchange    string    `json:"exchange"`    // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	Price       float64   `json:"price"`       // Giá gần nhất
	Bid         float64   `json:"bid"`         // Giá mua tốt nhất
	Ask         float64   `json:"ask"`         // Giá bán tốt nhất
	BidSize     float64   `json:"bidSize"`     // Khối lượng ở giá mua tốt nhất
	AskSize     float64   `json:"askSize"`     // Khối lượng ở giá bán tốt nhất
	High24h     float64   `json:"high24h"`     // Giá cao nhất trong 24h
	Low24h      float64   `json:"low24h"`      // Giá thấp nhất trong 24h
	Volume24h   float64   `json:"volume24h"`   // Khối lượng trong 24h
	QuoteVolume float64   `json:"quoteVolume"` // Khối lượng theo đồng báo giá trong 24h
	Timestamp   time.Time `json:"timestamp"`   // Thời điểm cập nhật
}

// GetTickerRequest đại diện cho request lấy dữ liệu ticker
type GetTickerRequest struct {
	Symbol   string `json:"symbol"`   // Symbol (ví dụ: BTC/USDT)
	Exchange string `json:"exchange"` // Sàn giao dịch (ví dụ: BINANCE_SPOT)
}

// GetTickerResponse đại diện cho response trả về dữ liệu ticker
type GetTickerResponse struct {
	Ticker TickerDTO `json:"ticker"` // Dữ liệu ticker
}

// GetTickersRequest đại diện cho request lấy dữ liệu ticker cho nhiều symbol
type GetTickersRequest struct {
	Symbols  []string `json:"symbols"`  // Danh sách symbol
	Exchange string   `json:"exchange"` // Sàn giao dịch
}

// GetTickersResponse đại diện cho response trả về dữ liệu ticker cho nhiều symbol
type GetTickersResponse struct {
	Tickers []TickerDTO `json:"tickers"` // Danh sách dữ liệu ticker
}
