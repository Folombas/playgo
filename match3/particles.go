package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Particle represents a visual effect particle.
type Particle struct {
	X, Y     float64
	VX, VY   float64
	Life     float64 // 0.0 to 1.0
	Decay    float64
	Size     float64
	Color    color.RGBA
	Rotation float64
	Spin     float64
	Gravity  float64
}

// ParticleSystem manages active particles.
type ParticleSystem struct {
	Particles []*Particle
}

// NewParticleSystem creates a new particle system.
func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		Particles: make([]*Particle, 0, 200),
	}
}

// Emit spawns particles at a position.
func (ps *ParticleSystem) Emit(x, y float64, count int, colorOverride *color.RGBA) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 50 + rand.Float64()*150

		c := tileColors[rand.Intn(len(tileColors))]
		if colorOverride != nil {
			c = *colorOverride
		}

		particle := &Particle{
			X:        x,
			Y:        y,
			VX:       math.Cos(angle) * speed,
			VY:       math.Sin(angle) * speed,
			Life:     1.0,
			Decay:    0.8 + rand.Float64()*1.2,
			Size:     3 + rand.Float64()*5,
			Color:    c,
			Rotation: rand.Float64() * math.Pi * 2,
			Spin:     (rand.Float64() - 0.5) * 10,
			Gravity:  200 + rand.Float64()*100,
		}
		ps.Particles = append(ps.Particles, particle)
	}
}

// EmitSparkle spawns sparkle particles (small, bright, short-lived).
func (ps *ParticleSystem) EmitSparkle(x, y float64, count int) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 30 + rand.Float64()*80

		particle := &Particle{
			X:        x,
			Y:        y,
			VX:       math.Cos(angle) * speed,
			VY:       math.Sin(angle)*speed - 50,
			Life:     1.0,
			Decay:    1.5 + rand.Float64()*1.0,
			Size:     2 + rand.Float64()*3,
			Color:    color.RGBA{0xFF, 0xFF, 0xAA, 0xFF},
			Rotation: 0,
			Spin:     0,
			Gravity:  -50, // float up slightly
		}
		ps.Particles = append(ps.Particles, particle)
	}
}

// Update advances all particles. Returns true if all particles are dead.
func (ps *ParticleSystem) Update(dt float64) {
	for i := len(ps.Particles) - 1; i >= 0; i-- {
		p := ps.Particles[i]

		p.VY += p.Gravity * dt
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.Rotation += p.Spin * dt
		p.Life -= p.Decay * dt

		if p.Life <= 0 {
			ps.Particles = append(ps.Particles[:i], ps.Particles[i+1:]...)
		}
	}
}

// Draw renders all active particles.
func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for _, p := range ps.Particles {
		alpha := uint8(float64(p.Color.A) * p.Life)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		op.GeoM.Rotate(p.Rotation)
		op.GeoM.Translate(-p.Size/2, -p.Size/2)

		// Create a small colored square
		img := ebiten.NewImage(int(p.Size), int(p.Size))
		vector.DrawFilledRect(img, 0, 0, float32(p.Size), float32(p.Size), c, true)

		op.ColorScale.ScaleAlpha(float32(p.Life))
		screen.DrawImage(img, op)
	}
}

// IsEmpty returns true if no particles are alive.
func (ps *ParticleSystem) IsEmpty() bool {
	return len(ps.Particles) == 0
}
