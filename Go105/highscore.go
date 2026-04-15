package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
)

const highScoreFile = "highscore.json"

// loadHighScore загружает рекорд из файла
func (g *Game) loadHighScore() {
	path := filepath.Join("save", highScoreFile)
	data, err := os.ReadFile(path)
	if err != nil {
		g.highScore = 0
		return
	}

	var record struct {
		Score int `json:"score"`
	}
	if err := json.Unmarshal(data, &record); err != nil {
		g.highScore = 0
		return
	}
	g.highScore = record.Score
}

// saveHighScore сохраняет рекорд в файл
func (g *Game) saveHighScore() {
	if g.score <= g.highScore {
		return
	}

	g.highScore = g.score

	// Создать папку save
	os.MkdirAll("save", 0755)

	path := filepath.Join("save", highScoreFile)
	data, err := json.Marshal(struct {
		Score int `json:"score"`
	}{Score: g.score})
	if err != nil {
		return
	}

	os.WriteFile(path, data, 0644)
}

// randomInt возвращает случайное число в диапазоне [min, max)
func randomInt(min, max int) int {
	return min + int(rand.Intn(max-min))
}
