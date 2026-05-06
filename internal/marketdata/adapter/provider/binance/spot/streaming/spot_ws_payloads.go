// spot_ws_payloads.go
package binance

// Trade event: <symbol>@trade
type TradeEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	TradeID   int64  `json:"t"`
	Price     string `json:"p"`
	Qty       string `json:"q"`
	TradeTime int64  `json:"T"`
	IsMaker   bool   `json:"m"`
	Ignore    bool   `json:"M"`
}

// AggTrade event: <symbol>@aggTrade
type AggTradeEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	AggTradeID int64 `json:"a"`
	Price     string `json:"p"`
	Qty       string `json:"q"`
	FirstID   int64  `json:"f"`
	LastID    int64  `json:"l"`
	TradeTime int64  `json:"T"`
	IsMaker   bool   `json:"m"`
	Ignore    bool   `json:"M"`
}

// Kline event: <symbol>@kline_<interval>
type KlineEvent struct {
	EventType string     `json:"e"`
	EventTime int64      `json:"E"`
	Symbol    string     `json:"s"`
	K         KlineInner `json:"k"`
}

type KlineInner struct {
	StartTime int64  `json:"t"`
	CloseTime int64 `json:"T"`
	Symbol    string `json:"s"`
	Interval  string `json:"i"`
	FirstID   int64  `json:"f"`
	LastID    int64  `json:"L"`
	Open      string `json:"o"`
	Close     string `json:"c"`
	High      string `json:"h"`
	Low       string `json:"l"`
	Volume    string `json:"v"`
	Trades    int64  `json:"n"`
	IsClosed  bool   `json:"x"`
	QuoteVol  string `json:"q"`
	TakerBase string `json:"V"`
	TakerQuote string `json:"Q"`
	Ignore    string `json:"B"`
}

// BookTicker event: <symbol>@bookTicker
type BookTickerEvent struct {
	UpdateID int64  `json:"u"`
	Symbol   string `json:"s"`
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
}

// Partial depth (top levels): <symbol>@depth<levels> OR ...@100ms
type DepthPartialEvent struct {
	LastUpdateID int64       `json:"lastUpdateId"`
	Bids         [][2]string `json:"bids"`
	Asks         [][2]string `json:"asks"`
}

// Diff depth updates: <symbol>@depth OR ...@100ms
type DepthDiffEvent struct {
	EventType string     `json:"e"`
	EventTime int64      `json:"E"`
	Symbol    string     `json:"s"`
	FirstU    int64      `json:"U"`
	FinalU    int64      `json:"u"`
	Bids      [][2]string `json:"b"`
	Asks      [][2]string `json:"a"`
}

// MiniTicker: <symbol>@miniTicker
type MiniTickerEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Close     string `json:"c"`
	Open      string `json:"o"`
	High      string `json:"h"`
	Low       string `json:"l"`
	BaseVol   string `json:"v"`
	QuoteVol  string `json:"q"`
}

// 24hr Ticker: <symbol>@ticker
type Ticker24hEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	PriceChange string `json:"p"`
	PriceChangePercent string `json:"P"`
	WeightedAvg string `json:"w"`
	PrevClose string `json:"x"`
	LastPrice string `json:"c"`
	LastQty   string `json:"Q"`
	BidPrice  string `json:"b"`
	BidQty    string `json:"B"`
	AskPrice  string `json:"a"`
	AskQty    string `json:"A"`
	OpenPrice string `json:"o"`
	HighPrice string `json:"h"`
	LowPrice  string `json:"l"`
	BaseVol   string `json:"v"`
	QuoteVol  string `json:"q"`
	OpenTime  int64  `json:"O"`
	CloseTime int64  `json:"C"`
	FirstID   int64  `json:"F"`
	LastID    int64  `json:"L"`
	Trades    int64  `json:"n"`
}

// AvgPrice: <symbol>@avgPrice
type AvgPriceEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Interval  string `json:"i"`
	AvgPrice  string `json:"w"`
	LastTradeTime int64 `json:"T"`
}
