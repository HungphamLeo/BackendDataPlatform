package logging

import (
	"go.uber.org/zap/zapcore"
	"net/http"
	"sync"
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

// InitLogger initializes the global logger instance.
func InitLogger() {
	once.Do(func() {
		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout"} // Can be extended for Loki
		cfg.ErrorOutputPaths = []string{"stderr"}

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zapcore.AddSync(zapcore.Lock(os.Stdout)),
			zapcore.DebugLevel,
		)

		logger = zap.New(core, zap.Hooks(prometheusHook))
	})
}

// GetLogger returns the global logger instance.
func GetLogger() *zap.Logger {
	if logger == nil {
		InitLogger()
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

// NewLogger returns a new instance of Logger
func NewLogger() Logger {
	return &zapLogger{logger: GetLogger()}
}