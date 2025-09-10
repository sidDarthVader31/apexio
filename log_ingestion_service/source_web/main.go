package main

import (
	"context"
	"fmt"
	"os"
	"sourceweb/config"
	"sourceweb/logDistributor"
	logger "sourceweb/logging"
	datastream "sourceweb/service/dataStream"

	"github.com/gin-gonic/gin"
)

var Routev1 *gin.RouterGroup

func main() {
	r := gin.Default() //init gin
	lg := logger.NewLogger()
	Routev1 = r.Group("/api/v1")
	config.InitEnv() //init config
	lc := logDistributor.NewLogController(lg)
	initRoutes(Routev1, lc) // init routes
	initDataStream(context.Background(), lg)
	r.Run(fmt.Sprintf(":%s", config.Config.PORT))
}
func initDataStream(ctx context.Context, lg *logger.Logger) {
	DataStreamService, dataStreamError := datastream.CreateDataStream(config.Config.MESSAGE_BROKER, lg)

	if dataStreamError != nil {
		lg.Error("Error initializing data stream", dataStreamError)
		os.Exit(1)
	}
	connectErr := DataStreamService.Connect(ctx)
	if connectErr != nil {
		lg.Error("Error connecting to data stream", connectErr, map[string]interface{}{"dataStream": config.Config.MESSAGE_BROKER})
		os.Exit(1)
	}
}
