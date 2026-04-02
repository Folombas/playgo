// Package level - система генерации уровней для Cyber City Runner
// Go365 Day 92 - Киберпанк-платформер
package level

import (
	"math/rand"

	"cyber_city/internal/entity"
)

// TileType - тип тайла
type TileType string

const (
	TileEmpty    TileType = "empty"
	TileGround   TileType = "ground"
	TilePlatform TileType = "platform"
	TileWall     TileType = "wall"
	TileLadder   TileType = "ladder"
	TileDoor     TileType = "door"
	TileHazard   TileType = "hazard"
)

// Platform - платформа
type Platform struct {
	X, Y, Width, Height float64
	Type                TileType
	Solid               bool
	Transform           *entity.Transform
	Color               int // 0=ground, 1=platform, 2=wall
}

// NewPlatform создаёт платформу
func NewPlatform(x, y, w, h float64, tileType TileType, solid bool, color int) *Platform {
	return &Platform{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
		Type:    tileType,
		Solid:   solid,
		Transform: entity.NewTransform(x, y, w, h),
		Color:   color,
	}
}

// LevelData - данные уровня
type LevelData struct {
	Platforms []*Platform
	Enemies   []*EnemySpawn
	Items     []*ItemSpawn
	Terminals []*TerminalSpawn
	Doors     []*DoorSpawn
	Hazards   []*HazardSpawn
	Width     float64
	Height    float64
	Name      string
	ExitX     float64
	ExitY     float64
	Theme     LevelTheme
}

// EnemySpawn - точка спавна врага
type EnemySpawn struct {
	X, Y      float64
	Type      string
	PatrolMin float64
	PatrolMax float64
}

// ItemSpawn - точка спавна предмета
type ItemSpawn struct {
	X, Y     float64
	ItemType string
}

// TerminalSpawn - точка спавна терминала
type TerminalSpawn struct {
	X, Y        float64
	HackLevel   int
	Reward      string // "data", "door_open", "turret_disable"
	TargetID    int    // ID цели (двери, турели)
}

// DoorSpawn - точка спавна двери
type DoorSpawn struct {
	X, Y       float64
	Locked     bool
	KeycardReq bool
	Open       bool
	ID         int
}

// HazardSpawn - точка опасности
type HazardSpawn struct {
	X, Y, Width, Height float64
	Type                string // "spikes", "laser", "electric"
	Damage              int
	Active              bool
}

// LevelTheme - тема уровня
type LevelTheme int

const (
	ThemeSlums LevelTheme = iota
	ThemeIndustrial
	ThemeLowerCity
	ThemeMidCity
	ThemeHub
	ThemeLabs
	ThemeMilitary
	ThemeServer
	ThemeRooftops
	ThemeTower
)

// LevelGenerator - генератор уровней
type LevelGenerator struct {
	rng *rand.Rand
}

// NewLevelGenerator создаёт генератор уровней
func NewLevelGenerator(rng *rand.Rand) *LevelGenerator {
	return &LevelGenerator{
		rng: rng,
	}
}

// GenerateLevel генерирует уровень
func (g *LevelGenerator) GenerateLevel(levelNum int) *LevelData {
	levelWidth := 3000.0 + float64(levelNum)*500
	levelHeight := 800.0

	data := &LevelData{
		Platforms: make([]*Platform, 0),
		Enemies:   make([]*EnemySpawn, 0),
		Items:     make([]*ItemSpawn, 0),
		Terminals: make([]*TerminalSpawn, 0),
		Doors:     make([]*DoorSpawn, 0),
		Hazards:   make([]*HazardSpawn, 0),
		Width:     levelWidth,
		Height:    levelHeight,
		Name:      getLevelName(levelNum),
		Theme:     getLevelTheme(levelNum),
	}

	// Генерация
	g.generateTerrain(data, levelNum)
	g.generatePlatforms(data, levelNum)
	g.generateEnemies(data, levelNum)
	g.generateItems(data, levelNum)
	g.generateTerminals(data, levelNum)
	g.generateDoors(data, levelNum)
	g.generateHazards(data, levelNum)

	// Выход в конце уровня
	data.ExitX = levelWidth - 100
	data.ExitY = levelHeight - 150

	return data
}

// generateTerrain генерирует ландшафт
func (g *LevelGenerator) generateTerrain(data *LevelData, levelNum int) {
	// Пол уровня
	groundY := data.Height - 50

	// Основной пол (с возможными ямами)
	x := 0.0
	segmentWidth := 200.0
	for x < data.Width {
		// Ямы после 3 уровня
		if levelNum >= 3 && x > 300 && x < data.Width-300 && g.rng.Float64() < 0.15 {
			// Пропускаем сегмент (яма)
			x += segmentWidth * 1.5
			continue
		}

		data.Platforms = append(data.Platforms, NewPlatform(
			x, groundY, segmentWidth, 50,
			TileGround, true, 0,
		))
		x += segmentWidth
	}

	// Стартовая платформа
	data.Platforms = append(data.Platforms, NewPlatform(
		0, groundY-100, 200, 20,
		TilePlatform, true, 1,
	))

	// Левая и правая стены
	data.Platforms = append(data.Platforms,
		NewPlatform(-50, 0, 50, data.Height, TileWall, true, 2),
		NewPlatform(data.Width, 0, 50, data.Height, TileWall, true, 2),
	)
}

// generatePlatforms генерирует платформы
func (g *LevelGenerator) generatePlatforms(data *LevelData, levelNum int) {
	platformCount := 8 + levelNum*2

	for i := 0; i < platformCount; i++ {
		x := 300.0 + g.rng.Float64()*(data.Width-600)
		y := data.Height - 150.0 - g.rng.Float64()*(data.Height-300)
		width := 80.0 + g.rng.Float64()*120

		// Некоторые платформы движутся (после 5 уровня)
		if levelNum >= 5 && g.rng.Float64() < 0.2 {
			// Движущаяся платформа (помечаем цветом)
			data.Platforms = append(data.Platforms, NewPlatform(x, y, width, 20, TilePlatform, true, 3))
		} else {
			data.Platforms = append(data.Platforms, NewPlatform(x, y, width, 20, TilePlatform, true, 1))
		}
	}

	// Стены для стена-рана
	wallCount := 3 + levelNum
	for i := 0; i < wallCount; i++ {
		x := 400.0 + g.rng.Float64()*(data.Width-700)
		y := data.Height - 400.0 - g.rng.Float64()*200
		height := 150.0 + g.rng.Float64()*100

		data.Platforms = append(data.Platforms, NewPlatform(x, y, 40, height, TileWall, true, 2))
	}
}

// generateEnemies генерирует врагов
func (g *LevelGenerator) generateEnemies(data *LevelData, levelNum int) {
	// Доступные типы врагов по уровням
	enemyTypes := []string{"soldier"}
	if levelNum >= 2 {
		enemyTypes = append(enemyTypes, "drone")
	}
	if levelNum >= 3 {
		enemyTypes = append(enemyTypes, "turret")
	}
	if levelNum >= 5 {
		enemyTypes = append(enemyTypes, "robot")
	}
	if levelNum >= 7 {
		enemyTypes = append(enemyTypes, "elite")
	}

	enemyCount := 4 + levelNum*2
	for i := 0; i < enemyCount; i++ {
		x := 400.0 + g.rng.Float64()*(data.Width-700)
		y := data.Height - 100.0

		// Размещаем на платформах
		for _, p := range data.Platforms {
			if p.Type == TilePlatform && p.Y < y-50 {
				if x >= p.X && x <= p.X+p.Width {
					y = p.Y - 48
					break
				}
			}
		}

		enemyType := enemyTypes[g.rng.Intn(len(enemyTypes))]
		patrolMin := x - 100
		patrolMax := x + 100

		data.Enemies = append(data.Enemies, &EnemySpawn{
			X:         x,
			Y:         y,
			Type:      enemyType,
			PatrolMin: patrolMin,
			PatrolMax: patrolMax,
		})
	}
}

// generateItems генерирует предметы
func (g *LevelGenerator) generateItems(data *LevelData, levelNum int) {
	itemTypes := []struct {
		Type   string
		Weight int
	}{
		{entity.ItemEnergy, 40},
		{entity.ItemStimpack, 30},
		{entity.ItemArmor, 25},
		{entity.ItemData, 50},
		{entity.ItemGrenade, 15},
		{entity.ItemEMP, 10},
		{entity.ItemKeycard, 5},
	}

	itemCount := 6 + levelNum*2
	for i := 0; i < itemCount; i++ {
		x := 300.0 + g.rng.Float64()*(data.Width-600)
		y := data.Height - 150.0 - g.rng.Float64()*(data.Height-300)

		// Выбираем предмет по весу
		itemType := g.selectItem(itemTypes)

		data.Items = append(data.Items, &ItemSpawn{
			X:        x,
			Y:        y,
			ItemType: itemType,
		})
	}
}

// generateTerminals генерирует терминалы
func (g *LevelGenerator) generateTerminals(data *LevelData, levelNum int) {
	terminalCount := 2 + levelNum/2
	for i := 0; i < terminalCount; i++ {
		x := 500.0 + g.rng.Float64()*(data.Width-800)
		y := data.Height - 120

		data.Terminals = append(data.Terminals, &TerminalSpawn{
			X:       x,
			Y:       y,
			HackLevel: 1 + levelNum/3,
			Reward:  "data",
		})
	}
}

// generateDoors генерирует двери
func (g *LevelGenerator) generateDoors(data *LevelData, levelNum int) {
	// Двери-препятствия
	doorCount := 1 + levelNum/3
	for i := 0; i < doorCount; i++ {
		x := 600.0 + float64(i)*800.0
		y := data.Height - 150

		data.Doors = append(data.Doors, &DoorSpawn{
			X:          x,
			Y:          y,
			Locked:     i%2 == 0,
			KeycardReq: i%2 == 0,
			Open:       false,
			ID:         i,
		})

		// Стена рядом с дверью
		if i%2 == 0 {
			data.Platforms = append(data.Platforms, NewPlatform(x-50, y-200, 50, 200, TileWall, true, 2))
			data.Platforms = append(data.Platforms, NewPlatform(x+50, y-200, 50, 200, TileWall, true, 2))
		}
	}
}

// generateHazards генерирует опасности
func (g *LevelGenerator) generateHazards(data *LevelData, levelNum int) {
	if levelNum < 2 {
		return
	}

	hazardCount := 2 + levelNum/2
	for i := 0; i < hazardCount; i++ {
		x := 400.0 + g.rng.Float64()*(data.Width-700)
		y := data.Height - 50

		hazardType := "spikes"
		if levelNum >= 4 {
			types := []string{"spikes", "laser", "electric"}
			hazardType = types[g.rng.Intn(len(types))]
		}

		data.Hazards = append(data.Hazards, &HazardSpawn{
			X:      x,
			Y:      y,
			Width:  60,
			Height: 20,
			Type:   hazardType,
			Damage: 20 + levelNum*5,
			Active: true,
		})
	}
}

// selectItem выбирает предмет по весу
func (g *LevelGenerator) selectItem(items []struct {
	Type   string
	Weight int
}) string {
	total := 0
	for _, it := range items {
		total += it.Weight
	}

	roll := g.rng.Intn(total)
	current := 0
	for _, it := range items {
		current += it.Weight
		if roll < current {
			return it.Type
		}
	}
	return items[0].Type
}

// getLevelName возвращает название уровня
func getLevelName(levelNum int) string {
	names := []string{
		"Трущобы",
		"Промзона",
		"Нижний город",
		"Средний уровень",
		"Транспортный хаб",
		"Лаборатории",
		"Военная зона",
		"Серверная",
		"Крыши",
		"Zaibatsu Tower",
	}

	if levelNum <= len(names) {
		return names[levelNum-1]
	}
	return "Неизвестный сектор"
}

// getLevelTheme возвращает тему уровня
func getLevelTheme(levelNum int) LevelTheme {
	if levelNum <= 10 {
		return LevelTheme(levelNum - 1)
	}
	return ThemeTower
}
