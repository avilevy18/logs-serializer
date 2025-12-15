package main

import (
	"flag"
	"fmt"
	"runtime"
)

func main() {
	stdout := flag.Bool("stdout", false, "print logs and metrics to standard out")
	logs := flag.Bool("logs", false, "run only the logs server")
	metrics := flag.Bool("metrics", false, "run only the metrics server")
	all := flag.Bool("all", true, "run both servers")
	flag.Parse()

	if *all || *logs {
		logsServer, err := NewLoggingTestServer(*stdout)
		if err != nil {
			panic(err)
		}
		go func() {
			fmt.Printf("starting logs server at %s\n", logsServer.Endpoint)
			logsServer.Serve()
		}()
		defer logsServer.Shutdown()
	}

	if *all || *metrics {
		metricsServer, err := NewMetricTestServer(*stdout)
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
