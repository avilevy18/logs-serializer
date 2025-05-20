package main

func main() {
	testServer, err := NewLoggingTestServer()
	if err != nil {
		panic(err)
	}
	testServer.Serve()
	defer testServer.Shutdown()
}
