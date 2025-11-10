package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func setCORSHeaders(w http.ResponseWriter) {

	w.Header().Set("Access-Control-Allow-Origin", "https://gencompivotcontroller.onrender.com") // "*") //
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, `Error writing response`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func sendToClient(id string, message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	client, exists := clients[id]
	if !exists {
		log.Printf("⚠️ No client connected with ID %s\n", id)
		return
	}

	select {
	case client.send <- message:
		log.Printf("📡 Sent '%s' to %s\n", message, id)
	default:
		log.Printf("⚠️ Skipping %s (channel full)\n", id)
	}
}
