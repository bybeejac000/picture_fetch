package websocket

import (
	"fmt"

	"github.com/gorilla/websocket"
)

func handleMessage(messageType int, msg []byte) error {
	return nil
}

type WebSocketInjectPicturesMessage struct {
	MessageType int      `json:"messageType"`
	Message     []string `json:"message"`
}

func InjectPictures(conn *websocket.Conn, messageType int, msg []string) error {

	if conn == nil {
		fmt.Printf("WebSocket connection is not established\n")
		return fmt.Errorf("WebSocket connection is not established")
	}

	payload := WebSocketInjectPicturesMessage{
		MessageType: messageType,
		Message:     msg,
	}

	if err := conn.WriteJSON(payload); err != nil {
		fmt.Printf("WebSocket write error: %v\n", err)
		return err
	}

	fmt.Printf("Message %v sent successfully\n", msg)
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
