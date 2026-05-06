package event

import (
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// DomainEvent là interface cho tất cả các sự kiện domain
type DomainEvent interface {
	// EventType trả về loại sự kiện
	EventType() string
	
	// AggregateID trả về ID của aggregate root
	AggregateID() string
	
	// OccurredAt trả về thời điểm xảy ra sự kiện
	OccurredAt() time.Time
	
	// Source trả về nguồn của sự kiện
	Source() string
}

// BaseDomainEvent chứa các trường chung cho tất cả các sự kiện domain
type BaseDomainEvent struct {
	EventName  string    // Tên sự kiện
	EntityID   string    // ID của entity
	Timestamp  time.Time // Thời điểm xảy ra sự kiện
	SourceName string    // Nguồn của sự kiện
}

// EventType trả về loại sự kiện
func (e BaseDomainEvent) EventType() string {
	return e.EventName
}

// AggregateID trả về ID của aggregate root
func (e BaseDomainEvent) AggregateID() string {
	return e.EntityID
}

// OccurredAt trả về thời điểm xảy ra sự kiện
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// Source trả về nguồn của sự kiện
func (e BaseDomainEvent) Source() string {
	return e.SourceName
}

// PriceUpdated là sự kiện khi giá được cập nhật
type PriceUpdated struct {
	BaseDomainEvent
	TickerID  string             // ID của ticker
	Symbol    value_object.Symbol // Symbol
	OldPrice  float64            // Giá cũ
	NewPrice  float64            // Giá mới
	Quantity  float64            // Khối lượng
	TradeID   string             // ID của giao dịch
}

// NewPriceUpdated tạo một sự kiện PriceUpdated mới
func NewPriceUpdated(
	tickerID string,
	symbol value_object.Symbol,
	oldPrice, newPrice, quantity float64,
	tradeID string,
	timestamp time.Time,
	source string,
) *PriceUpdated {
	return &PriceUpdated{
		BaseDomainEvent: BaseDomainEvent{
			EventName:  "PriceUpdated",
			EntityID:   tickerID,
			Timestamp:  timestamp,
			SourceName: source,
		},
		TickerID:  tickerID,
		Symbol:    symbol,
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
		Quantity:  quantity,
		TradeID:   tradeID,
	}
}

// QuoteUpdated là sự kiện khi báo giá được cập nhật
type QuoteUpdated struct {
	BaseDomainEvent
	TickerID  string             // ID của ticker
	Symbol    value_object.Symbol // Symbol
	OldQuote  value_object.Quote  // Báo giá cũ
	NewQuote  value_object.Quote  // Báo giá mới
}

// NewQuoteUpdated tạo một sự kiện QuoteUpdated mới
func NewQuoteUpdated(
	tickerID string,
	symbol value_object.Symbol,
	oldQuote, newQuote value_object.Quote,
	timestamp time.Time,
	source string,
) *QuoteUpdated {
	return &QuoteUpdated{
		BaseDomainEvent: BaseDomainEvent{
			EventName:  "QuoteUpdated",
			EntityID:   tickerID,
			Timestamp:  timestamp,
			SourceName: source,
		},
		TickerID:  tickerID,
		Symbol:    symbol,
		OldQuote:  oldQuote,
		NewQuote:  newQuote,
	}
}

// NewTrade là sự kiện khi có giao dịch mới
type NewTrade struct {
	BaseDomainEvent
	TickerID  string             // ID của ticker
	Symbol    value_object.Symbol // Symbol
	TradeID   string             // ID của giao dịch
	Price     float64            // Giá
	Quantity  float64            // Khối lượng
	Side      string             // Phía (BUY hoặc SELL)
}

// NewNewTrade tạo một sự kiện NewTrade mới
func NewNewTrade(
	tickerID string,
	symbol value_object.Symbol,
	tradeID string,
	price, quantity float64,
	side string,
	timestamp time.Time,
	source string,
) *NewTrade {
	return &NewTrade{
		BaseDomainEvent: BaseDomainEvent{
			EventName:  "NewTrade",
			EntityID:   tickerID,
			Timestamp:  timestamp,
			SourceName: source,
		},
		TickerID:  tickerID,
		Symbol:    symbol,
		TradeID:   tradeID,
		Price:     price,
		Quantity:  quantity,
		Side:      side,
	}
}

// OrderBookUpdated là sự kiện khi sổ lệnh được cập nhật
type OrderBookUpdated struct {
	BaseDomainEvent
	TickerID     string             // ID của ticker
	Symbol       value_object.Symbol // Symbol
	LastUpdateID int64              // ID cập nhật cuối cùng
}

// NewOrderBookUpdated tạo một sự kiện OrderBookUpdated mới
func NewOrderBookUpdated(
	tickerID string,
	symbol value_object.Symbol,
	lastUpdateID int64,
	timestamp time.Time,
	source string,
) *OrderBookUpdated {
	return &OrderBookUpdated{
		BaseDomainEvent: BaseDomainEvent{
			EventName:  "OrderBookUpdated",
			EntityID:   tickerID,
			Timestamp:  timestamp,
			SourceName: source,
		},
		TickerID:     tickerID,
		Symbol:       symbol,
		LastUpdateID: lastUpdateID,
	}
}

// CandleClosed là sự kiện khi nến đóng
type CandleClosed struct {
	BaseDomainEvent
	TickerID   string                // ID của ticker
	Symbol     value_object.Symbol    // Symbol
	Timeframe  value_object.Timeframe // Khung thời gian
	OpenTime   time.Time             // Thời điểm mở
	CloseTime  time.Time             // Thời điểm đóng
	Open       float64               // Giá mở
	High       float64               // Giá cao nhất
	Low        float64               // Giá thấp nhất
	Close      float64               // Giá đóng
	Volume     float64               // Khối lượng
	TradeCount int64                 // Số lượng giao dịch
}

// NewCandleClosed tạo một sự kiện CandleClosed mới
func NewCandleClosed(
	tickerID string,
	symbol value_object.Symbol,
	timeframe value_object.Timeframe,
	openTime, closeTime time.Time,
	open, high, low, close, volume float64,
	tradeCount int64,
	timestamp time.Time,
	source string,
) *CandleClosed {
	return &CandleClosed{
		BaseDomainEvent: BaseDomainEvent{
			EventName:  "CandleClosed",
			EntityID:   tickerID,
			Timestamp:  timestamp,
			SourceName: source,
		},
		TickerID:   tickerID,
		Symbol:     symbol,
		Timeframe:  timeframe,
		OpenTime:   openTime,
		CloseTime:  closeTime,
		Open:       open,
		High:       high,
		Low:        low,
		Close:      close,
		Volume:     volume,
		TradeCount: tradeCount,
	}
}

// AlertTriggered là sự kiện khi cảnh báo được kích hoạt
type AlertTriggered struct {
	BaseDomainEvent
	TickerID    string             // ID của ticker
	Symbol      value_object.Symbol // Symbol
	AlertID     string             // ID của cảnh báo
	AlertType   string             // Loại cảnh báo
	Condition   string             // Điều kiện
	CurrentValue float64           // Giá trị hiện tại
	ThresholdValue float64         // Giá trị ngưỡng
}

// NewAlertTriggered tạo một sự kiện AlertTriggered mới
func NewAlertTriggered(
	tickerID string,
	symbol value_object.Symbol,
	alertID, alertType, condition string,
	currentValue, thresholdValue float64,
	timestamp time.Time,
	source string,
) *AlertTriggered {
	return &AlertTriggered{
		BaseDomainEvent: BaseDomainEvent{
			EventName:  "AlertTriggered",
			EntityID:   tickerID,
			Timestamp:  timestamp,
			SourceName: source,
		},
		TickerID:       tickerID,
		Symbol:         symbol,
		AlertID:        alertID,
		AlertType:      alertType,
		Condition:      condition,
		CurrentValue:   currentValue,
		ThresholdValue: thresholdValue,
	}
}
