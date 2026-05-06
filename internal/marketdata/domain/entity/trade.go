package entity

import (
	"fmt"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// Trade đại diện cho một giao dịch
type Trade struct {
	ID           string               // ID của giao dịch
	Symbol       value_object.Symbol  // Symbol của giao dịch
	Price        float64              // Giá giao dịch
	Quantity     float64              // Khối lượng giao dịch
	QuoteQuantity float64             // Khối lượng theo đồng báo giá (price * quantity)
	Timestamp    time.Time            // Thời điểm giao dịch
	IsBuyerMaker bool                 // Người mua là maker (true) hay taker (false)
	IsBestMatch  bool                 // Giao dịch có phải là best match không
	Source       string               // Nguồn dữ liệu (BINANCE_SPOT, BINANCE_FUTURES, ...)
}

// NewTrade tạo một Trade mới
func NewTrade(
	id string,
	symbol value_object.Symbol,
	price, quantity float64,
	timestamp time.Time,
	isBuyerMaker, isBestMatch bool,
	source string,
) *Trade {
	quoteQuantity := price * quantity
	
	return &Trade{
		ID:            id,
		Symbol:        symbol,
		Price:         price,
		Quantity:      quantity,
		QuoteQuantity: quoteQuantity,
		Timestamp:     timestamp,
		IsBuyerMaker:  isBuyerMaker,
		IsBestMatch:   isBestMatch,
		Source:        source,
	}
}

// Validate kiểm tra tính hợp lệ của giao dịch
func (t *Trade) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("trade ID cannot be empty")
	}
	if t.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if t.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if t.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}
	if t.Source == "" {
		return fmt.Errorf("source cannot be empty")
	}
	return nil
}

// IsBuy kiểm tra xem giao dịch có phải là mua không
func (t *Trade) IsBuy() bool {
	// Trong Binance, nếu isBuyerMaker = true, thì người mua là maker
	// Điều này có nghĩa là người bán là taker, và giao dịch được xem là "sell"
	// Ngược lại, nếu isBuyerMaker = false, thì người mua là taker, và giao dịch được xem là "buy"
	return !t.IsBuyerMaker
}

// IsSell kiểm tra xem giao dịch có phải là bán không
func (t *Trade) IsSell() bool {
	return t.IsBuyerMaker
}

// Side trả về phía của giao dịch (BUY hoặc SELL)
func (t *Trade) Side() string {
	if t.IsBuy() {
		return "BUY"
	}
	return "SELL"
}

// String trả về biểu diễn chuỗi của Trade
func (t *Trade) String() string {
	return fmt.Sprintf("Trade %s: %s %s %.8f @ %.8f = %.8f, Time=%s, Maker=%t, Source=%s",
		t.ID, t.Symbol.String(), t.Side(), t.Quantity, t.Price, t.QuoteQuantity,
		t.Timestamp.Format(time.RFC3339), t.IsBuyerMaker, t.Source)
}
