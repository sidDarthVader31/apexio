package main

import (
	"context"
	"fmt"
	"log"
	"net"

	collogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"github.com/sidDarthVader31/apexio/pkg/schema"
	"google.golang.org/grpc"
)

type otlpGRPCServer struct {
	collogsv1.UnimplementedLogsServiceServer
	pub     *publisher
	metrics *ingestMetrics
}

func (s *otlpGRPCServer) Export(ctx context.Context, req *collogsv1.ExportLogsServiceRequest) (*collogsv1.ExportLogsServiceResponse, error) {
	events, err := schemaLogEventsFromOTLP(req)
	if err != nil {
		s.metrics.incOTLPErr()
		return nil, err
	}
	if len(events) == 0 {
		s.metrics.incOTLPOK(0)
		return &collogsv1.ExportLogsServiceResponse{}, nil
	}
	if err := s.pub.publishEvents(ctx, events); err != nil {
		s.metrics.incOTLPErr()
		return nil, fmt.Errorf("publish failed: %w", err)
	}
	s.metrics.incOTLPOK(uint64(len(events)))
	return &collogsv1.ExportLogsServiceResponse{}, nil
}

func schemaLogEventsFromOTLP(req *collogsv1.ExportLogsServiceRequest) ([]schema.LogEvent, error) {
	return schema.LogEventsFromOTLP(req)
}

func startOTLPGRPC(addr string, pub *publisher, metrics *ingestMetrics) (*grpc.Server, error) {
	if addr == "" {
		return nil, nil
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := grpc.NewServer()
	collogsv1.RegisterLogsServiceServer(srv, &otlpGRPCServer{pub: pub, metrics: metrics})
	go func() {
		log.Printf("otlp grpc listening on %s", addr)
		if err := srv.Serve(lis); err != nil {
			log.Printf("otlp grpc stopped: %v", err)
		}
	}()
	return srv, nil
}
