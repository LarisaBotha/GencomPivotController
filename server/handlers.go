package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func handleTest(w http.ResponseWriter, r *http.Request) {

	body := []byte("Hello from Test")

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	// CORS
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	////////////////////////////////////////

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

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

func handleUpdate(w http.ResponseWriter, r *http.Request) {

	// CORS
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	////////////////////////////////////////

	var req PivotUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Find the pivot by IMEI
	var pivotID uuid.UUID
	err := DB.QueryRow(ctx, `SELECT id FROM pivots WHERE imei=$1`, req.IMEI).Scan(&pivotID)
	if err != nil {
		http.Error(w, fmt.Sprintf("pivot not found for imei %s", req.IMEI), http.StatusNotFound)
		return
	}

	// Update the status
	_, err = DB.Exec(ctx, `
        INSERT INTO pivot_status (pivot_id, position_deg, speed_pct, direction, wet, status, last_update)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
        ON CONFLICT (pivot_id)
        DO UPDATE SET
            position_deg = EXCLUDED.position_deg,
            speed_pct = EXCLUDED.speed_pct,
            direction = EXCLUDED.direction,
            wet = EXCLUDED.wet,
            status = EXCLUDED.status,
            last_update = NOW()
    `, pivotID, req.PositionDeg, req.SpeedPct, req.Direction, req.Wet, req.Status)
	if err != nil {
		log.Println("Error updating pivot status:", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Fetch queued (unacknowledged) commands
	rows, err := DB.Query(ctx, `
        SELECT id, command, payload
        FROM pivot_command_queue
        WHERE pivot_id=$1 AND acknowledged=false
        ORDER BY created_at ASC
    `, pivotID)
	if err != nil {
		log.Println("Error querying commands:", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var commands []PivotCommand
	for rows.Next() {
		var cmd PivotCommand
		var payload []byte
		rows.Scan(&cmd.ID, &cmd.Command, &payload)
		cmd.Payload = payload
		commands = append(commands, cmd)
	}

	// Mark commands as acknowledged
	if len(commands) > 0 {
		_, err = DB.Exec(ctx, `
            UPDATE pivot_command_queue
            SET acknowledged=true, acknowledged_at=$2
            WHERE pivot_id=$1 AND acknowledged=false
        `, pivotID, time.Now())
		if err != nil {
			log.Println("Error marking commands acknowledged:", err)
		}
	}

	// Return
	writeJSON(w, http.StatusOK, PivotUpdateResponse{Commands: commands})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {

	// CORS
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	////////////////////////////////////////

	imei := r.URL.Query().Get("imei")
	if imei == "" {
		http.Error(w, "imei required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Find pivot by IMEI
	var pivotID string
	err := DB.QueryRow(ctx, `SELECT id FROM pivots WHERE imei=$1`, imei).Scan(&pivotID)
	if err != nil {
		http.Error(w, "pivot not found", http.StatusNotFound)
		return
	}

	// Fetch current status into struct
	var status PivotStatus
	err = DB.QueryRow(ctx, `
        SELECT position_deg, speed_pct, direction::text, wet, status
        FROM pivot_status
        WHERE pivot_id=$1
    `, pivotID).Scan(&status.PositionDeg, &status.SpeedPct, &status.Direction, &status.Wet, &status.Status)
	if err != nil {
		http.Error(w, "pivot status not found", http.StatusNotFound)
		return
	}

	// Return
	writeJSON(w, http.StatusOK, status)
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	// CORS
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	////////////////////////////////////////

	var req RegisterCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.IMEI == "" || req.Command == "" {
		http.Error(w, "imei and command are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Find the pivot by IMEI
	var pivotID uuid.UUID
	err := DB.QueryRow(ctx, `SELECT id FROM pivots WHERE imei=$1`, req.IMEI).Scan(&pivotID)
	if err != nil {
		http.Error(w, "pivot not found", http.StatusNotFound)
		return
	}

	// Convert payload to JSON, allow nil
	var payloadJSON []byte
	if req.Payload != nil {
		payloadJSON, err = json.Marshal(req.Payload)
		if err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
	} else {
		payloadJSON = nil
	}

	// Insert command into the queue
	_, err = DB.Exec(ctx, `
        INSERT INTO pivot_command_queue (pivot_id, command, payload)
        VALUES ($1, $2, $3)
    `, pivotID, req.Command, payloadJSON)
	if err != nil {
		log.Println("Error inserting command:", err)
		http.Error(w, "failed to register command", http.StatusInternalServerError)
		return
	}

	sendToClient(req.IMEI, req.Command)

	// Respond
	w.WriteHeader(http.StatusOK)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	// CORS
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	////////////////////////////////////////

	var req RegisterPivotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.IMEI == "" {
		http.Error(w, "IMEI is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Check if pivot with this IMEI already exists
	var existingID string
	err := DB.QueryRow(ctx, `SELECT id FROM pivots WHERE imei=$1`, req.IMEI).Scan(&existingID)
	if err == nil {
		// Pivot already exists
		w.WriteHeader(http.StatusAccepted)
		return
	} else if err != pgx.ErrNoRows {
		log.Println("Error checking existing pivot:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Insert new pivot
	pivotID := uuid.New()
	_, err = DB.Exec(ctx, `
        INSERT INTO pivots (id, imei)
        VALUES ($1, $2)
    `, pivotID, req.IMEI)
	if err != nil {
		log.Println("Error inserting pivot:", err)
		http.Error(w, "failed to register pivot", http.StatusInternalServerError)
		return
	}

	// Return
	w.WriteHeader(http.StatusOK)
}
