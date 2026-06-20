package realtime

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"photo_fetch/internal/faces"
	"time"

	"github.com/gorilla/websocket"
)

type unifiEvent struct {
	Type string    `json:"type"`
	Item eventItem `json:"item"`
}

type eventItem struct {
	ID               string   `json:"id"`
	ModelKey         string   `json:"modelKey"`
	Type             string   `json:"type"`
	Start            int64    `json:"start"`
	End              int64    `json:"end"`
	Device           string   `json:"device"`
	SmartDetectTypes []string `json:"smartDetectTypes"`
}

func (m *Manager) connectWithReconnect() {
	url := fmt.Sprintf("wss://%s/proxy/protect/integration/v1/subscribe/events", m.cfg.DoorbellHost)

	dialer := websocket.Dialer{
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	headers.Set("X-API-KEY", m.cfg.UnifiAPIKey)

	for {
		conn, _, err := dialer.Dial(url, headers)
		if err != nil {
			log.Printf("connection failed, retrying in 5s: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("connected to UniFi Protect")

		m.unifiReadLoop(conn)

		conn.Close()

		log.Println("disconnected, reconnecting in 5s")
		time.Sleep(5 * time.Second)
	}
}

func (m *Manager) unifiReadLoop(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("websocket read error: %v", err)
			return
		}

		var event unifiEvent
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("json unmarshal error: %v", err)
			continue
		}

		if len(event.Item.SmartDetectTypes) == 0 || event.Item.End == 0 {
			continue
		}

		if event.Item.SmartDetectTypes[0] != "person" {
			continue
		}

		log.Println("person detected, capturing frames")
		filePath, err := m.faces.Capture()
		if err != nil {
			log.Printf("error capturing frames: %v", err)
			continue
		}

		facesList, detected := m.faces.Process(filePath)
		injectUrlList := []string{}
		if detected {
			url := fmt.Sprintf("http://%s:%s/photo?file=%s", "localhost", m.cfg.GoListenPort, filePath)
			injectUrlList = append(injectUrlList, url)
			for _, faceInfo := range facesList {
				log.Printf("detected face: %s", faceInfo.Name)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				cancel()
				faceImages, err := m.faces.PullPhotosByFace(ctx, []faces.FaceInfo{faceInfo})
				if err != nil {
					log.Printf("error pulling photos by face: %v", err)
					continue
				}
				injectUrlList = append(injectUrlList, faceImages...)
			}
			m.inject(1, injectUrlList)
			log.Printf("injected new photo for person: %s and possibly others", facesList[0].Name)
		}
	}
}
