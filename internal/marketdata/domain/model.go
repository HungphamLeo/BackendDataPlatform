package sharedkernel

import (
    "fmt"
)

// Money: Đối tượng giá trị (Value Object) - không thay đổi được
type Money struct {
    Amount   int64   // Số tiền (tính theo cent, ví dụ: 1000 = 10.00 USD)
    Currency string  // Loại tiền (USD, VND, EUR)
}

// NewMoney: Tạo Money mới
func NewMoney(amount int64, currency string) (*Money, error) {
    if amount < 0 {
        return nil, fmt.Errorf("amount cannot be negative: %d", amount)
    }
    if currency == "" {
        return nil, fmt.Errorf("currency is required")
    }
    return &Money{
        Amount:   amount,
        Currency: currency,
    }, nil
}

// Add: Cộng hai Money (tạo object mới, không thay đổi hiện tại)
func (m *Money) Add(other *Money) (*Money, error) {
    if m.Currency != other.Currency {
        return nil, fmt.Errorf("cannot add different currencies: %s + %s", m.Currency, other.Currency)
    }
    return &Money{
        Amount:   m.Amount + other.Amount,
        Currency: m.Currency,
    }, nil
}

// Subtract: Trừ hai Money
func (m *Money) Subtract(other *Money) (*Money, error) {
    if m.Currency != other.Currency {
        return nil, fmt.Errorf("cannot subtract different currencies")
    }
    if m.Amount < other.Amount {
        return nil, fmt.Errorf("insufficient balance: %d < %d", m.Amount, other.Amount)
    }
    return &Money{
        Amount:   m.Amount - other.Amount,
        Currency: m.Currency,
    }, nil
}

// IsGreaterThan: So sánh xem có nhiều hơn không
func (m *Money) IsGreaterThan(other *Money) bool {
    if m.Currency != other.Currency {
        return false
    }
    return m.Amount > other.Amount
}

// IsEqual: Hai Money có bằng nhau không
func (m *Money) IsEqual(other *Money) bool {
    return m.Amount == other.Amount && m.Currency == other.Currency
}

// String: Để in ra (ví dụ: "10.00 USD")
func (m *Money) String() string {
    return fmt.Sprintf("%.2f %s", float64(m.Amount)/100, m.Currency)
}

// IsEqual: Hai Symbol có bằng nhau không
func (s *Symbol) IsEqual(other *Symbol) bool {
    return s.Code == other.Code
}


**Cách dùng:**
```go
// Tạo 10 USD
money1, _ := sharedkernel.NewMoney(1000, "USD")

// Tạo 5 USD
money2, _ := sharedkernel.NewMoney(500, "USD")

// Cộng lại
total, _ := money1.Add(money2)
fmt.Println(total)  // "15.00 USD"

// Trừ
remaining, _ := money1.Subtract(money2)
fmt.Println(remaining)  // "5.00 USD"

// So sánh
if money1.IsGreaterThan(money2) {
    fmt.Println("money1 nhiều hơn")
}
```

---