package datastream

import (
	"context"
	"encoding/json"
	"fmt"
	"log-processor/config"
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


func getNewkafkaStream() (*KafkaStream, error){
  return &KafkaStream{
    maxRetries: config.Config.KAFKA_MAX_RETRIES,
    workers: config.Config.KAFKA_WORKERS,
    retryBackoff: time.Second,
  }, nil
}

func (k *KafkaStream) Connect(ctx context.Context) error{
  KafkaConnector,err := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":    config.Config.KAFKA_HOST,
     "group.id":            config.Config.KAFKA_GROUP_ID,
     "auto.offset.reset":    config.Config.KAFKA_OFFSET_RESET,
  })
  if err != nil {
    fmt.Println("issue while connecting to kafka", err)
    return err
  }
  fmt.Println("connected to kafka:", KafkaConnector)
  k.Consumer = KafkaConnector
  return nil
}

func (k *KafkaStream) Consume(ctx context.Context, topics []string){
  err := k.Consumer.SubscribeTopics([]string {"logs.ingestion.raw.v1"}, nil)
  if err != nil{
    fmt.Println("issue subsribing to kafka topics:",err)
    os.Exit(1)
  }
  k.topics = topics

  // setup signal handling for graceful shutdown
  sigChan := make(chan os.Signal, k.workers)
  signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
  messageChan := make(chan *kafka.Message, k.workers)

  var wg sync.WaitGroup
  consumeCtx, cancel := context.WithCancel(ctx) 
  defer cancel()
  for i := 0; i< k.workers; i++{
    wg.Add(1)
    go k.messageWorker(messageChan, &wg)
  }
  // Goroutine to handle signals
  go func() {
       <-sigChan
       fmt.Println("Received interrupt signal, shutting down...")
       cancel() // Cancel the context to stop consuming
  }() 
  
  for {
    select {
      case <- consumeCtx.Done():
      close(messageChan)
      wg.Done()
      return 
    default: 
      msg, err := k.Consumer.ReadMessage(time.Second)
      if err!=nil{
          // fmt.Println("error reading message:", err)
          continue
        }
      select {
        case messageChan<- msg:
        case <- consumeCtx.Done():
          close(messageChan)
          wg.Wait()
          return
        }
    }
  }
}



func (k * KafkaStream) messageWorker(messageChan <- chan *kafka.Message, wg *sync.WaitGroup){
  defer wg.Done()
  for msg := range(messageChan){
    err := k.processMessageWithRetry(msg)
    if err !=nil{
      fmt.Printf("failed to process message after all retries:%v", err)
    }
  }
}

func (k * KafkaStream) processMessageWithRetry(msg *kafka.Message) error {
  var error error
  for attempt :=0 ; attempt < k.maxRetries; attempt++{
    fmt.Printf("trying for attemp no:%v", attempt)
    error = k.processMessage(msg)
    if error == nil {
      fmt.Println("processed message successfully")
     return nil 
    }else{
      fmt.Println("issue while processing message", error)
    }
    if attempt < k.maxRetries {
      time.Sleep(k.retryBackoff)
    }
  }
  return fmt.Errorf("failed to process message after %d retries :%w", k.maxRetries, error)
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
    fmt.Println("error while inserting data to elastic:", error)
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
