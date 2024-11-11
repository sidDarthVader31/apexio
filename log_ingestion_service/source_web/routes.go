package main

import "github.com/gin-gonic/gin"

func initRoutes(route *gin.RouterGroup) {
  route.POST("/log", ingestData)
}
