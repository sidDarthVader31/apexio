package datastream

import (
	"context"
	"encoding/json"
	"fmt"
	"log-processor/datastore/models"
	"os"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)



type KafkaStream struct{
  Consumer *kafka.Consumer
  configMap map[string]string
}


func getNewkafkaStream(configMap map[string]string) (*KafkaStream, error){
  return &KafkaStream{configMap: configMap}, nil
}



func (k *KafkaStream) Connect(context context.Context, config map[string]string) error{
  KafkaConnector,err := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":    "localhost:9092",
     "group.id":             "foo",
     "auto.offset.reset":    "smallest",
  })
  if err != nil {
    fmt.Println("issue while connecting to kafka", err)
    return err
  }
  k.Consumer = KafkaConnector
  return nil
}

func (k *KafkaStream) Consume(context context.Context, topics []string){
  err := k.Consumer.SubscribeTopics(topics, nil)
  if err != nil{
    fmt.Println("issue with connecting to kafka:",err)
    os.Exit(1)
  }
  run := true
  for run == true{
    ev := k.Consumer.Poll(100)
    switch e := ev.(type){
    case *kafka.Message:
      fmt.Println("received kafka message")
      fmt.Println("log:", string(e.Value))
       var logData models.LogInfo
      err := json.Unmarshal(e.Value, &logData)
      if err!=nil{
        fmt.Println("error while processing log:", err)
      }
      logData.Insert()
    case kafka.Error:
      fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
    run = false
    }
  }
}

func (k *KafkaStream) Close(){
  k.Consumer.Close()
}
