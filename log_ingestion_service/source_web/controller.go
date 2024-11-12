package main

import (
	"encoding/json"
	"fmt"
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
  Request_id string                 `json:"requestId"` 
  Client_ip  string                 `json:"clientIp"` 
  User_agent string                 `json:"userAgent"` 
  Request_method string             `json:"requestMethod"` 
  Request_path string               `json:"requestPath"` 
  Response_status string            `json:"responseStatus"` 
  Response_duration string          `json:"responseDuration"` 
  Extra map[string] string          `json:"extra"` 
}

type source struct{
  Host string                       `json:"host"` 
  Service string                    `json:"service"` 
  Environment string                `json:"environment"` 
  Extra map[string] string          `json:"extra"` 
}
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
