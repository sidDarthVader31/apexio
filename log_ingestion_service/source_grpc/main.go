package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"source_grpc/config"
	"source_grpc/constants"
	"source_grpc/services/dataservice"
	"google.golang.org/grpc"
)

var DataStreamService *dataservice.DataStreamService;
var err error


type logServer struct {
  UnimplementedLoggingServiceServer
}

 func (s *logServer) IngestLog( ctx context.Context, in *LogRequest) (*LogResponse, error) {
  fmt.Println("in:", in)
  value, err := json.Marshal(in.Entry)
  if err != nil {
    return nil, err
  }
   _, err = DataStreamService.ProduceMessage(context.Background(), value, constants.LogTopic)
  if err != nil {
    res := LogResponse{
      Message: "error ingesting log",
      Success: false,
    }
    return &res, nil
  }
  res := LogResponse{
    Message: "log ingested successfully",
    Success: true,
  }
  return &res, nil;
}

func main(){
  flag.Parse()
  config.InitEnv()
  DataStreamService, err = dataservice.GetDataStreamService(config.Config.MESSAGE_BROKER)
  if err!= nil{
    log.Fatalf("Error connecting with kafka : %v", err)
    os.Exit(1)
  }
   connectErr := DataStreamService.Connect(context.Background())
  if connectErr != nil{
    fmt.Println("error connecting to data service:", connectErr)
  }

  lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *&config.Config.PORT))
  if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
  s := grpc.NewServer()
  if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

  RegisterLoggingServiceServer(s, &logServer{})

}

