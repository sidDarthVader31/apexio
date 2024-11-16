package main

import (
	"fmt"
	"os"
	"github.com/gin-gonic/gin"
  ds "log-processor/datastore"
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
  _,err := connectKafka()
  if err!=nil{
    fmt.Println("error connecting to kafka")
    os.Exit(1)
  }
  ds.ConnectToElastic()
  getLogs()
  fmt.Println("go consumer running")
  r.Run()
}
