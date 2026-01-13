package xtb

// Streaming command names (cmd_stream_execute in xtb_connection.txt)
const (
	StreamGetTickPrices  = "getTickPrices"
	StreamGetCandles     = "getCandles"
	StreamGetTrades      = "getTrades"
	StreamGetBalance     = "getBalance"
	StreamGetTradeStatus = "getTradeStatus"
	StreamGetProfits     = "getProfits"
	StreamGetNews        = "getNews"

	StreamStopTickPrices  = "stopTickPrices"
	StreamStopCandles     = "stopCandles"
	StreamStopTrades      = "stopTrades"
	StreamStopBalance     = "stopBalance"
	StreamStopTradeStatus = "stopTradeStatus"
	StreamStopProfits     = "stopProfits"
	StreamStopNews        = "stopNews"
)
