package dataservice

import (
	"context"
	"errors"
	"sync"
)

// ------------------------------------   data stream
// service interface and utils

type IDataStreamService interface{
  Connect(context context.Context) error
  ProduceMessage(context context.Context, message []byte, topicName string) (bool, error)
  Close()
}


type DataStreamService struct{
  service IDataStreamService
}


// data stream service methods
func (d *DataStreamService) Connect(context context.Context) error{
  return d.service.Connect(context)
}

func (d *DataStreamService) ProduceMessage(context context.Context, message []byte, topicName string) (bool, error){
  return d.service.ProduceMessage(context, message, topicName)
}

func (d *DataStreamService) Close(){
 d.service.Close()
}



// --------------------------------------- data stream
// service factory

var (
    StreamService *DataStreamService
    once          sync.Once // Ensure initialization is thread-safe
)


func GetDataStreamService(messaseService string) (*DataStreamService, error){
	var service IDataStreamService
   once.Do(func() {
        StreamService = &DataStreamService{}
    })
	var err error


  switch(messaseService){
  case "KAFKA":
    service, err = getNewkafkaService()
    if err!=nil{
      return nil, err
    }
    StreamService.service = service
    return StreamService, nil
  default:
    return nil, errors.New("Invalid message service option")
  }
}
