package logger

import (
	"log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log struct {
	Message  string
	Metadata map[string]interface{}
}

type Logger struct {
	base *zap.Logger
}

func (l *Logger) Info(message string, metadata ...any) {
	l.base.Info(message, convertMeta(nil, metadata)...)
}
func (l *Logger) Error(message string, err error, metadata ...any) {
	l.base.Error(message, convertMeta(err, metadata)...)
}
func (l *Logger) Debug(message string, metadata ...any) {
	l.base.Debug(message, convertMeta(nil, metadata)...)
}
func (l *Logger) Warn(message string, metadata ...any) {
	l.base.Warn(message, convertMeta(nil, metadata)...)
}
func (l *Logger) Fatal(message string, metadata ...any) {
	l.base.Fatal(message, convertMeta(nil, metadata)...)
}

func NewLogger() *Logger {
	return &Logger{
		base: initZap(),
	}
}

// this should be in a separate file zap.go
func initZap() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	var err error
	base, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		log.Fatal("Unable to initialize ZAP")
	}
	return base
}

// helper to convert metadata → zap fields
func convertMeta(err error, meta ...any) []zap.Field {
	fields := make([]zap.Field, 0, len(meta)+1)
	for _, m := range meta {
		if m != nil {
			fields = append(fields, zap.Any("metadata", m))
		}
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	return fields
}
