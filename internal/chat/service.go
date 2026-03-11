package chat

import (
	"sync"
	"time"

	"n8n-chat/internal/models"
)

// Service сервис чата
type Service struct {
	rooms      map[string]*models.Room
	messages   map[string][]*models.Message // roomID -> messages
	messagesMu sync.RWMutex
	roomsMu    sync.RWMutex
}

// NewService создаёт новый сервис чата
func NewService() *Service {
	service := &Service{
		rooms:    make(map[string]*models.Room),
		messages: make(map[string][]*models.Message),
	}

	// Создаём дефолтные комнаты
	service.createDefaultRooms()

	return service
}

// createDefaultRooms создаёт комнаты по умолчанию
func (s *Service) createDefaultRooms() {
	defaultRooms := []struct {
		name        string
		description string
		isPrivate   bool
	}{
		{"general", "Общий чат для всех", false},
		{"random", "Свободные темы", false},
		{"tech", "Технические обсуждения", false},
		{"go-lang", "Go программирование", false},
		{"news", "Новости и объявления", false},
	}

	for _, room := range defaultRooms {
		s.CreateRoom(room.name, room.description, room.isPrivate, "")
	}
}

// CreateRoom создаёт новую комнату
func (s *Service) CreateRoom(name, description string, isPrivate bool, ownerID string) *models.Room {
	s.roomsMu.Lock()
	defer s.roomsMu.Unlock()

	room := &models.Room{
		ID:          generateID(),
		Name:        name,
		Description: description,
		IsPrivate:   isPrivate,
		OwnerID:     ownerID,
		MemberCount: 0,
		CreatedAt:   time.Now(),
	}

	s.rooms[room.ID] = room
	s.messages[room.ID] = make([]*models.Message, 0)

	return room
}

// GetRoom получает комнату по ID
func (s *Service) GetRoom(roomID string) (*models.Room, error) {
	s.roomsMu.RLock()
	defer s.roomsMu.RUnlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return nil, nil
	}

	return room, nil
}

// GetAllRooms получает все комнаты
func (s *Service) GetAllRooms() []*models.Room {
	s.roomsMu.RLock()
	defer s.roomsMu.RUnlock()

	rooms := make([]*models.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

// AddMessage добавляет сообщение в комнату
func (s *Service) AddMessage(roomID, userID, username, content, messageType string) *models.Message {
	s.messagesMu.Lock()
	defer s.messagesMu.Unlock()

	message := &models.Message{
		ID:        generateID(),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		Content:   content,
		Timestamp: time.Now(),
		Type:      messageType,
	}

	if _, exists := s.messages[roomID]; !exists {
		s.messages[roomID] = make([]*models.Message, 0)
	}

	s.messages[roomID] = append(s.messages[roomID], message)

	// Ограничиваем историю 100 сообщениями
	if len(s.messages[roomID]) > 100 {
		s.messages[roomID] = s.messages[roomID][1:]
	}

	return message
}

// GetMessages получает сообщения из комнаты
func (s *Service) GetMessages(roomID string, limit int) []*models.Message {
	s.messagesMu.RLock()
	defer s.messagesMu.RUnlock()

	messages, exists := s.messages[roomID]
	if !exists {
		return []*models.Message{}
	}

	// Возвращаем последние limit сообщений
	if len(messages) > limit {
		return messages[len(messages)-limit:]
	}

	return messages
}

// generateID генерирует уникальный ID
func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}
