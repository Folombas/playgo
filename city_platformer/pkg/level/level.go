// Package level - генерация и управление уровнями
// Go365 Day 88 - Food Platformer
package level

import (
	"math/rand"
)

// Platform - платформа
type Platform struct {
	X, Y, Width, Height float64
	Type                string // "counter", "floor", "shelf", "stove"
}

// Collectible - коллекционный предмет (еда)
type Collectible struct {
	X, Y      float64
	Width     float64
	Height    float64
	Type      string // "fruit", "vegetable", "meat", "dairy", "bakery", "junk"
	TypeInt   int    // числовой тип для entity
	Value     int
	Collected bool
	AnimFrame float64
}

// Level - данные уровня
type Level struct {
	Platforms    []Platform
	Collectibles []Collectible
	Width        float64
	Height       float64
	Name         string
}

// GenerateLevel - генерация уровня (кухня)
func GenerateLevel(levelNum int, rng *rand.Rand) *Level {
	l := &Level{
		Platforms:    make([]Platform, 0),
		Collectibles: make([]Collectible, 0),
		Width:        3000 + float64(levelNum*500),
		Height:       720,
		Name:         getLevelName(levelNum),
	}
	
	// Пол кухни
	l.Platforms = append(l.Platforms, Platform{
		X: 0, Y: 670, Width: l.Width, Height: 50, Type: "floor",
	})
	
	// Кухонные столы/стойки
	counterCount := 8 + levelNum*2
	for i := 0; i < counterCount; i++ {
		x := float64(200 + i*250 + rng.Intn(50))
		y := float64(550 - rng.Intn(150))
		width := float64(120 + rng.Intn(80))
		
		l.Platforms = append(l.Platforms, Platform{
			X: x, Y: y, Width: width, Height: 50, Type: "counter",
		})
	}
	
	// Полки
	shelfCount := 5 + levelNum
	for i := 0; i < shelfCount; i++ {
		x := float64(400 + i*500)
		y := float64(400 - rng.Intn(100))
		
		l.Platforms = append(l.Platforms, Platform{
			X: x, Y: y, Width: 180, Height: 20, Type: "shelf",
		})
	}
	
	// Плиты
	stoveCount := 2 + levelNum/2
	for i := 0; i < stoveCount; i++ {
		x := float64(600 + i*800)
		
		l.Platforms = append(l.Platforms, Platform{
			X: x, Y: 620, Width: 100, Height: 50, Type: "stove",
		})
	}
	
	// Генерация еды
	l.generateFood(rng, levelNum)
	
	return l
}

// getLevelName - название уровня
func getLevelName(levelNum int) string {
	names := []string{
		"Начало Кухни",
		"Холодильник",
		"Склад Продуктов",
		"Горячий Цех",
		"Кондитерская",
		"Гриль-Бар",
		"Суши-Станция",
		"Пекарня",
		"Морозильник",
		"Логово Гнилого Шефа",
	}
	
	if levelNum <= len(names) {
		return names[levelNum-1]
	}
	return "Кухня " + string(rune('0'+levelNum))
}

// generateFood - генерация еды
func (l *Level) generateFood(rng *rand.Rand, levelNum int) {
	// Фрукты
	fruitCount := 8 + levelNum*2
	for i := 0; i < fruitCount; i++ {
		x := float64(300 + i*200 + rng.Intn(100))
		y := float64(600 - rng.Intn(200))
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 28, Height: 28,
			Type: "fruit", TypeInt: 0, Value: 10,
		})
	}
	
	// Овощи
	vegCount := 6 + levelNum
	for i := 0; i < vegCount; i++ {
		x := float64(400 + i*300)
		y := float64(500 - rng.Intn(150))
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 28, Height: 28,
			Type: "vegetable", TypeInt: 1, Value: 12,
		})
	}
	
	// Мясо
	meatCount := 4 + levelNum
	for i := 0; i < meatCount; i++ {
		x := float64(500 + i*400)
		y := float64(450 - rng.Intn(100))
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 28, Height: 28,
			Type: "meat", TypeInt: 2, Value: 15,
		})
	}
	
	// Молочные продукты
	dairyCount := 3 + levelNum
	for i := 0; i < dairyCount; i++ {
		x := float64(350 + i*450)
		y := float64(400 - rng.Intn(120))
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 28, Height: 28,
			Type: "dairy", TypeInt: 3, Value: 13,
		})
	}
	
	// Выпечка
	bakeryCount := 4 + levelNum
	for i := 0; i < bakeryCount; i++ {
		x := float64(450 + i*350)
		y := float64(380 - rng.Intn(100))
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 28, Height: 28,
			Type: "bakery", TypeInt: 4, Value: 18,
		})
	}
	
	// Сладости (бонус)
	sweetCount := 2 + levelNum
	for i := 0; i < sweetCount; i++ {
		x := float64(400 + i*500)
		y := float64(300 - rng.Intn(80))
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 28, Height: 28,
			Type: "sweet", TypeInt: 7, Value: 25,
		})
	}
	
	// Вредная еда (штраф)
	junkCount := 2 + levelNum
	for i := 0; i < junkCount; i++ {
		x := float64(350 + i*400)
		y := float64(550 - rng.Intn(100))
		
		l.Collectibles = append(l.Collectibles, Collectible{
			X: x, Y: y, Width: 28, Height: 28,
			Type: "junk", TypeInt: 5, Value: -10,
		})
	}
}

// CheckCollection - проверка сбора еды
func (l *Level) CheckCollection(playerX, playerY, playerWidth, playerHeight float64) (int, int) {
	score := 0
	healthChange := 0
	
	for i := range l.Collectibles {
		if !l.Collectibles[i].Collected {
			if checkCollision(
				playerX, playerY, playerWidth, playerHeight,
				l.Collectibles[i].X, l.Collectibles[i].Y, l.Collectibles[i].Width, l.Collectibles[i].Height,
			) {
				l.Collectibles[i].Collected = true
				score += l.Collectibles[i].Value
				
				// Лечение за полезную еду
				if l.Collectibles[i].TypeInt == 0 || l.Collectibles[i].TypeInt == 1 {
					healthChange = 5
				} else if l.Collectibles[i].TypeInt == 5 {
					// Штраф за вредную еду
					healthChange = -10
				}
			}
		}
	}
	
	return score, healthChange
}

// checkCollision - проверка коллизии
func checkCollision(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}
