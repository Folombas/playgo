// Package entity содержит игровые сущности
package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Entity представляет базовую сущность в игре
type Entity struct {
	X        float64
	Y        float64
	Width    float64
	Height   float64
	Speed    float64
	HP       float64
	MaxHP    float64
	Damage   float64
	Color    color.Color
	IsActive bool
}

// NewEntity создаёт новую сущность
func NewEntity(x, y, width, height, speed, hp, damage float64, c color.Color) *Entity {
	return &Entity{
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		Speed:    speed,
		HP:       hp,
		MaxHP:    hp,
		Damage:   damage,
		Color:    c,
		IsActive: true,
	}
}

// Draw отрисовывает сущность
func (e *Entity) Draw(screen *ebiten.Image) {
	if !e.IsActive {
		return
	}
	vector.DrawFilledRect(
		screen,
		float32(e.X-e.Width/2),
		float32(e.Y-e.Height/2),
		float32(e.Width),
		float32(e.Height),
		e.Color,
		true,
	)
}

// Bounds возвращает границы сущности для коллизий
func (e *Entity) Bounds() (left, right, top, bottom float64) {
	return e.X - e.Width/2, e.X + e.Width/2, e.Y - e.Height/2, e.Y + e.Height/2
}

// DistanceTo вычисляет расстояние до другой сущности
func (e *Entity) DistanceTo(other *Entity) float64 {
	dx := e.X - other.X
	dy := e.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// TakeDamage наносит урон сущности
func (e *Entity) TakeDamage(damage float64) {
	e.HP -= damage
	if e.HP <= 0 {
		e.IsActive = false
	}
}

// Heal лечит сущность
func (e *Entity) Heal(amount float64) {
	e.HP += amount
	if e.HP > e.MaxHP {
		e.HP = e.MaxHP
	}
}

// HPPercent возвращает процент здоровья
func (e *Entity) HPPercent() float64 {
	return e.HP / e.MaxHP
}
