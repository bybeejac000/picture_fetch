package websocket

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"photo_fetch/config"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func createWebsocketInjectPictures() {
	http.HandleFunc("/injectPictures", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("WebSocket upgrade failed: %v\n", err)
			return
		}
		defer conn.Close()

		readLoop(conn, handleMessage)
	})
}

func connectWithReconnect() {
	url := fmt.Sprintf("wss://%s/proxy/protect/integration/v1/subscribe/events", config.DOORBELL_HOST)

	dialer := websocket.Dialer{
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	headers.Set("X-API-KEY", config.UNIFI_API_KEY)

	for {
		conn, _, err := dialer.Dial(url, headers)
		if err != nil {
			log.Printf("Connection failed, retrying in 5s: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Connected to UniFi Protect")

		unifiWebsocketReadLoop(conn)

		conn.Close()

		log.Println("Disconnected, reconnecting in 5s...")
		time.Sleep(5 * time.Second)
	}
}

func InitializeSockets() {
	createWebsocketInjectPictures()
	connectWithReconnect()
	fmt.Printf("Websockets initialized.\n")

}
