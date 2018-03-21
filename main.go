package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

const (
	DefaultPort     = "8080"
	DefaultWaitTime = 1 * time.Second
)

type handler struct {
	waitTime time.Duration
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	waitParam := r.FormValue("wait")
	if waitParam == "" {
		waitParam = "10us"
	}
	waitTime, err := time.ParseDuration(waitParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "hello ")
	time.Sleep(waitTime)
	fmt.Fprintf(w, "world!\n")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// waitTime, err := time.ParseDuration(os.Getenv("WAIT_TIME"))
	// if err != nil {
	// 	waitTime = DefaultWaitTime
	// }

	signals := make(chan os.Signal, 1)
	signal.Notify(signals)

	handler := &handler{}

	runner := http_server.New(fmt.Sprintf(":%s", port), handler)
	runner = sigmon.New(runner, syscall.SIGTERM, syscall.SIGINT)
	process := ifrit.Invoke(runner)
	fmt.Println("Serving on port: " + port)
	<-process.Wait()
}
