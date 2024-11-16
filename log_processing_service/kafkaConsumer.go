package main

import (
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// kafka topic namespace
// logs.{serviceName}.{logType}.{version}
//eg - logs.ingestion.raw.v1 -> for logs to be sent for
//processing



func getLogs(){
  err := KafkaConnector.SubscribeTopics([]string {logTopic}, nil)
  if err != nil{
    fmt.Println("issue with connecting to kafka:",err)
    os.Exit(1)
  }
  run := true
  for run == true{
    ev := KafkaConnector.Poll(100)
    switch e := ev.(type){
    case *kafka.Message:
    fmt.Println("received kafka message")
      fmt.Println("log:", string(e.Value))
        //basic preprocessing
        //ingest to elastick
    case kafka.Error:
      fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
    run = false
    }
  }

}
