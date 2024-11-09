package main

import "github.com/gin-gonic/gin"


func main(){
  r := gin.Default()
  routev1 := r.Group("/api/v1")
  routev1.GET("/", func(ctx *gin.Context) {
    ctx.JSON(200, gin.H{
      "message": "fetched successfully",
    })
  })
  r.Run()
}
