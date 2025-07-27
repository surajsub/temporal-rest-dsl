package logger

import (
	"go.uber.org/zap"
)

// ZapAdapter adapts Zap logger for Temporal's log.Logger interface
type ZapAdapter struct {
	zapLogger *zap.Logger
}

// NewZapAdapter creates a new ZapAdapter
func NewZapAdapter(zapLogger *zap.Logger) *ZapAdapter {
	return &ZapAdapter{zapLogger: zapLogger}
}

// Debug logs a debug message
func (z *ZapAdapter) Debug(msg string, keyvals ...any) {
	z.zapLogger.Debug(msg, zapFields(keyvals)...)
}

// Info logs an info message
func (z *ZapAdapter) Info(msg string, keyvals ...any) {
	z.zapLogger.Info(msg, zapFields(keyvals)...)
}

// Warn logs a warning message
func (z *ZapAdapter) Warn(msg string, keyvals ...any) {
	z.zapLogger.Warn(msg, zapFields(keyvals)...)
}

// Error logs an error message
func (z *ZapAdapter) Error(msg string, keyvals ...any) {
	z.zapLogger.Error(msg, zapFields(keyvals)...)
}

// zapFields converts key-value pairs into Zap fields
func zapFields(keyvals []any) []zap.Field {
	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals)-1; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = "unknown_key"
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}
	return fields
}
