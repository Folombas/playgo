// Package level - генерация и управление уровнями
// Go365 Day 91 - City Platformer
package level

import (
	"math/rand"
)

// Platform - платформа
type Platform struct {
	X, Y, Width, Height float64
	Type                string // ground, building, rubble, metal
}

// Level - данные уровня
type Level struct {
	Platforms []*Platform
	Items     []*LevelItem
	Enemies   []*LevelEnemy
	Width     float64
	Height    float64
	Name      string
	ExitX     float64
}

// LevelItem - предмет на уровне
type LevelItem struct {
	X, Y      float64
	Width     float64
	Height    float64
	Type      string // medkit, ammo, food, parts
	Value     int
	Collected bool
}

// LevelEnemy - враг на уровне
type LevelEnemy struct {
	X, Y   float64
	Type   string // mutant, robot, zombie
	Active bool
}

// GenerateLevel - генерация уровня
func GenerateLevel(levelNum int, rng *rand.Rand) *Level {
	l := &Level{
		Platforms: make([]*Platform, 0),
		Items:     make([]*LevelItem, 0),
		Enemies:   make([]*LevelEnemy, 0),
		Width:     3000 + float64(levelNum*500),
		Height:    720,
		Name:      getLevelName(levelNum),
	}

	// Генерация пола
	l.generateGround(rng)

	// Генерация платформ
	l.generatePlatforms(rng, levelNum)

	// Генерация предметов
	l.generateItems(rng, levelNum)

	// Генерация врагов
	l.generateEnemies(rng, levelNum)

	// Точка выхода
	l.ExitX = l.Width - 200

	return l
}

// generateGround - генерация земли
func (l *Level) generateGround(rng *rand.Rand) {
	// Основная земля
	l.Platforms = append(l.Platforms, &Platform{
		X: 0, Y: 670, Width: l.Width, Height: 50, Type: "ground",
	})

	// Ямы (разрывы в земле)
	pitCount := 2 + rng.Intn(5)
	for i := 0; i < pitCount; i++ {
		pitX := 500 + rng.Intn(int(l.Width)-1000)
		pitWidth := 80 + rng.Intn(100)

		// Добавляем платформы до и после ямы
		if pitX > 100 {
			l.Platforms = append(l.Platforms, &Platform{
				X: float64(pitX - pitWidth), Y: 670, Width: float64(pitWidth), Height: 50, Type: "ground",
			})
		}
	}
}

// generatePlatforms - генерация платформ
func (l *Level) generatePlatforms(rng *rand.Rand, levelNum int) {
	// Здания (высокие платформы)
	buildingCount := 5 + levelNum*2
	for i := 0; i < buildingCount; i++ {
		x := float64(200 + i*250 + rng.Intn(100))
		y := float64(500 - rng.Intn(150))
		width := float64(100 + rng.Intn(100))
		height := float64(50 + rng.Intn(50))

		l.Platforms = append(l.Platforms, &Platform{
			X: x, Y: y, Width: width, Height: height, Type: "building",
		})
	}

	// Обломки (низкие платформы)
	rubbleCount := 8 + levelNum*3
	for i := 0; i < rubbleCount; i++ {
		x := float64(300 + i*200 + rng.Intn(80))
		y := float64(600 - rng.Intn(100))
		width := float64(60 + rng.Intn(60))

		l.Platforms = append(l.Platforms, &Platform{
			X: x, Y: y, Width: width, Height: 32, Type: "rubble",
		})
	}

	// Металлические конструкции
	metalCount := 3 + levelNum
	for i := 0; i < metalCount; i++ {
		x := float64(400 + i*400)
		y := float64(350 - rng.Intn(100))
		width := float64(150 + rng.Intn(50))

		l.Platforms = append(l.Platforms, &Platform{
			X: x, Y: y, Width: width, Height: 20, Type: "metal",
		})
	}
}

// generateItems - генерация предметов
func (l *Level) generateItems(rng *rand.Rand, levelNum int) {
	// Аптечки
	medkitCount := 2 + levelNum/2
	for i := 0; i < medkitCount; i++ {
		x := float64(300 + i*400 + rng.Intn(100))
		y := float64(600 - rng.Intn(200))

		l.Items = append(l.Items, &LevelItem{
			X: x, Y: y, Type: "medkit", Value: 25,
		})
	}

	// Патроны
	ammoCount := 4 + levelNum
	for i := 0; i < ammoCount; i++ {
		x := float64(250 + i*300 + rng.Intn(80))
		y := float64(550 - rng.Intn(150))

		l.Items = append(l.Items, &LevelItem{
			X: x, Y: y, Type: "ammo", Value: 20,
		})
	}

	// Еда
	foodCount := 3 + levelNum
	for i := 0; i < foodCount; i++ {
		x := float64(350 + i*350 + rng.Intn(60))
		y := float64(500 - rng.Intn(120))

		l.Items = append(l.Items, &LevelItem{
			X: x, Y: y, Type: "food", Value: 10,
		})
	}

	// Детали (очки)
	partsCount := 5 + levelNum*2
	for i := 0; i < partsCount; i++ {
		x := float64(200 + i*250 + rng.Intn(50))
		y := float64(450 - rng.Intn(100))

		l.Items = append(l.Items, &LevelItem{
			X: x, Y: y, Type: "parts", Value: 50,
		})
	}
}

// generateEnemies - генерация врагов
func (l *Level) generateEnemies(rng *rand.Rand, levelNum int) {
	// Мутанты
	mutantCount := 3 + levelNum
	for i := 0; i < mutantCount; i++ {
		x := float64(400 + i*350 + rng.Intn(100))
		y := float64(630) // На земле

		l.Enemies = append(l.Enemies, &LevelEnemy{
			X: x, Y: y, Type: "mutant", Active: true,
		})
	}

	// Роботы (на платформах)
	robotCount := 2 + levelNum/2
	for i := 0; i < robotCount; i++ {
		x := float64(500 + i*450 + rng.Intn(80))
		y := float64(550 - rng.Intn(150))

		l.Enemies = append(l.Enemies, &LevelEnemy{
			X: x, Y: y, Type: "robot", Active: true,
		})
	}

	// Зомби
	zombieCount := 4 + levelNum*2
	for i := 0; i < zombieCount; i++ {
		x := float64(350 + i*300 + rng.Intn(60))
		y := float64(630)

		l.Enemies = append(l.Enemies, &LevelEnemy{
			X: x, Y: y, Type: "zombie", Active: true,
		})
	}
}

// getLevelName - название уровня
func getLevelName(levelNum int) string {
	names := []string{
		"Разрушенная улица",
		"Заброшенный завод",
		"Тоннель метро",
		"Крыша небоскрёба",
		"Мёртвый город",
		"Промзона",
		"Трущобы",
		"Центр города",
		"Военная база",
		"Точка эвакуации",
	}

	if levelNum <= len(names) {
		return names[levelNum-1]
	}
	return "Зона " + string(rune('0'+levelNum%10))
}

// CheckPlatformCollision - проверка коллизии с платформами
func (l *Level) CheckPlatformCollision(x, y, width, height float64) (*Platform, bool) {
	for _, p := range l.Platforms {
		if checkCollision(x, y, width, height, p.X, p.Y, p.Width, p.Height) {
			return p, true
		}
	}
	return nil, false
}

// CheckItemCollection - проверка сбора предметов
func (l *Level) CheckItemCollection(x, y, width, height float64) *LevelItem {
	for _, item := range l.Items {
		if !item.Collected {
			if checkCollision(x, y, width, height, item.X, item.Y, item.Width, item.Height) {
				item.Collected = true
				return item
			}
		}
	}
	return nil
}

// CheckExitReach - проверка достижения выхода
func (l *Level) CheckExitReach(x, y, width, height float64) bool {
	return x+width > l.ExitX
}

// checkCollision - проверка коллизии AABB
func checkCollision(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}
