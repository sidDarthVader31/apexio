package datastore

import (
	"context"
	"errors"
	"fmt"
	"log-processor/constants"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

/**
function to initialize index with correct mapping
**/ 
func InitIndex() error{
  mapping := `{
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
                  "type": "text"
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
    }`
  req := esapi.IndicesCreateRequest{
    Index: constants.LOGS_INDEX_NAME,
    Body: strings.NewReader(mapping),
  }
  res, err := req.Do(context.Background(), Es)

  if err != nil{
    fmt.Println("error while creating index:", err)
    return err
  }else if res.StatusCode != 400{
    fmt.Println("error received while creating index", res)
    return errors.New(res.String())
  }else{
    fmt.Println("successfully created index:", res)
    return nil
  }
 }


