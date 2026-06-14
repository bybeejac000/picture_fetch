package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"photo_fetch/picture_processing"

	"github.com/gorilla/websocket"
)

type UnifiEvent struct {
	Type string    `json:"type"` // "add" or "update"
	Item EventItem `json:"item"`
}

type EventItem struct {
	ID               string   `json:"id"`
	ModelKey         string   `json:"modelKey"`         // always "event"
	Type             string   `json:"type"`             // "motion", "smartDetectZone", "smartAudioDetect"
	Start            int64    `json:"start"`            // unix ms timestamp
	End              int64    `json:"end"`              // unix ms timestamp (may be missing on "add")
	Device           string   `json:"device"`           // device ID
	SmartDetectTypes []string `json:"smartDetectTypes"` // ["person"], ["alrmSpeak"], etc
}

func unifiWebsocketReadLoop(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v\n", err)
			return // exits loop, triggers reconnect
		}

		var event UnifiEvent
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("JSON unmarshal error: %v\n", err)
			continue // skip bad message, keep listening
		}

		// Guard: skip events with no smart detect types
		if len(event.Item.SmartDetectTypes) == 0 {
			continue
		}

		if event.Item.SmartDetectTypes[0] != "person" {
			continue
		}

		fmt.Println("Person detected, capturing frames")
		picture_processing.CaptureFrames()
		picture_processing.ProcessImages()
	}
}
