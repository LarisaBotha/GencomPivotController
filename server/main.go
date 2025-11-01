package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Client struct {
	id   string
	send chan string
}

var (
	clients   = make(map[string]*Client)
	clientsMu sync.Mutex
)

func main() {
	http.HandleFunc("/api/ping", handlePing)
	http.HandleFunc("/api/sse", handleSSE)
	http.HandleFunc("/api/start", handleStart)
	http.HandleFunc("/api/stop", handleStop)

	fmt.Println("🚀 Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
