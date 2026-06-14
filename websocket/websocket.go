package websocket

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var conn *websocket.Conn

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

func InitializeSockets() {
	createWebsocketInjectPictures()
	fmt.Printf("Websockets initialized.\n")

}
