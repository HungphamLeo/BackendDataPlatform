package dto

import (
	"time"
)

// TradeDTO đại diện cho dữ liệu giao dịch được trả về cho client
type TradeDTO struct {
	ID           string    `json:"id"`           // ID của giao dịch
	Symbol       string    `json:"symbol"`       // Symbol (ví dụ: BTC/USDT)
	Exchange     string    `json:"exchange"`     // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	Price        float64   `json:"price"`        // Giá giao dịch
	Quantity     float64   `json:"quantity"`     // Khối lượng giao dịch
	QuoteQuantity float64  `json:"quoteQuantity"` // Khối lượng theo đồng báo giá
	Timestamp    time.Time `json:"timestamp"`    // Thời điểm giao dịch
	Side         string    `json:"side"`         // Phía (BUY hoặc SELL)
	IsBuyerMaker bool      `json:"isBuyerMaker"` // Người mua là maker hay không
}

// GetTradesRequest đại diện cho request lấy dữ liệu giao dịch
type GetTradesRequest struct {
	Symbol    string     `json:"symbol"`    // Symbol (ví dụ: BTC/USDT)
	Exchange  string     `json:"exchange"`  // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	StartTime *time.Time `json:"startTime"` // Thời điểm bắt đầu (tùy chọn)
	EndTime   *time.Time `json:"endTime"`   // Thời điểm kết thúc (tùy chọn)
	Limit     int        `json:"limit"`     // Số lượng giao dịch tối đa (tùy chọn)
}

// GetTradesResponse đại diện cho response trả về dữ liệu giao dịch
type GetTradesResponse struct {
	Trades []TradeDTO `json:"trades"` // Danh sách giao dịch
}

// GetHistoricalTradesRequest đại diện cho request lấy dữ liệu giao dịch lịch sử
type GetHistoricalTradesRequest struct {
	Symbol    string     `json:"symbol"`    // Symbol (ví dụ: BTC/USDT)
	Exchange  string     `json:"exchange"`  // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	FromID    string     `json:"fromId"`    // ID giao dịch bắt đầu (tùy chọn)
	StartTime *time.Time `json:"startTime"` // Thời điểm bắt đầu (tùy chọn)
	EndTime   *time.Time `json:"endTime"`   // Thời điểm kết thúc (tùy chọn)
	Limit     int        `json:"limit"`     // Số lượng giao dịch tối đa (tùy chọn)
}

// GetHistoricalTradesResponse đại diện cho response trả về dữ liệu giao dịch lịch sử
type GetHistoricalTradesResponse struct {
	Trades []TradeDTO `json:"trades"` // Danh sách giao dịch
}

// StreamTradesRequest đại diện cho request stream dữ liệu giao dịch
type StreamTradesRequest struct {
	Symbol   string `json:"symbol"`   // Symbol (ví dụ: BTC/USDT)
	Exchange string `json:"exchange"` // Sàn giao dịch (ví dụ: BINANCE_SPOT)
}

// StreamTradesResponse đại diện cho response stream dữ liệu giao dịch
type StreamTradesResponse struct {
	Trade TradeDTO `json:"trade"` // Dữ liệu giao dịch
}
