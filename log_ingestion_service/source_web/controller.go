package main

import (
	"encoding/json"
	"net/http"
	"github.com/gin-gonic/gin"
)

type LogInfo struct {
	Id        uint                    `json:"id"`
	Metadata  metadata                `json:"metadata"` 
  Timestamp uint64                  `json:"timestamp"`
  Loglevel  string                  `json:"logLevel"`
  Message   string                  `json:"message"`
  Source    source                  `json:"source"`
}


type metadata struct {
  request_id string                 `json:"requestId"` 
  client_ip  string                 `json:"clientIp"` 
  user_agent string                 `json:"userAgent"` 
  request_method string             `json:"requestMethod"` 
  request_path string               `json:"requestPath"` 
  response_status string            `json:"responseStatus"` 
  response_duration string          `json:"responseDuration"` 
  extra map[string] string          `json:"extra"` 
}

type source struct{
  host string                       `json:"host"` 
  service string                    `json:"service"` 
  environment string                `json:"environment"` 
  extra map[string] string          `json:"extra"` 
}
func ingestData(c *gin.Context) {
  var logInfo LogInfo
  if err := c.BindJSON(&logInfo); err != nil {
        return
    }
  //ingest loginfo to kafka
  value, err := json.Marshal(logInfo)
  if err!= nil{
    c.IndentedJSON(http.StatusInternalServerError, err)
  }
  ingestToKafka(value, logTopic)
 c.IndentedJSON(http.StatusCreated, logInfo)
}
