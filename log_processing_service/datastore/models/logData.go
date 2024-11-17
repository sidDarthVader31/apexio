package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log-processor/datastore"

	"github.com/elastic/go-elasticsearch/v8/esapi"
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

