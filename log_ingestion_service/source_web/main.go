// main file for REST based web service to ingest logs
package main

import (
	"context"
	"fmt"
	"os"
	"sourceweb/config"
	datastream "sourceweb/service/dataStream"
	logger "sourceweb/logging"
	"github.com/gin-gonic/gin"
)
var Routev1 *gin.RouterGroup;


func main(){
  r := gin.Default()
  config.InitEnv()

  initRoutes(Routev1)

	logging.InitLogger()

  DataStreamService, dataStreamError := datastream.CreateDataStream(config.Config.MESSAGE_BROKER)
  
  connectErr := DataStreamService.Connect(context.Background())
  if connectErr !=nil{
    fmt.Println("error connecting to kafka", connectErr)
    os.Exit(1)
  }

  if dataStreamError!=nil{
    fmt.Println("error connecting to kafka:", dataStreamError)
    os.Exit(1)
  }

  r.Run(fmt.Sprintf(":%s", config.Config.PORT))
}
func initDataStream(ctx context.Context){
	DataStreamService, dataStreamError := datastream.CreateDataStream(config.Config.MESSAGE_BROKER)

	if(dataStreamError !=nil){
		logger.Error(fmt.Sprintf("Error initializing data stream: %e", dataStreamError)
	}
	connectErr := DataStreamService.Connect(ctx)
	if connectErr!=nil{
		logging.Logger.Fatal(fmt.Sprintf("error connecting to kafka %v", connectErr.Error()))
	}
}
