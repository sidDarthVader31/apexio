// main file for REST based web service to ingest logs
package main

import (
	"context"
	"fmt"
	"os"
	datastream "sourceweb/service/dataStream"

	"github.com/gin-gonic/gin"
)
var Routev1 *gin.RouterGroup;
func main(){
  r := gin.Default()
  Routev1 = r.Group("/api/v1")
  Routev1.GET("/health", func(ctx *gin.Context) {
    ctx.JSON(200, gin.H{
      "message": "health check passed",
    })
  })
  initRoutes(Routev1)
  DataStreamService, dataStreamError := datastream.CreateDataStream("KAFKA", map[string]string{"baseurl":"http://localhost:9092"})
  DataStreamService.Connect(context.Background(), map[string]string{})

  if dataStreamError!=nil{
    fmt.Println("error connecting to kafka:", dataStreamError)
    os.Exit(1)
  }
  r.Run()
}
