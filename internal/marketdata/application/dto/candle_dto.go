package dto

import (
	"time"
)

// CandleDTO đại diện cho dữ liệu nến được trả về cho client
type CandleDTO struct {
	Symbol      string    `json:"symbol"`      // Symbol (ví dụ: BTC/USDT)
	Exchange    string    `json:"exchange"`    // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	Timeframe   string    `json:"timeframe"`   // Khung thời gian (ví dụ: 1m, 5m, 1h)
	OpenTime    time.Time `json:"openTime"`    // Thời điểm mở nến
	CloseTime   time.Time `json:"closeTime"`   // Thời điểm đóng nến
	Open        float64   `json:"open"`        // Giá mở
	High        float64   `json:"high"`        // Giá cao nhất
	Low         float64   `json:"low"`         // Giá thấp nhất
	Close       float64   `json:"close"`       // Giá đóng
	Volume      float64   `json:"volume"`      // Khối lượng
	QuoteVolume float64   `json:"quoteVolume"` // Khối lượng theo đồng báo giá
	TradeCount  int64     `json:"tradeCount"`  // Số lượng giao dịch
	Closed      bool      `json:"closed"`      // Nến đã đóng hay chưa
}

// GetCandlesRequest đại diện cho request lấy dữ liệu nến
type GetCandlesRequest struct {
	Symbol    string     `json:"symbol"`    // Symbol (ví dụ: BTC/USDT)
	Exchange  string     `json:"exchange"`  // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	Timeframe string     `json:"timeframe"` // Khung thời gian (ví dụ: 1m, 5m, 1h)
	StartTime *time.Time `json:"startTime"` // Thời điểm bắt đầu (tùy chọn)
	EndTime   *time.Time `json:"endTime"`   // Thời điểm kết thúc (tùy chọn)
	Limit     int        `json:"limit"`     // Số lượng nến tối đa (tùy chọn)
}

// GetCandlesResponse đại diện cho response trả về dữ liệu nến
type GetCandlesResponse struct {
	Candles []CandleDTO `json:"candles"` // Danh sách nến
}

// StreamCandlesRequest đại diện cho request stream dữ liệu nến
type StreamCandlesRequest struct {
	Symbol     string   `json:"symbol"`     // Symbol (ví dụ: BTC/USDT)
	Exchange   string   `json:"exchange"`   // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	Timeframes []string `json:"timeframes"` // Danh sách khung thời gian (ví dụ: 1m, 5m, 1h)
}

// StreamCandlesResponse đại diện cho response stream dữ liệu nến
type StreamCandlesResponse struct {
	Candle CandleDTO `json:"candle"` // Dữ liệu nến
}
