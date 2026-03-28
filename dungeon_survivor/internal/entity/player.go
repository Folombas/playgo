// Package entity содержит игровые сущности
package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Player представляет игрока
type Player struct {
	X         float64
	Y         float64
	Width     float64
	Height    float64
	MoveSpeed float64
	Direction float64 // Угол направления
	AnimFrame float64
}

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64) *Player {
	return &Player{
		X:         x,
		Y:         y,
		Width:     40,
		Height:    40,
		MoveSpeed: 5.0,
		Direction: 0,
		AnimFrame: 0,
	}
}

// Update обновляет состояние игрока
func (p *Player) Update() {
	p.AnimFrame += 0.1
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY

	// Тело (милый круглый персонаж)
	vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(p.Width/2), color.RGBA{255, 200, 150, 255}, true)

	// Глаза
	eyeOffset := math.Cos(p.AnimFrame * 0.1) * 2
	vector.DrawFilledCircle(screen, float32(screenX-8+eyeOffset), float32(screenY-5), 5, color.White, true)
	vector.DrawFilledCircle(screen, float32(screenX+8+eyeOffset), float32(screenY-5), 5, color.White, true)

	// Зрачки
	vector.DrawFilledCircle(screen, float32(screenX-7+eyeOffset), float32(screenY-5), 2, color.Black, true)
	vector.DrawFilledCircle(screen, float32(screenX+9+eyeOffset), float32(screenY-5), 2, color.Black, true)

	// Румянец
	vector.DrawFilledCircle(screen, float32(screenX-12), float32(screenY+3), 4, color.RGBA{255, 150, 150, 180}, true)
	vector.DrawFilledCircle(screen, float32(screenX+12), float32(screenY+3), 4, color.RGBA{255, 150, 150, 180}, true)

	// Улыбка (дуга из нескольких точек)
	for i := -8; i <= 8; i += 2 {
		yOffset := float32(5 + i*i/20)
		vector.DrawFilledRect(screen, float32(screenX+float64(i)), float32(screenY)+yOffset, 2, 2, color.RGBA{150, 50, 50, 200}, true)
	}

	// Ножки (анимация ходьбы)
	if p.AnimFrame > 0 {
		legOffset := math.Sin(p.AnimFrame) * 5
		vector.DrawFilledRect(screen, float32(screenX-8), float32(screenY+18+legOffset), 6, 8, color.RGBA{200, 100, 50, 255}, true)
		vector.DrawFilledRect(screen, float32(screenX+2), float32(screenY+18-legOffset), 6, 8, color.RGBA{200, 100, 50, 255}, true)
	}
}
