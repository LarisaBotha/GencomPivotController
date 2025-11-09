package main

import (
	"log"
	"net/http"
)

func setCORSHeaders(w http.ResponseWriter) {

	w.Header().Set("Access-Control-Allow-Origin", "*") //"https://gencompivotcontroller.onrender.com")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
