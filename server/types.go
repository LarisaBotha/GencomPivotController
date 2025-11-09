package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type PivotStatus struct {
	PositionDeg float64
	SpeedPct    float64
	Direction   string
	Wet         bool
	Status      string
}

type PivotCommand struct {
	ID      uuid.UUID       `json:"id"`
	Command string          `json:"command"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type PivotUpdateRequest struct {
	IMEI        string  `json:"imei"`
	PositionDeg float64 `json:"position"`
	SpeedPct    float64 `json:"speed"`
	Direction   string  `json:"direction"`
	Wet         bool    `json:"wet"`
	Status      string  `json:"status"`
}

type PivotUpdateResponse struct {
	Commands []PivotCommand `json:"commands"`
}

type RegisterCommandRequest struct {
	IMEI    string                 `json:"imei"`
	Command string                 `json:"command"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type RegisterCommandResponse struct {
	ID string `json:"id"`
}

type RegisterPivotRequest struct {
	IMEI string `json:"imei"`
}
