package main

import (
	"encoding/json"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

// kafka topic namespace
// logs.{serviceName}.{logType}.{version}
//eg - logs.ingestion.raw.v1 -> for logs to be sent for
//ingestion

const logTopic = "logs.ingestion.raw.v1"


func ingestToKafka[T interface{}](data T){
  value, err := json.Marshal(data)
  if err!=nil{
    panic(err)
  }
  topic := logTopic
  kafkaConnector.Produce(&kafka.Message{
  TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
  Value: value,
 }, nil)
}
