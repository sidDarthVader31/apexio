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
  r := gin.Default() //init gin
	Routev1 = r.Group("/api/v1")
  config.InitEnv() //init config 
  initRoutes(Routev1) // init routes
	logger.InitLogger() // init logger 
	initDataStream(context.Background())
  r.Run(fmt.Sprintf(":%s", config.Config.PORT)) 
}
func initDataStream(ctx context.Context){
	DataStreamService, dataStreamError := datastream.CreateDataStream(config.Config.MESSAGE_BROKER)

	if dataStreamError !=nil{
		logger.Error("Error initializing data stream", dataStreamError)
		os.Exit(1)
	}
	connectErr := DataStreamService.Connect(ctx)
	if connectErr!=nil{
		logger.Error("Error connecting to data stream", connectErr, map[string]interface{} {"dataStream": config.Config.MESSAGE_BROKER})
		os.Exit(1)
	}
}
