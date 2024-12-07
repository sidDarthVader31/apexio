package datastream

import (
	"context"
	"errors"
	"fmt"
	"sync"
)



type IDataStream interface{
  Connect(context context.Context) error
  Consume(context context.Context, topics []string) 
  Close()
}

type DataStreamService struct {
  service IDataStream
}


func (d *DataStreamService) Connect(context context.Context) error{
  return d.service.Connect(context)
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

func CreateDataStream(context  context.Context, serviceName string) (*DataStreamService, error){
   once.Do(func() {
        StreamService = &DataStreamService{}
    })

  switch (serviceName){
  case "KAFKA":
    fmt.Println("getting kafka service")
    service, err := getNewkafkaStream()
    if err != nil {
      fmt.Println("error while getting kafka service:", err)
      return nil, err
    }
    StreamService.service = service
    return StreamService, nil
  default:
    return nil, errors.New("Invalid stream name ")
  } 
}
