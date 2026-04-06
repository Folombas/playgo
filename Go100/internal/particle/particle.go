package particle

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Particle represents a visual particle effect
type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    int
	MaxLife int
	Color   color.Color
	Size    float64
	Type    ParticleType
}

type ParticleType int

const (
	PTExplosion ParticleType = iota
	PTHit
	PTDeath
	PTCoin
	PTLevelUp
)

// System manages all particles
type System struct {
	particles []*Particle
}

// NewSystem creates a new particle system
func NewSystem() *System {
	return &System{
		particles: make([]*Particle, 0),
	}
}

// Emit spawns particles
func (ps *System) Emit(x, y float64, ptype ParticleType, count int) {
	for i := 0; i < count; i++ {
		p := &Particle{
			X:    x + (rand.Float64()-0.5)*10,
			Y:    y + (rand.Float64()-0.5)*10,
			Life: 20 + rand.Intn(30),
			Size: 2 + rand.Float64()*4,
		}

		angle := rand.Float64() * math.Pi * 2
		speed := 1 + rand.Float64()*4

		switch ptype {
		case PTExplosion:
			c := rand.Intn(3)
			switch c {
			case 0:
				p.Color = color.RGBA{255, 200, 50, 255}
			case 1:
				p.Color = color.RGBA{255, 100, 0, 255}
			case 2:
				p.Color = color.RGBA{255, 255, 200, 255}
			}
			p.VX = math.Cos(angle) * speed * 1.5
			p.VY = math.Sin(angle) * speed * 1.5
		case PTHit:
			p.Color = color.RGBA{255, 255, 255, 200}
			p.VX = (rand.Float64() - 0.5) * 3
			p.VY = (rand.Float64() - 0.5) * 3
			p.Size = 1 + rand.Float64()*2
		case PTDeath:
			p.Color = color.RGBA{255, 50, 50, 255}
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle)*speed - 2
			p.Size = 3 + rand.Float64()*4
		case PTCoin:
			p.Color = color.RGBA{255, 215, 0, 255}
			p.VX = (rand.Float64() - 0.5) * 2
			p.VY = -2 - rand.Float64()*3
			p.Size = 2 + rand.Float64()*3
		case PTLevelUp:
			c := rand.Intn(3)
			switch c {
			case 0:
				p.Color = color.RGBA{255, 215, 0, 255}
			case 1:
				p.Color = color.RGBA{100, 255, 100, 255}
			case 2:
				p.Color = color.RGBA{100, 150, 255, 255}
			}
			p.VX = math.Cos(angle) * speed * 0.8
			p.VY = math.Sin(angle)*speed*0.8 - 1
		}

		p.MaxLife = p.Life
		ps.particles = append(ps.particles, p)
	}
}

// Update updates all particles
func (ps *System) Update() {
	active := make([]*Particle, 0, len(ps.particles))
	for _, p := range ps.particles {
		p.X += p.VX
		p.Y += p.VY
		p.Life--

		if p.Type == PTDeath || p.Type == PTExplosion {
			p.VY += 0.1
		}

		p.VX *= 0.98
		p.VY *= 0.98

		if p.Life > 0 {
			active = append(active, p)
		}
	}
	ps.particles = active
}

// Draw renders all particles
func (ps *System) Draw(screen *ebiten.Image) {
	for _, p := range ps.particles {
		alpha := float64(p.Life) / float64(p.MaxLife)
		size := p.Size * alpha

		if col, ok := p.Color.(color.RGBA); ok {
			c := color.RGBA{col.R, col.G, col.B, uint8(float64(col.A) * alpha)}
			vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), float32(size), c, false)
		}
	}
}

// IsActive returns true if there are active particles
func (ps *System) IsActive() bool {
	return len(ps.particles) > 0
}
