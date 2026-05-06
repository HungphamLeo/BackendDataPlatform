package logging

import "go.uber.org/zap"

// LogTopic định nghĩa các khu vực nghiệp vụ trong hệ thống để phân loại log
type LogTopic string

const (
	TopicMarketData LogTopic = "MARKET_DATA"
	TopicOrder      LogTopic = "ORDER"
	TopicAccount    LogTopic = "ACCOUNT"
	TopicSystem     LogTopic = "SYSTEM"
)

// EventLogSchema là chuẩn Schema thống nhất cho toàn bộ Log Event trên hệ thống
type EventLogSchema struct {
	EventID   string      `json:"event_id,omitempty"` // ID của trace, request hoặc transaction nếu có
	Topic     LogTopic    `json:"topic"`              // Chủ đề log
	Service   string      `json:"service"`            // Tên microservice phát ra log
	Action    string      `json:"action"`             // Hành động (Ví dụ: "SUBSCRIBE_WS", "PLACE_ORDER")
	Status    string      `json:"status"`             // Trạng thái: "SUCCESS", "FAILED", "PENDING"
	Payload   interface{} `json:"payload,omitempty"`  // Data chi tiết hoặc error message
}

// LogEvent là hàm Helper giúp ép chuẩn các Dev khác trong team phải ghi log đúng định dạng
func LogEvent(logger Logger, level string, schema EventLogSchema, msg string) {
	fields := []zap.Field{
		zap.String("topic", string(schema.Topic)),
		zap.String("service", schema.Service),
		zap.String("action", schema.Action),
		zap.String("status", schema.Status),
	}

	if schema.EventID != "" {
		fields = append(fields, zap.String("event_id", schema.EventID))
	}
	if schema.Payload != nil {
		fields = append(fields, zap.Any("payload", schema.Payload))
	}

	switch level {
	case "DEBUG":
		logger.Debug(msg, fields...)
	case "WARN":
		logger.Warn(msg, fields...)
	case "ERROR":
		logger.Error(msg, fields...)
	case "FATAL":
		logger.Fatal(msg, fields...)
	default: // INFO là mặc định
		logger.Info(msg, fields...)
	}
}