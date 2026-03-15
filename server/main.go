package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// PlayerProgress представляет прогресс игрока
type PlayerProgress struct {
	PlayerID       string         `json:"playerId"`
	TotalCrystals  int            `json:"totalCrystals"`
	CompletedLevels []string      `json:"completedLevels"`
	LevelData      map[string]LevelData `json:"levelData"`
	TotalTime      int64          `json:"totalTime"` // в секундах
	LastSave       time.Time      `json:"lastSave"`
}

// LevelData данные по конкретному уровню
type LevelData struct {
	CrystalsCollected int  `json:"crystalsCollected"`
	Completed         bool `json:"completed"`
	BestTime          int64 `json:"bestTime"`
}

// LeaderboardEntry запись в таблице лидеров
type LeaderboardEntry struct {
	PlayerID      string `json:"playerId"`
	TotalCrystals int    `json:"totalCrystals"`
	CompletedLevels int  `json:"completedLevels"`
}

// Server структура сервера
type Server struct {
	db       *sql.DB
	mu       sync.RWMutex
}

// NewServer создаёт новый сервер
func NewServer(dbPath string) (*Server, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Создаём таблицы
	if err := initDB(db); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	return &Server{db: db}, nil
}

// initDB инициализирует базу данных
func initDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS players (
		id TEXT PRIMARY KEY,
		total_crystals INTEGER DEFAULT 0,
		completed_levels TEXT DEFAULT '[]',
		level_data TEXT DEFAULT '{}',
		total_time INTEGER DEFAULT 0,
		last_save TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS leaderboard (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		player_id TEXT,
		total_crystals INTEGER,
		completed_levels INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(query)
	return err
}

// SaveProgressHandler обрабатывает сохранение прогресса
func (s *Server) SaveProgressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var progress PlayerProgress
	if err := json.NewDecoder(r.Body).Decode(&progress); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Сохраняем в БД
	completedLevelsJSON, _ := json.Marshal(progress.CompletedLevels)
	levelDataJSON, _ := json.Marshal(progress.LevelData)

	query := `
	INSERT OR REPLACE INTO players (id, total_crystals, completed_levels, level_data, total_time, last_save)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		progress.PlayerID,
		progress.TotalCrystals,
		string(completedLevelsJSON),
		string(levelDataJSON),
		progress.TotalTime,
		time.Now(),
	)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обновляем leaderboard
	s.updateLeaderboard(progress)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Progress saved",
	})
}

// GetProgressHandler возвращает прогресс игрока
func (s *Server) GetProgressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playerID := r.URL.Query().Get("playerId")
	if playerID == "" {
		playerID = "default"
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var progress PlayerProgress
	var completedLevelsJSON, levelDataJSON string

	query := `
	SELECT id, total_crystals, completed_levels, level_data, total_time, last_save
	FROM players WHERE id = ?
	`
	err := s.db.QueryRow(query, playerID).Scan(
		&progress.PlayerID,
		&progress.TotalCrystals,
		&completedLevelsJSON,
		&levelDataJSON,
		&progress.TotalTime,
		&progress.LastSave,
	)

	if err == sql.ErrNoRows {
		// Игрок не найден, возвращаем пустой прогресс
		progress = PlayerProgress{
			PlayerID:      playerID,
			LevelData:     make(map[string]LevelData),
			CompletedLevels: []string{},
		}
	} else if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	} else {
		// Парсим JSON обратно
		json.Unmarshal([]byte(completedLevelsJSON), &progress.CompletedLevels)
		json.Unmarshal([]byte(levelDataJSON), &progress.LevelData)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

// GetLeaderboardHandler возвращает топ-10 игроков
func (s *Server) GetLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, total_crystals,
	       json_array_length(completed_levels) as completed_count
	FROM players
	ORDER BY total_crystals DESC
	LIMIT 10
	`

	rows, err := s.db.Query(query)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leaderboard []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.PlayerID, &entry.TotalCrystals, &entry.CompletedLevels); err != nil {
			continue
		}
		leaderboard = append(leaderboard, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

// updateLeaderboard обновляет таблицу лидеров
func (s *Server) updateLeaderboard(progress PlayerProgress) {
	query := `
	INSERT INTO leaderboard (player_id, total_crystals, completed_levels)
	VALUES (?, ?, ?)
	`
	s.db.Exec(query,
		progress.PlayerID,
		progress.TotalCrystals,
		len(progress.CompletedLevels),
	)
}

// Middleware для логирования
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// Middleware для CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Определяем путь к БД
	dbPath := "./game_data.db"
	
	// Создаём сервер
	server, err := NewServer(dbPath)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}
	defer server.db.Close()

	// Создаём mux
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/save", server.SaveProgressHandler)
	mux.HandleFunc("/api/progress", server.GetProgressHandler)
	mux.HandleFunc("/api/leaderboard", server.GetLeaderboardHandler)

	// Раздаём статические файлы из client/
	clientDir := "../client"
	if _, err := os.Stat(clientDir); os.IsNotExist(err) {
		clientDir = "./client"
	}
	
	// Абсолютный путь
	absClientDir, err := filepath.Abs(clientDir)
	if err != nil {
		log.Fatal("Failed to get absolute path:", err)
	}
	
	// Проверяем существование директории
	if _, err := os.Stat(absClientDir); err == nil {
		fs := http.FileServer(http.Dir(absClientDir))
		mux.Handle("/", fs)
		log.Printf("Serving client files from: %s", absClientDir)
	} else {
		log.Printf("Client directory not found: %s", absClientDir)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Purple Lord: Digital Odyssey\n\nServer is running. Client files not found."))
		})
	}

	// Применяем middleware
	handler := loggingMiddleware(corsMiddleware(mux))

	// Запускаем сервер
	port := ":3000"
	log.Printf("🟣 Purple Lord Server starting on http://localhost%s", port)
	log.Printf("📁 Database: %s", dbPath)
	
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal("Server failed:", err)
	}
}
