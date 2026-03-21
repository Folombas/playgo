package game

import (
	"encoding/json"
	"log"
	"os"
)

// SaveData - структура для сохранения
type SaveData struct {
	PlayerX         float64
	PlayerY         float64
	PlayerVY        float64
	Score           int
	Coins           int
	Lives           int
	MaxHealth       int
	CurrentHealth   int
	Stats           PlayerStats
	CollectedCoins  []bool
	DefeatedEnemies []bool
	OpenedChests    []bool
	BlocksMined     int
	EnemiesDefeated int
}

// SaveGame сохраняет прогресс
func SaveGame(player *Player, coins []Coin, enemies []Enemy, chests interface{}, blocksMined, enemiesDefeated int) bool {
	collectedCoins := make([]bool, len(coins))
	for i, coin := range coins {
		collectedCoins[i] = coin.collected
	}

	defeatedEnemies := make([]bool, len(enemies))
	for i, enemy := range enemies {
		defeatedEnemies[i] = !enemy.alive
	}

	// Simplified chest handling
	openedChests := make([]bool, 0)

	saveData := SaveData{
		PlayerX:         player.x,
		PlayerY:         player.y,
		PlayerVY:        player.vy,
		Score:           player.score,
		Coins:           player.coins,
		Lives:           player.lives,
		MaxHealth:       player.maxHealth,
		CurrentHealth:   player.currentHealth,
		Stats:           player.stats,
		CollectedCoins:  collectedCoins,
		DefeatedEnemies: defeatedEnemies,
		OpenedChests:    openedChests,
		BlocksMined:     blocksMined,
		EnemiesDefeated: enemiesDefeated,
	}

	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		log.Printf("Save error: %v", err)
		return false
	}

	err = os.WriteFile("savegame.json", data, 0644)
	if err != nil {
		log.Printf("Save write error: %v", err)
		return false
	}

	return true
}

// LoadGame загружает сохранение
func LoadGame(player *Player, coins []Coin, enemies []Enemy) (*SaveData, bool) {
	data, err := os.ReadFile("savegame.json")
	if err != nil {
		return nil, false
	}

	var saveData SaveData
	err = json.Unmarshal(data, &saveData)
	if err != nil {
		log.Printf("Load error: %v", err)
		return nil, false
	}

	player.x = saveData.PlayerX
	player.y = saveData.PlayerY
	player.vy = saveData.PlayerVY
	player.score = saveData.Score
	player.coins = saveData.Coins
	player.lives = saveData.Lives
	player.maxHealth = saveData.MaxHealth
	player.currentHealth = saveData.CurrentHealth
	player.stats = saveData.Stats

	for i, collected := range saveData.CollectedCoins {
		if i < len(coins) {
			coins[i].collected = collected
		}
	}

	for i, defeated := range saveData.DefeatedEnemies {
		if i < len(enemies) {
			enemies[i].alive = !defeated
		}
	}

	return &saveData, true
}
