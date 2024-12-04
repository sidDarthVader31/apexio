package main

import (
	"context"
	"fmt"
	"log-processor/config"
	"log-processor/constants"
	ds "log-processor/datastore"
	datastream "log-processor/service/dataStream"
	"os"

	"github.com/gin-gonic/gin"
)

var Routev1 *gin.RouterGroup;
func main(){
  fmt.Println("starting kafka consumer")
 r := gin.Default()
  Routev1 = r.Group("/api/v1")
  Routev1.GET("/health", func(ctx *gin.Context) {
    ctx.JSON(200, gin.H{
      "message": "health check passed",
    })
  })
  config.InitEnv()
  fmt.Println("message broker:", config.Config.MESSAGE_BROKER)
  DataStreamService, errorData := datastream.CreateDataStream(context.Background(),make(map[string]string), config.Config.MESSAGE_BROKER)
  if errorData != nil{
    fmt.Println("error connecting to message broker, shutting down")
    os.Exit(1)
  }
  err := DataStreamService.Connect(context.Background(), make(map[string]string))
  if err != nil{
    fmt.Println("errow connecting to data stream,::", err)
    os.Exit(1)
  }
  ds.ConnectToElastic()
  DataStreamService.Consume(context.Background(), constants.LogTopic)
  fmt.Println("kafka consumer running")
  r.Run(fmt.Sprintf(":%s", config.Config.PORT))
}
