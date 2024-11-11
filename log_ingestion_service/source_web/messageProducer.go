package main

import (
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

// kafka topic namespace
// logs.{serviceName}.{logType}.{version}
//eg - logs.ingestion.raw.v1 -> for logs to be sent for
//ingestion

const logTopic = "logs.ingestion.raw.v1"


func ingestToKafka(value []byte, topic string) (bool){
  kafkaConnector.Produce(&kafka.Message{
  TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
  Value: value,
 }, nil)
  return true;
}
