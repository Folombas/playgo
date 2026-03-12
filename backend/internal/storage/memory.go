package storage

import (
	"sync"
	"time"

	"github.com/playgo/backend/internal/model"
)

// Storage interface for data persistence
type Storage interface {
	SavePlayer(player *model.Player) error
	GetPlayer(id string) (*model.Player, error)
	GetLeaderboard(limit int) ([]model.LeaderboardEntry, error)
	SaveUpgrade(upgrade *model.PlayerUpgrade) error
	GetUpgrades(playerID string) ([]model.PlayerUpgrade, error)
}

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
	mu       sync.RWMutex
	players  map[string]*model.Player
	upgrades map[string][]*model.PlayerUpgrade
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		players:  make(map[string]*model.Player),
		upgrades: make(map[string][]*model.PlayerUpgrade),
	}
}

// SavePlayer saves a player to storage
func (s *MemoryStorage) SavePlayer(player *model.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	player.UpdatedAt = time.Now().Unix()
	if player.CreatedAt == 0 {
		player.CreatedAt = player.UpdatedAt
	}
	s.players[player.ID] = player
	return nil
}

// GetPlayer retrieves a player by ID
func (s *MemoryStorage) GetPlayer(id string) (*model.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	player, exists := s.players[id]
	if !exists {
		return nil, nil
	}
	// Return a copy
	playerCopy := *player
	return &playerCopy, nil
}

// GetLeaderboard returns top players by score
func (s *MemoryStorage) GetLeaderboard(limit int) ([]model.LeaderboardEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]model.LeaderboardEntry, 0, len(s.players))
	for _, player := range s.players {
		entries = append(entries, model.LeaderboardEntry{
			Username: player.Username,
			Score:    player.Score,
			Level:    player.Level,
		})
	}

	// Sort by score (descending)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Score > entries[i].Score {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Add ranks and limit
	result := make([]model.LeaderboardEntry, 0, limit)
	for i := 0; i < len(entries) && i < limit; i++ {
		entry := entries[i]
		entry.Rank = i + 1
		result = append(result, entry)
	}

	return result, nil
}

// SaveUpgrade saves a player upgrade
func (s *MemoryStorage) SaveUpgrade(upgrade *model.PlayerUpgrade) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	upgrades := s.upgrades[upgrade.PlayerID]
	found := false
	for i, u := range upgrades {
		if u.UpgradeID == upgrade.UpgradeID {
			upgrades[i].Count = upgrade.Count
			found = true
			break
		}
	}

	if !found {
		upgrades = append(upgrades, upgrade)
	}
	s.upgrades[upgrade.PlayerID] = upgrades
	return nil
}

// GetUpgrades retrieves all upgrades for a player
func (s *MemoryStorage) GetUpgrades(playerID string) ([]model.PlayerUpgrade, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	upgrades := s.upgrades[playerID]
	result := make([]model.PlayerUpgrade, len(upgrades))
	for i, u := range upgrades {
		result[i] = *u
	}
	return result, nil
}
