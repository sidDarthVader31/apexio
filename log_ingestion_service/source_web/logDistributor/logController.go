package logDistributor

import (
	"fmt"
	"net/http"
	"sourceweb/constants"
	"github.com/gin-gonic/gin"
)

func IngestData(c *gin.Context) {
  var logInfo LogInfo
  if err := c.BindJSON(&logInfo); err != nil {
    fmt.Println("error while creating json", err)
        c.IndentedJSON(400, err)
    }
  success, err := ingestLogs(logInfo, constants.LogTopic)
  if success == false {
    fmt.Println("error while ingeting logs :", err)
    c.IndentedJSON(http.StatusNotFound, err)
  }
 c.IndentedJSON(http.StatusCreated, logInfo)
}
