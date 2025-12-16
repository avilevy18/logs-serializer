package main

import (
	"context"
	"net"
	"strings"
	"sync"

	logpb "cloud.google.com/go/logging/apiv2/loggingpb"
	metricpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/avilevy18/logs-serializer/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type LogsTestServer struct {
	lis                     net.Listener
	srv                     *grpc.Server
	Endpoint                string
	userAgent               string
	writeLogEntriesRequests []*logpb.WriteLogEntriesRequest
	mu                      sync.Mutex
	logger                  logs.StructuredLogger
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
	md, _ := metadata.FromIncomingContext(ctx)
	f.logsTestServer.logger.Infow("received log", "request", request, "metadata", md)
	f.logsTestServer.appendWriteLogEntriesRequest(ctx, request)
	return &logpb.WriteLogEntriesResponse{}, nil
}

func NewLoggingTestServer(stdout bool, path string) (*LogsTestServer, error) {
	var logger logs.StructuredLogger
	var err error
	if stdout {
		logger = logs.NewStdoutLogger()
	} else {
		logger, err = logs.NewFileLogger(path, "log")
		if err != nil {
			return nil, err
		}
	}

	srv := grpc.NewServer()
	lis, err := net.Listen("tcp", "localhost:18888")
	if err != nil {
		return nil, err
	}
	testServer := &LogsTestServer{
		Endpoint: lis.Addr().String(),
		lis:      lis,
		srv:      srv,
		logger:   logger,
	}
	logpb.RegisterLoggingServiceV2Server(
		srv,
		&fakeLoggingServiceServer{logsTestServer: testServer},
	)

	return testServer, nil
}

type MetricTestServer struct {
	lis                      net.Listener
	srv                      *grpc.Server
	Endpoint                 string
	userAgent                string
	createTimeSeriesRequests []*metricpb.CreateTimeSeriesRequest
	mu                       sync.Mutex
	logger                   logs.StructuredLogger
}

func (m *MetricTestServer) Shutdown() {
	m.srv.GracefulStop()
}

func (m *MetricTestServer) Serve() {
	//nolint:errcheck
	m.srv.Serve(m.lis)
}

func (m *MetricTestServer) CreateTimeSeriesRequests() []*metricpb.CreateTimeSeriesRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	reqs := m.createTimeSeriesRequests
	m.createTimeSeriesRequests = nil
	return reqs
}

// Pops out the UserAgent from the most recent CreateTimeSeries.
func (m *MetricTestServer) UserAgent() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	ua := m.userAgent
	m.userAgent = ""
	return ua
}

func (m *MetricTestServer) appendCreateTimeSeriesRequest(ctx context.Context, req *metricpb.CreateTimeSeriesRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createTimeSeriesRequests = append(m.createTimeSeriesRequests, req)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		m.userAgent = strings.Join(md.Get("User-Agent"), ";")
	}
}

type fakeMetricServiceServer struct {
	metricpb.UnimplementedMetricServiceServer
	metricTestServer *MetricTestServer
}

func (f *fakeMetricServiceServer) CreateTimeSeries(
	ctx context.Context,
	request *metricpb.CreateTimeSeriesRequest,
) (*emptypb.Empty, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	f.metricTestServer.logger.Infow("received metric", "request", request, "metadata", md)
	f.metricTestServer.appendCreateTimeSeriesRequest(ctx, request)
	return &emptypb.Empty{}, nil
}

func NewMetricTestServer(stdout bool, path string) (*MetricTestServer, error) {
	var logger logs.StructuredLogger
	var err error
	if stdout {
		logger = logs.NewStdoutLogger()
	} else {
		logger, err = logs.NewFileLogger(path, "metric")
		if err != nil {
			return nil, err
		}
	}
	srv := grpc.NewServer()
	lis, err := net.Listen("tcp", "localhost:18889")
	if err != nil {
		return nil, err
	}
	testServer := &MetricTestServer{
		Endpoint: lis.Addr().String(),
		lis:      lis,
		srv:      srv,
		logger:   logger,
	}
	metricpb.RegisterMetricServiceServer(
		srv,
		&fakeMetricServiceServer{metricTestServer: testServer},
	)

	return testServer, nil
}
