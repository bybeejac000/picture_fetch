package realtime

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"photo_fetch/internal/config"
	"photo_fetch/internal/faces"
)

type Manager struct {
	cfg        *config.Config
	faces      *faces.Processor
	upgrader   websocket.Upgrader
	mu         sync.Mutex
	injectConn *websocket.Conn
}

type injectMessage struct {
	MessageType int      `json:"messageType"`
	Message     []string `json:"message"`
}

func NewManager(cfg *config.Config, processor *faces.Processor) *Manager {
	return &Manager{
		cfg:   cfg,
		faces: processor,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (m *Manager) Register(mux *http.ServeMux) {
	mux.HandleFunc("/injectPictures", func(w http.ResponseWriter, r *http.Request) {
		conn, err := m.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("websocket upgrade failed: %v", err)
			return
		}
		m.setInjectConn(conn)
		log.Println("websocket connection established for inject pictures")
		m.readLoop(conn)
	})
}

func (m *Manager) Start() {
	go m.connectWithReconnect()
	log.Println("websockets initialized")
}

func (m *Manager) setInjectConn(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.injectConn = conn
}

func (m *Manager) inject(messageType int, msg []string) error {
	m.mu.Lock()
	conn := m.injectConn
	m.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("websocket connection is not established")
	}

	payload := injectMessage{
		MessageType: messageType,
		Message:     msg,
	}

	if err := conn.WriteJSON(payload); err != nil {
		log.Printf("websocket write error: %v", err)
		return err
	}

	log.Printf("message %v sent successfully", msg)
	return nil
}

func (m *Manager) readLoop(conn *websocket.Conn) {
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("websocket read error: %v", err)
			break
		}

		log.Printf("received message: %s message type: %d", msg, messageType)

		if err := m.handleMessage(messageType, msg); err != nil {
			log.Printf("message handling error: %v", err)
			break
		}
	}
}

func (m *Manager) handleMessage(messageType int, msg []byte) error {
	return nil
}
