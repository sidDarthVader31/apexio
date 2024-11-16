package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)


var KafkaConnector *kafka.Consumer
func connectKafka() (*kafka.Consumer, error){
  var err error
  KafkaConnector,err = kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":    "localhost:9092",
     "group.id":             "foo",
     "auto.offset.reset":    "smallest",
  })
  
  if err!=nil{
    fmt.Printf("error connecting to kafka %v, shutting down the server", err)
    return nil, err
  }
  fmt.Println("connected to kafka ")
  return KafkaConnector, nil
}
