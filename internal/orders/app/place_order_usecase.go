package app

import (
    "context"
    "fmt"
    "time"
    "github.com/BackendDataPlatform/internal/orders/domain"
    "github.com/BackendDataPlatform/internal/accounts/domain"
    "github.com/BackendDataPlatform/internal/sharedkernel"
    "github.com/BackendDataPlatform/internal/platform/idempotency"
)

// PlaceOrderCommand: Input từ client (HTTP request)
type PlaceOrderCommand struct {
    UserID   string
    Symbol   string  // "AAPL"
    Side     string  // "BUY" hoặc "SELL"
    Price    int64   // 18000 = 180.00
    Quantity float64 // 10
}

// PlaceOrderResult: Output trả về client (HTTP response)
type PlaceOrderResult struct {
    OrderID   string `json:"order_id"`
    Symbol    string `json:"symbol"`
    Side      string `json:"side"`
    Price     int64  `json:"price"`
    Quantity  float64 `json:"quantity"`
    Status    string `json:"status"`
    CreatedAt string `json:"created_at"`
}

// PlaceOrderUseCase: Use case chính - Đặt lệnh mua/bán
type PlaceOrderUseCase struct {
    orderRepo      domain.OrderRepository
    accountRepo    domain.AccountRepository  // từ accounts service
    eventPublisher domain.EventPublisher
    idempotency    idempotency.IdempotencyStore
}

// NewPlaceOrderUseCase: Constructor
func NewPlaceOrderUseCase(
    orderRepo domain.OrderRepository,
    accountRepo domain.AccountRepository,
    eventPublisher domain.EventPublisher,
    idempotency idempotency.IdempotencyStore,
) *PlaceOrderUseCase {
    return &PlaceOrderUseCase{
        orderRepo:      orderRepo,
        accountRepo:    accountRepo,
        eventPublisher: eventPublisher,
        idempotency:    idempotency,
    }
}

// Execute: Thực hiện case - Đặt lệnh
func (uc *PlaceOrderUseCase) Execute(
    ctx context.Context,
    cmd *PlaceOrderCommand,
    idempotencyKey string,
) (*PlaceOrderResult, error) {
    // ✅ BƯỚC 1: Kiểm tra idempotency
    // Nếu client gửi 2 request giống nhau, chỉ xử lý 1 lần
    cachedResult, err := uc.idempotency.Get(ctx, idempotencyKey)
    if err == nil && cachedResult != nil {
        return cachedResult.(*PlaceOrderResult), nil
    }
    
    // ✅ BƯỚC 2: Validate input
    if cmd.UserID == "" || cmd.Symbol == "" || cmd.Price <= 0 || cmd.Quantity <= 0 {
        return nil, fmt.Errorf("invalid input: missing required fields")
    }
    
    // ✅ BƯỚC 3: Lấy tài khoản
    account, err := uc.accountRepo.GetByID(ctx, cmd.UserID)
    if err != nil {
        return nil, fmt.Errorf("account not found: %w", err)
    }
    
    // ✅ BƯỚC 4: Tạo Money object
    orderPrice, err := sharedkernel.NewMoney(cmd.Price, "USD")
    if err != nil {
        return nil, fmt.Errorf("invalid price: %w", err)
    }
    
    // ✅ BƯỚC 5: Tạo Symbol object
    symbol, err := sharedkernel.NewSymbol(cmd.Symbol)
    if err != nil {
        return nil, fmt.Errorf("invalid symbol: %w", err)
    }
    
    // ✅ BƯỚC 6: Tính tổng giá (price * quantity)
    totalPrice := orderPrice.Amount * int64(cmd.Quantity)
    totalMoney, _ := sharedkernel.NewMoney(totalPrice, "USD")
    
    // ✅ BƯỚC 7: Kiểm tra có đủ tiền không
    if !account.AvailableBalance.IsGreaterThan(totalMoney) && !account.AvailableBalance.IsEqual(totalMoney) {
        return nil, fmt.Errorf("insufficient balance: need %v, have %v",
            totalMoney, account.AvailableBalance)
    }
    
    // ✅ BƯỚC 8: Tạo order mới (domain logic)
    order, err := domain.NewOrder(cmd.UserID, symbol, cmd.Side, orderPrice, cmd.Quantity)
    if err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }
    
    // ✅ BƯỚC 9: Validate order
    if err := order.Validate(); err != nil {
        return nil, fmt.Errorf("invalid order: %w", err)
    }
    
    // ✅ BƯỚC 10: Lưu order vào database
    if err := uc.orderRepo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to save order: %w", err)
    }
    
    // ✅ BƯỚC 11: Reserve balance (giữ tiền)
    if err := account.Reserve(totalMoney); err != nil {
        return nil, fmt.Errorf("failed to reserve balance: %w", err)
    }
    
    // ✅ BƯỚC 12: Lưu account (với balance đã bị giữ)
    if err := uc.accountRepo.Update(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to update account: %w", err)
    }
    
    // ✅ BƯỚC 13: Tạo event (sự kiện) để gửi
    event := map[string]interface{}{
        "order_id": order.ID,
        "user_id": order.UserID,
        "symbol": order.Symbol.String(),
        "side": order.Side,
        "price": order.Price.Amount,
        "quantity": order.Quantity,
        "created_at": order.CreatedAt.Unix(),
    }
    
    // ✅ BƯỚC 14: Gửi event (Kafka)
    if err := uc.eventPublisher.Publish(ctx, "orders.placed", event); err != nil {
        fmt.Printf("Warning: failed to publish event: %v\n", err)
        // Không return error, event sẽ retry
    }
    
    // ✅ BƯỚC 15: Tạo result để trả về
    result := &PlaceOrderResult{
        OrderID:   order.ID,
        Symbol:    order.Symbol.String(),
        Side:      order.Side,
        Price:     order.Price.Amount,
        Quantity:  order.Quantity,
        Status:    string(order.Status),
        CreatedAt: order.CreatedAt.Format(time.RFC3339),
    }
    
    // ✅ BƯỚC 16: Lưu vào cache (idempotency)
    _ = uc.idempotency.Set(ctx, idempotencyKey, result, 24*time.Hour)
    
    return result, nil
}