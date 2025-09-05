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
func  Info(message string, metadata ...map[string]interface{}) {
	base.Info(message, convertMeta(metadata[0])...)
}
func Error(message string,err error, metadata ...map[string] interface{}) {
	base.Error(message, convertMeta(metadata[0])...)
}
func  Debug(message string, metadata ...map[string]interface{} ) {
	base.Debug(message, convertMeta(metadata[0])...)
}
func  Warn(message string, metadata ...map[string]interface{}) {
	base.Warn(message, convertMeta(metadata[0])...)
}
func  Fatal(message string, metadata ...map[string]interface{}) {
	base.Fatal(message, convertMeta(metadata[0])...)
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
func convertMeta(meta map[string]interface{}, err ...error) []zap.Field {
	fields := make([]zap.Field, 0, len(meta))
	for k, v := range meta {
		fields = append(fields, zap.Any(k, v))
	}

	if len(err)>0 {
		fields = append(fields, zap.Error(err[0]))
	}
	return fields
}
