package datastream

import (
	"context"
	"fmt"
	"sync"
)

type DataStreamServiceInterface interface{
  Connect(ctx context.Context, config map[string]string) error
  ProduceMessage(ctx context.Context, message []byte, topicName string) (bool, error)
  Close()
}

type DataStreamService struct {
  service DataStreamServiceInterface
}

var (
    StreamService *DataStreamService
    once          sync.Once // Ensure initialization is thread-safe
)



func CreateDataStream(provider string, config map[string]string) (*DataStreamService, error) {
	var service DataStreamServiceInterface
   once.Do(func() {
        StreamService = &DataStreamService{}
    })
	var err error

	switch provider {
	case "KAFKA":
		service, err = NewKafkaService(config)
    StreamService.service = service
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err != nil {
		return nil, err
	}

	return StreamService, err 
}

func(ds *DataStreamService) Connect(ctx context.Context, config map[string]string) (error){
  return ds.service.Connect(ctx, config)
}

func (ds *DataStreamService) ProduceMessage(ctx context.Context, message []byte, topicName string) (bool, error){
  return ds.service.ProduceMessage(ctx, message, topicName)
}

func (ds *DataStreamService) Close(){
  ds.service.Close()
}



