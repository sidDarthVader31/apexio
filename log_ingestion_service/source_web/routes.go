package main

import (
	"sourceweb/logDistributor"
	"github.com/gin-gonic/gin"
)

func initRoutes(route *gin.RouterGroup) {
  route.POST("/log", logDistributor.IngestData)
}
