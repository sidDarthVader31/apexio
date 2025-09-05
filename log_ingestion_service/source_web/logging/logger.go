package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log struct{
	Message string
	Metadata map[string]interface{}
}

var base *zap.Logger;
func  Info(entry Log) {
	base.Info(entry.Message, convertMeta(entry.Metadata)...)
}
func Error(entry Log) {
	base.Error(entry.Message, convertMeta(entry.Metadata)...)
}
func  Debug(entry Log) {
	base.Debug(entry.Message, convertMeta(entry.Metadata)...)
}
func  Warn(entry Log) {
	base.Warn(entry.Message, convertMeta(entry.Metadata)...)
}
func  Fatal(entry Log) {
	base.Fatal(entry.Message, convertMeta(entry.Metadata)...)
}

func InitLogger() {
	initZap()
}

// this should be in a separate file zap.go
func initZap(){
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	var err error
	base, err = cfg.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	if(err!=nil){
		log.Fatal("Unable to initialize ZAP")
	}
}
 // helper to convert metadata → zap fields
func convertMeta(meta map[string]interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(meta))
	for k, v := range meta {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
