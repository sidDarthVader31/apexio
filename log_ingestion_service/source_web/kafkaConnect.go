package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)


var KafkaConnector *kafka.Producer
func connectKafka() (bool,*kafka.Producer){
  var err error
  KafkaConnector,err =  kafka.NewProducer(&kafka.ConfigMap{ "bootstrap.servers":"localhost:9092",
    "client.id":"logProducer",
    "acks":"all",
  })
  if err!=nil{
    fmt.Printf("error connecting to kafka %v, shutting down the server", err)
    return false, nil
  }
  fmt.Println("connected to kafka ")
  return true, KafkaConnector
}
