package main

// Система частиц для визуальных эффектов
// Go365 Day 99 - Улучшение визуала

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var onePixelImg *ebiten.Image

func initOnePixelImg() {
	if onePixelImg == nil {
		onePixelImg = ebiten.NewImage(1, 1)
		onePixelImg.Fill(color.White)
	}
}

type Particle struct {
	X, Y       float64
	VX, VY     float64
	Life       int
	MaxLife    int
	Color      color.Color
	Size       float64
	Type       ParticleType
}

type ParticleType int

const (
	PTMoveHint ParticleType = iota
	PTCapture
	PTKingPromotion
	PTVictory
	PTTrail
)

type ParticleSystem struct {
	particles []*Particle
}

func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		particles: make([]*Particle, 0),
	}
}

func (ps *ParticleSystem) Emit(x, y float64, ptype ParticleType, count int) {
	for i := 0; i < count; i++ {
		p := &Particle{
			X:       x,
			Y:       y,
			Life:    30 + rand.Intn(30),
			Size:    2 + rand.Float64()*4,
			Type:    ptype,
			MaxLife: 60,
		}

		switch ptype {
		case PTMoveHint:
			p.Color = color.RGBA{100, 255, 100, 200}
			angle := rand.Float64() * math.Pi * 2
			speed := 0.5 + rand.Float64()*1.5
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed
		case PTCapture:
			p.Color = color.RGBA{255, 100, 50, 255}
			angle := rand.Float64() * math.Pi * 2
			speed := 1 + rand.Float64()*3
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed
		case PTKingPromotion:
			p.Color = color.RGBA{255, 215, 0, 255}
			angle := rand.Float64() * math.Pi * 2
			speed := 0.5 + rand.Float64()*2
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed - 1
		case PTVictory:
			c := rand.Intn(3)
			switch c {
			case 0:
				p.Color = color.RGBA{255, 215, 0, 255}
			case 1:
				p.Color = color.RGBA{255, 100, 100, 255}
			case 2:
				p.Color = color.RGBA{100, 200, 255, 255}
			}
			angle := rand.Float64() * math.Pi * 2
			speed := 1 + rand.Float64()*4
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed - 2
			p.Size = 3 + rand.Float64()*5
		case PTTrail:
			p.Color = color.RGBA{200, 200, 255, 150}
			p.VX = (rand.Float64() - 0.5) * 0.5
			p.VY = (rand.Float64() - 0.5) * 0.5
			p.Life = 20
			p.Size = 1 + rand.Float64()*2
		}

		p.MaxLife = p.Life
		ps.particles = append(ps.particles, p)
	}
}

func (ps *ParticleSystem) Update() {
	active := make([]*Particle, 0)
	for _, p := range ps.particles {
		p.X += p.VX
		p.Y += p.VY
		p.Life--

		// Гравитация для некоторых типов
		if p.Type == PTCapture || p.Type == PTVictory {
			p.VY += 0.05
		}

		if p.Life > 0 {
			active = append(active, p)
		}
	}
	ps.particles = active
}

func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for _, p := range ps.particles {
		alpha := float64(p.Life) / float64(p.MaxLife)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		op.GeoM.Scale(p.Size, p.Size)
		
		if col, ok := p.Color.(color.RGBA); ok {
			op.ColorScale.SetR(float32(col.R) / 255)
			op.ColorScale.SetG(float32(col.G) / 255)
			op.ColorScale.SetB(float32(col.B) / 255)
			op.ColorScale.SetA(float32(col.A) / 255 * alpha)
		}

		screen.DrawImage(onePixelImg, op)
	}
}

func (ps *ParticleSystem) IsActive() bool {
	return len(ps.particles) > 0
}
