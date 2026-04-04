package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/bomberman_go/internal/config"
)

// Explosion - взрыв
type Explosion struct {
	Cells   [][2]int // Клетки с взрывом
	Timer   float64
	MaxTime float64 // Общее время взрыва
	Frame   int     // Кадр анимации
}

// NewExplosion создает новый взрыв
func NewExplosion(bombX, bombY float64, radius int, grid *Grid) *Explosion {
	// Создаем временную бомбу для расчета клеток
	tempBomb := &Bomb{
		X:             bombX,
		Y:             bombY,
		ExplosionRadius: radius,
	}

	cells := tempBomb.GetExplosionCells(grid)

	// Разрушаем кирпичи
	for _, cell := range cells {
		grid.Destroy(cell[0], cell[1])
	}

	return &Explosion{
		Cells:   cells,
		Timer:   0.5, // Взрыв длится 0.5 секунды
		MaxTime: 0.5,
		Frame:   0,
	}
}

// Update обновляет взрыв. Возвращает true когда взрыв закончился
func (e *Explosion) Update() bool {
	e.Timer -= 1.0 / config.TPS

	// Обновляем кадр анимации
	elapsed := e.MaxTime - e.Timer
	progress := elapsed / e.MaxTime
	e.Frame = int(progress * 3)
	if e.Frame > 2 {
		e.Frame = 2
	}

	return e.Timer <= 0
}

// Draw отрисовывает взрыв
func (e *Explosion) Draw(screen *ebiten.Image, sprite *ebiten.Image) {
	if sprite == nil {
		return
	}

	// Прозрачность зависит от времени
	alpha := e.Timer / e.MaxTime

	for _, cell := range e.Cells {
		px := float64(cell[0] * config.TileSize)
		py := float64(cell[1] * config.TileSize)

		op := &ebiten.DrawImageOptions{}
		scale := float64(config.TileSize) / float64(sprite.Bounds().Dx())
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(px, py)
		
		// Уменьшаем прозрачность по мере затухания
		op.ColorM.Scale(1, 1, 1, alpha)

		screen.DrawImage(sprite, op)
	}
}

// Contains проверяет, содержит ли взрыв данную клетку
func (e *Explosion) Contains(x, y int) bool {
	for _, cell := range e.Cells {
		if cell[0] == x && cell[1] == y {
			return true
		}
	}
	return false
}
