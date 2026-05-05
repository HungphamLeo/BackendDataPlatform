package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"sync"
	"path/filepath"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
)

var (
	logger    *zap.Logger
	once      sync.Once
	logEntries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:   "log_entries_total",
			Help:   "Total number of log entries.",
		},
		[]string{"level"},
	)
)

func init() {
	prometheus.MustRegister(logEntries)
}

// InitLogger initializes the global logger instance with an optional file path.
func InitLogger(logFilePath string) {
	once.Do(func() {
		cfg := zap.NewProductionConfig()

		var cores []zapcore.Core
		
		// Ghi log ra Console (dễ nhìn khi chạy docker logs/local)
		consoleEncoder := zapcore.NewConsoleEncoder(cfg.EncoderConfig)
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel))

		// Ghi log ra File JSON (nếu cấu hình path)
		if logFilePath != "" {
			// Tự động tạo folder nếu chưa tồn tại
			if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err == nil {
				if file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
					fileEncoder := zapcore.NewJSONEncoder(cfg.EncoderConfig)
					cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zapcore.InfoLevel))
				}
			}
		}

		core := zapcore.NewTee(cores...)
		// AddCaller để biết chính xác dòng code nào đang in ra log
		logger = zap.New(core, zap.Hooks(prometheusHook), zap.AddCaller())
	})
}

// GetLogger returns the global logger instance.
func GetLogger() *zap.Logger {
	if logger == nil {
		InitLogger("") // Mặc định chỉ xuất Console nếu chưa init với file
	}
	return logger
}

// Sync flushes any buffered log entries.
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

func prometheusHook(entry zapcore.Entry) error {
	logEntries.WithLabelValues(entry.Level.String()).Inc()
	return nil
}

func StartPrometheusEndpoint() {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9090", nil)
}

// Logger interface to abstract logging functionality

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
}

type zapLogger struct {
	logger *zap.Logger
}

func (z *zapLogger) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}

func (z *zapLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

func (z *zapLogger) Warn(msg string, fields ...zap.Field) {
	z.logger.Warn(msg, fields...)
}

func (z *zapLogger) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
}

func (z *zapLogger) Fatal(msg string, fields ...zap.Field) {
	z.logger.Fatal(msg, fields...)
}

// With cho phép clone Logger hiện tại kèm theo các Context Metadata cố định
func (z *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{logger: z.logger.With(fields...)}
}

// NewLogger returns a new instance of Logger
func NewLogger() Logger {
	return &zapLogger{logger: GetLogger()}
}