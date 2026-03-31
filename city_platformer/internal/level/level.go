// Package level - система бесконечных уровней
// Go365 Day 91 - Sunny Adventure: Бесконечный мир
package level

import (
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"sunny_adventure/internal/entity"
	"sunny_adventure/internal/sprite"
)

// TileType - тип тайла
type TileType string

const (
	TileEmpty    TileType = "empty"
	TileGround   TileType = "ground"
	TileGrass    TileType = "grass"
	TileBrick    TileType = "brick"
	TileBox      TileType = "box"
	TileLadder   TileType = "ladder"
	TileCastle   TileType = "castle"
	TileCandy    TileType = "candy"
	TileIce      TileType = "ice"
)

// Platform - платформа
type Platform struct {
	X, Y, Width, Height float64
	Type                TileType
	Solid               bool
	Transform           entity.Transform
}

// LevelEnemy - данные врага на уровне
type LevelEnemy struct {
	X, Y   float64
	Type   string
	Active bool
}

// LevelFriend - данные друга на уровне
type LevelFriend struct {
	X, Y  float64
	Type  string
}

// LevelCloud - данные облачка на уровне
type LevelCloud struct {
	X, Y float64
	Num  int // 1, 2, или 3
}

// LevelItem - данные предмета на уровне
type LevelItem struct {
	X, Y      float64
	Type      string
	Value     int
	Collected bool
}

// LevelDecor - данные декорации на уровне
type LevelDecor struct {
	X, Y float64
	Type string
}

// LevelData - данные чанка уровня
type LevelData struct {
	Platforms []*Platform
	Enemies   []*LevelEnemy
	Friends   []*LevelFriend
	Clouds    []*LevelCloud
	Items     []*LevelItem
	Decors    []*LevelDecor
	Width     int
	Height    int
	TileSize  int
	Name      string
	ExitX     float64
	ExitY     float64
	Background *ebiten.Image
	Theme      LevelTheme
	ChunkNum   int
}

// LevelTheme - тема чанка
type LevelTheme int

const (
	ThemeGrass LevelTheme = iota
	ThemeMushroom
	ThemeCastle
	ThemeCandy
	ThemeIce
)

// LevelGenerator - генератор бесконечных уровней
type LevelGenerator struct {
	rng         *rand.Rand
	TileSize    int
	spriteSheet *sprite.SpriteSheet
	seed        int64
	ChunkWidth  int
	chunkCache  map[int]*LevelData
	// Пулы для оптимизации
	platformPool []*Platform
	enemyPool    []*LevelEnemy
	friendPool   []*LevelFriend
	cloudPool    []*LevelCloud
	itemPool     []*LevelItem
	decorPool    []*LevelDecor
}

// NewLevelGenerator создаёт генератор бесконечных уровней
func NewLevelGenerator(rng *rand.Rand, tileSize int, ss *sprite.SpriteSheet) *LevelGenerator {
	return &LevelGenerator{
		rng:         rng,
		TileSize:    tileSize,
		spriteSheet: ss,
		seed:        time.Now().UnixNano(),
		ChunkWidth:  50,
		chunkCache:  make(map[int]*LevelData),
		// Предварительно выделяем пулы
		platformPool: make([]*Platform, 0, 100),
		enemyPool:    make([]*LevelEnemy, 0, 50),
		friendPool:   make([]*LevelFriend, 0, 30),
		cloudPool:    make([]*LevelCloud, 0, 30),
		itemPool:     make([]*LevelItem, 0, 50),
		decorPool:    make([]*LevelDecor, 0, 50),
	}
}

// GenerateChunk генерирует чанок мира (оптимизировано)
func (g *LevelGenerator) GenerateChunk(chunkNum int) *LevelData {
	// Проверяем кэш
	if chunk, ok := g.chunkCache[chunkNum]; ok {
		return chunk
	}

	// Детерминированный RNG для чанка
	chunkRng := rand.New(rand.NewSource(g.seed + int64(chunkNum)))

	levelWidth := g.ChunkWidth
	levelHeight := 10

	data := &LevelData{
		Platforms: make([]*Platform, 0, 30),
		Enemies:   make([]*LevelEnemy, 0, 20),
		Friends:   make([]*LevelFriend, 0, 15),
		Clouds:    make([]*LevelCloud, 0, 15),
		Items:     make([]*LevelItem, 0, 30),
		Decors:    make([]*LevelDecor, 0, 30),
		Width:     levelWidth,
		Height:    levelHeight,
		TileSize:  g.TileSize,
		Name:      getChunkName(chunkNum),
		Theme:     getChunkTheme(chunkNum),
		ChunkNum:  chunkNum,
	}

	// Упрощённая генерация для производительности
	g.generateTerrainFast(data, chunkNum, chunkRng)
	g.generatePlatformsFast(data, chunkNum, chunkRng)
	g.generateEnemiesFast(data, chunkNum, chunkRng)
	g.generateFriendsFast(data, chunkNum, chunkRng)
	g.generateCloudsFast(data, chunkNum, chunkRng)
	g.generateItemsFast(data, chunkNum, chunkRng)
	g.generateDecorsFast(data, chunkNum, chunkRng)

	// Выход в конце чанка (кроме первого)
	if chunkNum > 0 {
		data.ExitX = float64(levelWidth-3) * float64(g.TileSize)
		data.ExitY = float64(levelHeight-3) * float64(g.TileSize)
	}

	data.Background = g.spriteSheet.GetBackground()

	// Кэш - храним только 3 последних чанка
	g.chunkCache[chunkNum] = data
	if len(g.chunkCache) > 3 {
		delete(g.chunkCache, chunkNum-3)
	}

	return data
}

// generateTerrainFast - быстрая генерация ландшафта
func (g *LevelGenerator) generateTerrainFast(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	groundY := data.Height - 2

	for x := 0; x < data.Width; x++ {
		// Простая проверка на яму
		if x > 5 && x < data.Width-5 && rng.Float64() < 0.1 {
			continue
		}

		data.Platforms = append(data.Platforms, &Platform{
			X:      float64(x) * ts,
			Y:      float64(groundY) * ts,
			Width:  ts,
			Height: ts * 2,
			Type:    TileGround,
			Solid:   true,
			Transform: entity.Transform{X: float64(x) * ts, Y: float64(groundY) * ts, Width: ts, Height: ts * 2},
		})
	}
}

// generatePlatformsFast - быстрая генерация платформ
func (g *LevelGenerator) generatePlatformsFast(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	count := 5 + chunkNum

	for i := 0; i < count; i++ {
		x := float64(rng.Intn(data.Width-10) + 5) * ts
		y := float64(rng.Intn(data.Height-6) + 3) * ts
		width := float64(3 + rng.Intn(3)) * ts

		data.Platforms = append(data.Platforms, &Platform{
			X:      x,
			Y:      y,
			Width:  width,
			Height: ts,
			Type:    TileGrass,
			Solid:   true,
			Transform: entity.Transform{X: x, Y: y, Width: width, Height: ts},
		})
	}
}

// generateEnemiesFast - быстрая генерация врагов
func (g *LevelGenerator) generateEnemiesFast(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	count := 3 + chunkNum

	for i := 0; i < count; i++ {
		x := float64(rng.Intn(data.Width-10) + 5) * ts
		y := float64(data.Height-3) * ts

		data.Enemies = append(data.Enemies, &LevelEnemy{
			X:      x,
			Y:      y,
			Type:   "snake",
			Active: true,
		})
	}
}

// generateFriendsFast - быстрая генерация друзей
func (g *LevelGenerator) generateFriendsFast(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	friendTypes := []string{"bee", "ladybug", "frog"}
	count := 4 + chunkNum/2

	for i := 0; i < count; i++ {
		x := float64(rng.Intn(data.Width-10) + 5) * ts
		y := float64(data.Height-3) * ts

		data.Friends = append(data.Friends, &LevelFriend{
			X:    x,
			Y:    y,
			Type: friendTypes[rng.Intn(len(friendTypes))],
		})
	}
}

// generateCloudsFast - быстрая генерация облачков
func (g *LevelGenerator) generateCloudsFast(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	count := 3 + chunkNum/2

	for i := 0; i < count; i++ {
		x := float64(rng.Intn(data.Width-10) + 5) * ts
		y := float64(rng.Intn(data.Height-6) + 2) * ts

		data.Clouds = append(data.Clouds, &LevelCloud{
			X:   x,
			Y:   y,
			Num: (i % 3) + 1,
		})
	}
}

// generateItemsFast - быстрая генерация предметов
func (g *LevelGenerator) generateItemsFast(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	itemTypes := []string{"coinGold", "coinSilver", "gemRed", "star"}
	count := 8 + chunkNum

	for i := 0; i < count; i++ {
		x := float64(rng.Intn(data.Width-10) + 5) * ts
		y := float64(rng.Intn(data.Height-6) + 3) * ts

		data.Items = append(data.Items, &LevelItem{
			X:         x,
			Y:         y,
			Type:      itemTypes[rng.Intn(len(itemTypes))],
			Value:     10,
			Collected: false,
		})
	}
}

// generateDecorsFast - быстрая генерация декораций
func (g *LevelGenerator) generateDecorsFast(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	decorTypes := []string{"hill_large", "mushroomRed", "bush", "rock", "flagBlue"}
	count := 5 + chunkNum

	for i := 0; i < count; i++ {
		x := float64(rng.Intn(data.Width-5)) * ts
		y := float64(data.Height-2) * ts

		data.Decors = append(data.Decors, &LevelDecor{
			X:    x,
			Y:    y,
			Type: decorTypes[rng.Intn(len(decorTypes))],
		})
	}
}

// GetChunkAtX возвращает чанок для позиции X
func (g *LevelGenerator) GetChunkAtX(x float64) *LevelData {
	chunkNum := int(x) / (g.ChunkWidth * g.TileSize)
	if x < 0 {
		chunkNum = -1
	}
	return g.GenerateChunk(chunkNum)
}

// generateTerrain генерирует ландшафт
func (g *LevelGenerator) generateTerrain(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	groundY := data.Height - 2

	for x := 0; x < data.Width; x++ {
		if g.isPit(x, data.Width, chunkNum, rng) {
			continue
		}

		tileType := TileGround
		switch data.Theme {
		case ThemeCastle:
			tileType = TileCastle
		case ThemeCandy:
			tileType = TileCandy
		case ThemeIce:
			tileType = TileIce
		}

		data.Platforms = append(data.Platforms, &Platform{
			X:      float64(x) * ts,
			Y:      float64(groundY) * ts,
			Width:  ts,
			Height: ts * 2,
			Type:    tileType,
			Solid:   true,
			Transform: entity.Transform{X: float64(x) * ts, Y: float64(groundY) * ts, Width: ts, Height: ts * 2},
		})
	}

	// Границы чанка
	data.Platforms = append(data.Platforms,
		&Platform{X: 0, Y: 0, Width: 5, Height: float64(data.Height) * ts, Type: TileBrick, Solid: true, Transform: entity.Transform{X: 0, Y: 0, Width: 5, Height: float64(data.Height) * ts}},
	)
}

// isPit - проверка на яму
func (g *LevelGenerator) isPit(x, levelWidth, chunkNum int, rng *rand.Rand) bool {
	if x < 5 || x > levelWidth-5 {
		return false
	}

	pitChance := 0.08 + float64(chunkNum)*0.01
	if pitChance > 0.2 {
		pitChance = 0.2
	}

	return rng.Float64() < pitChance
}

// generatePlatforms генерирует платформы
func (g *LevelGenerator) generatePlatforms(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)
	platformCount := 5 + chunkNum*2

	for i := 0; i < platformCount; i++ {
		x := rng.Intn(data.Width-10) + 5
		y := rng.Intn(data.Height-6) + 3
		width := 2 + rng.Intn(4)

		tileType := TileGrass
		switch data.Theme {
		case ThemeMushroom:
			tileType = TileBox
		case ThemeCastle:
			tileType = TileBrick
		case ThemeCandy:
			tileType = TileCandy
		case ThemeIce:
			tileType = TileIce
		}

		data.Platforms = append(data.Platforms, &Platform{
			X:      float64(x) * ts,
			Y:      float64(y) * ts,
			Width:  float64(width) * ts,
			Height: ts,
			Type:    tileType,
			Solid:   true,
			Transform: entity.Transform{X: float64(x) * ts, Y: float64(y) * ts, Width: float64(width) * ts, Height: ts},
		})
	}

	// Лестницы
	ladderCount := 2 + chunkNum/2
	for i := 0; i < ladderCount; i++ {
		x := rng.Intn(data.Width-10) + 5
		height := 2 + rng.Intn(4)
		y := data.Height - 2 - height

		data.Platforms = append(data.Platforms, &Platform{
			X:      float64(x) * ts,
			Y:      float64(y) * ts,
			Width:  ts,
			Height: float64(height) * ts,
			Type:    TileLadder,
			Solid:   false,
			Transform: entity.Transform{X: float64(x) * ts, Y: float64(y) * ts, Width: ts, Height: float64(height) * ts},
		})
	}
}

// generateEnemies генерирует врагов
func (g *LevelGenerator) generateEnemies(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)

	enemyTypes := []string{"snake"}
	if chunkNum >= 2 {
		enemyTypes = append(enemyTypes, "spider")
	}
	if chunkNum >= 4 {
		enemyTypes = append(enemyTypes, "bat")
	}
	if chunkNum >= 6 {
		enemyTypes = append(enemyTypes, "wind", "storm")
	}

	enemyCount := 3 + chunkNum*2
	for i := 0; i < enemyCount; i++ {
		x := rng.Intn(data.Width-10) + 5
		y := data.Height - 3

		if rng.Float32() < 0.4 {
			for _, p := range data.Platforms {
				if p.Solid && p.Type != TileLadder && p.Width >= ts*2 {
					x = int(p.X/ts) + rng.Intn(int(p.Width/ts))
					y = int(p.Y/ts) - 1
					break
				}
			}
		}

		enemyType := enemyTypes[rng.Intn(len(enemyTypes))]

		data.Enemies = append(data.Enemies, &LevelEnemy{
			X:      float64(x) * ts,
			Y:      float64(y) * ts,
			Type:   enemyType,
			Active: true,
		})
	}
}

// generateFriends генерирует друзей
func (g *LevelGenerator) generateFriends(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)

	friendTypes := []string{"bee", "ladybug", "frog", "snail", "ghost"}

	friendCount := 4 + chunkNum
	for i := 0; i < friendCount; i++ {
		x := rng.Intn(data.Width-10) + 5
		y := data.Height - 3

		friendType := friendTypes[rng.Intn(len(friendTypes))]

		data.Friends = append(data.Friends, &LevelFriend{
			X:    float64(x) * ts,
			Y:    float64(y) * ts,
			Type: friendType,
		})
	}
}

// generateClouds генерирует облачка
func (g *LevelGenerator) generateClouds(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)

	cloudCount := 3 + chunkNum
	for i := 0; i < cloudCount; i++ {
		x := rng.Intn(data.Width-10) + 5
		y := rng.Intn(data.Height-6) + 2
		cloudNum := (i % 3) + 1

		data.Clouds = append(data.Clouds, &LevelCloud{
			X:   float64(x) * ts,
			Y:   float64(y) * ts,
			Num: cloudNum,
		})
	}
}

// generateItems генерирует предметы
func (g *LevelGenerator) generateItems(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)

	itemTypes := []struct {
		Type   string
		Value  int
		Weight int
	}{
		{"coinGold", 10, 50},
		{"coinSilver", 5, 40},
		{"coinBronze", 3, 30},
		{"gemRed", 25, 20},
		{"gemBlue", 30, 15},
		{"star", 50, 5},
		{"mushroomRed", 0, 3},
	}

	itemCount := 8 + chunkNum*2
	for i := 0; i < itemCount; i++ {
		x := rng.Intn(data.Width-10) + 5
		y := rng.Intn(data.Height-6) + 3

		itemType := g.selectItem(itemTypes, rng)

		data.Items = append(data.Items, &LevelItem{
			X:         float64(x) * ts,
			Y:         float64(y) * ts,
			Type:      itemType.Type,
			Value:     itemType.Value,
			Collected: false,
		})
	}
}

// selectItem выбирает предмет
func (g *LevelGenerator) selectItem(items []struct {
	Type   string
	Value  int
	Weight int
}, rng *rand.Rand) struct {
	Type   string
	Value  int
	Weight int
} {
	total := 0
	for _, it := range items {
		total += it.Weight
	}

	roll := rng.Intn(total)
	current := 0
	for _, it := range items {
		current += it.Weight
		if roll < current {
			return it
		}
	}
	return items[0]
}

// generateDecors генерирует декорации
func (g *LevelGenerator) generateDecors(data *LevelData, chunkNum int, rng *rand.Rand) {
	ts := float64(g.TileSize)

	decorTypes := []string{
		"hill_large", "hill_small",
		"mushroomRed", "mushroomBrown",
		"bush", "rock", "plant",
		"flagBlue", "flagGreen", "flagRed",
		"box", "boxCoin",
		"fence", "sign",
	}

	decorCount := 8 + chunkNum*2
	for i := 0; i < decorCount; i++ {
		x := rng.Intn(data.Width-5) * int(ts)
		y := data.Height - 2

		decorType := decorTypes[rng.Intn(len(decorTypes))]

		data.Decors = append(data.Decors, &LevelDecor{
			X:    float64(x),
			Y:    float64(y),
			Type: decorType,
		})
	}
}

// getChunkName возвращает название чанка
func getChunkName(chunkNum int) string {
	names := []string{
		"Начало пути",
		"Зелёные луга",
		"Грибной лес",
		"Цветочная долина",
		"Хрустальные пещеры",
		"Конфетные горы",
		"Ледяная пустошь",
		"Забытые руины",
		"Небесные острова",
		"Бесконечность",
	}

	if chunkNum < len(names) {
		return names[chunkNum]
	}
	return "Неизведанные земли #" + string(rune('0'+chunkNum/10)) + string(rune('0'+chunkNum%10))
}

// getChunkTheme возвращает тему чанка
func getChunkTheme(chunkNum int) LevelTheme {
	if chunkNum <= 2 {
		return ThemeGrass
	} else if chunkNum <= 4 {
		return ThemeMushroom
	} else if chunkNum <= 6 {
		return ThemeCastle
	} else if chunkNum <= 8 {
		return ThemeCandy
	}
	return ThemeIce
}

// CheckExitReach проверяет достижение выхода
func (data *LevelData) CheckExitReach(x, y, width, height float64) bool {
	if data.ExitX == 0 {
		return false
	}
	return x+width > data.ExitX && y+height > data.ExitY
}
