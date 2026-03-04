// spot_ws_streams.go
package binance

import (
	"fmt"
	"strings"
)

type DepthSpeed string

const (
	DepthSpeed1000ms DepthSpeed = ""       // default
	DepthSpeed100ms  DepthSpeed = "@100ms" // suffix
)

func NormalizeSymbolLower(symbol string) string {
	return strings.ToLower(symbol)
}

// Trades
func StreamTrade(symbol string) string {
	return fmt.Sprintf("%s@trade", NormalizeSymbolLower(symbol))
}

// Aggregate trades
func StreamAggTrade(symbol string) string {
	return fmt.Sprintf("%s@aggTrade", NormalizeSymbolLower(symbol))
}

// Klines/Candles
// interval examples: "1s","1m","3m","5m","15m","30m","1h","2h","4h","6h","8h","12h","1d","3d","1w","1M"
func StreamKline(symbol, interval string) string {
	return fmt.Sprintf("%s@kline_%s", NormalizeSymbolLower(symbol), interval)
}

// Book ticker (best bid/ask)
func StreamBookTicker(symbol string) string {
	return fmt.Sprintf("%s@bookTicker", NormalizeSymbolLower(symbol))
}

// Avg price
func StreamAvgPrice(symbol string) string {
	return fmt.Sprintf("%s@avgPrice", NormalizeSymbolLower(symbol))
}

// Mini ticker
func StreamMiniTicker(symbol string) string {
	return fmt.Sprintf("%s@miniTicker", NormalizeSymbolLower(symbol))
}

// 24hr ticker (single symbol)
func StreamTicker24h(symbol string) string {
	return fmt.Sprintf("%s@ticker", NormalizeSymbolLower(symbol))
}

// Rolling window ticker (1h/4h/1d)
func StreamRollingWindowTicker(symbol, windowSize string) string {
	return fmt.Sprintf("%s@ticker_%s", NormalizeSymbolLower(symbol), windowSize)
}

// Diff depth (order book updates to manage local order book)
func StreamDepthDiff(symbol string, speed DepthSpeed) string {
	return fmt.Sprintf("%s@depth%s", NormalizeSymbolLower(symbol), string(speed))
}

// Partial depth top levels: levels = 5,10,20
func StreamDepthPartial(symbol string, levels int, speed DepthSpeed) string {
	return fmt.Sprintf("%s@depth%d%s", NormalizeSymbolLower(symbol), levels, string(speed))
}

// All-market streams (examples)
func StreamAllMiniTickerArr() string {
	return "!miniTicker@arr"
}

func StreamAllRollingWindowArr(windowSize string) string {
	return fmt.Sprintf("!ticker_%s@arr", windowSize)
}
