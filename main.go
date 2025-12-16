package main

import (
	"flag"
	"fmt"
	"runtime"
)

func main() {
	stdout := flag.Bool("stdout", false, "print logs and metrics to standard out")
	logsFlag := flag.Bool("logs", false, "run only the logs server")
	metricsFlag := flag.Bool("metrics", false, "run only the metrics server")
	allFlag := flag.Bool("all", true, "run both servers")
	path := flag.String("path", "/tmp/log-serializer", "the path to write the files to")
	flag.Parse()

	if *allFlag || *logsFlag {
		logsServer, err := NewLoggingTestServer(*stdout, *path)
		if err != nil {
			panic(err)
		}
		go func() {
			fmt.Printf("starting logs server at %s\n", logsServer.Endpoint)
			logsServer.Serve()
		}()
		defer logsServer.Shutdown()
	}

	if *allFlag || *metricsFlag {
		metricsServer, err := NewMetricTestServer(*stdout, *path)
		if err != nil {
			panic(err)
		}
		go func() {
			fmt.Printf("starting metrics server at %s\n", metricsServer.Endpoint)
			metricsServer.Serve()
		}()
		defer metricsServer.Shutdown()
	}

	runtime.Goexit()
}
