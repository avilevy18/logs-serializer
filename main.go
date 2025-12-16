package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
)

func main() {
	stdout := flag.Bool("stdout", false, "print logs and metrics to standard out")
	logsFlag := flag.Bool("logs", false, "run only the logs server")
	metricsFlag := flag.Bool("metrics", false, "run only the metrics server")
	allFlag := flag.Bool("all", true, "run both servers")
	path := flag.String("path", "/tmp/log-serializer", "the path to write the files to")
	host := flag.String("host", "0.0.0.0", "the host to run the servers on")
	logsPort := flag.Int("logs-port", 18888, "the port to run the logs server on")
	metricsPort := flag.Int("metrics-port", 18889, "the port to run the metrics server on")
	flag.Parse()

	// If logs or metrics are specified, don't run all.
	if *logsFlag || *metricsFlag {
		*allFlag = false
	}

	var wg sync.WaitGroup

	var logsServer *LogsTestServer
	var metricsServer *MetricTestServer
	var err error

	if *allFlag || *logsFlag {
		logsServer, err = NewLoggingTestServer(*stdout, *path, *host, *logsPort)
		if err != nil {
			panic(err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("starting logs server at %s\n", logsServer.Endpoint)
			logsServer.Serve()
		}()
	}

	if *allFlag || *metricsFlag {
		metricsServer, err = NewMetricTestServer(*stdout, *path, *host, *metricsPort)
		if err != nil {
			panic(err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("starting metrics server at %s\n", metricsServer.Endpoint)
			metricsServer.Serve()
		}()
	}

	if logsServer == nil && metricsServer == nil {
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	if logsServer != nil {
		logsServer.Shutdown()
	}
	if metricsServer != nil {
		metricsServer.Shutdown()
	}

	wg.Wait()
}
