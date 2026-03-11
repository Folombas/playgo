package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"n8n-chat/internal/auth"
	"n8n-chat/internal/chat"
	"n8n-chat/internal/models"
	chatws "n8n-chat/internal/websocket"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var (
	hub           *chatws.Hub
	authService   *auth.Service
	chatService   *chat.Service
	upgrader      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	jwtCookieName = "auth_token"
)

func main() {
	// Инициализация сервисов
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "n8n-chat-super-secret-key-change-in-production"
	}

	authService = auth.NewService(jwtSecret)
	chatService = chat.NewService()
	hub = chatws.NewHub()

	// Запускаем хаб
	go hub.Run()

	// Настройка CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// HTTP роуты
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/api/rooms", roomsHandler)
	http.HandleFunc("/api/messages", messagesHandler)
	http.HandleFunc("/api/users", usersHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Логирование
	loggedHandler := loggingMiddleware(c.Handler(http.DefaultServeMux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 N8N Chat Server starting on http://localhost:%s\n", port)
	fmt.Println("✨ Features:")
	fmt.Println("  - Real-time messaging via WebSocket")
	fmt.Println("  - Multiple chat rooms")
	fmt.Println("  - JWT authentication")
	fmt.Println("  - User presence tracking")
	fmt.Println("  - Typing indicators")
	log.Fatal(http.ListenAndServe(":"+port, loggedHandler))
}

// Middleware для логирования
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// Главная страница
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	tmpl.Execute(w, nil)
}

// WebSocket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем аутентификацию
	token, err := r.Cookie(jwtCookieName)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := authService.ValidateToken(token.Value)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Получаем комнату из query params
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = "general"
	}

	// Upgrader с проверкой origin
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	// Создаём клиента
	client := &chatws.Client{
		ID:       claims.UserID,
		UserID:   claims.UserID,
		Username: claims.Username,
		RoomID:   roomID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	// Регистрируем и обрабатываем
	hub.Register <- client
	hub.HandleClient(client)
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, token, err := authService.Login(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     jwtCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// Register handler
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := authService.Register(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	token, err := authService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     jwtCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// Rooms handler
func roomsHandler(w http.ResponseWriter, r *http.Request) {
	rooms := chatService.GetAllRooms()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// Messages handler
func messagesHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = "general"
	}

	messages := chatService.GetMessages(roomID, 50)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// Users handler
func usersHandler(w http.ResponseWriter, r *http.Request) {
	users := authService.GetAllUsers()
	// Убираем пароли
	safeUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		safeUsers[i] = map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"avatar":    user.Avatar,
			"is_online": user.IsOnline,
			"last_seen": user.LastSeen,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeUsers)
}

// Helper для получения username из cookie
func getUsernameFromCookie(r *http.Request) string {
	token, err := r.Cookie(jwtCookieName)
	if err != nil {
		return ""
	}

	claims, err := authService.ValidateToken(token.Value)
	if err != nil {
		return ""
	}

	return claims.Username
}
