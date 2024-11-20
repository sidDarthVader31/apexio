package logDistributor

import (
	"context"
	"encoding/json"
	"fmt"
	"sourceweb/constants"
	datastream "sourceweb/service/dataStream"
)





func ingestLogs(logData LogInfo, logTopic string) (bool, error) {
  value, err := json.Marshal(logData)
  if err!= nil {
    fmt.Println("error converting to json", err)
    return false, err
  }
  success, err := datastream.StreamService.ProduceMessage(context.Background(),value, constants.LogTopic)
  return success, err
}

