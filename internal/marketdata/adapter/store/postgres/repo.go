package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/aggregate"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/storage"
)

// ✅ REPOSITORY IMPLEMENTATION (Infrastructure)
type PostgresTickerRepository struct {
	db storage.Database
}

func NewPostgresTickerRepository(db storage.Database) *PostgresTickerRepository {
	return &PostgresTickerRepository{db: db}
}

// Save ticker to PostgreSQL
func (r *PostgresTickerRepository) Save(ctx context.Context, ticker *aggregate.Ticker) error {
	query := `
		INSERT INTO tickers (id, symbol, exchange, last_price, last_update)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			last_price = $4,
			last_update = $5
	`

	_, err := r.db.Exec(ctx, query,
		ticker.ID,
		ticker.Symbol.String(),
		ticker.Exchange,
		ticker.LastPrice.MidPrice(),
		ticker.LastUpdate,
	)

	if err != nil {
		return fmt.Errorf("failed to save ticker: %w", err)
	}

	return nil
}

// Get ticker from PostgreSQL
func (r *PostgresTickerRepository) GetBySymbol(
	ctx context.Context,
	symbol value_object.Symbol,
) (*aggregate.Ticker, error) {
	query := `
		SELECT id, symbol, exchange, last_price, last_update
		FROM tickers
		WHERE symbol = $1
	`

	row := r.db.QueryRow(ctx, query, symbol.String())

	var tickerID string
	var symbolStr string
	var exchange string
	var lastPrice float64
	var lastUpdate sql.NullTime

	if err := row.Scan(&tickerID, &symbolStr, &exchange, &lastPrice, &lastUpdate); err != nil {
		if err == sql.ErrNoRows {
			// Create new ticker
			return aggregate.NewTicker(symbol, exchange), nil
		}
		return nil, fmt.Errorf("failed to query ticker: %w", err)
	}

	ticker := aggregate.NewTicker(symbol, exchange)
	return ticker, nil
}
