package entity

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/bomberman_go/internal/config"
)

// TileType - тип тайла
type TileType int

const (
	TileEmpty TileType = iota
	TileStone          // Неразрушаемая стена
	TileBrick          // Разрушаемая стена
)

// Grid - игровое поле
type Grid struct {
	Tiles [][]TileType
}

// NewGrid создает новое игровое поле
func NewGrid() *Grid {
	g := &Grid{
		Tiles: make([][]TileType, config.GridHeight),
	}

	// Инициализируем пустое поле
	for y := range g.Tiles {
		g.Tiles[y] = make([]TileType, config.GridWidth)
	}

	// Создаем сетку из неразрушаемых стен (каждая вторая клетка)
	for y := 0; y < config.GridHeight; y += 2 {
		for x := 0; x < config.GridWidth; x += 2 {
			g.Tiles[y][x] = TileStone
		}
	}

	// Случайно размещаем разрушаемые стены
	rand.Seed(42) // Фиксированный seed для воспроизводимости
	for y := 1; y < config.GridHeight-1; y++ {
		for x := 1; x < config.GridWidth-1; x++ {
			// Не ставим стены на стартовой позиции игрока (1,1), (1,2), (2,1)
			if (x <= 2 && y <= 2) {
				continue
			}

			// 30% шанс появления кирпича
			if rand.Float32() < 0.3 {
				g.Tiles[y][x] = TileBrick
			}
		}
	}

	return g
}

// IsWalkable проверяет, можно ли пройти через тайл
func (g *Grid) IsWalkable(x, y int) bool {
	if x < 0 || x >= config.GridWidth || y < 0 || y >= config.GridHeight {
		return false
	}
	return g.Tiles[y][x] == TileEmpty
}

// IsDestructible проверяет, является ли тайл разрушаемым
func (g *Grid) IsDestructible(x, y int) bool {
	if x < 0 || x >= config.GridWidth || y < 0 || y >= config.GridHeight {
		return false
	}
	return g.Tiles[y][x] == TileBrick
}

// Destroy разрушает тайл (кирпич)
func (g *Grid) Destroy(x, y int) {
	if x >= 0 && x < config.GridWidth && y >= 0 && y < config.GridHeight {
		if g.Tiles[y][x] == TileBrick {
			g.Tiles[y][x] = TileEmpty
		}
	}
}

// Draw отрисовывает игровое поле
func (g *Grid) Draw(screen *ebiten.Image, sprites map[string]*ebiten.Image) {
	grassSprite := sprites["grass"]
	stoneSprite := sprites["stone"]
	brickSprite := sprites["brick"]

	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			px := float64(x * config.TileSize)
			py := float64(y * config.TileSize)

			// Рисуем траву как фон
			if grassSprite != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(px, py)
				screen.DrawImage(grassSprite, op)
			}

			// Рисуем стены
			switch g.Tiles[y][x] {
			case TileStone:
				if stoneSprite != nil {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(px, py)
					screen.DrawImage(stoneSprite, op)
				}
			case TileBrick:
				if brickSprite != nil {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(px, py)
					screen.DrawImage(brickSprite, op)
				}
			}
		}
	}
}

// HasPowerUp проверяет, есть ли бонус под разрушаемой стеной
func (g *Grid) HasPowerUp(x, y int) bool {
	// Пока просто случайный шанс
	rand.Seed(int64(x*1000 + y))
	return rand.Float32() < 0.2 // 20% шанс бонуса
}
