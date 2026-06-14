package websocket

import (
	"fmt"

	"github.com/gorilla/websocket"
)

func handleMessage(messageType int, msg []byte) error {
	return nil
}

func injectPictures(conn *websocket.Conn, messageType int, msg []byte) error {

	if err := conn.WriteMessage(messageType, msg); err != nil {
		fmt.Printf("WebSocket write error: %v\n", err)
		return err
	}

	fmt.Printf("Message %s sent successfully\n", string(msg))
	return nil
}

func readLoop(conn *websocket.Conn, handleMessage func(int, []byte) error) {
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("WebSocket read error: %v\n", err)
			break
		}

		fmt.Print("Received message: ", string(msg), " Message type:", messageType, "\n")

		if err := handleMessage(messageType, msg); err != nil {
			fmt.Printf("Message handling error: %v\n", err)
			break
		}
	}
}
