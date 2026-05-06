// futures_ws_streams.go
package streaming

import (
	"fmt"
	"strings"
)

func FuturesNormalizeSymbolLower(symbol string) string {
	return strings.ToLower(symbol)
}

// aggTrade
func FuturesStreamAggTrade(symbol string) string {
	return fmt.Sprintf("%s@aggTrade", FuturesNormalizeSymbolLower(symbol))
}

// markPrice: <symbol>@markPrice OR <symbol>@markPrice@1s
func FuturesStreamMarkPrice(symbol string, oneSecond bool) string {
	if oneSecond {
		return fmt.Sprintf("%s@markPrice@1s", FuturesNormalizeSymbolLower(symbol))
	}
	return fmt.Sprintf("%s@markPrice", FuturesNormalizeSymbolLower(symbol))
}

// markPrice all-market: !markPrice@arr OR !markPrice@arr@1s
func FuturesStreamAllMarkPriceArr(oneSecond bool) string {
	if oneSecond {
		return "!markPrice@arr@1s"
	}
	return "!markPrice@arr"
}

// klines: <symbol>@kline_<interval>
func FuturesStreamKline(symbol, interval string) string {
	return fmt.Sprintf("%s@kline_%s", FuturesNormalizeSymbolLower(symbol), interval)
}

// contractType: perpetual, current_quarter, next_quarter, tradifi_perpetual
// stream: <pair>_<contractType>@continuousKline_<interval>
func FuturesStreamContinuousKline(pair, contractType, interval string) string {
	return fmt.Sprintf("%s_%s@continuousKline_%s",
		FuturesNormalizeSymbolLower(pair),
		FuturesNormalizeSymbolLower(contractType),
		interval,
	)
}

// bookTicker: <symbol>@bookTicker
func FuturesStreamBookTicker(symbol string) string {
	return fmt.Sprintf("%s@bookTicker", FuturesNormalizeSymbolLower(symbol))
}

// all bookTicker: !bookTicker
func FuturesStreamAllBookTicker() string {
	return "!bookTicker"
}
