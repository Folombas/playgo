// Package level - генерация и управление уровнями
// Go365 Day 88
package level

import (
	"math/rand"
)

// Platform - платформа
type Platform struct {
	X, Y, Width, Height float64
	Type                string // "ground", "stone", "brick", "metal"
}

// Collectible - коллекционный предмет
type Collectible struct {
	X, Y     float64
	Width    float64
	Height   float64
	Type     string // "coin", "gem", "heart"
	Value    int
	Collected bool
	AnimFrame float64
}

// Level - данные уровня
type Level struct {
	Platforms   []Platform
	Collectibles []Collectible
	Width       float64
	Height      float64
}

// GenerateLevel - генерация уровня
func GenerateLevel(levelNum int, rng *rand.Rand) *Level {
	l := &Level{
		Platforms:    make([]Platform, 0),
		Collectibles: make([]Collectible, 0),
		Width:        3000 + float64(levelNum*500),
		Height:       720,
	}
	
	// Земля (основная платформа)
	l.Platforms = append(l.Platforms, Platform{
		X: 0, Y: 682, Width: l.Width, Height: 50, Type: "ground",
	})
	
	// Генерация платформ
	platformCount := 10 + levelNum*3
	for i := 0; i < platformCount; i++ {
		x := float64(200 + i*180 + rng.Intn(80))
		y := float64(550 - rng.Intn(250))
		width := float64(100 + rng.Intn(100))
		
		platformType := "stone"
		if levelNum > 2 && rng.Float64() < 0.3 {
			platformType = "metal"
		}
		
		l.Platforms = append(l.Platforms, Platform{
			X: x, Y: y, Width: width, Height: 24, Type: platformType,
		})
	}
	
	// Здания (высокие платформы)
	buildingCount := 2 + levelNum
	for i := 0; i < buildingCount; i++ {
		x := float64(500 + i*600)
		l.Platforms = append(l.Platforms, Platform{
			X: x, Y: 562, Width: 180, Height: 120, Type: "brick",
		})
	}
	
	// Генерация коллекционных предметов
	l.generateCollectibles(rng)
	
	return l
}

// generateCollectibles - генерация предметов
func (l *Level) generateCollectibles(rng *rand.Rand) {
	// Монеты
	coinCount := 15
	for i := 0; i < coinCount; i++ {
		x := float64(300 + i*200 + rng.Intn(100))
		y := float64(600 - rng.Intn(200))
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 24, Height: 24,
			Type: "coin", Value: 10,
		})
	}
	
	// Драгоценные камни (редкие)
	gemCount := 3
	for i := 0; i < gemCount; i++ {
		x := float64(400 + i*500)
		y := float64(400 - rng.Intn(100))
		
		gemType := "gem_blue"
		if i == 0 {
			gemType = "gem_red"
		} else if i == 1 {
			gemType = "gem_green"
		}
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 28, Height: 28,
			Type: gemType, Value: 50,
		})
	}
}

// CheckCollection - проверка сбора предмета
func (l *Level) CheckCollection(playerX, playerY, playerWidth, playerHeight float64) int {
	score := 0
	
	for i := range l.Collectibles {
		if !l.Collectibles[i].Collected {
			if checkCollision(
				playerX, playerY, playerWidth, playerHeight,
				l.Collectibles[i].X, l.Collectibles[i].Y, l.Collectibles[i].Width, l.Collectibles[i].Height,
			) {
				l.Collectibles[i].Collected = true
				score += l.Collectibles[i].Value
			}
		}
	}
	
	return score
}

// checkCollision - проверка коллизии
func checkCollision(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}
