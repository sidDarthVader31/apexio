package datastream

import (
	"context"
	"errors"
	"fmt"
	"sourceweb/constants"
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
func NewKafkaService(config map[string]string) (*KafkaService, error) {
	return &KafkaService{config: config}, nil
}

func (k *KafkaService) Connect(ctx context.Context, config map[string]string) error {
	// Kafka connection logic
  kafkaConnector, err := kafka.NewProducer(&kafka.ConfigMap{ "bootstrap.servers":"localhost:9092",
    "client.id":"logProducer",
    "acks":"all",
    "retries": 5,
    "message.send.max.retries": 5,
    "delivery.timeout.ms": 100000,
    "linger.ms":5,
  })
  if err!=nil{
    fmt.Printf("error connecting to kafka %v, shutting down the server", err)
    k.batchProcess= batchProcess{}
    return err
  }
  fmt.Println("connected to kafka ")

  //create a batchprocessor
  k.batchProcess = batchProcess{
    producer: kafkaConnector,
    batchSize: constants.BatchSize,
    batchWindow: constants.BufferTime,
    buffer: make([][]byte, 0, constants.BatchSize),
    logChan : make(chan []byte, constants.BatchSize*2), // buffer channel to handle spike in traffic 
    done: make(chan struct{}),
  }
  go k.batchProcess.processLogs() 
  return nil;
}

func (k *KafkaService) ProduceMessage(ctx context.Context, message []byte, topicName string) (bool, error) {
  // Kafka produce logic
  k.batchProcess.topicName = topicName
  select {
  case k.batchProcess.logChan <-message:
    return true, nil
  default:
    return false, errors.New("log channel is full")
  }
}

func (k *KafkaService) Close(){
}

func(b * batchProcess) processLogs(){
  ticker := time.NewTicker(b.batchWindow)
  defer func() {
        ticker.Stop()
        fmt.Println("Exiting processLogs goroutine.")
    }()
  for {
    select{
    //when new log comes in 
    case log := <-b.logChan:
      b.bufferMu.Lock()
      b.buffer = append(b.buffer, log)
      if len(b.buffer)>= b.batchSize{
      //flush logs 
        b.flush()
      }
      b.bufferMu.Unlock()
    case <-ticker.C:
      b.bufferMu.Lock()
      if len(b.buffer) > 0 {
        //flush
        b.flush()
      }
    b.bufferMu.Unlock()
    case <-b.done:
      b.bufferMu.Lock()
      if len(b.buffer) > 0{
        b.flush()
      }
      b.bufferMu.Unlock()
      return;
    }
  }
}


func (b *batchProcess) flush(){
  if len(b.buffer) == 0{
    return 
  }
  for _,v := range b.buffer{
    b.producer.Produce(&kafka.Message{
    TopicPartition: kafka.TopicPartition{Topic: &b.topicName, Partition: kafka.PartitionAny},
    Value:v,
  }, nil)
  }
  b.buffer = make([][]byte, b.batchSize)
}
