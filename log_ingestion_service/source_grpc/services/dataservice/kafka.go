package dataservice

import (
	"context"
	"errors"
	"fmt"
	"source_grpc/constants"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaService struct{
  config map[string]string
  batchProcess batchProcess
}

type batchProcess struct {
  producer *kafka.Producer
  batchSize int
  batchWindow time.Duration
  buffer [] []byte
  bufferMu sync.Mutex
  done chan struct{}
  logChan chan []byte
  topicName string
}


func getNewkafkaService(config map[string]string) (*KafkaService, error){
  return &KafkaService{config: config,}, nil
}


func (k *KafkaService) Connect(context context.Context, config map[string]string) error{
  kafkaConnector, err := kafka.NewProducer(&kafka.ConfigMap{ "bootstrap.servers":"localhost:9092",
    "client.id":"logProducer",
    "acks":"all",
    "retries": 5,
    "message.send.max.retries": 5,
    "delivery.timeout.ms": 100000,
    "linger.ms":5,
  })
  if err!=nil{
    return err
  }
  fmt.Println("Kafka connected")
  //configure kafkaservice and batch batchProcess
  k.batchProcess = batchProcess{
    producer: kafkaConnector,
    batchSize: constants.BatchSize,
    batchWindow: constants.BufferTime,
    buffer: make([][]byte, 0, constants.BatchSize),
    logChan: make(chan []byte, constants.BatchSize*2),
    done: make(chan struct{}),
  }
  //start batch processor go routin
  go k.batchProcess.processLogs()
  return nil
}

func (k *KafkaService)ProduceMessage(context context.Context, message []byte, topicName string) (bool, error){
  k.batchProcess.topicName = topicName
  select{
    case  k.batchProcess.logChan <- message :
      return true,nil
    default:
      return false, errors.New("log channel is full")
  }
}

func (k *KafkaService)Close(){
  k.batchProcess.producer.Close()
}
//------------------------- batchProcess functions
//---------------


func (b *batchProcess)processLogs(){
  //timer
  ticker := time.NewTicker(b.batchWindow)
  defer func(){
    ticker.Stop()
    fmt.Println("exiting process logs go routine")
  }()
  select {
  case log:= <- b.logChan:
    //incase of log incoming
    b.bufferMu.Lock()
    b.buffer = append(b.buffer, log)
    if len(b.buffer)  >= b.batchSize{
      b.flush()
    }
    b.bufferMu.Unlock()
  case <- ticker.C:
    b.bufferMu.Lock()
    if(len(b.buffer) >= b.batchSize){
      //flush logs 
      b.flush()
    }
    b.bufferMu.Unlock()
  case <- b.done:
    b.bufferMu.Lock()
    if len(b.buffer) > 0 {
      b.flush()
    }
  }
}


func (b *batchProcess) flush(){
  if len(b.buffer) == 0{
    return
  }
  for _,v := range b.buffer{
    b.producer.Produce(&kafka.Message{
      TopicPartition: kafka.TopicPartition{
        Topic: &b.topicName,
        Partition: kafka.PartitionAny,
      },
      Value: v,
    }, nil)
  }
  b.buffer = make([][]byte, b.batchSize)
}

