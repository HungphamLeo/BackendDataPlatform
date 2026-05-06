package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/application/port/out"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/aggregate"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// TickerEntity là entity trong database
type TickerEntity struct {
	ID           string    `gorm:"column:id;primaryKey"`
	Symbol       string    `gorm:"column:symbol;index:idx_ticker_symbol_exchange,unique"`
	Exchange     string    `gorm:"column:exchange;index:idx_ticker_symbol_exchange,unique"`
	LastPrice    float64   `gorm:"column:last_price"`
	BestBid      float64   `gorm:"column:best_bid"`
	BestAsk      float64   `gorm:"column:best_ask"`
	BestBidSize  float64   `gorm:"column:best_bid_size"`
	BestAskSize  float64   `gorm:"column:best_ask_size"`
	High24h      float64   `gorm:"column:high_24h"`
	Low24h       float64   `gorm:"column:low_24h"`
	Volume24h    float64   `gorm:"column:volume_24h"`
	QuoteVolume  float64   `gorm:"column:quote_volume"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName trả về tên bảng trong database
func (TickerEntity) TableName() string {
	return "tickers"
}

// TickerRepository triển khai TickerRepositoryPort
type TickerRepository struct {
	db *gorm.DB
}

// NewTickerRepository tạo mới TickerRepository
func NewTickerRepository(db *gorm.DB) *TickerRepository {
	return &TickerRepository{
		db: db,
	}
}

// Đảm bảo TickerRepository triển khai interface TickerRepositoryPort
var _ out.TickerRepositoryPort = (*TickerRepository)(nil)

// FindByID tìm Ticker theo ID
func (r *TickerRepository) FindByID(ctx context.Context, id aggregate.TickerID) (*aggregate.Ticker, error) {
	var entity TickerEntity
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return r.mapToDomain(&entity)
}

// FindBySymbol tìm Ticker theo Symbol và Exchange
func (r *TickerRepository) FindBySymbol(ctx context.Context, symbol value_object.Symbol, exchange string) (*aggregate.Ticker, error) {
	var entity TickerEntity
	if err := r.db.WithContext(ctx).Where("symbol = ? AND exchange = ?", symbol.String(), exchange).First(&entity).Error; err != nil {
		return nil, err
	}
	return r.mapToDomain(&entity)
}

// FindAll tìm tất cả Ticker
func (r *TickerRepository) FindAll(ctx context.Context) ([]*aggregate.Ticker, error) {
	var entities []TickerEntity
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	
	tickers := make([]*aggregate.Ticker, len(entities))
	for i, entity := range entities {
		ticker, err := r.mapToDomain(&entity)
		if err != nil {
			return nil, err
		}
		tickers[i] = ticker
	}
	
	return tickers, nil
}

// FindByExchange tìm tất cả Ticker theo Exchange
func (r *TickerRepository) FindByExchange(ctx context.Context, exchange string) ([]*aggregate.Ticker, error) {
	var entities []TickerEntity
	if err := r.db.WithContext(ctx).Where("exchange = ?", exchange).Find(&entities).Error; err != nil {
		return nil, err
	}
	
	tickers := make([]*aggregate.Ticker, len(entities))
	for i, entity := range entities {
		ticker, err := r.mapToDomain(&entity)
		if err != nil {
			return nil, err
		}
		tickers[i] = ticker
	}
	
	return tickers, nil
}

// Save lưu Ticker
func (r *TickerRepository) Save(ctx context.Context, ticker *aggregate.Ticker) error {
	entity := r.mapToEntity(ticker)
	
	// Sử dụng Upsert để tạo mới hoặc cập nhật
	return r.db.WithContext(ctx).Save(&entity).Error
}

// Delete xóa Ticker
func (r *TickerRepository) Delete(ctx context.Context, id aggregate.TickerID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&TickerEntity{}).Error
}

// mapToDomain chuyển đổi từ entity sang domain model
func (r *TickerRepository) mapToDomain(entity *TickerEntity) (*aggregate.Ticker, error) {
	symbol, err := value_object.NewSymbol(entity.Symbol)
	if err != nil {
		return nil, fmt.Errorf("invalid symbol: %w", err)
	}
	
	return &aggregate.Ticker{
		ID:           aggregate.TickerID(entity.ID),
		Symbol:       symbol,
		Exchange:     entity.Exchange,
		LastPrice:    entity.LastPrice,
		BestBid:      entity.BestBid,
		BestAsk:      entity.BestAsk,
		BestBidSize:  entity.BestBidSize,
		BestAskSize:  entity.BestAskSize,
		High24h:      entity.High24h,
		Low24h:       entity.Low24h,
		Volume24h:    entity.Volume24h,
		QuoteVolume:  entity.QuoteVolume,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}, nil
}

// mapToEntity chuyển đổi từ domain model sang entity
func (r *TickerRepository) mapToEntity(ticker *aggregate.Ticker) *TickerEntity {
	return &TickerEntity{
		ID:           string(ticker.ID),
		Symbol:       ticker.Symbol.String(),
		Exchange:     ticker.Exchange,
		LastPrice:    ticker.LastPrice,
		BestBid:      ticker.BestBid,
		BestAsk:      ticker.BestAsk,
		BestBidSize:  ticker.BestBidSize,
		BestAskSize:  ticker.BestAskSize,
		High24h:      ticker.High24h,
		Low24h:       ticker.Low24h,
		Volume24h:    ticker.Volume24h,
		QuoteVolume:  ticker.QuoteVolume,
		CreatedAt:    ticker.CreatedAt,
		UpdatedAt:    ticker.UpdatedAt,
	}
}
