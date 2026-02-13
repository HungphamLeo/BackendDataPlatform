package sharedkernel

import (
    "fmt"
    "regexp"
)

// Symbol: Mã chứng chỉ (AAPL, GOOGL, ...)
type Symbol struct {
    Code string
}

// NewSymbol: Tạo Symbol mới (phải đúng format)
func NewSymbol(code string) (*Symbol, error) {
    if len(code) < 1 || len(code) > 10 {
        return nil, fmt.Errorf("symbol must be 1-10 characters")
    }
    
    // Kiểm tra chỉ chứa chữ cái
    if matched, _ := regexp.MatchString(`^[A-Z0-9]+$`, code); !matched {
        return nil, fmt.Errorf("symbol must be uppercase letters/numbers only")
    }
    
    return &Symbol{Code: code}, nil
}

// String: Để in ra
func (s *Symbol) String() string {
    return s.Code
}

// IsEqual: Hai Symbol có bằng nhau không
func (s *Symbol) IsEqual(other *Symbol) bool {
    return s.Code == other.Code
}