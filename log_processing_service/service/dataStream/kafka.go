package datastream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log-processor/datastore/models"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaStream struct{
  Consumer *kafka.Consumer
  configMap map[string]string
  topics []string
  workers int
  maxRetries int
  retryBackoff time.Duration
}

type ConsumerConfig struct{
  BootStrapServers string
  GroupId string
  Topics []string
  AutoOffsetReset string
  WorkerCount string
  maxRetries int
  RetryBackoff time.Duration
}


func getNewkafkaStream(configMap map[string]string) (*KafkaStream, error){
  return &KafkaStream{configMap: configMap}, nil
}

func (k *KafkaStream) Connect(context context.Context, config map[string]string) error{
  KafkaConnector,err := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":    "localhost:9092",
     "group.id":             "foo",
     "auto.offs<et.reset":    "smallest",
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

  // setup signal handling for graceful shutdown
  sigChan := make(chan os.Signal, 1)
  signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
  messageChan := make(chan *kafka.Message, k.workers)

  var wg sync.WaitGroup

  for i := 0; i< k.workers; i++{
    wg.Add(1)
    go k.messageWorker(messageChan, &wg)
  }
  
  for {
    select {

    case <- context.Done():
    close(messageChan)
    wg.Wait()
    default: 
    msg, err := k.Consumer.ReadMessage(100 * time.Millisecond)
    if err!=nil{
        fmt.Println("error reading message:", err)
      }
      messageChan <- msg
    }
  }
}



func (k * KafkaStream) messageWorker(messageChan <- chan *kafka.Message, wg *sync.WaitGroup){
  defer wg.Done()
  for msg := range(messageChan){
    err := k.processMessageWithRetry(msg)
    if err !=nil{
      fmt.Printf("failed to process message after all retries:", err)
    }
  }
}

func (k * KafkaStream) processMessageWithRetry(msg *kafka.Message) error {
  var error error
  for attempt :=0 ; attempt < k.maxRetries; attempt++{
    error = k.processMessage(msg)
    if error != nil {
      return nil 
    }
    if attempt < k.maxRetries {
      time.Sleep(k.retryBackoff)
    }
  }
  return fmt.Errorf("failed to process message after &d retries :%w", k.maxRetries, error)
}


func(k *KafkaStream)processMessage(msg *kafka.Message) error{
  var logData models.LogInfo
  err := json.Unmarshal(msg.Value, &logData)
  if err != nil {
    return err
  }
  //insert 
  error := logData.Insert()
  if error != nil {
    return error
  }
  //commit to kafka 
   _, errC:= k.Consumer.CommitMessage(msg)
  if errC != nil {
    return errC
  }
  return nil
}

func (k *KafkaStream) Close(){
  k.Consumer.Close()
}
