package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// LeaderboardEntry представляет запись в таблице рекордов
type LeaderboardEntry struct {
	PlayerName string    `json:"player_name"`
	Score      int       `json:"score"`
	Level      int       `json:"level"`
	Moves      int       `json:"moves"`
	Date       time.Time `json:"date"`
}

// Leaderboard таблица рекордов
type Leaderboard struct {
	entries []LeaderboardEntry
	maxSize int
	filePath string
}

// NewLeaderboard создаёт новую таблицу рекордов
func NewLeaderboard() *Leaderboard {
	lb := &Leaderboard{
		entries: make([]LeaderboardEntry, 0),
		maxSize: 10, // Топ-10
		filePath: getLeaderboardFilePath(),
	}
	
	lb.Load()
	
	return lb
}

// AddEntry добавляет запись в таблицу рекордов
func (lb *Leaderboard) AddEntry(entry LeaderboardEntry) {
	lb.entries = append(lb.entries, entry)
	
	// Сортируем по очкам (по убыванию)
	sort.Slice(lb.entries, func(i, j int) bool {
		return lb.entries[i].Score > lb.entries[j].Score
	})
	
	// Оставляем только топ-10
	if len(lb.entries) > lb.maxSize {
		lb.entries = lb.entries[:lb.maxSize]
	}
	
	lb.Save()
}

// GetEntries возвращает все записи в таблице
func (lb *Leaderboard) GetEntries() []LeaderboardEntry {
	return lb.entries
}

// GetRank возвращает ранг игрока (по очкам)
func (lb *Leaderboard) GetRank(score int) int {
	for i, entry := range lb.entries {
		if entry.Score <= score {
			return i + 1
		}
	}
	return len(lb.entries) + 1
}

// Load загружает таблицу рекордов
func (lb *Leaderboard) Load() error {
	data, err := os.ReadFile(lb.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read leaderboard: %w", err)
	}
	
	err = json.Unmarshal(data, &lb.entries)
	if err != nil {
		fmt.Printf("Corrupted leaderboard file, starting fresh: %v\n", err)
		lb.entries = make([]LeaderboardEntry, 0)
		return nil
	}
	
	return nil
}

// Save сохраняет таблицу рекордов
func (lb *Leaderboard) Save() error {
	data, err := json.MarshalIndent(lb.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal leaderboard: %w", err)
	}
	
	dir := filepath.Dir(lb.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	err = os.WriteFile(lb.filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write leaderboard: %w", err)
	}
	
	return nil
}

// getLeaderboardFilePath возвращает путь к файлу таблицы рекордов
func getLeaderboardFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "leaderboard.json"
	}
	
	saveDir := filepath.Join(homeDir, ".match3_game")
	return filepath.Join(saveDir, "leaderboard.json")
}
