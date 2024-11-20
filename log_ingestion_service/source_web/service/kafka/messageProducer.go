package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// kafka topic namespace
// logs.{serviceName}.{logType}.{version}
//eg - logs.ingestion.raw.v1 -> for logs to be sent for
//ingestion
func IngestToKafka(logEntry[]byte, topic string) (bool, error){ err := KafkaConnector.Produce(&kafka.Message{
  TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
  Value: logEntry,
 }, nil)
  if(err != nil){
    return false, err
  }
  return true, nil
}
