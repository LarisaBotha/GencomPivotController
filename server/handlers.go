package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func handleSSE(w http.ResponseWriter, r *http.Request) {

	setCORSHeaders(w)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		id = "controller"
	}

	client := &Client{
		id:   id,
		send: make(chan string),
	}

	clientsMu.Lock()
	clients[id] = client
	clientsMu.Unlock()

	log.Printf("✅ Client connected: %s\n", id)

	// Remove client on disconnect
	defer func() {
		clientsMu.Lock()
		delete(clients, id)
		clientsMu.Unlock()
		close(client.send)
		log.Printf("❌ Client disconnected: %s\n", id)
	}()

	// Send initial connected message
	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func handleStart(w http.ResponseWriter, r *http.Request) {

	// CORS
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	////////////////////////////////////////

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PivotCommand
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		http.Error(w, "Missing pivot ID", http.StatusBadRequest)
		return
	}

	sendToClient(req.ID, "start")
	w.WriteHeader(http.StatusOK)
}

func handleStop(w http.ResponseWriter, r *http.Request) {

	// CORS
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	////////////////////////////////////////

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PivotCommand
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		http.Error(w, "Missing pivot ID", http.StatusBadRequest)
		return
	}

	sendToClient(req.ID, "stop")
	w.WriteHeader(http.StatusOK)
}
