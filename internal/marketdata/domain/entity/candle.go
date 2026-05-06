package entity

import (
	"fmt"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// Candle đại diện cho một nến OHLCV (Open, High, Low, Close, Volume)
type Candle struct {
	Symbol     value_object.Symbol     // Symbol của nến
	Timeframe  value_object.Timeframe  // Khung thời gian của nến
	OpenTime   time.Time               // Thời điểm mở nến
	CloseTime  time.Time               // Thời điểm đóng nến
	Open       float64                 // Giá mở
	High       float64                 // Giá cao nhất
	Low        float64                 // Giá thấp nhất
	Close      float64                 // Giá đóng
	Volume     float64                 // Khối lượng giao dịch
	QuoteVolume float64                // Khối lượng giao dịch theo đồng báo giá
	TradeCount int64                   // Số lượng giao dịch
	Closed     bool                    // Nến đã đóng hay chưa
	Source     string                  // Nguồn dữ liệu (BINANCE_SPOT, BINANCE_FUTURES, ...)
}

// NewCandle tạo một Candle mới
func NewCandle(
	symbol value_object.Symbol,
	timeframe value_object.Timeframe,
	openTime time.Time,
	open, high, low, close, volume, quoteVolume float64,
	tradeCount int64,
	source string,
) *Candle {
	// Tính closeTime dựa trên openTime và timeframe
	closeTime := openTime.Add(timeframe.ToDuration())
	
	return &Candle{
		Symbol:      symbol,
		Timeframe:   timeframe,
		OpenTime:    openTime,
		CloseTime:   closeTime,
		Open:        open,
		High:        high,
		Low:         low,
		Close:       close,
		Volume:      volume,
		QuoteVolume: quoteVolume,
		TradeCount:  tradeCount,
		Closed:      false,
		Source:      source,
	}
}

// NewEmptyCandle tạo một Candle mới với giá trị mặc định
func NewEmptyCandle(
	symbol value_object.Symbol,
	timeframe value_object.Timeframe,
	openTime time.Time,
	source string,
) *Candle {
	return NewCandle(
		symbol,
		timeframe,
		openTime,
		0, 0, 0, 0, 0, 0, 0,
		source,
	)
}

// Update cập nhật nến với một giá mới
func (c *Candle) Update(price, volume, quoteVolume float64) {
	// Nếu nến đã đóng, không cập nhật nữa
	if c.Closed {
		return
	}
	
	// Nếu là giá đầu tiên
	if c.Open == 0 {
		c.Open = price
		c.High = price
		c.Low = price
		c.Close = price
		c.Volume = volume
		c.QuoteVolume = quoteVolume
		c.TradeCount = 1
		return
	}
	
	// Cập nhật giá cao nhất và thấp nhất
	if price > c.High {
		c.High = price
	}
	if price < c.Low || c.Low == 0 {
		c.Low = price
	}
	
	// Cập nhật giá đóng và khối lượng
	c.Close = price
	c.Volume += volume
	c.QuoteVolume += quoteVolume
	c.TradeCount++
}

// Close đánh dấu nến đã đóng
func (c *Candle) Close() {
	c.Closed = true
}

// Validate kiểm tra tính hợp lệ của nến
func (c *Candle) Validate() error {
	if c.Open <= 0 {
		return fmt.Errorf("open price must be positive")
	}
	if c.High <= 0 {
		return fmt.Errorf("high price must be positive")
	}
	if c.Low <= 0 {
		return fmt.Errorf("low price must be positive")
	}
	if c.Close <= 0 {
		return fmt.Errorf("close price must be positive")
	}
	if c.High < c.Low {
		return fmt.Errorf("high price cannot be less than low price")
	}
	if c.High < c.Open || c.High < c.Close {
		return fmt.Errorf("high price must be greater than or equal to open and close prices")
	}
	if c.Low > c.Open || c.Low > c.Close {
		return fmt.Errorf("low price must be less than or equal to open and close prices")
	}
	if c.Volume < 0 {
		return fmt.Errorf("volume cannot be negative")
	}
	if c.QuoteVolume < 0 {
		return fmt.Errorf("quote volume cannot be negative")
	}
	if c.TradeCount < 0 {
		return fmt.Errorf("trade count cannot be negative")
	}
	if c.OpenTime.IsZero() {
		return fmt.Errorf("open time cannot be zero")
	}
	if c.CloseTime.IsZero() {
		return fmt.Errorf("close time cannot be zero")
	}
	if c.CloseTime.Before(c.OpenTime) {
		return fmt.Errorf("close time cannot be before open time")
	}
	if c.Source == "" {
		return fmt.Errorf("source cannot be empty")
	}
	return nil
}

// IsGreen kiểm tra xem nến có xanh không (close > open)
func (c *Candle) IsGreen() bool {
	return c.Close > c.Open
}

// IsRed kiểm tra xem nến có đỏ không (close < open)
func (c *Candle) IsRed() bool {
	return c.Close < c.Open
}

// IsDoji kiểm tra xem nến có phải là doji không (close ≈ open)
func (c *Candle) IsDoji() bool {
	// Doji khi chênh lệch giữa open và close < 0.1% của giá
	threshold := c.Open * 0.001
	return abs(c.Close-c.Open) < threshold
}

// BodySize trả về kích thước thân nến
func (c *Candle) BodySize() float64 {
	return abs(c.Close - c.Open)
}

// UpperShadow trả về bóng trên của nến
func (c *Candle) UpperShadow() float64 {
	if c.IsGreen() {
		return c.High - c.Close
	}
	return c.High - c.Open
}

// LowerShadow trả về bóng dưới của nến
func (c *Candle) LowerShadow() float64 {
	if c.IsGreen() {
		return c.Open - c.Low
	}
	return c.Close - c.Low
}

// Range trả về biên độ của nến (high - low)
func (c *Candle) Range() float64 {
	return c.High - c.Low
}

// OHLC trả về map chứa giá OHLC
func (c *Candle) OHLC() map[string]float64 {
	return map[string]float64{
		"open":  c.Open,
		"high":  c.High,
		"low":   c.Low,
		"close": c.Close,
	}
}

// String trả về biểu diễn chuỗi của Candle
func (c *Candle) String() string {
	return fmt.Sprintf("%s %s: O=%.8f H=%.8f L=%.8f C=%.8f V=%.8f QV=%.8f TC=%d T=%s-%s",
		c.Symbol.String(), c.Timeframe.String(),
		c.Open, c.High, c.Low, c.Close,
		c.Volume, c.QuoteVolume, c.TradeCount,
		c.OpenTime.Format(time.RFC3339), c.CloseTime.Format(time.RFC3339))
}

// Helper function
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
