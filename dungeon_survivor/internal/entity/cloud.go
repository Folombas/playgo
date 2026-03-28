// Package entity содержит игровые сущности
package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Cloud представляет облако
type Cloud struct {
	X     float64
	Y     float64
	Speed float64
	Size  float64
	Puffs []CloudPuff
}

// CloudPuff представляет часть облака
type CloudPuff struct {
	X      float64
	Y      float64
	Radius float64
}

// NewCloud создаёт новое облако
func NewCloud(x, y float64, rng *rand.Rand) *Cloud {
	size := 40 + rng.Float64()*40
	puffs := make([]CloudPuff, 3+int(rng.Float64()*3))

	for i := range puffs {
		puffs[i] = CloudPuff{
			X:      rng.Float64() * size * 1.5,
			Y:      rng.Float64() * size * 0.3,
			Radius: size * (0.4 + rng.Float64()*0.3),
		}
	}

	return &Cloud{
		X:     x,
		Y:     y,
		Speed: 0.2 + rng.Float64()*0.3,
		Size:  size,
		Puffs: puffs,
	}
}

// Update обновляет облако
func (c *Cloud) Update() {
	c.X += c.Speed
	if c.X > 1400 {
		c.X = -200
	}
}

// Draw отрисовывает облако
func (c *Cloud) Draw(screen *ebiten.Image) {
	for _, puff := range c.Puffs {
		vector.DrawFilledCircle(
			screen,
			float32(c.X+puff.X),
			float32(c.Y+puff.Y),
			float32(puff.Radius),
			color.RGBA{255, 255, 255, 220},
			true,
		)
	}
}

// Particle представляет частицу
type Particle struct {
	X      float64
	Y      float64
	VX     float64
	VY     float64
	Life   float64
	MaxLife float64
	Color  color.Color
	Size   float64
}

// NewParticle создаёт новую частицу
func NewParticle(x, y, vx, vy float64, c color.Color) *Particle {
	return &Particle{
		X:      x,
		Y:      y,
		VX:     vx,
		VY:     vy + 1, // Гравитация
		Life:   1.0,
		MaxLife: 1.0,
		Color:  c,
		Size:   3 + rand.Float64()*4,
	}
}

// Update обновляет частицу
func (p *Particle) Update() {
	p.X += p.VX
	p.Y += p.VY
	p.VY += 0.1 // Гравитация
	p.Life -= 0.02
}

// Draw отрисовывает частицу
func (p *Particle) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if p.Life <= 0 {
		return
	}

	screenX := p.X - cameraX
	screenY := p.Y - cameraY

	alpha := uint8(p.Life * 255)
	vector.DrawFilledRect(
		screen,
		float32(screenX-p.Size/2),
		float32(screenY-p.Size/2),
		float32(p.Size),
		float32(p.Size),
		color.RGBA{255, 255, 255, alpha},
		true,
	)
}

// Animal представляет животное
type Animal struct {
	X         float64
	Y         float64
	Type      string
	Width     float64
	Height    float64
	VX        float64
	VY        float64
	AnimFrame float64
	Direction float64
	ChangeDir float64
}

// NewAnimal создаёт новое животное
func NewAnimal(animalType string, x, y float64, rng *rand.Rand) *Animal {
	var width, height float64
	switch animalType {
	case "bunny":
		width, height = 30, 25
	case "bird":
		width, height = 20, 15
	case "butterfly":
		width, height = 15, 15
	default:
		width, height = 30, 25
	}

	return &Animal{
		X:         x,
		Y:         y,
		Type:      animalType,
		Width:     width,
		Height:    height,
		VX:        (rng.Float64() - 0.5) * 2,
		VY:        (rng.Float64() - 0.5) * 2,
		AnimFrame: 0,
		Direction: 0,
		ChangeDir: rng.Float64() * 100,
	}
}

// Update обновляет животное
func (a *Animal) Update() {
	a.AnimFrame += 0.15

	// Движение
	a.X += a.VX
	a.Y += a.VY

	// Случайная смена направления
	a.ChangeDir--
	if a.ChangeDir <= 0 {
		a.VX = (rand.Float64() - 0.5) * 2
		a.VY = (rand.Float64() - 0.5) * 2
		a.ChangeDir = 50 + rand.Float64()*100
	}

	// Направление движения
	if a.VX > 0 {
		a.Direction = 1
	} else if a.VX < 0 {
		a.Direction = -1
	}

	// Ограничение (простое)
	if a.X < 0 {
		a.X = 0
		a.VX = -a.VX
	}
	if a.Y < 0 {
		a.Y = 0
		a.VY = -a.VY
	}
}

// Draw отрисовывает животное
func (a *Animal) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := a.X - cameraX
	screenY := a.Y - cameraY

	switch a.Type {
	case "bunny":
		a.drawBunny(screen, screenX, screenY)
	case "bird":
		a.drawBird(screen, screenX, screenY)
	case "butterfly":
		a.drawButterfly(screen, screenX, screenY)
	}
}

// drawBunny отрисовывает кролика
func (a *Animal) drawBunny(screen *ebiten.Image, x, y float64) {
	// Тело
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(a.Width/2), color.RGBA{255, 255, 255, 255}, true)

	// Уши
	earWiggle := math.Sin(a.AnimFrame) * 3
	vector.DrawFilledRect(screen, float32(x-10), float32(y-20+earWiggle), 8, 20, color.RGBA{255, 255, 255, 255}, true)
	vector.DrawFilledRect(screen, float32(x+2), float32(y-20-earWiggle), 8, 20, color.RGBA{255, 255, 255, 255}, true)

	// Внутренняя часть ушей
	vector.DrawFilledRect(screen, float32(x-8), float32(y-18+earWiggle), 4, 12, color.RGBA{255, 180, 180, 255}, true)
	vector.DrawFilledRect(screen, float32(x+4), float32(y-18-earWiggle), 4, 12, color.RGBA{255, 180, 180, 255}, true)

	// Глаза
	vector.DrawFilledCircle(screen, float32(x-5), float32(y-3), 3, color.Black, true)
	vector.DrawFilledCircle(screen, float32(x+5), float32(y-3), 3, color.Black, true)

	// Нос
	vector.DrawFilledCircle(screen, float32(x), float32(y+5), 3, color.RGBA{255, 150, 150, 255}, true)

	// Хвост
	vector.DrawFilledCircle(screen, float32(x-a.Direction*15), float32(y+5), 6, color.White, true)
}

// drawBird отрисовывает птичку
func (a *Animal) drawBird(screen *ebiten.Image, x, y float64) {
	// Крылья (анимация)
	wingOffset := math.Sin(a.AnimFrame*2) * 10

	// Тело
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(a.Width/2), color.RGBA{100, 180, 255, 255}, true)

	// Крыло
	wingX := x - 15
	wingY := y + wingOffset
	vector.DrawFilledRect(screen, float32(wingX), float32(wingY), 15, 8, color.RGBA{80, 150, 220, 255}, true)

	// Клюв
	vector.StrokeLine(screen, float32(x+a.Direction*10), float32(y), float32(x+a.Direction*18), float32(y+3), 2, color.RGBA{255, 150, 50, 255}, false)

	// Глаз
	vector.DrawFilledCircle(screen, float32(x+a.Direction*5), float32(y-3), 2, color.Black, true)
}

// drawButterfly отрисовывает бабочку
func (a *Animal) drawButterfly(screen *ebiten.Image, x, y float64) {
	// Крылья (анимация)
	wingSpread := math.Sin(a.AnimFrame*3) * 5

	// Левое крыло
	leftWingX := x - 8 + wingSpread
	leftWingY := y
	vector.DrawFilledCircle(screen, float32(leftWingX), float32(leftWingY), 6, color.RGBA{255, 100, 200, 255}, true)

	// Правое крыло
	rightWingX := x + 8 - wingSpread
	rightWingY := y
	vector.DrawFilledCircle(screen, float32(rightWingX), float32(rightWingY), 6, color.RGBA{255, 100, 200, 255}, true)

	// Тело
	vector.DrawFilledRect(screen, float32(x-2), float32(y-8), 4, 16, color.RGBA{50, 50, 50, 255}, true)

	// Усики
	vector.StrokeLine(screen, float32(x-2), float32(y-8), float32(x-5), float32(y-14), 1, color.Black, false)
	vector.StrokeLine(screen, float32(x+2), float32(y-8), float32(x+5), float32(y-14), 1, color.Black, false)
}
