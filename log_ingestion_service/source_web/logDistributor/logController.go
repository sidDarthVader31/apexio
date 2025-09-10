package logDistributor

import (
	"net/http"
	"sourceweb/constants"
	logger "sourceweb/logging"

	"github.com/gin-gonic/gin"
)

type LogController struct {
	Logger *logger.Logger
}

func NewLogController(logger *logger.Logger) *LogController {
	return &LogController{Logger: logger}
}

func (lc *LogController) IngestData(c *gin.Context) {
	var logInfo LogInfo
	if err := c.BindJSON(&logInfo); err != nil {
		lc.Logger.Error(string(JSON_CONVERSION_FAILED), err, logInfo)
		c.IndentedJSON(400, err)
	}
	
	success, errIngest := ingestLogs(logInfo, constants.LogTopic)
	if success == false {
		lc.Logger.Error(string(ERROR_INGESTING_LOGS), errIngest, logInfo)
		c.IndentedJSON(http.StatusNotFound, errIngest)
	}
	c.IndentedJSON(http.StatusCreated, logInfo)
}
