package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log-processor/constants"
	"log-processor/datastore"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

/**
elastic search index mapping -
{
  "mappings": {
    "properties": {
      "id": {
        "type": "long"
      },
      "metadata": {
        "properties": {
          "requestId": {
            "type": "keyword"
          },
          "clientIp": {
            "type": "string"
          },
          "userAgent": {
            "type": "text",
            "fields": {
              "keyword": {
                "type": "keyword",
                "ignore_above": 256
              }
            }
          },
          "requestMethod": {
            "type": "keyword"
          },
          "requestPath": {
            "type": "text",
            "fields": {
              "keyword": {
                "type": "keyword",
                "ignore_above": 256
              }
            }
          },
          "responseStatus": {
            "type": "integer"
          },
          "responseDuration": {
            "type": "float"
          },
          "extra": {
            "type": "object",
            "dynamic": true
          }
        }
      },
      "timestamp": {
        "type": "date",
        "format": "epoch_millis"
      },
      "logLevel": {
        "type": "keyword"
      },
      "message": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "source": {
        "properties": {
          "host": {
            "type": "keyword"
          },
          "service": {
            "type": "keyword"
          },
          "environment": {
            "type": "keyword"
          },
          "extra": {
            "type": "object",
            "dynamic": true
          }
        }
      }
    }
  }
}
**/


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

func (l *LogInfo) Insert() error{
  data, e := json.Marshal(l)
  fmt.Println("adding to elasticsearch")
  if e!= nil{
    return e
  }
  req := esapi.IndexRequest{
		Index:      constants.LOGS_INDEX_NAME, // Index name
		Body:       bytes.NewReader(data),
		Refresh:    "true", // Make the document immediately available for search
	}
  res, err := req.Do(context.Background(), datastore.Es)
  if err != nil {
    fmt.Println("error while inserting to elastic search:", err)
    return err
	}
	defer res.Body.Close()
	// Check if the response is successful
	if res.IsError() {
    fmt.Println("error while elastic search:", res.String())
    return errors.New(res.String())
	} else {
		fmt.Printf("Document indexed successfully: %s\n", res.String())
	}
  return nil
}
