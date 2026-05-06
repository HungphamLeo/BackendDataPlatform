package value_object

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeframeUnit đại diện cho đơn vị của timeframe
type TimeframeUnit string

const (
	Second TimeframeUnit = "s"
	Minute TimeframeUnit = "m"
	Hour   TimeframeUnit = "h"
	Day    TimeframeUnit = "d"
	Week   TimeframeUnit = "w"
	Month  TimeframeUnit = "M"
)

// Timeframe đại diện cho một khung thời gian đã được chuẩn hóa
// Ví dụ: 1m, 5m, 1h, 1d
type Timeframe struct {
	Value int          // Giá trị số
	Unit  TimeframeUnit // Đơn vị (s, m, h, d, w, M)
}

// NewTimeframe tạo một Timeframe mới
func NewTimeframe(value int, unit TimeframeUnit) Timeframe {
	return Timeframe{
		Value: value,
		Unit:  unit,
	}
}

// ParseTimeframe phân tích một chuỗi thành Timeframe
// Hỗ trợ định dạng: "1m", "5m", "1h", "1d", "1w", "1M"
func ParseTimeframe(s string) (Timeframe, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return Timeframe{}, fmt.Errorf("invalid timeframe format: %s", s)
	}
	
	// Tách giá trị số và đơn vị
	var unitStr string
	var valueStr string
	
	// Tìm vị trí của ký tự đầu tiên không phải số
	for i, c := range s {
		if c < '0' || c > '9' {
			valueStr = s[:i]
			unitStr = s[i:]
			break
		}
	}
	
	if valueStr == "" || unitStr == "" {
		return Timeframe{}, fmt.Errorf("invalid timeframe format: %s", s)
	}
	
	// Chuyển đổi giá trị số
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return Timeframe{}, fmt.Errorf("invalid timeframe value: %s", valueStr)
	}
	
	// Xác định đơn vị
	var unit TimeframeUnit
	switch unitStr {
	case "s", "sec", "second", "seconds":
		unit = Second
	case "m", "min", "minute", "minutes":
		unit = Minute
	case "h", "hr", "hour", "hours":
		unit = Hour
	case "d", "day", "days":
		unit = Day
	case "w", "week", "weeks":
		unit = Week
	case "M", "mo", "month", "months":
		unit = Month
	default:
		return Timeframe{}, fmt.Errorf("invalid timeframe unit: %s", unitStr)
	}
	
	return NewTimeframe(value, unit), nil
}

// String trả về biểu diễn chuỗi của Timeframe theo định dạng chuẩn
func (t Timeframe) String() string {
	return fmt.Sprintf("%d%s", t.Value, t.Unit)
}

// ToExchangeFormat chuyển đổi Timeframe sang định dạng của một sàn cụ thể
func (t Timeframe) ToExchangeFormat(exchange string) string {
	switch strings.ToUpper(exchange) {
	case "BINANCE", "BINANCE_SPOT", "BINANCE_FUTURES":
		return t.String() // 1m, 5m, 1h, 1d, 1w, 1M
	case "FTX":
		// FTX sử dụng: 60, 300, 3600, 86400 (seconds)
		seconds := t.ToSeconds()
		return fmt.Sprintf("%d", seconds)
	case "KRAKEN":
		// Kraken sử dụng: 1, 5, 60, 1440 (minutes)
		minutes := t.ToMinutes()
		return fmt.Sprintf("%d", minutes)
	default:
		return t.String()
	}
}

// ToSeconds chuyển đổi Timeframe sang số giây
func (t Timeframe) ToSeconds() int {
	switch t.Unit {
	case Second:
		return t.Value
	case Minute:
		return t.Value * 60
	case Hour:
		return t.Value * 60 * 60
	case Day:
		return t.Value * 24 * 60 * 60
	case Week:
		return t.Value * 7 * 24 * 60 * 60
	case Month:
		// Giả định 1 tháng = 30 ngày
		return t.Value * 30 * 24 * 60 * 60
	default:
		return 0
	}
}

// ToMinutes chuyển đổi Timeframe sang số phút
func (t Timeframe) ToMinutes() int {
	return t.ToSeconds() / 60
}

// ToDuration chuyển đổi Timeframe sang time.Duration
func (t Timeframe) ToDuration() time.Duration {
	return time.Duration(t.ToSeconds()) * time.Second
}

// CommonTimeframes trả về danh sách các timeframe phổ biến
func CommonTimeframes() []Timeframe {
	return []Timeframe{
		NewTimeframe(1, Minute),
		NewTimeframe(3, Minute),
		NewTimeframe(5, Minute),
		NewTimeframe(15, Minute),
		NewTimeframe(30, Minute),
		NewTimeframe(1, Hour),
		NewTimeframe(2, Hour),
		NewTimeframe(4, Hour),
		NewTimeframe(6, Hour),
		NewTimeframe(8, Hour),
		NewTimeframe(12, Hour),
		NewTimeframe(1, Day),
		NewTimeframe(3, Day),
		NewTimeframe(1, Week),
		NewTimeframe(1, Month),
	}
}

// Equals kiểm tra hai Timeframe có bằng nhau không
func (t Timeframe) Equals(other Timeframe) bool {
	// Chuyển đổi cả hai về seconds để so sánh
	return t.ToSeconds() == other.ToSeconds()
}
