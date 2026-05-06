package dto

import (
	"time"
)

// OrderBookEntryDTO đại diện cho một mức giá trong sổ lệnh
type OrderBookEntryDTO struct {
	Price    float64 `json:"price"`    // Giá
	Quantity float64 `json:"quantity"` // Khối lượng
}

// OrderBookDTO đại diện cho dữ liệu sổ lệnh được trả về cho client
type OrderBookDTO struct {
	Symbol       string             `json:"symbol"`       // Symbol (ví dụ: BTC/USDT)
	Exchange     string             `json:"exchange"`     // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	Bids         []OrderBookEntryDTO `json:"bids"`        // Danh sách giá mua
	Asks         []OrderBookEntryDTO `json:"asks"`        // Danh sách giá bán
	LastUpdateID int64              `json:"lastUpdateId"` // ID cập nhật cuối cùng
	Timestamp    time.Time          `json:"timestamp"`    // Thời điểm cập nhật
}

// GetOrderBookRequest đại diện cho request lấy dữ liệu sổ lệnh
type GetOrderBookRequest struct {
	Symbol   string `json:"symbol"`   // Symbol (ví dụ: BTC/USDT)
	Exchange string `json:"exchange"` // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	Depth    int    `json:"depth"`    // Độ sâu của sổ lệnh (tùy chọn)
}

// GetOrderBookResponse đại diện cho response trả về dữ liệu sổ lệnh
type GetOrderBookResponse struct {
	OrderBook OrderBookDTO `json:"orderBook"` // Dữ liệu sổ lệnh
}

// StreamOrderBookRequest đại diện cho request stream dữ liệu sổ lệnh
type StreamOrderBookRequest struct {
	Symbol   string `json:"symbol"`   // Symbol (ví dụ: BTC/USDT)
	Exchange string `json:"exchange"` // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	Depth    int    `json:"depth"`    // Độ sâu của sổ lệnh (tùy chọn)
}

// StreamOrderBookResponse đại diện cho response stream dữ liệu sổ lệnh
type StreamOrderBookResponse struct {
	OrderBook OrderBookDTO `json:"orderBook"` // Dữ liệu sổ lệnh
}

// OrderBookDiffDTO đại diện cho dữ liệu thay đổi sổ lệnh
type OrderBookDiffDTO struct {
	Symbol       string             `json:"symbol"`       // Symbol (ví dụ: BTC/USDT)
	Exchange     string             `json:"exchange"`     // Sàn giao dịch (ví dụ: BINANCE_SPOT)
	FirstUpdateID int64             `json:"firstUpdateId"` // ID cập nhật đầu tiên
	LastUpdateID int64              `json:"lastUpdateId"` // ID cập nhật cuối cùng
	Bids         []OrderBookEntryDTO `json:"bids"`        // Danh sách giá mua thay đổi
	Asks         []OrderBookEntryDTO `json:"asks"`        // Danh sách giá bán thay đổi
	Timestamp    time.Time          `json:"timestamp"`    // Thời điểm cập nhật
}

// StreamOrderBookDiffResponse đại diện cho response stream dữ liệu thay đổi sổ lệnh
type StreamOrderBookDiffResponse struct {
	OrderBookDiff OrderBookDiffDTO `json:"orderBookDiff"` // Dữ liệu thay đổi sổ lệnh
}
