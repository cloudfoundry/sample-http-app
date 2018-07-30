package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	defaultPort = "8080"
)

// connManager keeps a record of active connections and their states
type connManager struct {
	activeConns map[net.Conn]*atomic.Value
	mu          sync.Mutex
}

// setState is a callback called from the http server when the connection state changes
func (cm *connManager) setState(nc net.Conn, state http.ConnState) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.activeConns == nil {
		cm.activeConns = make(map[net.Conn]*atomic.Value)
	}
	switch state {
	case http.StateNew:
		cm.activeConns[nc] = &atomic.Value{}
	case http.StateHijacked, http.StateClosed:
		delete(cm.activeConns, nc)
	}
	if v, ok := cm.activeConns[nc]; ok {
		v.Store(state)
	}
}

// closeIdleConns closes idle connections and reports if there are still
// any in-flight connections
func (cm *connManager) closeIdleConns() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	inflight := false
	for nc, v := range cm.activeConns {
		state, ok := v.Load().(http.ConnState)
		if !ok || state == http.StateNew || state == http.StateActive {
			inflight = true
			continue
		}
		nc.Close()
		delete(cm.activeConns, nc)
	}
	return inflight
}

type handler struct{}

func newHandler() *handler {
	return &handler{}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d, err := time.ParseDuration(r.FormValue("wait"))
	if err != nil {
		d = 10 * time.Microsecond
	}
	fmt.Fprintf(w, "hello ")
	time.Sleep(d)
	fmt.Fprintf(w, "world!\n")
}

func main() {
	// Register signal handler for SIGTERM and SIGINT
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	connMgr := new(connManager)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	addr := fmt.Sprintf(":%s", port)

	server := http.Server{
		Addr:      addr,
		Handler:   newHandler(),
		ConnState: connMgr.setState, // register callback when the connection state changes
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error starting server:", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()
	fmt.Println("Serving on port: " + port)

	select {
	case err := <-errCh:
		fmt.Fprintln(os.Stderr, "error starting server:", err)
		os.Exit(1)
	case <-signals:
		// It is required that the listener is closed as soon as the signal is
		// received to prevent any new traffic from getting in
		listener.Close()

		// busy loop until all connections are closed
		ticker := time.NewTicker(10 * time.Millisecond)
		for {
			if stillActive := connMgr.closeIdleConns(); !stillActive {
				return
			}
			<-ticker.C
		}
	}
}
