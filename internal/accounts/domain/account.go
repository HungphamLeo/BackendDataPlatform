package domain

import (
	"fmt"
	"time"

	"github.com/BackendDataPlatform/internal/sharedkernel"
)

// Account: Entity - Tài khoản
type Account struct {
	ID               string
	UserID           string
	Balance          *sharedkernel.Money // Tổng số dư
	ReservedBalance  *sharedkernel.Money // Số dư bị giữ (đặt order)
	AvailableBalance *sharedkernel.Money // Số dư có thể dùng
	Status           string              // ACTIVE, FROZEN, CLOSED
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Version          int
}

// NewAccount: Tạo account mới
func NewAccount(userID string, initialBalance *sharedkernel.Money) (*Account, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	if initialBalance == nil {
		return nil, fmt.Errorf("initialBalance is required")
	}

	return &Account{
		ID:               fmt.Sprintf("ACC-%s", userID)[:20],
		UserID:           userID,
		Balance:          initialBalance,
		ReservedBalance:  &sharedkernel.Money{0, initialBalance.Currency},
		AvailableBalance: initialBalance,
		Status:           "ACTIVE",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Version:          1,
	}, nil
}

// Deposit: Nạp tiền vào tài khoản
func (a *Account) Deposit(amount *sharedkernel.Money) error {
	if a.Status != "ACTIVE" {
		return fmt.Errorf("account is not active: %s", a.Status)
	}

	newBalance, err := a.Balance.Add(amount)
	if err != nil {
		return fmt.Errorf("failed to add balance: %w", err)
	}

	newAvailable, err := a.AvailableBalance.Add(amount)
	if err != nil {
		return fmt.Errorf("failed to update available balance: %w", err)
	}

	a.Balance = newBalance
	a.AvailableBalance = newAvailable
	a.UpdatedAt = time.Now()
	a.Version++

	return nil
}

// Withdraw: Rút tiền từ tài khoản
func (a *Account) Withdraw(amount *sharedkernel.Money) error {
	if a.Status != "ACTIVE" {
		return fmt.Errorf("account is not active")
	}

	// Kiểm tra có đủ tiền không
	if !a.AvailableBalance.IsGreaterThan(amount) && !a.AvailableBalance.IsEqual(amount) {
		return fmt.Errorf("insufficient balance")
	}

	newBalance, _ := a.Balance.Subtract(amount)
	newAvailable, _ := a.AvailableBalance.Subtract(amount)

	a.Balance = newBalance
	a.AvailableBalance = newAvailable
	a.UpdatedAt = time.Now()
	a.Version++

	return nil
}

// Reserve: Giữ tiền (để đặt order)
// Ví dụ: balance = 1000, reserve = 500 → available = 500
func (a *Account) Reserve(amount *sharedkernel.Money) error {
	if !a.AvailableBalance.IsGreaterThan(amount) && !a.AvailableBalance.IsEqual(amount) {
		return fmt.Errorf("insufficient available balance to reserve")
	}

	newReserved, _ := a.ReservedBalance.Add(amount)
	newAvailable, _ := a.AvailableBalance.Subtract(amount)

	a.ReservedBalance = newReserved
	a.AvailableBalance = newAvailable
	a.UpdatedAt = time.Now()
	a.Version++

	return nil
}

// Release: Thả tiền (hủy order)
func (a *Account) Release(amount *sharedkernel.Money) error {
	if !a.ReservedBalance.IsGreaterThan(amount) && !a.ReservedBalance.IsEqual(amount) {
		return fmt.Errorf("insufficient reserved balance to release")
	}

	newReserved, _ := a.ReservedBalance.Subtract(amount)
	newAvailable, _ := a.AvailableBalance.Add(amount)

	a.ReservedBalance = newReserved
	a.AvailableBalance = newAvailable
	a.UpdatedAt = time.Now()
	a.Version++

	return nil
}
