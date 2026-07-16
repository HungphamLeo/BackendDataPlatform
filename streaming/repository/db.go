package repository

import (
	"context"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/HungphamLeo/BackendDataPlatform/streaming/models"
)

// DB wraps gorm.DB for the streaming service.
type DB struct {
	db *gorm.DB
}

// New opens a MySQL connection, auto-migrates streaming tables and returns a DB.
func New(dsn string) (*DB, error) {
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := gormDB.AutoMigrate(
		&models.MarketTicker{},
		&models.MarketOrderBook{},
		&models.MarketKline{},
	); err != nil {
		return nil, err
	}

	return &DB{db: gormDB}, nil
}

// SaveTicker upserts a ticker record.
func (r *DB) SaveTicker(ctx context.Context, t *models.MarketTicker) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// SaveOrderBook inserts an orderbook snapshot.
func (r *DB) SaveOrderBook(ctx context.Context, ob *models.MarketOrderBook) error {
	return r.db.WithContext(ctx).Create(ob).Error
}

// SaveKline inserts a kline record.
func (r *DB) SaveKline(ctx context.Context, k *models.MarketKline) error {
	return r.db.WithContext(ctx).Create(k).Error
}

// GetLatestTicker fetches the most recent ticker for symbol+exchange.
func (r *DB) GetLatestTicker(ctx context.Context, symbol, exchange string) (*models.MarketTicker, error) {
	var t models.MarketTicker
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND exchange = ?", symbol, exchange).
		Order("timestamp DESC").
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetLatestOrderBook fetches the most recent orderbook snapshot.
func (r *DB) GetLatestOrderBook(ctx context.Context, symbol, exchange string) (*models.MarketOrderBook, error) {
	var ob models.MarketOrderBook
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND exchange = ?", symbol, exchange).
		Order("timestamp DESC").
		First(&ob).Error
	if err != nil {
		return nil, err
	}
	return &ob, nil
}

// GetKlines fetches klines for a given symbol/exchange/interval within a time range.
func (r *DB) GetKlines(ctx context.Context, symbol, exchange, interval string, from, to time.Time, limit int) ([]models.MarketKline, error) {
	query := r.db.WithContext(ctx).
		Where("symbol = ? AND exchange = ? AND interval = ?", symbol, exchange, interval).
		Where("timestamp BETWEEN ? AND ?", from, to).
		Order("timestamp ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var klines []models.MarketKline
	if err := query.Find(&klines).Error; err != nil {
		return nil, err
	}
	return klines, nil
}
