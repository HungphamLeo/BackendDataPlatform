package entity

import (
	"fmt"
	"sort"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// OrderBookEntry đại diện cho một mức giá trong sổ lệnh
type OrderBookEntry struct {
	Price    float64 // Giá
	Quantity float64 // Khối lượng
}

// OrderBook đại diện cho sổ lệnh của một symbol
type OrderBook struct {
	Symbol      value_object.Symbol  // Symbol của sổ lệnh
	Bids        []OrderBookEntry     // Danh sách giá mua (bid), sắp xếp giảm dần theo giá
	Asks        []OrderBookEntry     // Danh sách giá bán (ask), sắp xếp tăng dần theo giá
	LastUpdateID int64               // ID cập nhật cuối cùng
	Timestamp   time.Time            // Thời điểm cập nhật
	Source      string               // Nguồn dữ liệu (BINANCE_SPOT, BINANCE_FUTURES, ...)
}

// NewOrderBook tạo một OrderBook mới
func NewOrderBook(
	symbol value_object.Symbol,
	bids, asks []OrderBookEntry,
	lastUpdateID int64,
	timestamp time.Time,
	source string,
) *OrderBook {
	// Sắp xếp bids giảm dần theo giá
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].Price > bids[j].Price
	})
	
	// Sắp xếp asks tăng dần theo giá
	sort.Slice(asks, func(i, j int) bool {
		return asks[i].Price < asks[j].Price
	})
	
	return &OrderBook{
		Symbol:      symbol,
		Bids:        bids,
		Asks:        asks,
		LastUpdateID: lastUpdateID,
		Timestamp:   timestamp,
		Source:      source,
	}
}

// NewEmptyOrderBook tạo một OrderBook mới với danh sách rỗng
func NewEmptyOrderBook(
	symbol value_object.Symbol,
	source string,
) *OrderBook {
	return NewOrderBook(
		symbol,
		[]OrderBookEntry{},
		[]OrderBookEntry{},
		0,
		time.Now(),
		source,
	)
}

// BestBid trả về giá mua tốt nhất (giá cao nhất)
func (ob *OrderBook) BestBid() (OrderBookEntry, error) {
	if len(ob.Bids) == 0 {
		return OrderBookEntry{}, fmt.Errorf("no bids available")
	}
	return ob.Bids[0], nil
}

// BestAsk trả về giá bán tốt nhất (giá thấp nhất)
func (ob *OrderBook) BestAsk() (OrderBookEntry, error) {
	if len(ob.Asks) == 0 {
		return OrderBookEntry{}, fmt.Errorf("no asks available")
	}
	return ob.Asks[0], nil
}

// Spread trả về chênh lệch giữa giá bán tốt nhất và giá mua tốt nhất
func (ob *OrderBook) Spread() (float64, error) {
	bestBid, err := ob.BestBid()
	if err != nil {
		return 0, err
	}
	
	bestAsk, err := ob.BestAsk()
	if err != nil {
		return 0, err
	}
	
	return bestAsk.Price - bestBid.Price, nil
}

// SpreadPercentage trả về chênh lệch dưới dạng phần trăm
func (ob *OrderBook) SpreadPercentage() (float64, error) {
	bestBid, err := ob.BestBid()
	if err != nil {
		return 0, err
	}
	
	bestAsk, err := ob.BestAsk()
	if err != nil {
		return 0, err
	}
	
	if bestBid.Price == 0 {
		return 0, fmt.Errorf("bid price is zero")
	}
	
	return (bestAsk.Price - bestBid.Price) / bestBid.Price * 100, nil
}

// MidPrice trả về giá trung bình giữa giá mua tốt nhất và giá bán tốt nhất
func (ob *OrderBook) MidPrice() (float64, error) {
	bestBid, err := ob.BestBid()
	if err != nil {
		return 0, err
	}
	
	bestAsk, err := ob.BestAsk()
	if err != nil {
		return 0, err
	}
	
	return (bestBid.Price + bestAsk.Price) / 2, nil
}

// BidVolume trả về tổng khối lượng của các lệnh mua
func (ob *OrderBook) BidVolume() float64 {
	var total float64
	for _, bid := range ob.Bids {
		total += bid.Quantity
	}
	return total
}

// AskVolume trả về tổng khối lượng của các lệnh bán
func (ob *OrderBook) AskVolume() float64 {
	var total float64
	for _, ask := range ob.Asks {
		total += ask.Quantity
	}
	return total
}

// BidAskRatio trả về tỷ lệ giữa khối lượng mua và khối lượng bán
func (ob *OrderBook) BidAskRatio() float64 {
	askVolume := ob.AskVolume()
	if askVolume == 0 {
		return 0
	}
	return ob.BidVolume() / askVolume
}

// UpdateBid cập nhật hoặc thêm một mức giá mua
func (ob *OrderBook) UpdateBid(price, quantity float64) {
	// Nếu quantity = 0, xóa mức giá này
	if quantity == 0 {
		for i, bid := range ob.Bids {
			if bid.Price == price {
				ob.Bids = append(ob.Bids[:i], ob.Bids[i+1:]...)
				break
			}
		}
		return
	}
	
	// Cập nhật nếu mức giá đã tồn tại
	for i, bid := range ob.Bids {
		if bid.Price == price {
			ob.Bids[i].Quantity = quantity
			return
		}
	}
	
	// Thêm mức giá mới và sắp xếp lại
	ob.Bids = append(ob.Bids, OrderBookEntry{Price: price, Quantity: quantity})
	sort.Slice(ob.Bids, func(i, j int) bool {
		return ob.Bids[i].Price > ob.Bids[j].Price
	})
}

// UpdateAsk cập nhật hoặc thêm một mức giá bán
func (ob *OrderBook) UpdateAsk(price, quantity float64) {
	// Nếu quantity = 0, xóa mức giá này
	if quantity == 0 {
		for i, ask := range ob.Asks {
			if ask.Price == price {
				ob.Asks = append(ob.Asks[:i], ob.Asks[i+1:]...)
				break
			}
		}
		return
	}
	
	// Cập nhật nếu mức giá đã tồn tại
	for i, ask := range ob.Asks {
		if ask.Price == price {
			ob.Asks[i].Quantity = quantity
			return
		}
	}
	
	// Thêm mức giá mới và sắp xếp lại
	ob.Asks = append(ob.Asks, OrderBookEntry{Price: price, Quantity: quantity})
	sort.Slice(ob.Asks, func(i, j int) bool {
		return ob.Asks[i].Price < ob.Asks[j].Price
	})
}

// ApplyDiff áp dụng các thay đổi từ một diff update
func (ob *OrderBook) ApplyDiff(bids, asks []OrderBookEntry, lastUpdateID int64, timestamp time.Time) {
	// Cập nhật bids
	for _, bid := range bids {
		ob.UpdateBid(bid.Price, bid.Quantity)
	}
	
	// Cập nhật asks
	for _, ask := range asks {
		ob.UpdateAsk(ask.Price, ask.Quantity)
	}
	
	// Cập nhật metadata
	ob.LastUpdateID = lastUpdateID
	ob.Timestamp = timestamp
}

// Validate kiểm tra tính hợp lệ của sổ lệnh
func (ob *OrderBook) Validate() error {
	// Kiểm tra xem có bids và asks không
	if len(ob.Bids) == 0 {
		return fmt.Errorf("no bids available")
	}
	if len(ob.Asks) == 0 {
		return fmt.Errorf("no asks available")
	}
	
	// Kiểm tra xem giá bán tốt nhất có lớn hơn giá mua tốt nhất không
	bestBid, _ := ob.BestBid()
	bestAsk, _ := ob.BestAsk()
	if bestBid.Price >= bestAsk.Price {
		return fmt.Errorf("best bid (%f) must be less than best ask (%f)", bestBid.Price, bestAsk.Price)
	}
	
	// Kiểm tra xem bids có được sắp xếp giảm dần không
	for i := 1; i < len(ob.Bids); i++ {
		if ob.Bids[i].Price > ob.Bids[i-1].Price {
			return fmt.Errorf("bids must be sorted in descending order by price")
		}
	}
	
	// Kiểm tra xem asks có được sắp xếp tăng dần không
	for i := 1; i < len(ob.Asks); i++ {
		if ob.Asks[i].Price < ob.Asks[i-1].Price {
			return fmt.Errorf("asks must be sorted in ascending order by price")
		}
	}
	
	// Kiểm tra các giá trị khác
	if ob.LastUpdateID <= 0 {
		return fmt.Errorf("lastUpdateID must be positive")
	}
	if ob.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}
	if ob.Source == "" {
		return fmt.Errorf("source cannot be empty")
	}
	
	return nil
}

// String trả về biểu diễn chuỗi của OrderBook
func (ob *OrderBook) String() string {
	bestBid, _ := ob.BestBid()
	bestAsk, _ := ob.BestAsk()
	
	return fmt.Sprintf("OrderBook %s: Bid=%f (Size=%f), Ask=%f (Size=%f), Spread=%.8f, LastUpdateID=%d, Time=%s, Source=%s",
		ob.Symbol.String(), bestBid.Price, bestBid.Quantity, bestAsk.Price, bestAsk.Quantity,
		bestAsk.Price-bestBid.Price, ob.LastUpdateID, ob.Timestamp.Format(time.RFC3339), ob.Source)
}
