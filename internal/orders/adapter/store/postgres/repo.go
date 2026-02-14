package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/BackendDataPlatform/internal/orders/domain"
)

// PostgresOrderRepository: Implement OrderRepository với PostgreSQL
type PostgresOrderRepository struct {
	db *sql.DB
	mu sync.RWMutex // Để thread-safe
}

// NewPostgresOrderRepository: Constructor
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

// Save: Lưu order vào database
func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	query := `
        INSERT INTO orders (id, user_id, symbol, side, price, quantity, status, created_at, updated_at, version)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (id) DO NOTHING
        RETURNING id
    `

	var returnedID string
	err := r.db.QueryRowContext(
		ctx,
		query,
		order.ID,
		order.UserID,
		order.Symbol.String(),
		order.Side,
		order.Price.Amount,
		order.Quantity,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
		order.Version,
	).Scan(&returnedID)

	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	return nil
}

// GetByID: Lấy order từ database
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
        SELECT id, user_id, symbol, side, price, quantity, status, created_at, updated_at, version
        FROM orders
        WHERE id = $1
    `

	var (
		o         domain.Order
		symbolStr string
		statusStr string
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID,
		&o.UserID,
		&symbolStr,
		&o.Side,
		&o.Price,
		&o.Quantity,
		&statusStr,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.Version,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Chuyển string → object
	symbol, _ := sharedkernel.NewSymbol(symbolStr)
	o.Symbol = symbol
	o.Status = domain.OrderStatus(statusStr)

	return &o, nil
}

// GetByUserID: Lấy tất cả order của user
func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
        SELECT id, user_id, symbol, side, price, quantity, status, created_at, updated_at, version
        FROM orders
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var (
			o         domain.Order
			symbolStr string
			statusStr string
		)

		err := rows.Scan(
			&o.ID,
			&o.UserID,
			&symbolStr,
			&o.Side,
			&o.Price,
			&o.Quantity,
			&statusStr,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Version,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		symbol, _ := sharedkernel.NewSymbol(symbolStr)
		o.Symbol = symbol
		o.Status = domain.OrderStatus(statusStr)

		orders = append(orders, &o)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %w", err)
	}

	return orders, nil
}

// Update: Cập nhật order (với optimistic locking)
func (r *PostgresOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
        UPDATE orders
        SET status = $1, updated_at = $2, version = $3
        WHERE id = $4 AND version = $5
        RETURNING id
    `

	var returnedID string
	err := r.db.QueryRowContext(
		ctx,
		query,
		order.Status,
		order.UpdatedAt,
		order.Version,
		order.ID,
		order.Version-1, // Kiểm tra version cũ
	).Scan(&returnedID)

	if err == sql.ErrNoRows {
		return fmt.Errorf("order not found or version conflict")
	}
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// Delete: Xóa order
func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM orders WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}
