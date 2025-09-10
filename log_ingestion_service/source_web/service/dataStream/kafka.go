package datastream

import (
	"context"
	"errors"
	"fmt"
	"sourceweb/config"
	"sourceweb/constants"
	logger "sourceweb/logging"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaService struct {
	batchProcess batchProcess
	logger       *logger.Logger
}

type batchProcess struct {
	producer    *kafka.Producer
	batchSize   int
	batchWindow time.Duration
	buffer      map[string][][]byte
	bufferMu    sync.Mutex
	done        chan struct{}
	logChan     chan logChanStruct
}

type logChanStruct struct {
	topicName string
	message   []byte
}

func NewKafkaService(lg *logger.Logger) (*KafkaService, error) {
	return &KafkaService{logger: lg}, nil
}

func (k *KafkaService) Connect(ctx context.Context) error {
	k.logger.Info("Kafka Config:", getKafkaConfig())
	// Kafka connection logic
	kafkaConnector, err := kafka.NewProducer(getKafkaConfig())

	if err != nil {
		fmt.Printf("error connecting to kafka %v, shutting down the server\n", err)
		k.batchProcess = batchProcess{}
		return err
	} else {
		fmt.Println("connected to kafka ", kafkaConnector)
	}

	//create a batchprocessor
	k.batchProcess = batchProcess{
		producer:    kafkaConnector,
		batchSize:   constants.BatchSize,
		batchWindow: constants.BufferTime,
		buffer:      make(map[string][][]byte, constants.BatchSize),
		logChan:     make(chan logChanStruct, constants.BatchSize*2), // buffer channel to handle spike in traffic
		done:        make(chan struct{}),
	}
	go k.batchProcess.processLogs()
	return nil
}

func (k *KafkaService) ProduceMessage(ctx context.Context, message []byte, topicName string) (bool, error) {
	// Kafka produce logic
	select {
	case k.batchProcess.logChan <- logChanStruct{
		topicName: topicName,
		message:   message,
	}:
		return true, nil
default:
		return false, errors.New("log channel is full")
	}
}

func (k *KafkaService) Close() {
	k.batchProcess.producer.Close()
}

func (b *batchProcess) processLogs() {
	ticker := time.NewTicker(b.batchWindow)
	defer func() {
		ticker.Stop()
		fmt.Println("Exiting processLogs goroutine.")
	}()
	for {
		select {
		//when new log comes in
		case log := <-b.logChan:
			b.bufferMu.Lock()
			// Ensure buffer map is initialized
			if b.buffer == nil {
				b.buffer = make(map[string][][]byte)
			}
			b.buffer[log.topicName] = append(b.buffer[log.topicName], log.message)

			// Check if any topic's buffer is full
			for topic, msgs := range b.buffer {
				if len(msgs) >= b.batchSize {
					b.flush(topic)
				}
			}
			b.bufferMu.Unlock()

		case <-ticker.C:
			b.bufferMu.Lock()
			// Flush non-empty topic buffers
			for topic, msgs := range b.buffer {
				if len(msgs) > 0 {
					b.flush(topic)
				}
			}
			b.bufferMu.Unlock()

		case <-b.done:
			b.bufferMu.Lock()
			// Flush all topic buffers on shutdown
			for topic, msgs := range b.buffer {
				if len(msgs) > 0 {
					b.flush(topic)
				}
			}
			b.bufferMu.Unlock()
			return
		}
	}
}

func (b *batchProcess) flush(topic string) {
	topicBuffer := b.buffer[topic]
	if len(topicBuffer) == 0 {
		return
	}

	for _, v := range topicBuffer {
		if v == nil {
			fmt.Println("v is nil")
			continue
		}
		fmt.Println("sending message:", len(v))
		err := b.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: v,
		}, nil)
		fmt.Println("message sent")

		if err != nil {
			fmt.Printf("error producing message to topic %s: %v\n", topic, err)
		}
	}
	// Remove the flushed topic's buffer
	delete(b.buffer, topic)
}

func getKafkaConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":        config.Config.KAFKA_HOST,
		"client.id":                config.Config.MESSAGE_BROKER_CLIENTID,
		"acks":                     config.Config.MESSSAGE_BROKER_ACKS,
		"retries":                  config.Config.MESSAGE_BROKER_RETRIES,
		"message.send.max.retries": config.Config.MESSAGE_BROKER_MAX_RETRIES,
		"delivery.timeout.ms":      config.Config.MESSAGE_BROKER_TIMEOUT,
		"linger.ms":                config.Config.MESSAGE_BROKER_LINGER_MS,
		"log_level":                config.Config.MESSAGE_BROKER_LOG_LEVEL,
	}
}
