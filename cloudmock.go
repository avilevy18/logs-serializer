package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net"
	"os"
	"strings"
	"sync"

	logpb "cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type LogsTestServer struct {
	lis net.Listener
	srv *grpc.Server
	// Endpoint where the gRPC server is listening
	Endpoint                string
	userAgent               string
	writeLogEntriesRequests []*logpb.WriteLogEntriesRequest
	mu                      sync.Mutex
}

func (l *LogsTestServer) Shutdown() {
	l.srv.GracefulStop()
}

func (l *LogsTestServer) Serve() {
	//nolint:errcheck
	l.srv.Serve(l.lis)
}

func (l *LogsTestServer) CreateWriteLogEntriesRequests() []*logpb.WriteLogEntriesRequest {
	l.mu.Lock()
	defer l.mu.Unlock()
	reqs := l.writeLogEntriesRequests
	l.writeLogEntriesRequests = nil
	return reqs
}

// Pops out the UserAgent from the most recent CreateWriteLogEntries.
func (l *LogsTestServer) UserAgent() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	ua := l.userAgent
	l.userAgent = ""
	return ua
}

func (l *LogsTestServer) appendWriteLogEntriesRequest(ctx context.Context, req *logpb.WriteLogEntriesRequest) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writeLogEntriesRequests = append(l.writeLogEntriesRequests, req)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		l.userAgent = strings.Join(md.Get("User-Agent"), ";")
	}
}

type fakeLoggingServiceServer struct {
	logpb.UnimplementedLoggingServiceV2Server
	logsTestServer *LogsTestServer
}

func (f *fakeLoggingServiceServer) WriteLogEntries(
		ctx context.Context,
		request *logpb.WriteLogEntriesRequest,
) (*logpb.WriteLogEntriesResponse, error) {
	b, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return &logpb.WriteLogEntriesResponse{}, nil
	}
	fileId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	fmt.Printf("writing to: %s.json")
	err = os.WriteFile(fmt.Sprintf("%s.json", fileId.String()), b, 0644)
	if err != nil {
		return nil, err
	}
	f.logsTestServer.appendWriteLogEntriesRequest(ctx, request)
	return &logpb.WriteLogEntriesResponse{}, nil
}

func NewLoggingTestServer() (*LogsTestServer, error) {
	srv := grpc.NewServer()
	lis, err := net.Listen("tcp", "localhost:18888")
	if err != nil {
		return nil, err
	}
	testServer := &LogsTestServer{
		Endpoint: lis.Addr().String(),
		lis:      lis,
		srv:      srv,
	}
	logpb.RegisterLoggingServiceV2Server(
		srv,
		&fakeLoggingServiceServer{logsTestServer: testServer},
	)

	return testServer, nil
}
