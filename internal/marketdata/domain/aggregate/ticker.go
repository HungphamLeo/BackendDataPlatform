package aggregate

import (
	"fmt"
	"sync"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/event"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// TickerID là value object đại diện cho ID của Ticker
type TickerID string

// Ticker là Aggregate Root quản lý thông tin giá của một symbol
type Ticker struct {
	id           TickerID                // ID của ticker
	symbol       value_object.Symbol     // Symbol của ticker
	exchange     string                  // Sàn giao dịch (BINANCE_SPOT, BINANCE_FUTURES, ...)
	lastPrice    float64                 // Giá gần nhất
	lastQuantity float64                 // Khối lượng gần nhất
	lastTrade    *entity.Trade           // Giao dịch gần nhất
	lastQuote    value_object.Quote      // Báo giá gần nhất
	orderBook    *entity.OrderBook       // Sổ lệnh hiện tại
	candles      map[string]*entity.Candle // Các nến theo timeframe
	lastUpdate   time.Time               // Thời điểm cập nhật gần nhất
	domainEvents []event.DomainEvent     // Các sự kiện domain chưa được publish
	mu           sync.RWMutex            // Mutex để đảm bảo thread-safe
}

// NewTicker tạo một Ticker mới
func NewTicker(symbol value_object.Symbol, exchange string) *Ticker {
	return &Ticker{
		id:           TickerID(fmt.Sprintf("%s:%s", symbol.String(), exchange)),
		symbol:       symbol,
		exchange:     exchange,
		lastPrice:    0,
		lastQuantity: 0,
		candles:      make(map[string]*entity.Candle),
		lastUpdate:   time.Time{},
		domainEvents: []event.DomainEvent{},
	}
}

// ID trả về ID của ticker
func (t *Ticker) ID() TickerID {
	return t.id
}

// Symbol trả về symbol của ticker
func (t *Ticker) Symbol() value_object.Symbol {
	return t.symbol
}

// Exchange trả về sàn giao dịch của ticker
func (t *Ticker) Exchange() string {
	return t.exchange
}

// LastPrice trả về giá gần nhất
func (t *Ticker) LastPrice() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastPrice
}

// LastQuantity trả về khối lượng gần nhất
func (t *Ticker) LastQuantity() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastQuantity
}

// LastTrade trả về giao dịch gần nhất
func (t *Ticker) LastTrade() *entity.Trade {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastTrade
}

// LastQuote trả về báo giá gần nhất
func (t *Ticker) LastQuote() value_object.Quote {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastQuote
}

// OrderBook trả về sổ lệnh hiện tại
func (t *Ticker) OrderBook() *entity.OrderBook {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.orderBook
}

// LastUpdate trả về thời điểm cập nhật gần nhất
func (t *Ticker) LastUpdate() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastUpdate
}

// GetCandle trả về nến theo timeframe
func (t *Ticker) GetCandle(timeframe value_object.Timeframe) *entity.Candle {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.candles[timeframe.String()]
}

// GetCandles trả về tất cả các nến
func (t *Ticker) GetCandles() map[string]*entity.Candle {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	// Tạo bản sao để tránh race condition
	result := make(map[string]*entity.Candle, len(t.candles))
	for k, v := range t.candles {
		result[k] = v
	}
	
	return result
}

// UpdatePrice cập nhật giá mới từ một giao dịch
func (t *Ticker) UpdatePrice(trade *entity.Trade) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Validate trade
	if err := trade.Validate(); err != nil {
		return fmt.Errorf("invalid trade: %w", err)
	}
	
	// Kiểm tra xem symbol có khớp không
	if !t.symbol.Equals(trade.Symbol) {
		return fmt.Errorf("symbol mismatch: expected %s, got %s", t.symbol.String(), trade.Symbol.String())
	}
	
	// Kiểm tra xem exchange có khớp không
	if t.exchange != trade.Source {
		return fmt.Errorf("exchange mismatch: expected %s, got %s", t.exchange, trade.Source)
	}
	
	// Kiểm tra xem thời gian có mới hơn không
	if !t.lastUpdate.IsZero() && !trade.Timestamp.After(t.lastUpdate) {
		return nil // Bỏ qua giao dịch cũ
	}
	
	// Cập nhật giá và khối lượng
	oldPrice := t.lastPrice
	t.lastPrice = trade.Price
	t.lastQuantity = trade.Quantity
	t.lastTrade = trade
	t.lastUpdate = trade.Timestamp
	
	// Cập nhật nến
	t.updateCandles(trade)
	
	// Tạo sự kiện PriceUpdated
	priceUpdatedEvent := &event.PriceUpdated{
		TickerID:   string(t.id),
		Symbol:     t.symbol,
		OldPrice:   oldPrice,
		NewPrice:   trade.Price,
		Quantity:   trade.Quantity,
		TradeID:    trade.ID,
		Timestamp:  trade.Timestamp,
		Source:     t.exchange,
	}
	
	// Tạo sự kiện NewTrade
	newTradeEvent := &event.NewTrade{
		TickerID:   string(t.id),
		Symbol:     t.symbol,
		TradeID:    trade.ID,
		Price:      trade.Price,
		Quantity:   trade.Quantity,
		Side:       trade.Side(),
		Timestamp:  trade.Timestamp,
		Source:     t.exchange,
	}
	
	// Thêm sự kiện vào danh sách
	t.domainEvents = append(t.domainEvents, priceUpdatedEvent, newTradeEvent)
	
	return nil
}

// UpdateQuote cập nhật báo giá mới
func (t *Ticker) UpdateQuote(quote value_object.Quote) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Validate quote
	if err := quote.Validate(); err != nil {
		return fmt.Errorf("invalid quote: %w", err)
	}
	
	// Kiểm tra xem symbol có khớp không
	if !t.symbol.Equals(quote.Symbol) {
		return fmt.Errorf("symbol mismatch: expected %s, got %s", t.symbol.String(), quote.Symbol.String())
	}
	
	// Kiểm tra xem exchange có khớp không
	if t.exchange != quote.Source {
		return fmt.Errorf("exchange mismatch: expected %s, got %s", t.exchange, quote.Source)
	}
	
	// Kiểm tra xem thời gian có mới hơn không
	if !t.lastUpdate.IsZero() && !quote.Timestamp.After(t.lastUpdate) {
		return nil // Bỏ qua báo giá cũ
	}
	
	// Cập nhật báo giá
	oldQuote := t.lastQuote
	t.lastQuote = quote
	t.lastUpdate = quote.Timestamp
	
	// Tạo sự kiện QuoteUpdated
	quoteUpdatedEvent := &event.QuoteUpdated{
		TickerID:   string(t.id),
		Symbol:     t.symbol,
		OldQuote:   oldQuote,
		NewQuote:   quote,
		Timestamp:  quote.Timestamp,
		Source:     t.exchange,
	}
	
	// Thêm sự kiện vào danh sách
	t.domainEvents = append(t.domainEvents, quoteUpdatedEvent)
	
	return nil
}

// UpdateOrderBook cập nhật sổ lệnh mới
func (t *Ticker) UpdateOrderBook(orderBook *entity.OrderBook) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Validate order book
	if err := orderBook.Validate(); err != nil {
		return fmt.Errorf("invalid order book: %w", err)
	}
	
	// Kiểm tra xem symbol có khớp không
	if !t.symbol.Equals(orderBook.Symbol) {
		return fmt.Errorf("symbol mismatch: expected %s, got %s", t.symbol.String(), orderBook.Symbol.String())
	}
	
	// Kiểm tra xem exchange có khớp không
	if t.exchange != orderBook.Source {
		return fmt.Errorf("exchange mismatch: expected %s, got %s", t.exchange, orderBook.Source)
	}
	
	// Kiểm tra xem lastUpdateID có mới hơn không
	if t.orderBook != nil && orderBook.LastUpdateID <= t.orderBook.LastUpdateID {
		return nil // Bỏ qua order book cũ
	}
	
	// Cập nhật order book
	oldOrderBook := t.orderBook
	t.orderBook = orderBook
	t.lastUpdate = orderBook.Timestamp
	
	// Tạo sự kiện OrderBookUpdated
	orderBookUpdatedEvent := &event.OrderBookUpdated{
		TickerID:     string(t.id),
		Symbol:       t.symbol,
		LastUpdateID: orderBook.LastUpdateID,
		Timestamp:    orderBook.Timestamp,
		Source:       t.exchange,
	}
	
	// Thêm sự kiện vào danh sách
	t.domainEvents = append(t.domainEvents, orderBookUpdatedEvent)
	
	// Cập nhật báo giá từ order book
	bestBid, err := orderBook.BestBid()
	if err == nil {
		bestAsk, err := orderBook.BestAsk()
		if err == nil {
			quote := value_object.NewQuote(
				t.symbol,
				bestBid.Price,
				bestAsk.Price,
				bestBid.Quantity,
				bestAsk.Quantity,
				orderBook.Timestamp,
				t.exchange,
			)
			t.lastQuote = quote
		}
	}
	
	return nil
}

// CreateCandle tạo một nến mới cho một timeframe
func (t *Ticker) CreateCandle(timeframe value_object.Timeframe, openTime time.Time) *entity.Candle {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	candle := entity.NewEmptyCandle(t.symbol, timeframe, openTime, t.exchange)
	t.candles[timeframe.String()] = candle
	
	return candle
}

// CloseCandle đóng một nến
func (t *Ticker) CloseCandle(timeframe value_object.Timeframe) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	candle, exists := t.candles[timeframe.String()]
	if !exists {
		return fmt.Errorf("candle not found for timeframe: %s", timeframe.String())
	}
	
	// Validate candle
	if err := candle.Validate(); err != nil {
		return fmt.Errorf("invalid candle: %w", err)
	}
	
	// Đóng nến
	candle.Close()
	
	// Tạo sự kiện CandleClosed
	candleClosedEvent := &event.CandleClosed{
		TickerID:   string(t.id),
		Symbol:     t.symbol,
		Timeframe:  timeframe,
		OpenTime:   candle.OpenTime,
		CloseTime:  candle.CloseTime,
		Open:       candle.Open,
		High:       candle.High,
		Low:        candle.Low,
		Close:      candle.Close,
		Volume:     candle.Volume,
		TradeCount: candle.TradeCount,
		Timestamp:  time.Now(),
		Source:     t.exchange,
	}
	
	// Thêm sự kiện vào danh sách
	t.domainEvents = append(t.domainEvents, candleClosedEvent)
	
	return nil
}

// updateCandles cập nhật các nến với một giao dịch mới
func (t *Ticker) updateCandles(trade *entity.Trade) {
	for _, candle := range t.candles {
		// Kiểm tra xem giao dịch có thuộc về nến này không
		if trade.Timestamp.After(candle.OpenTime) && (trade.Timestamp.Before(candle.CloseTime) || trade.Timestamp.Equal(candle.CloseTime)) {
			// Cập nhật nến
			candle.Update(trade.Price, trade.Quantity, trade.QuoteQuantity)
		}
	}
}

// DomainEvents trả về các sự kiện domain chưa được publish
func (t *Ticker) DomainEvents() []event.DomainEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	// Tạo bản sao để tránh race condition
	result := make([]event.DomainEvent, len(t.domainEvents))
	copy(result, t.domainEvents)
	
	return result
}

// ClearDomainEvents xóa các sự kiện domain đã được publish
func (t *Ticker) ClearDomainEvents() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.domainEvents = []event.DomainEvent{}
}

// String trả về biểu diễn chuỗi của Ticker
func (t *Ticker) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	return fmt.Sprintf("Ticker %s: LastPrice=%.8f, LastQuantity=%.8f, LastUpdate=%s, Exchange=%s",
		t.symbol.String(), t.lastPrice, t.lastQuantity, t.lastUpdate.Format(time.RFC3339), t.exchange)
}
