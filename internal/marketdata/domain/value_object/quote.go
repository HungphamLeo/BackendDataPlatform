package value_object

import "fmt"

// ✅ VALUE OBJECT: Quote (immutable, no ID)
// Represents Bid/Ask prices at a moment
type Quote struct {
	Bid       float64
	Ask       float64
	Spread    float64
	Timestamp int64 // Unix nanoseconds
}

// Value objects are compared by VALUE, not identity
func (q Quote) Equals(other Quote) bool {
	return q.Bid == other.Bid && q.Ask == other.Ask
}

func (q Quote) Validate() error {
	if q.Bid <= 0 {
		return fmt.Errorf("bid must be > 0, got %f", q.Bid)
	}
	if q.Ask <= 0 {
		return fmt.Errorf("ask must be > 0, got %f", q.Ask)
	}
	if q.Bid > q.Ask {
		return fmt.Errorf("bid must be < ask")
	}
	return nil
}

func (q Quote) MidPrice() float64 {
	return (q.Bid + q.Ask) / 2
}
