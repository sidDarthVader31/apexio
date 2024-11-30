package datastream

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

type kafkaBatch struct{
  value *kafka.Message
  topic []string
  buffer []*kafka.Message
  bufferSize int
}


func (k *kafkaBatch) consumeMessages(){
}

func (k *kafkaBatch) flush(){

}

