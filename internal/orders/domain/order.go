package domain

import (
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/BackendDataPlatform/internal/sharedkernel"
)

// OrderStatus: Trạng thái đơn hàng
type OrderStatus string

const (
    OrderPending   OrderStatus = "PENDING"    // Chờ thực hiện
    OrderFilled    OrderStatus = "FILLED"     // Đã thực hiện
    OrderPartial   OrderStatus = "PARTIAL"    // Thực hiện một phần
    OrderCancelled OrderStatus = "CANCELLED"  // Đã hủy
)

// Order: Entity chính - Đơn hàng
type Order struct {
    ID        string
    UserID    string
    Symbol    *sharedkernel.Symbol
    Side      string  // "BUY" hoặc "SELL"
    Price     *sharedkernel.Money
    Quantity  float64
    Status    OrderStatus
    CreatedAt time.Time
    UpdatedAt time.Time
    Version   int  // Optimistic locking
}

// NewOrder: Factory để tạo order mới
func NewOrder(
    userID string,
    symbol *sharedkernel.Symbol,
    side string,
    price *sharedkernel.Money,
    quantity float64,
) (*Order, error) {
    // Validate input
    if userID == "" {
        return nil, fmt.Errorf("userID is required")
    }
    if side != "BUY" && side != "SELL" {
        return nil, fmt.Errorf("side must be BUY or SELL")
    }
    if quantity <= 0 {
        return nil, fmt.Errorf("quantity must be positive")
    }
    
    return &Order{
        ID:        "ORDER-" + uuid.New().String()[:8],
        UserID:    userID,
        Symbol:    symbol,
        Side:      side,
        Price:     price,
        Quantity:  quantity,
        Status:    OrderPending,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Version:   1,
    }, nil
}

// Validate: Kiểm tra order có hợp lệ không
func (o *Order) Validate() error {
    if o.ID == "" {
        return fmt.Errorf("order ID is required")
    }
    if o.UserID == "" {
        return fmt.Errorf("user ID is required")
    }
    if o.Symbol == nil {
        return fmt.Errorf("symbol is required")
    }
    if o.Price == nil {
        return fmt.Errorf("price is required")
    }
    if o.Quantity <= 0 {
        return fmt.Errorf("quantity must be positive")
    }
    return nil
}

// GetTotalValue: Tính tổng giá trị order
func (o *Order) GetTotalValue() *sharedkernel.Money {
    // totalPrice = price * quantity
    singlePrice := o.Price.Amount
    totalAmount := singlePrice * int64(o.Quantity)
    total, _ := sharedkernel.NewMoney(totalAmount, o.Price.Currency)
    return total
}

// CanCancel: Có thể hủy order không?
func (o *Order) CanCancel() bool {
    return o.Status == OrderPending || o.Status == OrderPartial
}

// Cancel: Hủy order
func (o *Order) Cancel() error {
    if !o.CanCancel() {
        return fmt.Errorf("cannot cancel order in %s status", o.Status)
    }
    o.Status = OrderCancelled
    o.UpdatedAt = time.Now()
    o.Version++
    return nil
}

// Fill: Thực hiện order
func (o *Order) Fill() error {
    if o.Status == OrderPending || o.Status == OrderPartial {
        o.Status = OrderFilled
        o.UpdatedAt = time.Now()
        o.Version++
        return nil
    }
    return fmt.Errorf("cannot fill order in %s status", o.Status)
}