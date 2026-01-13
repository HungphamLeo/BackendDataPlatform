package ticker

import "time"

// Domain model for Tick / Candle simplified for DDD
type Ticker struct {
	Symbol            string
	AskPrice          float64
	BidPrice          float64
	BestBid           float64
	BestAsk           float64
	BidSize           float64
	AskSize           float64
	LastPrice         float64
	Spread            float64
	Timestamp         time.Time
	Raw               map[string]interface{} // keep raw if needed
}

type Candle struct {
	StartTime time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	QuoteId   int
	EndTime   time.Time
	Raw       map[string]interface{}
}