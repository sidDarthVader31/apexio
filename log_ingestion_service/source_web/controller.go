package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)
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
func ingestData(c *gin.Context) {
  var logInfo LogInfo
  if err := c.BindJSON(&logInfo); err != nil {
        return
    }
  //ingest loginfo to kafka
  value, err := json.Marshal(logInfo)
  if err!= nil{
    fmt.Println("error converting to json", err)
    c.IndentedJSON(http.StatusInternalServerError, err)
  }
  success := ingestToKafka(value, logTopic)
  if success == false {
    c.IndentedJSON(http.StatusNotFound, logInfo)
  }
 c.IndentedJSON(http.StatusCreated, logInfo)
}
