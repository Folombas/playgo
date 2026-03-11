package models

import (
	"time"
)

// User представляет пользователя чата
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Никогда не отправляем клиенту
	Avatar    string    `json:"avatar"`
	IsOnline  bool      `json:"is_online"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}

// Message представляет сообщение в чате
type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "text", "system", "file"
}

// Room представляет комнату чата
type Room struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"is_private"`
	OwnerID     string    `json:"owner_id"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuthRequest запрос аутентификации
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse ответ аутентификации
type AuthResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresAt int64  `json:"expires_at"`
}

// RegisterRequest запрос регистрации
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// WSMessage сообщение WebSocket
type WSMessage struct {
	Type    string      `json:"type"`    // "message", "join", "leave", "typing", "heartbeat"
	RoomID  string      `json:"room_id"` // ID комнаты
	Payload interface{} `json:"payload"` // Данные сообщения
}

// TypingStatus статус набора текста
type TypingStatus struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	RoomID   string `json:"room_id"`
	IsTyping bool   `json:"is_typing"`
}

// OnlineUsers список пользователей онлайн
type OnlineUsers struct {
	RoomID string   `json:"room_id"`
	Users  []string `json:"users"` // Список username
}
