package constants
import "time"
const LogTopic = "logs.ingestion.raw.v1"
const BufferLength = 1000000
const BufferTime = time.Second*1
const BatchSize = 5
