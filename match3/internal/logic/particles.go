package logic

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// ParticleType определяет тип частицы
type ParticleType int

const (
	ParticleExplosion ParticleType = iota
	ParticleSparkle
	ParticleFire
	ParticleRainbow
	ParticleBomb
)

// Particle представляет частицу для визуальных эффектов
type Particle struct {
	X, Y       float64
	VX, VY     float64
	Life       float64
	MaxLife    float64
	Size       float64
	Color      color.RGBA
	Type       ParticleType
	Rotation   float64
	RotSpeed   float64
	Gravity    float64
	FadeOut    bool
	Shrink     bool
}

// ParticleSystem управит системой частиц
type ParticleSystem struct {
	particles []Particle
	rng       *rand.Rand
}

// NewParticleSystem создаёт новую систему частиц
func NewParticleSystem(rng *rand.Rand) *ParticleSystem {
	return &ParticleSystem{
		particles: make([]Particle, 0, 100),
		rng:       rng,
	}
}

// Emit создаёт новые частицы
func (ps *ParticleSystem) Emit(x, y float64, particleType ParticleType, count int, options ...ParticleOption) {
	for i := 0; i < count; i++ {
		p := Particle{
			X:       x,
			Y:       y,
			Life:    1.0,
			MaxLife: 1.0,
			Type:    particleType,
			Gravity: 0.3,
		}

		// Применение опций
		for _, opt := range options {
			opt(&p)
		}

		// Настройки по типу
		switch particleType {
		case ParticleExplosion:
			angle := float64(i) * math.Pi * 2 / float64(count)
			speed := 2.0 + ps.rng.Float64()*3
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed - 2
			p.Size = 3 + ps.rng.Float64()*5
			p.FadeOut = true
			p.Shrink = true
		case ParticleSparkle:
			angle := ps.rng.Float64() * math.Pi * 2
			speed := 1.0 + ps.rng.Float64()*2
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed - 1
			p.Size = 2 + ps.rng.Float64()*3
			p.RotSpeed = (ps.rng.Float64() - 0.5) * 0.2
			p.FadeOut = true
		case ParticleFire:
			p.VX = (ps.rng.Float64() - 0.5) * 2
			p.VY = -2 - ps.rng.Float64()*3
			p.Size = 4 + ps.rng.Float64()*6
			p.Gravity = -0.1 // Огонь поднимается
			p.FadeOut = true
			p.Shrink = true
		case ParticleRainbow:
			angle := float64(i) * math.Pi * 2 / float64(count)
			speed := 1.5 + ps.rng.Float64()*2
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed
			p.Size = 4 + ps.rng.Float64()*4
			p.Rotation = float64(i) * 60
			p.RotSpeed = 0.1
			p.FadeOut = true
		case ParticleBomb:
			angle := float64(i) * math.Pi * 2 / float64(count)
			speed := 4.0 + ps.rng.Float64()*4
			p.VX = math.Cos(angle) * speed
			p.VY = math.Sin(angle) * speed
			p.Size = 5 + ps.rng.Float64()*7
			p.Gravity = 0.1
			p.FadeOut = true
			p.Shrink = true
		}

		ps.particles = append(ps.particles, p)
	}
}

// Update обновляет все частицы
func (ps *ParticleSystem) Update() {
	for i := len(ps.particles) - 1; i >= 0; i-- {
		p := &ps.particles[i]

		// Обновление позиции
		p.X += p.VX
		p.Y += p.VY
		p.VY += p.Gravity

		// Обновление вращения
		p.Rotation += p.RotSpeed

		// Обновление жизни
		p.Life -= 1.0 / 60.0

		// Обновление размера
		if p.Shrink {
			p.Size *= 0.95
		}

		// Удаление мёртвых частиц
		if p.Life <= 0 || p.Size < 0.5 {
			ps.particles = append(ps.particles[:i], ps.particles[i+1:]...)
		}
	}
}

// Draw рисует все частицы
func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for _, p := range ps.particles {
		alpha := uint8(255 * (p.Life / p.MaxLife))
		if p.FadeOut {
			alpha = uint8(255 * math.Pow(p.Life/p.MaxLife, 2))
		}

		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}

		// Рисование частицы
		if p.Type == ParticleRainbow {
			// Радужные частицы
			hue := math.Mod(p.Rotation+float64(p.Life*100), 360)
			rgb := hslToRgb(hue, 1.0, 0.5)
			c = color.RGBA{rgb[0], rgb[1], rgb[2], alpha}
		}

		circle := ps.createCircle(p.Size*2, c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-p.Size, p.Y-p.Size)
		if p.Rotation != 0 {
			op.GeoM.Rotate(p.Rotation)
		}
		screen.DrawImage(circle, op)
	}
}

// createCircle создаёт изображение круга
func (ps *ParticleSystem) createCircle(diameter float64, c color.RGBA) *ebiten.Image {
	size := int(math.Ceil(diameter))
	if size < 2 {
		size = 2
	}
	img := ebiten.NewImage(size, size)
	center := float64(size) / 2
	radius := diameter / 2

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			if math.Sqrt(dx*dx+dy*dy) <= radius {
				img.Set(x, y, c)
			}
		}
	}

	return img
}

// Clear очищает все частицы
func (ps *ParticleSystem) Clear() {
	ps.particles = ps.particles[:0]
}

// Count возвращает количество активных частиц
func (ps *ParticleSystem) Count() int {
	return len(ps.particles)
}

// ParticleOption - функциональная опция для настройки частиц
type ParticleOption func(*Particle)

// WithColor устанавливает цвет частицы
func WithColor(c color.RGBA) ParticleOption {
	return func(p *Particle) {
		p.Color = c
	}
}

// WithGravity устанавливает гравитацию частицы
func WithGravity(g float64) ParticleOption {
	return func(p *Particle) {
		p.Gravity = g
	}
}

// WithSpeed устанавливает начальную скорость
func WithSpeed(speed float64) ParticleOption {
	return func(p *Particle) {
		p.VX = speed
		p.VY = speed
	}
}

// WithLife устанавливает время жизни
func WithLife(life float64) ParticleOption {
	return func(p *Particle) {
		p.Life = life
		p.MaxLife = life
	}
}

// WithSize устанавливает размер частицы
func WithSize(size float64) ParticleOption {
	return func(p *Particle) {
		p.Size = size
	}
}
