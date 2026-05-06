package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/HungphamLeo/BackendDataPlatform/internal/platform/config"
)

// Fields là alias cho map[string]interface{} để sử dụng với logger
type Fields map[string]interface{}

// Logger định nghĩa interface cho logging
type Logger interface {
	Debug(msg string, fields Fields)
	Info(msg string, fields Fields)
	Warn(msg string, fields Fields)
	Error(msg string, fields Fields)
	Fatal(msg string, fields Fields)
	With(fields Fields) Logger
}

// zapLogger triển khai Logger interface sử dụng zap
type zapLogger struct {
	logger *zap.Logger
}

// NewLogger tạo mới Logger từ cấu hình
func NewLogger(cfg *config.LoggingConfig) (Logger, error) {
	// Tạo encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	// Tạo encoder dựa trên format
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}
	
	// Tạo writer dựa trên output
	var writer zapcore.WriteSyncer
	if cfg.Output == "file" {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writer = zapcore.AddSync(file)
	} else {
		writer = zapcore.AddSync(os.Stdout)
	}
	
	// Tạo level dựa trên cấu hình
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}
	
	// Tạo core
	core := zapcore.NewCore(encoder, writer, level)
	
	// Tạo logger
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	
	return &zapLogger{logger: logger}, nil
}

// Debug logs a message at debug level
func (l *zapLogger) Debug(msg string, fields Fields) {
	if fields == nil {
		l.logger.Debug(msg)
		return
	}
	l.logger.Debug(msg, fieldsToZapFields(fields)...)
}

// Info logs a message at info level
func (l *zapLogger) Info(msg string, fields Fields) {
	if fields == nil {
		l.logger.Info(msg)
		return
	}
	l.logger.Info(msg, fieldsToZapFields(fields)...)
}

// Warn logs a message at warn level
func (l *zapLogger) Warn(msg string, fields Fields) {
	if fields == nil {
		l.logger.Warn(msg)
		return
	}
	l.logger.Warn(msg, fieldsToZapFields(fields)...)
}

// Error logs a message at error level
func (l *zapLogger) Error(msg string, fields Fields) {
	if fields == nil {
		l.logger.Error(msg)
		return
	}
	l.logger.Error(msg, fieldsToZapFields(fields)...)
}

// Fatal logs a message at fatal level
func (l *zapLogger) Fatal(msg string, fields Fields) {
	if fields == nil {
		l.logger.Fatal(msg)
		return
	}
	l.logger.Fatal(msg, fieldsToZapFields(fields)...)
}

// With returns a logger with the specified fields
func (l *zapLogger) With(fields Fields) Logger {
	if fields == nil {
		return l
	}
	return &zapLogger{logger: l.logger.With(fieldsToZapFields(fields)...)}
}

// fieldsToZapFields chuyển đổi Fields sang zap.Field
func fieldsToZapFields(fields Fields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}
