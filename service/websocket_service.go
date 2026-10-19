package service

import (
	"log"
	"net/http"
	"sync"

	"pet-community/model"

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
	ID       uint   `json:"id"`
	ToID     uint   `json:"toId"`
	FromID   uint   `json:"fromId"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	MediaURL string `json:"mediaUrl,omitempty"`
	Duration int    `json:"duration,omitempty"`
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

// SendMessage 推送完整私信（含媒体字段），供在线接收方即时展示
func (ws *WebSocketService) SendMessage(msg *model.Message) {
	if msg == nil {
		return
	}
	ws.messageCh <- &MessagePayload{
		ID:       msg.ID,
		ToID:     msg.ToID,
		FromID:   msg.FromID,
		Content:  msg.Content,
		Type:     msg.Type,
		MediaURL: msg.MediaURL,
		Duration: msg.Duration,
	}
}
