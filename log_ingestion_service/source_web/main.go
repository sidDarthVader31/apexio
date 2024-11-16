//main file for REST based web service to ingest logs 
package main

import (
	"fmt"
	"os"

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
  _,err := connectKafka()
  if err!=nil{
    fmt.Println("error connecting to kafka")
    os.Exit(1)
  }
  r.Run()
}