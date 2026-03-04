// futures_ws_payloads.go
package binance

// Aggregate Trade (Futures): <symbol>@aggTrade
type FuturesAggTradeEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	AggTradeID int64 `json:"a"`
	Price     string `json:"p"`
	Qty       string `json:"q"`  // Quantity with all market trades
	NormalQty string `json:"nq"` // Normal quantity without RPI-involving trades
	FirstID   int64  `json:"f"`
	LastID    int64  `json:"l"`
	TradeTime int64  `json:"T"`
	IsMaker   bool   `json:"m"`
}

// Mark Price Update: <symbol>@markPrice or @1s
type FuturesMarkPriceEvent struct {
	EventType  string `json:"e"` // "markPriceUpdate"
	EventTime  int64  `json:"E"`
	Symbol     string `json:"s"`
	MarkPrice  string `json:"p"`
	IndexPrice string `json:"i"`
	EstSettlePrice string `json:"P"`
	FundingRate string `json:"r"`
	NextFundingTime int64 `json:"T"`
}

// Kline: <symbol>@kline_<interval>
type FuturesKlineEvent struct {
	EventType string          `json:"e"` // "kline"
	EventTime int64           `json:"E"`
	Symbol    string          `json:"s"`
	K         FuturesKlineInner `json:"k"`
}

type FuturesKlineInner struct {
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

// Continuous Contract Kline: <pair>_<contractType>@continuousKline_<interval>
type FuturesContinuousKlineEvent struct {
	EventType string `json:"e"` // "continuous_kline"
	EventTime int64  `json:"E"`
	Pair      string `json:"ps"` // Pair e.g. BTCUSDT
	ContractType string `json:"ct"` // PERPETUAL / CURRENT_QUARTER / ...
	K         FuturesContinuousKlineInner `json:"k"`
}

type FuturesContinuousKlineInner struct {
	StartTime int64  `json:"t"`
	CloseTime int64 `json:"T"`
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

// BookTicker: <symbol>@bookTicker or !bookTicker
type FuturesBookTickerEvent struct {
	EventType string `json:"e"` // "bookTicker"
	UpdateID  int64  `json:"u"`
	EventTime int64  `json:"E"`
	TransTime int64  `json:"T"`
	Symbol    string `json:"s"`
	BidPrice  string `json:"b"`
	BidQty    string `json:"B"`
	AskPrice  string `json:"a"`
	AskQty    string `json:"A"`
}
