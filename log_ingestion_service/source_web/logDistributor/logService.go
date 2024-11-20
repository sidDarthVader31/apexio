package logDistributor

import (
	"encoding/json"
	"fmt"
	"sourceweb/service/kafka"
)





func ingestLogs(logData LogInfo, logTopic string) (bool, error) {
  value, err := json.Marshal(logData)
  if err!= nil {
    fmt.Println("error converting to json", err)
    return false, err
  }
  success, err := kafka.IngestToKafka(value, logTopic)
  return success, err
}

