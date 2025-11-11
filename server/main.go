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
	InitDB()
	defer CloseDB()

	http.HandleFunc("/api/ping", handlePing)
	http.HandleFunc("/api/test", handleTest)
	http.HandleFunc("/api/sse", handleSSE)
	http.HandleFunc("/api/register", handleRegister)
	http.HandleFunc("/api/update", handleUpdate)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/command", handleCommand)

	fmt.Println("🚀 Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
