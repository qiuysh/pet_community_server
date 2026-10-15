package service

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketService struct {
	clients   sync.Map
	mu        sync.Mutex
	messageCh chan *MessagePayload
}

type MessagePayload struct {
	ToID    uint   `json:"toId"`
	FromID  uint   `json:"fromId"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

func NewWebSocketService() *WebSocketService {
	ws := &WebSocketService{
		messageCh: make(chan *MessagePayload, 100),
	}
	go ws.messageDispatcher()
	return ws
}

func (ws *WebSocketService) messageDispatcher() {
	for payload := range ws.messageCh {
		ws.broadcastToUser(payload.ToID, payload)
	}
}

func (ws *WebSocketService) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	userID := r.Context().Value("user_id").(uint)

	ws.clients.Store(userID, conn)

	log.Printf("User %d connected via WebSocket", userID)

	defer func() {
		ws.clients.Delete(userID)
		conn.Close()
		log.Printf("User %d disconnected from WebSocket", userID)
	}()

	for {
		var msg struct {
			Type string `json:"type"`
		}
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}
	}
}

func (ws *WebSocketService) broadcastToUser(userID uint, payload *MessagePayload) {
	if conn, ok := ws.clients.Load(userID); ok {
		ws.mu.Lock()
		defer ws.mu.Unlock()
		if err := conn.(*websocket.Conn).WriteJSON(payload); err != nil {
			log.Printf("Failed to send message to user %d: %v", userID, err)
			ws.clients.Delete(userID)
		}
	}
}

func (ws *WebSocketService) SendMessage(toID, fromID uint, content, msgType string) {
	ws.messageCh <- &MessagePayload{
		ToID:    toID,
		FromID:  fromID,
		Content: content,
		Type:    msgType,
	}
}
