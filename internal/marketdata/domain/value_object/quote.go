package value_object

import (
	"fmt"
	"time"
)

// Quote đại diện cho một báo giá bid/ask
type Quote struct {
	Symbol    Symbol    // Symbol của báo giá
	Bid       float64   // Giá mua (bid)
	Ask       float64   // Giá bán (ask)
	BidSize   float64   // Khối lượng ở giá mua
	AskSize   float64   // Khối lượng ở giá bán
	Timestamp time.Time // Thời điểm báo giá
	Source    string    // Nguồn báo giá (BINANCE_SPOT, BINANCE_FUTURES, ...)
}

// NewQuote tạo một Quote mới
func NewQuote(symbol Symbol, bid, ask, bidSize, askSize float64, timestamp time.Time, source string) Quote {
	return Quote{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		BidSize:   bidSize,
		AskSize:   askSize,
		Timestamp: timestamp,
		Source:    source,
	}
}

// Validate kiểm tra tính hợp lệ của Quote
func (q Quote) Validate() error {
	if q.Bid <= 0 {
		return fmt.Errorf("bid price must be positive")
	}
	if q.Ask <= 0 {
		return fmt.Errorf("ask price must be positive")
	}
	if q.Ask < q.Bid {
		return fmt.Errorf("ask price cannot be less than bid price")
	}
	if q.BidSize < 0 {
		return fmt.Errorf("bid size cannot be negative")
	}
	if q.AskSize < 0 {
		return fmt.Errorf("ask size cannot be negative")
	}
	if q.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}
	if q.Source == "" {
		return fmt.Errorf("source cannot be empty")
	}
	return nil
}

// MidPrice trả về giá trung bình giữa bid và ask
func (q Quote) MidPrice() float64 {
	return (q.Bid + q.Ask) / 2
}

// Spread trả về chênh lệch giữa ask và bid
func (q Quote) Spread() float64 {
	return q.Ask - q.Bid
}

// SpreadPercentage trả về chênh lệch dưới dạng phần trăm
func (q Quote) SpreadPercentage() float64 {
	if q.Bid == 0 {
		return 0
	}
	return (q.Ask - q.Bid) / q.Bid * 100
}

// IsStale kiểm tra xem báo giá có cũ không (quá 5 phút)
func (q Quote) IsStale() bool {
	return time.Since(q.Timestamp) > 5*time.Minute
}

// String trả về biểu diễn chuỗi của Quote
func (q Quote) String() string {
	return fmt.Sprintf("%s: Bid=%f (Size=%f), Ask=%f (Size=%f), Time=%s, Source=%s",
		q.Symbol.String(), q.Bid, q.BidSize, q.Ask, q.AskSize, q.Timestamp.Format(time.RFC3339), q.Source)
}
