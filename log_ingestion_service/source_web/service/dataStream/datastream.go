package datastream

import (
	"context"
	"fmt"
	logger "sourceweb/logging"
	"sync"
)

type DataStreamServiceInterface interface {
	Connect(ctx context.Context) error
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

// ----------------------------- factory to create data
// stream ---------------------
func CreateDataStream(provider string, logger *logger.Logger) (*DataStreamService, error) {
	var service DataStreamServiceInterface
	once.Do(func() {
		StreamService = &DataStreamService{}
	})
	var err error

	switch provider {
	case "KAFKA":
		service, err = NewKafkaService(logger)
		StreamService.service = service
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err != nil {
		return nil, err
	}

	return StreamService, err
}

// ------------- data stream service methods
func (ds *DataStreamService) Connect(ctx context.Context) error {
	return ds.service.Connect(ctx)
}

func (ds *DataStreamService) ProduceMessage(ctx context.Context, message []byte, topicName string) (bool, error) {
	return ds.service.ProduceMessage(ctx, message, topicName)
}

func (ds *DataStreamService) Close() {
	ds.service.Close()
}