package main

import (
	"fmt"
	"net/http"
	"time"
)

type handler struct {
	waitTime time.Duration
}

func NewHandler(waitTime time.Duration) *handler {
	return &handler{
		waitTime: waitTime,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello ")
	time.Sleep(h.waitTime)
	fmt.Fprintf(w, "world!\n")
}
