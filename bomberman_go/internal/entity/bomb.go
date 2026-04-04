package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/bomberman_go/internal/config"
)

// Bomb - бомба
type Bomb struct {
	X, Y            float64
	ExplosionRadius int
	Timer           float64
	Exploded        bool
	OnExplode       func()
	
	// Анимация пульсации
	pulseTimer float64
	pulseFrame int
}

// NewBomb создает новую бомбу
func NewBomb(x, y float64, radius int) *Bomb {
	return &Bomb{
		X:             x,
		Y:             y,
		ExplosionRadius: radius,
		Timer:         config.BombTimer,
		Exploded:      false,
		pulseTimer:    0,
		pulseFrame:    0,
	}
}

// Update обновляет бомбу. Возвращает true когда бомба взорвалась
func (b *Bomb) Update() bool {
	if b.Exploded {
		return true
	}

	b.Timer -= 1.0 / config.TPS
	
	// Анимация пульсации
	b.pulseTimer += 1.0 / config.TPS
	if b.pulseTimer >= 0.3 {
		b.pulseTimer = 0
		b.pulseFrame = (b.pulseFrame + 1) % 2
	}

	if b.Timer <= 0 {
		b.Exploded = true
		if b.OnExplode != nil {
			b.OnExplode()
		}
		return true
	}

	return false
}

// Draw отрисовывает бомбу
func (b *Bomb) Draw(screen *ebiten.Image, sprite *ebiten.Image) {
	if sprite == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	
	// Пульсация - легкое изменение размера
	scale := float64(config.TileSize) / float64(sprite.Bounds().Dx())
	if b.pulseFrame == 1 {
		scale *= 1.1
	}
	
	op.GeoM.Scale(scale, scale)
	
	// Центрируем бомбу
	offsetX := (float64(config.TileSize) - float64(sprite.Bounds().Dx())*scale) / 2
	offsetY := (float64(config.TileSize) - float64(sprite.Bounds().Dy())*scale) / 2
	op.GeoM.Translate(b.X+offsetX, b.Y+offsetY)

	screen.DrawImage(sprite, op)
}

// GetPosition возвращает позицию бомбы
func (b *Bomb) GetPosition() (float64, float64) {
	return b.X, b.Y
}

// GetExplosionCells возвращает все клетки, которые затронет взрыв
func (b *Bomb) GetExplosionCells(grid *Grid) [][2]int {
	cells := make([][2]int, 0)
	
	centerX := int(b.X / float64(config.TileSize))
	centerY := int(b.Y / float64(config.TileSize))

	// Центр взрыва
	cells = append(cells, [2]int{centerX, centerY})

	// 4 направления
	directions := [][2]int{
		{0, -1},  // Up
		{0, 1},   // Down
		{-1, 0},  // Left
		{1, 0},   // Right
	}

	for _, dir := range directions {
		for i := 1; i <= b.ExplosionRadius; i++ {
			x := centerX + dir[0]*i
			y := centerY + dir[1]*i

			// Проверяем границы
			if x < 0 || x >= config.GridWidth || y < 0 || y >= config.GridHeight {
				break
			}

			// Камень останавливает взрыв
			if grid.Tiles[y][x] == TileStone {
				break
			}

			cells = append(cells, [2]int{x, y})

			// Кирпич разрушается и останавливает взрыв
			if grid.Tiles[y][x] == TileBrick {
				break
			}
		}
	}

	return cells
}
