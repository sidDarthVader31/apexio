package main

import (
	"fmt"
	"os"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)


var kafkaConnector *kafka.Producer
func connectKafka() *kafka.Producer{
  kafkaConnector,err :=  kafka.NewProducer(&kafka.ConfigMap{
    "bootstrap.servers":"host",
    "client.id":"myproducer",
    "acks":"all",
  })
  if err!=nil{
    fmt.Printf("error connecting to kafka %v, shutting down the server", err)
    os.Exit(1)
  }
  return kafkaConnector
}
