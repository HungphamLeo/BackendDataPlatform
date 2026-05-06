package value_object

import (
	"fmt"
	"strings"
)

// Symbol đại diện cho một cặp giao dịch đã được chuẩn hóa
// Ví dụ: BTC/USDT, ETH/USD, EUR/USD
type Symbol struct {
	Base  string // Đồng cơ sở (BTC, ETH, EUR)
	Quote string // Đồng báo giá (USDT, USD)
}

// NewSymbol tạo một Symbol mới từ base và quote
func NewSymbol(base, quote string) Symbol {
	return Symbol{
		Base:  strings.ToUpper(base),
		Quote: strings.ToUpper(quote),
	}
}

// ParseSymbol phân tích một chuỗi thành Symbol
// Hỗ trợ nhiều định dạng: "BTC/USDT", "BTCUSDT", "BTC-USDT"
func ParseSymbol(s string) (Symbol, error) {
	s = strings.ToUpper(s)
	
	// Kiểm tra định dạng "BTC/USDT"
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) != 2 {
			return Symbol{}, fmt.Errorf("invalid symbol format: %s", s)
		}
		return NewSymbol(parts[0], parts[1]), nil
	}
	
	// Kiểm tra định dạng "BTC-USDT"
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) != 2 {
			return Symbol{}, fmt.Errorf("invalid symbol format: %s", s)
		}
		return NewSymbol(parts[0], parts[1]), nil
	}
	
	// Xử lý định dạng "BTCUSDT" - cần phân tách base và quote
	// Danh sách các quote phổ biến
	commonQuotes := []string{"USDT", "USD", "BTC", "ETH", "BNB", "BUSD", "DAI", "TUSD", "USDC", "EUR", "GBP", "JPY"}
	
	for _, quote := range commonQuotes {
		if strings.HasSuffix(s, quote) {
			base := strings.TrimSuffix(s, quote)
			if base != "" {
				return NewSymbol(base, quote), nil
			}
		}
	}
	
	return Symbol{}, fmt.Errorf("unable to parse symbol: %s", s)
}

// String trả về biểu diễn chuỗi của Symbol theo định dạng chuẩn "BTC/USDT"
func (s Symbol) String() string {
	return fmt.Sprintf("%s/%s", s.Base, s.Quote)
}

// ToExchangeFormat chuyển đổi Symbol sang định dạng của một sàn cụ thể
func (s Symbol) ToExchangeFormat(exchange string) string {
	switch strings.ToUpper(exchange) {
	case "BINANCE", "BINANCE_SPOT", "BINANCE_FUTURES":
		return s.Base + s.Quote // BTCUSDT
	case "FTX":
		return s.Base + "/" + s.Quote // BTC/USDT
	case "KRAKEN":
		// Kraken sử dụng "XXBT" cho BTC và "ZUSD" cho USD
		base := s.Base
		quote := s.Quote
		if base == "BTC" {
			base = "XXBT"
		}
		if quote == "USD" {
			quote = "ZUSD"
		}
		return base + quote // XXBTZUSD
	default:
		return s.String() // BTC/USDT
	}
}

// Equals kiểm tra hai Symbol có bằng nhau không
func (s Symbol) Equals(other Symbol) bool {
	return s.Base == other.Base && s.Quote == other.Quote
}
