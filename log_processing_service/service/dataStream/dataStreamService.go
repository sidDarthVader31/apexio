package datastream

import (
	"context"
	"errors"
	"sync"
)



type IDataStream interface{

  Connect(context context.Context, config map[string]string) error
  Consume(context context.Context, topics []string) 
  Close()
}

type DataStreamService struct {
  service IDataStream
}


func (d *DataStreamService) Connect(context context.Context, config map[string]string) error{
  return d.service.Connect(context, config)
}

func (d *DataStreamService) Consume(context context.Context, topicNames []string) {
 d.service.Consume(context, topicNames)
}

func (d *DataStreamService) Close(){
 d.service.Close()
}

var (
    StreamService *DataStreamService
    once          sync.Once // Ensure initialization is thread-safe
)

func CreateDataStream(context  context.Context, configMap map[string]string, serviceName string) (*DataStreamService, error){
   once.Do(func() {
        StreamService = &DataStreamService{}
    })

  switch (serviceName){
  case "KAFKA":
    service, err := getNewkafkaStream(configMap)
    if err != nil {
      return nil, err
    }
    StreamService.service = service
    return StreamService, nil
  default:
    return nil, errors.New("Invalid stream name ")
  } 
}
