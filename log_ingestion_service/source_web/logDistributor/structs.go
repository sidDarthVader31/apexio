package logDistributor
type LogInfo struct {
	Id        uint                    `json:"id"`
	Metadata  Metadata                `json:"metadata"` 
  Timestamp uint64                  `json:"timestamp"`
  Loglevel  string                  `json:"logLevel"`
  Message   string                  `json:"message"`
  Source    Source                  `json:"source"`
}


type Metadata struct {
  RequestId string                  `json:"requestId"` 
  ClientIp  string                  `json:"clientIp"` 
  UserAgent string                  `json:"userAgent"` 
  RequestMethod string              `json:"requestMethod"` 
  RequestPath string                `json:"requestPath"` 
  ResponseStatus int                `json:"responseStatus"` 
  ResponseDuration float32          `json:"responseDuration"` 
  Extra map[string] string          `json:"extra"` 
}

type Source struct{
  Host string                       `json:"host"` 
  Service string                    `json:"service"` 
  Environment string                `json:"environment"` 
  Extra map[string] string          `json:"extra"` 
}
type LogLevel string
// Constants for LogLevel
const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarn    LogLevel = "WARN"
	LogLevelError   LogLevel = "ERROR"
	LogLevelFatal   LogLevel = "FATAL"
)

