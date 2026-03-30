// Package level - система уровней и тайловых карт
// Go365 Day 90 - City Survivor
package level

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"city_platformer/internal/sprite"
)

// TileType - тип тайла
type TileType string

const (
	TileEmpty      TileType = "empty"
	TileGround     TileType = "ground"
	TileGrass      TileType = "grass"
	TileBrick      TileType = "brick"
	TileBox        TileType = "box"
	TileBoxCoin    TileType = "boxCoin"
	TileBoxItem    TileType = "boxItem"
	TileBoxEmpty   TileType = "boxEmpty"
	TileFence      TileType = "fence"
	TileLadder     TileType = "ladder"
	TileDoor       TileType = "door"
	TileBridge     TileType = "bridge"
	TilePlatform   TileType = "platform"
	TileSpike      TileType = "spike"
)

// Tile - отдельный тайл
type Tile struct {
	X, Y     int
	Type     TileType
	Solid    bool
	Dangerous bool
}

// Platform - платформа для коллизий
type Platform struct {
	X, Y, Width, Height float64
	Type                TileType
	Solid               bool
}

// LevelEnemy - данные врага на уровне
type LevelEnemy struct {
	X, Y   float64
	Type   string
	Active bool
}

// LevelItem - данные предмета на уровне
type LevelItem struct {
	X, Y      float64
	Type      string
	Value     int
	Collected bool
}

// LevelData - данные уровня
type LevelData struct {
	Tiles       [][]*Tile
	Platforms   []*Platform
	Enemies     []*LevelEnemy
	Items       []*LevelItem
	Width       int
	Height      int
	TileSize    int
	Name        string
	ExitX       float64
	ExitY       float64
	Background  *ebiten.Image
}

// NewLevelData создаёт новые данные уровня
func NewLevelData(width, height, tileSize int) *LevelData {
	return &LevelData{
		Tiles:     make([][]*Tile, height),
		Platforms: make([]*Platform, 0),
		Enemies:   make([]*LevelEnemy, 0),
		Items:     make([]*LevelItem, 0),
		Width:     width,
		Height:    height,
		TileSize:  tileSize,
	}
}

// LevelGenerator - генератор уровней
type LevelGenerator struct {
	rng        *rand.Rand
	tileSize   int
	spriteSheet *sprite.SpriteSheet
}

// NewLevelGenerator создаёт генератор уровней
func NewLevelGenerator(rng *rand.Rand, tileSize int, ss *sprite.SpriteSheet) *LevelGenerator {
	return &LevelGenerator{
		rng:         rng,
		tileSize:    tileSize,
		spriteSheet: ss,
	}
}

// GenerateLevel генерирует уровень
func (g *LevelGenerator) GenerateLevel(levelNum int) *LevelData {
	// Размер уровня в тайлах
	levelWidth := 50 + levelNum*10
	levelHeight := 10  // Уменьшил с 15 до 10 (пол будет на Y=512 вместо Y=832)

	data := NewLevelData(levelWidth, levelHeight, g.tileSize)
	data.Name = getLevelName(levelNum)

	// Генерация тайлов
	g.generateTerrain(data, levelNum)
	g.generatePlatforms(data, levelNum)
	g.generateEnemies(data, levelNum)
	g.generateItems(data, levelNum)

	// Установка выхода
	data.ExitX = float64(levelWidth-5) * float64(g.tileSize)
	data.ExitY = float64(levelHeight-3) * float64(g.tileSize)

	// Фон
	data.Background = g.spriteSheet.GetBackground()

	return data
}

// generateTerrain генерирует ландшафт
func (g *LevelGenerator) generateTerrain(data *LevelData, levelNum int) {
	ts := float64(g.tileSize)

	// Создаём пол (нижние ряды)
	groundY := data.Height - 2
	for x := 0; x < data.Width; x++ {
		// Пропускаем ямы
		if g.isPit(x, data.Width, levelNum) {
			continue
		}

		// Земля
		data.Platforms = append(data.Platforms, &Platform{
			X:      float64(x) * ts,
			Y:      float64(groundY) * ts,
			Width:  ts,
			Height: ts * 2,
			Type:   TileGround,
			Solid:  true,
		})
	}

	// Стены по краям
	data.Platforms = append(data.Platforms,
		&Platform{X: -10, Y: 0, Width: 10, Height: float64(data.Height) * ts, Type: TileBrick, Solid: true},
		&Platform{X: float64(data.Width) * ts, Y: 0, Width: 10, Height: float64(data.Height) * ts, Type: TileBrick, Solid: true},
	)
}

// isPit проверяет, должна ли здесь быть яма
func (g *LevelGenerator) isPit(x, levelWidth, levelNum int) bool {
	// Количество ям увеличивается с уровнем
	pitCount := 2 + levelNum/2
	if pitCount > 8 {
		pitCount = 8
	}

	// Генерируем позиции ям детерминированно
	for i := 0; i < pitCount; i++ {
		pitX := (levelWidth/3)*(i+1) + g.rng.Intn(5)
		if x >= pitX && x < pitX+2 {
			return true
		}
	}
	return false
}

// generatePlatforms генерирует платформы
func (g *LevelGenerator) generatePlatforms(data *LevelData, levelNum int) {
	ts := float64(g.tileSize)

	// Количество платформ растёт с уровнем
	platformCount := 5 + levelNum*2

	for i := 0; i < platformCount; i++ {
		x := g.rng.Intn(data.Width-10) + 5
		y := g.rng.Intn(data.Height-6) + 3
		width := 3 + g.rng.Intn(5)

		platformType := TileGrass
		if g.rng.Float32() < 0.3 {
			platformType = TileBrick
		} else if g.rng.Float32() < 0.2 {
			platformType = TileBox
		}

		data.Platforms = append(data.Platforms, &Platform{
			X:      float64(x) * ts,
			Y:      float64(y) * ts,
			Width:  float64(width) * ts,
			Height: ts,
			Type:   platformType,
			Solid:  true,
		})
	}

	// Добавляем лестницы
	ladderCount := 2 + levelNum/2
	for i := 0; i < ladderCount; i++ {
		x := g.rng.Intn(data.Width-10) + 5
		height := 3 + g.rng.Intn(5)
		y := data.Height - 2 - height

		data.Platforms = append(data.Platforms, &Platform{
			X:      float64(x) * ts,
			Y:      float64(y) * ts,
			Width:  ts,
			Height: float64(height) * ts,
			Type:   TileLadder,
			Solid:  false,
		})
	}
}

// generateEnemies генерирует врагов
func (g *LevelGenerator) generateEnemies(data *LevelData, levelNum int) {
	ts := g.tileSize

	// Типы врагов доступные для уровня
	enemyTypes := []string{"slime"}
	if levelNum >= 2 {
		enemyTypes = append(enemyTypes, "snail")
	}
	if levelNum >= 3 {
		enemyTypes = append(enemyTypes, "fish", "fly")
	}
	if levelNum >= 4 {
		enemyTypes = append(enemyTypes, "blocker")
	}

	enemyCount := 3 + levelNum*2
	for i := 0; i < enemyCount; i++ {
		x := g.rng.Intn(data.Width-10) + 5
		y := data.Height - 3 // На земле

		// Иногда размещаем на платформах
		if g.rng.Float32() < 0.3 {
			for _, p := range data.Platforms {
				if p.Solid && p.Type != TileLadder {
					x = int(p.X/float64(ts)) + g.rng.Intn(int(p.Width/float64(ts)))
					y = int(p.Y/float64(ts)) - 1
					break
				}
			}
		}

		enemyType := enemyTypes[g.rng.Intn(len(enemyTypes))]

		data.Enemies = append(data.Enemies, &LevelEnemy{
			X:      float64(x) * float64(ts),
			Y:      float64(y) * float64(ts),
			Type:   enemyType,
			Active: true,
		})
	}
}

// generateItems генерирует предметы
func (g *LevelGenerator) generateItems(data *LevelData, levelNum int) {
	ts := g.tileSize

	// Типы предметов
	itemTypes := []struct {
		Type  string
		Value int
		Weight int
	}{
		{"coinGold", 10, 50},
		{"coinSilver", 5, 40},
		{"coinBronze", 3, 30},
		{"gemRed", 25, 20},
		{"gemBlue", 30, 15},
		{"gemGreen", 35, 10},
		{"star", 50, 5},
	}

	itemCount := 8 + levelNum*3
	for i := 0; i < itemCount; i++ {
		x := g.rng.Intn(data.Width-10) + 5
		y := g.rng.Intn(data.Height-6) + 3

		// Выбираем тип предмета с учётом веса
		itemType := g.selectItemType(itemTypes)

		data.Items = append(data.Items, &LevelItem{
			X:         float64(x) * float64(ts),
			Y:         float64(y) * float64(ts),
			Type:      itemType.Type,
			Value:     itemType.Value,
			Collected: false,
		})
	}
}

// selectItemType выбирает тип предмета с учётом веса
func (g *LevelGenerator) selectItemType(itemTypes []struct {
	Type   string
	Value  int
	Weight int
}) struct {
	Type   string
	Value  int
	Weight int
} {
	totalWeight := 0
	for _, it := range itemTypes {
		totalWeight += it.Weight
	}

	roll := g.rng.Intn(totalWeight)
	current := 0
	for _, it := range itemTypes {
		current += it.Weight
		if roll < current {
			return it
		}
	}
	return itemTypes[0]
}

// getLevelName возвращает название уровня
func getLevelName(levelNum int) string {
	names := []string{
		"Окраины города",
		"Заброшенный район",
		"Промышленная зона",
		"Центральные улицы",
		"Метрополитен",
		"Подземный бункер",
		"Крыши небоскрёбов",
		"Военная база",
		"Зона отчуждения",
		"Точка эвакуации",
	}

	if levelNum <= len(names) {
		return names[levelNum-1]
	}
	return "Неизвестная зона"
}

// CheckPlatformCollision проверяет коллизию с платформами
func (data *LevelData) CheckPlatformCollision(x, y, width, height float64) (*Platform, bool) {
	for _, p := range data.Platforms {
		if !p.Solid {
			continue
		}
		if x < p.X+p.Width && x+width > p.X &&
			y < p.Y+p.Height && y+height > p.Y {
			return p, true
		}
	}
	return nil, false
}

// CheckItemCollection проверяет сбор предметов
func (data *LevelData) CheckItemCollection(x, y, width, height float64) *LevelItem {
	for _, item := range data.Items {
		if item.Collected {
			continue
		}
		if x < item.X+32 && x+width > item.X &&
			y < item.Y+32 && y+height > item.Y {
			item.Collected = true
			return item
		}
	}
	return nil
}

// CheckExitReach проверяет достижение выхода
func (data *LevelData) CheckExitReach(x, y, width, height float64) bool {
	return x+width > data.ExitX && y+height > data.ExitY
}

// GetTileSpriteName возвращает имя спрайта для тайла
func GetTileSpriteName(tileType TileType) string {
	switch tileType {
	case TileGround:
		return "dirt"
	case TileGrass:
		return "grass"
	case TileBrick:
		return "brickWall"
	case TileBox:
		return "box"
	case TileBoxCoin:
		return "boxCoin"
	case TileBoxItem:
		return "boxItem"
	case TileBoxEmpty:
		return "boxEmpty"
	case TileFence:
		return "fence"
	case TileLadder:
		return "ladder_mid"
	case TileDoor:
		return "door_closedTop"
	case TileBridge:
		return "bridge"
	default:
		return "dirt"
	}
}
