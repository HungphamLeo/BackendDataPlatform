package out

import (
	"context"
	"time"

	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/entity"
	"github.com/HungphamLeo/BackendDataPlatform/internal/marketdata/domain/value_object"
)

// MarketDataProviderPort định nghĩa port đầu ra cho việc lấy dữ liệu thị trường từ các sàn giao dịch
type MarketDataProviderPort interface {
	// Name trả về tên của provider
	Name() string
	
	// Connect kết nối đến provider
	Connect(ctx context.Context) error
	
	// Disconnect ngắt kết nối khỏi provider
	Disconnect(ctx context.Context) error
	
	// SubscribeTicker đăng ký nhận dữ liệu ticker
	SubscribeTicker(ctx context.Context, symbols []value_object.Symbol) error
	
	// SubscribeTrades đăng ký nhận dữ liệu giao dịch
	SubscribeTrades(ctx context.Context, symbols []value_object.Symbol) error
	
	// SubscribeOrderBook đăng ký nhận dữ liệu sổ lệnh
	SubscribeOrderBook(ctx context.Context, symbols []value_object.Symbol, depth int) error
	
	// SubscribeCandles đăng ký nhận dữ liệu nến
	SubscribeCandles(ctx context.Context, symbols []value_object.Symbol, timeframes []value_object.Timeframe) error
	
	// GetTicker lấy dữ liệu ticker
	GetTicker(ctx context.Context, symbol value_object.Symbol) (*value_object.Quote, error)
	
	// GetTickers lấy dữ liệu ticker cho nhiều symbol
	GetTickers(ctx context.Context, symbols []value_object.Symbol) (map[string]*value_object.Quote, error)
	
	// GetTrades lấy dữ liệu giao dịch
	GetTrades(ctx context.Context, symbol value_object.Symbol, limit int) ([]*entity.Trade, error)
	
	// GetHistoricalTrades lấy dữ liệu giao dịch lịch sử
	GetHistoricalTrades(ctx context.Context, symbol value_object.Symbol, fromID string, limit int) ([]*entity.Trade, error)
	
	// GetOrderBook lấy dữ liệu sổ lệnh
	GetOrderBook(ctx context.Context, symbol value_object.Symbol, depth int) (*entity.OrderBook, error)
	
	// GetCandles lấy dữ liệu nến
	GetCandles(
		ctx context.Context,
		symbol value_object.Symbol,
		timeframe value_object.Timeframe,
		startTime, endTime time.Time,
		limit int,
	) ([]*entity.Candle, error)
	
	// GetExchangeInfo lấy thông tin sàn giao dịch
	GetExchangeInfo(ctx context.Context) (map[string]interface{}, error)
	
	// IsConnected kiểm tra xem đã kết nối chưa
	IsConnected() bool
}
