package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)



type logServer struct {
  UnimplementedLoggingServiceServer
}

 func (s *logServer) IngestLog( ctx context.Context, in *LogRequest) (*LogResponse, error) {
  fmt.Println("in:", in)
  value, err := json.Marshal(in.Entry)
  if err != nil {
    return nil, err
  }
  ingestToKafka(value, logTopic);
  res := LogResponse{
    Message: "log ingested successfully",
    Success: true,
  }
  return &res, nil;
}

func main(){
  flag.Parse()
  connectKafka();
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










