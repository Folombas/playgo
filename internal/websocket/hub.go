package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"n8n-chat/internal/models"
)

// Client представляет подключенного клиента
type Client struct {
	ID       string
	UserID   string
	Username string
	RoomID   string
	Conn     *websocket.Conn
	Send     chan []byte
	mu       sync.RWMutex
}

// Hub управляет всеми клиентами
type Hub struct {
	clients     map[string]*Client            // clientID -> Client
	roomClients map[string]map[string]*Client // roomID -> clientID -> Client
	Register    chan *Client
	Unregister  chan *Client
	Broadcast   chan *BroadcastMessage
	mu          sync.RWMutex
}

// BroadcastMessage сообщение для рассылки
type BroadcastMessage struct {
	RoomID  string
	Message []byte
}

// NewHub создаёт новый хаб
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		roomClients: make(map[string]map[string]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan *BroadcastMessage, 256),
	}
}

// Run запускает хаб
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastToRoom(message.RoomID, message.Message)
		}
	}
}

// registerClient регистрирует клиента
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	// Добавляем в комнату
	if _, exists := h.roomClients[client.RoomID]; !exists {
		h.roomClients[client.RoomID] = make(map[string]*Client)
	}
	h.roomClients[client.RoomID][client.ID] = client

	log.Printf("🔌 Client connected: %s (room: %s)", client.Username, client.RoomID)
}

// unregisterClient отключает клиента
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client.ID]; exists {
		delete(h.clients, client.ID)
		close(client.Send)

		// Удаляем из комнаты
		if roomClients, exists := h.roomClients[client.RoomID]; exists {
			delete(roomClients, client.ID)
			log.Printf("🔌 Client disconnected: %s (room: %s)", client.Username, client.RoomID)
		}
	}
}

// broadcastToRoom отправляет сообщение всем в комнате
func (h *Hub) broadcastToRoom(roomID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomClients, exists := h.roomClients[roomID]; exists {
		for _, client := range roomClients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}

// BroadcastToRoom отправляет сообщение в комнату
func (h *Hub) BroadcastToRoom(roomID string, msgType string, payload interface{}) {
	message := models.WSMessage{
		Type:    msgType,
		RoomID:  roomID,
		Payload: payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Error marshaling message: %v", err)
		return
	}

	h.Broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: data,
	}
}

// GetOnlineUsersInRoom получает пользователей онлайн в комнате
func (h *Hub) GetOnlineUsersInRoom(roomID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0)
	if roomClients, exists := h.roomClients[roomID]; exists {
		for _, client := range roomClients {
			users = append(users, client.Username)
		}
	}

	return users
}

// HandleClient обрабатывает подключение клиента
func (h *Hub) HandleClient(client *Client) {
	defer func() {
		h.Unregister <- client
	}()

	// Отправляем приветственное сообщение
	h.BroadcastToRoom(client.RoomID, "user_joined", map[string]string{
		"user_id":  client.UserID,
		"username": client.Username,
	})

	// Отправляем список пользователей онлайн
	h.BroadcastToRoom(client.RoomID, "online_users", h.GetOnlineUsersInRoom(client.RoomID))

	// Читаем сообщения от клиента
	go client.readPump(h)

	// Пишем сообщения клиенту
	client.writePump()
}

// readPump читает сообщения от клиента
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}

		// Парсим сообщение
		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("❌ Error parsing message: %v", err)
			continue
		}

		// Обрабатываем сообщение
		c.handleMessage(&wsMsg, hub)
	}
}

// handleMessage обрабатывает входящее сообщение
func (c *Client) handleMessage(msg *models.WSMessage, hub *Hub) {
	switch msg.Type {
	case "message":
		// Сообщение в чат
		content, ok := msg.Payload.(string)
		if !ok {
			return
		}

		message := models.Message{
			ID:        generateID(),
			RoomID:    c.RoomID,
			UserID:    c.UserID,
			Username:  c.Username,
			Content:   content,
			Timestamp: time.Now(),
			Type:      "text",
		}

		hub.BroadcastToRoom(c.RoomID, "message", message)

	case "typing":
		// Статус набора текста
		isTyping, ok := msg.Payload.(bool)
		if !ok {
			return
		}

		typing := models.TypingStatus{
			UserID:   c.UserID,
			Username: c.Username,
			RoomID:   c.RoomID,
			IsTyping: isTyping,
		}

		hub.BroadcastToRoom(c.RoomID, "typing", typing)

	case "heartbeat":
		// Heartbeat для поддержания соединения
		hub.BroadcastToRoom(c.RoomID, "heartbeat", time.Now().Unix())
	}
}

// writePump пишет сообщения клиенту
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateID генерирует уникальный ID
func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}
