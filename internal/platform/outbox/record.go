package record

import (
	"encoding/json"
	"time"
)

// TickerRecord is the external message structure we will publish.
type TickerRecord struct {
	Symbol  string                 `json:"symbol"`
	AskPr   float64                `json:"askPr"`
	BidPr   float64                `json:"bidPr"`
	BestBid float64                `json:"bestBid"`
	BestAsk float64                `json:"bestAsk"`
	BidSz   float64                `json:"bidSz"`
	AskSz   float64                `json:"askSz"`
	LastPr  float64                `json:"lastPr"`
	Spread  float64                `json:"spread"`
	TS      int64                  `json:"ts"`
	Raw     map[string]interface{} `json:"raw,omitempty"`
}

type CandleRecord struct {
	StartTime int64   `json:"start"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	QuoteId   int     `json:"quoteId"`
	EndTime   int64   `json:"endTime"`
}

// TickerRecordFromRaw converts a generic data payload to TickerRecord
func TickerRecordFromRaw(raw interface{}) TickerRecord {
	// raw might be map[string]interface{} or a typed struct. We defensively parse.
	var out TickerRecord
	switch v := raw.(type) {
	case map[string]interface{}:
		out.Symbol = toString(v["symbol"])
		out.AskPr = toFloat(v["ask"])
		out.BidPr = toFloat(v["bid"])
		out.BestBid = toFloat(v["bestBid"])
		out.BestAsk = toFloat(v["bestAsk"])
		out.BidSz = toFloat(v["bidVolume"])
		out.AskSz = toFloat(v["askVolume"])
		out.Spread = toFloat(v["spreadRaw"])
		ts := int64(toFloat(v["timestamp"]))
		if ts == 0 {
			ts = time.Now().UnixMilli()
		}
		out.TS = ts
		last := (out.BestBid + out.BestAsk) / 2
		out.LastPr = last
		out.Raw = v
	default:
		// fallback: json marshal/unmarshal to try decode
		b, _ := json.Marshal(v)
		_ = json.Unmarshal(b, &out)
		if out.TS == 0 {
			out.TS = time.Now().UnixMilli()
		}
	}
	return out
}

func CandlesRecordFromRaw(raw interface{}) map[string][]CandleRecord {
	// raw -> map[timeframe][]CandleRecord
	out := map[string][]CandleRecord{}
	switch v := raw.(type) {
	case map[string]interface{}:
		// if data is {"1m": [[start, open, high, low, close, quoteId, volume, end], ...]}
		for k, vv := range v {
			switch arr := vv.(type) {
			case []interface{}:
				for _, item := range arr {
					switch it := item.(type) {
					case []interface{}:
						cr := CandleRecord{}
						if len(it) >= 8 {
							cr.StartTime = toInt64(it[0])
							cr.Open = toFloat(it[1])
							cr.High = toFloat(it[2])
							cr.Low = toFloat(it[3])
							cr.Close = toFloat(it[4])
							cr.QuoteId = int(toFloat(it[5]))
							cr.Volume = toFloat(it[6])
							cr.EndTime = toInt64(it[7])
							out[k] = append(out[k], cr)
						}
					}
				}
			}
		}
	default:
		// fallback empty
	}
	return out
}

// helper conversions
func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		// try parse
		var f float64
		_ = json.Unmarshal([]byte(`"`+t+`"`), &f)
		return f
	default:
		return 0
	}
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return ""
	default:
		return ""
	}
}

func toInt64(v interface{}) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	default:
		return time.Now().UnixMilli()
	}
}
