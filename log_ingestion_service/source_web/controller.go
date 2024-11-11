package main

import (
	"encoding/json"
	"net/http"
	"github.com/gin-gonic/gin"
)

type LogInfo struct {
	Id        uint                    `json:"id"`
	Metadata  map[string]string       `json:"metadata"` 
  Timestamp uint64                  `json:"timestamp"`
  Loglevel  string                  `json:"logLevel"`
  message   string                  `json:"message"`
  Source    map[string]string       `json:"source"`
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
