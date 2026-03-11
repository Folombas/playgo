package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"n8n-chat/internal/models"
)

// JWTClaims представляет claims для JWT токена
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Service сервис аутентификации
type Service struct {
	users      map[string]*models.User
	usersMu    sync.RWMutex
	jwtSecret  []byte
	tokenTTL   time.Duration
}

// NewService создаёт новый сервис аутентификации
func NewService(jwtSecret string) *Service {
	return &Service{
		users:     make(map[string]*models.User),
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  24 * time.Hour,
	}
}

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPassword проверяет пароль
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken генерирует JWT токен
func (s *Service) GenerateToken(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken валидирует JWT токен
func (s *Service) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// Register регистрирует нового пользователя
func (s *Service) Register(req models.RegisterRequest) (*models.User, error) {
	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	// Проверяем существует ли пользователь
	if _, exists := s.users[req.Username]; exists {
		return nil, errors.New("username already exists")
	}

	// Хешируем пароль
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Создаём пользователя
	user := &models.User{
		ID:        generateID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		Avatar:    "https://api.dicebear.com/7.x/avataaars/svg?seed=" + req.Username,
		IsOnline:  false,
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
	}

	s.users[user.ID] = user
	s.users[user.Username] = user // Индекс по username для быстрого поиска

	return user, nil
}

// Login аутентифицирует пользователя
func (s *Service) Login(req models.AuthRequest) (*models.User, string, error) {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	// Ищем пользователя по username
	user, exists := s.users[req.Username]
	if !exists {
		return nil, "", errors.New("invalid credentials")
	}

	// Проверяем пароль
	if !CheckPassword(req.Password, user.Password) {
		return nil, "", errors.New("invalid credentials")
	}

	// Генерируем токен
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// GetUserByID получает пользователя по ID
func (s *Service) GetUserByID(id string) (*models.User, error) {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByUsername получает пользователя по username
func (s *Service) GetUserByUsername(username string) (*models.User, error) {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// SetUserOnline устанавливает статус онлайн
func (s *Service) SetUserOnline(userID string, online bool) {
	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	if user, exists := s.users[userID]; exists {
		user.IsOnline = online
		user.LastSeen = time.Now()
	}
}

// GetAllUsers получает всех пользователей
func (s *Service) GetAllUsers() []*models.User {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	users := make([]*models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users
}

// generateID генерирует уникальный ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
