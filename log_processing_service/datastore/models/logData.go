package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log-processor/datastore"
	"net"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type LogInfo struct {
	Id        uint                    `json:"id"`
	Metadata  Metadata                `json:"metadata"` 
  Timestamp uint64                  `json:"timestamp"`
  Loglevel  LogLevel                  `json:"logLevel"`
  Message   string                  `json:"message"`
  Source    Source                 `json:"source"`
}


type Metadata struct {
  RequestId string                 `json:"requestId"` 
  ClientIp  net.IP                 `json:"clientIp"` 
  UserAgent string                 `json:"userAgent"` 
  RequestMethod string             `json:"requestMethod"` 
  RequestPath string               `json:"requestPath"` 
  ResponseStatus int            `json:"responseStatus"` 
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
func (l *LogInfo) Insert(){
  data, e := json.Marshal(l)

  if e!= nil{
    fmt.Println("error while converting to json:", e)
  }
  req := esapi.IndexRequest{
		Index:      "raw_logs", // Index name
		Body:       bytes.NewReader(data),
		Refresh:    "true", // Make the document immediately available for search
	}
  res, err := req.Do(context.Background(), datastore.Es)
  if err != nil {
		log.Fatalf("Error indexing document: %s", err)
	}
	defer res.Body.Close()
	// Check if the response is successful
	if res.IsError() {
		log.Printf("Error indexing document: %s", res.String())
	} else {
		fmt.Printf("Document indexed successfully: %s\n", res.String())
	}
}

