package logging

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log struct{
	Message string
	Metadata map[string]interface{}
}
type LoggerStruct struct {
	base *zap.Logger
}
func (l *LoggerStruct) Info(entry Log) {
	l.base.Info(entry.Message, convertMeta(entry.Metadata)...)
}
func (l *LoggerStruct) Error(entry Log) {
	l.base.Error(entry.Message, convertMeta(entry.Metadata)...)
}
func (l *LoggerStruct) Debug(entry Log) {
	l.base.Debug(entry.Message, convertMeta(entry.Metadata)...)
}
func (l *LoggerStruct) Warn(entry Log) {
	l.base.Warn(entry.Message, convertMeta(entry.Metadata)...)
}
func (l *LoggerStruct) Fatal(entry Log) {
	l.base.Fatal(entry.Message, convertMeta(entry.Metadata)...)
}


var Logger *LoggerStruct

func main() {
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

	base, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	if(err!=nil){
		log.Fatal("Unable to initialize ZAP")
	}
	Logger = &LoggerStruct{base: base}
}
 // helper to convert metadata → zap fields
func convertMeta(meta map[string]interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(meta))
	for k, v := range meta {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
