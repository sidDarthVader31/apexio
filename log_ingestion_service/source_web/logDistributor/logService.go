package logDistributor

import (
	"context"
	"encoding/json"
	"fmt"
	datastream "sourceweb/service/dataStream"
)

func ingestLogs(logData LogInfo, logTopic string) (bool, error) {
  value, err := json.Marshal(logData)
  if err!= nil {
    fmt.Println("error converting to json", err)
    return false, err
  }
  success, err := datastream.StreamService.ProduceMessage(context.Background(),value, logTopic)
  return success, err
}

