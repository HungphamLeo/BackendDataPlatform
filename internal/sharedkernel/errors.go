package sharedkernel

import "fmt"

// DomainError: Lỗi domain (sự kiện gì không hợp lệ trong logic kinh doanh)
type DomainError struct {
    Code    string
    Message string
}

func (e *DomainError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Error definitions
var (
    ErrInsufficientBalance = &DomainError{
        Code:    "INSUFFICIENT_BALANCE",
        Message: "Account balance is insufficient for this operation",
    }
    
    ErrInvalidOrder = &DomainError{
        Code:    "INVALID_ORDER",
        Message: "Order is invalid",
    }
    
    ErrOrderNotFound = &DomainError{
        Code:    "ORDER_NOT_FOUND",
        Message: "Order not found",
    }
)