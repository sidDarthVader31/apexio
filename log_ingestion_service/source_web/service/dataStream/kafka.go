package datastream

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)





type KafkaService struct{
  config map[string]string
  kafkaProducer *kafka.Producer
} 


func NewKafkaService(config map[string]string) (*KafkaService, error) {
	return &KafkaService{config: config}, nil
}



func (k *KafkaService) Connect(ctx context.Context, confi map[string]string) error {
	// Kafka connection logic
  kafkaConnector, err := kafka.NewProducer(&kafka.ConfigMap{ "bootstrap.servers":"localhost:9092",
    "client.id":"logProducer",
    "acks":"all",
    "compression":"snappy",
    "retries": 5,
    "retries.backoff.ms": 100,
    "enable.idempotence": true,
    "message.send.max.retries": 5,
    "delivery.timeout.ms": 100000,
    "linger.ms":5,
  })
  if err!=nil{
    fmt.Printf("error connecting to kafka %v, shutting down the server", err)
    k.kafkaProducer = nil;
    return err
  }
  fmt.Println("connected to kafka ")
  k.kafkaProducer = kafkaConnector
  k.Close()
  return nil;
}

func (k *KafkaService) ProduceMessage(ctx context.Context, message []byte, topicName string) (bool, error) {
  // Kafka produce logic
  err := k.kafkaProducer.Produce(&kafka.Message{
    TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
    Value: message,
  }, nil)
  if(err != nil){
    return false, err
  }
  return true, nil
}

func (k KafkaService) Close(){
 //close connection 
  defer k.kafkaProducer.Close()
}
