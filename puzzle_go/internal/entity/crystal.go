package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/puzzle_go/internal/config"
)

// CrystalType определяет тип кристалла
type CrystalType int

const (
	CrystalRed CrystalType = iota
	CrystalBlue
	CrystalGreen
	CrystalYellow
	CrystalViolet
	CrystalOrange
	CrystalBomb
	CrystalRainbow
	CrystalBeamH
	CrystalBeamV
	CrystalEmpty
)

// Crystal представляет игровой элемент
type Crystal struct {
	Type       CrystalType
	Col        int
	Row        int
	TargetCol  int
	TargetRow  int
	X          float64
	Y          float64
	TargetX    float64
	TargetY    float64
	Scale      float64
	Alpha      float64
	IsSelected bool
	IsMatched  bool
	IsFalling  bool
	Image      *ebiten.Image
}

// NewCrystal создает новый кристалл
func NewCrystal(crystalType CrystalType, col, row int) *Crystal {
	x := float64(config.BoardOffsetX + col*(config.CellSize+config.CellPadding))
	y := float64(config.BoardOffsetY + row*(config.CellSize+config.CellPadding))

	return &Crystal{
		Type:      crystalType,
		Col:       col,
		Row:       row,
		TargetCol: col,
		TargetRow: row,
		X:         x,
		Y:         y,
		TargetX:   x,
		TargetY:   y,
		Scale:     1.0,
		Alpha:     1.0,
	}
}

// Update обновляет состояние кристалла (анимации)
func (c *Crystal) Update(progress float64) {
	// Анимация перемещения
	c.X += (c.TargetX - c.X) * progress
	c.Y += (c.TargetY - c.Y) * progress

	// Анимация масштаба при выделении
	if c.IsSelected {
		c.Scale = 1.1
	} else {
		c.Scale += (1.0 - c.Scale) * progress
	}

	// Анимация исчезновения при совпадении
	if c.IsMatched {
		c.Scale *= 0.9
		c.Alpha *= 0.85
	}
}

// GetPosition возвращает текущую позицию
func (c *Crystal) GetPosition() (float64, float64) {
	return c.X, c.Y
}

// SetTarget устанавливает целевую позицию
func (c *Crystal) SetTarget(col, row int) {
	c.TargetCol = col
	c.TargetRow = row
	c.TargetX = float64(config.BoardOffsetX + col*(config.CellSize+config.CellPadding))
	c.TargetY = float64(config.BoardOffsetY + row*(config.CellSize+config.CellPadding))
}

// Contains проверяет, находится ли точка внутри кристалла
func (c *Crystal) Contains(mx, my float64) bool {
	size := float64(config.CellSize) * c.Scale
	return mx >= c.X && mx <= c.X+size &&
		my >= c.Y && my <= c.Y+size
}
