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

var (
	port = flag.Int("port", 50051, "The server port")
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
  DataStreamService.ProduceMessage(context.Background(), value, constants.LogTopic)
  // ingestToKafka(value, logTopic);
  res := LogResponse{
    Message: "log ingested successfully",
    Success: true,
  }
  return &res, nil;
}

func main(){
  flag.Parse()
  DataStreamService, err = dataservice.GetDataStreamService(config.Config.MESSAGE_BROKER, map[string]string{"baseUrl": config.Config.KAFKA_HOST})
  if err!= nil{
    log.Fatalf("Error connecting with kafka : %v", err)
    os.Exit(1)
  }
  DataStreamService.Connect(context.Background(), map[string]string{"baeUrl":config.Config.KAFKA_HOST})
  lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
  if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
  s := grpc.NewServer()
  RegisterLoggingServiceServer(s, &logServer{})
  if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

