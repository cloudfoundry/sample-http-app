package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DefaultPort     = "8080"
	DefaultWaitTime = 1 * time.Second
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	waitTime, err := time.ParseDuration(os.Getenv("WAIT_TIME"))
	if err != nil {
		waitTime = DefaultWaitTime
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	handler := NewHandler(waitTime)
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	fmt.Println("Serving on port: " + port)

	select {
	case err := <-errCh:
		fmt.Fprintln(os.Stderr, "error starting server:", err)
		os.Exit(1)
	case <-signals:
		ctx, cancel := context.WithTimeout(context.Background(), waitTime)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error shutting down server:", err)
			os.Exit(1)
		}
	}
}
