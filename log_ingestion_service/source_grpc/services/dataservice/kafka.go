package dataservice

import (
	"context"
	"errors"
	"fmt"
	"source_grpc/config"
	"source_grpc/constants"
	"sync"
	"time"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaService struct{
  batchProcess batchProcess
}

type batchProcess struct {
  producer *kafka.Producer
  batchSize int
  batchWindow time.Duration
  buffer map[string][][]byte
  bufferMu sync.Mutex
  done chan struct{}
  logChan chan logChanStruct
}

type logChanStruct struct{
  topicName string
  message []byte
}


func getNewkafkaService() (*KafkaService, error){
  return &KafkaService{}, nil
}


func (k *KafkaService) Connect(context context.Context) error{
  kafkaConnector, err := kafka.NewProducer(getKafkaConfig())
  if err!=nil{
    fmt.Println("error connecting to kafkaaaaa:", err)
    k.batchProcess = batchProcess{}
    return err
  }else{
    fmt.Println("Kafka connected")
  }

  //configure kafkaservice and batch batchProcess
  k.batchProcess = batchProcess{
    producer: kafkaConnector,
    batchSize: constants.BatchSize,
    batchWindow: constants.BufferTime,
    buffer: make(map[string][][]byte) ,
    logChan: make(chan logChanStruct, constants.BatchSize*2),
    done: make(chan struct{}),
  }
  //start batch processor go routin
  go k.batchProcess.processLogs()
  return nil
}

func (k *KafkaService)ProduceMessage(context context.Context, message []byte, topicName string) (bool, error){
  select{
    case  k.batchProcess.logChan <- logChanStruct{
    topicName: topicName,
    message: message,
  } :
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
  
  for{
    select { 
    case log:= <- b.logChan:
      //incase of log incoming
      b.bufferMu.Lock()
      //ensure buffer is initialized 
      if b.buffer == nil{
        b.buffer = make(map[string][][]byte)
      }
      b.buffer[log.topicName] = append(b.buffer[log.topicName], log.message)

      //check if any topic's buffer is full or not 
      for topic, msgs := range b.buffer{
        if len(msgs) >= b.batchSize{
          b.flush(topic)
        }
      }
      b.bufferMu.Unlock()
  
    case <- ticker.C:
      b.bufferMu.Lock()
      //flush non empty topic buffers 
      for topic, msgs := range b.buffer{
        if len(msgs)>= 0 {
          b.flush(topic)
        }
      }
      b.bufferMu.Unlock()

    case <- b.done:
      b.bufferMu.Lock()
      for topic, msgs := range b.buffer{
        if len(msgs) >= 0 {
          b.flush(topic)
        } 
      }
      b.bufferMu.Unlock()
      return
    }
  }
}


func (b *batchProcess) flush(topic string){
  topicBuffer := b.buffer[topic]
  if len(topicBuffer) == 0{
    return
  }
  for _, v:= range topicBuffer{
    if v == nil{
      continue
    }
    fmt.Println("sending message")
    err := b.producer.Produce(&kafka.Message{
      TopicPartition: kafka.TopicPartition{
        Topic: &topic,
        Partition: kafka.PartitionAny,
      },
      Value: v,
    }, nil)
    if err != nil{
      fmt.Printf("error producing message to topic &s: %v\n", topic, err)
    }
  }

}


func getKafkaConfig() *kafka.ConfigMap{
  return &kafka.ConfigMap{
    "bootstrap.servers": config.Config.KAFKA_HOST,
    "client.id": config.Config.MESSAGE_BROKER_CLIENTID,
    "acks": config.Config.MESSSAGE_BROKER_ACKS,
    "retries": config.Config.MESSAGE_BROKER_RETRIES,
    "message.send.max.retries": config.Config.MESSAGE_BROKER_MAX_RETRIES,
    "delivery.timeout.ms": config.Config.MESSAGE_BROKER_TIMEOUT,
    "linger.ms": config.Config.MESSAGE_BROKER_LINGER_MS,
    "log_level": config.Config.MESSAGE_BROKER_LOG_LEVEL,
  }
}

