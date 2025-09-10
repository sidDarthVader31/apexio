package main

import (
	"sourceweb/logDistributor"

	"github.com/gin-gonic/gin"
)

func initRoutes(route *gin.RouterGroup, lc *logDistributor.LogController) {
	route.POST("/log", lc.IngestData)
}
