package projectile

import (
	"image/color"
	"math"

	"towerdefense/internal/enemy"
	"towerdefense/internal/tower"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Projectile represents a projectile fired by a tower
type Projectile struct {
	X, Y       float64
	VX, VY     float64
	Damage     float64
	Speed      float64
	Target     *enemy.Enemy
	Type       int // Matches tower type
	Radius     float64 // For splash
	SlowFactor float64 // For slow towers
	SlowDuration int
	Alive      bool
}

// NewProjectile creates a new projectile
func NewProjectile(t *tower.Tower, target *enemy.Enemy) *Projectile {
	dx := target.X - t.X
	dy := target.Y - t.Y
	dist := math.Hypot(dx, dy)

	if dist == 0 {
		return nil
	}

	speed := 8.0
	stats := tower.GetType(t.Type)

	p := &Projectile{
		X:      t.X,
		Y:      t.Y,
		Damage: t.Damage,
		Speed:  speed,
		Target: target,
		Type:   t.Type,
		Alive:  true,
	}

	// Special properties based on tower type
	switch t.Type {
	case tower.TowerSlow:
		p.SlowFactor = 0.5
		p.SlowDuration = 60
		p.Radius = 20
	case tower.TowerSplash:
		p.Radius = 60
	case tower.TowerLaser:
		p.Speed = 100 // Instant hit
	}

	// Normalize direction
	p.VX = (dx / dist) * speed
	p.VY = (dy / dist) * speed

	return p
}

// Update updates projectile position
func (p *Projectile) Update() {
	if !p.Alive {
		return
	}

	// Homing missile - track target
	if p.Target != nil && p.Target.IsAlive() {
		dx := p.Target.X - p.X
		dy := p.Target.Y - p.Y
		dist := math.Hypot(dx, dy)

		if dist < 10 {
			// Hit!
			p.Hit()
			return
		}

		// Adjust velocity
		p.VX = (dx / dist) * p.Speed
		p.VY = (dy / dist) * p.Speed
	} else {
		// Target dead, continue in last direction
	}

	p.X += p.VX
	p.Y += p.VY

	// Out of bounds
	if p.X < -50 || p.X > 2000 || p.Y < -50 || p.Y > 1500 {
		p.Alive = false
	}
}

// Hit applies damage to target
func (p *Projectile) Hit() {
	p.Alive = false

	if p.Target == nil || !p.Target.IsAlive() {
		return
	}

	if p.Type == tower.TowerSplash {
		// Splash damage to all enemies in radius
		// This is handled by game loop
	} else {
		p.Target.TakeDamage(p.Damage)
		if p.Type == tower.TowerSlow {
			p.Target.ApplySlow(p.SlowFactor, p.SlowDuration)
		}
	}
}

// Draw renders the projectile
func (p *Projectile) Draw(screen *ebiten.Image) {
	if !p.Alive {
		return
	}

	var projColor color.Color
	var size float64

	switch p.Type {
	case tower.TowerBasic:
		projColor = color.RGBA{100, 255, 100, 255}
		size = 4
	case tower.TowerSniper:
		projColor = color.RGBA{100, 100, 255, 255}
		size = 6
	case tower.TowerSlow:
		projColor = color.RGBA{100, 200, 255, 255}
		size = 5
	case tower.TowerSplash:
		projColor = color.RGBA{255, 150, 50, 255}
		size = 7
	case tower.TowerLaser:
		projColor = color.RGBA{255, 50, 50, 255}
		size = 3
	}

	// Trail effect
	for i := 3; i > 0; i-- {
		alpha := uint8(100 - i*30)
		trailX := p.X - p.VX*float64(i)*0.3
		trailY := p.Y - p.VY*float64(i)*0.3
		vector.DrawFilledCircle(screen, trailX, trailY, size*0.6, color.RGBA{255, 255, 255, alpha}, true)
	}

	// Main projectile
	vector.DrawFilledCircle(screen, p.X, p.Y, size, projColor, true)
	vector.StrokeCircle(screen, p.X, p.Y, size, 1, color.White, true)
}

// GetColor returns projectile color
func (p *Projectile) GetColor() color.Color {
	switch p.Type {
	case tower.TowerBasic:
		return color.RGBA{100, 255, 100, 255}
	case tower.TowerSniper:
		return color.RGBA{100, 100, 255, 255}
	case tower.TowerSlow:
		return color.RGBA{100, 200, 255, 255}
	case tower.TowerSplash:
		return color.RGBA{255, 150, 50, 255}
	case tower.TowerLaser:
		return color.RGBA{255, 50, 50, 255}
	default:
		return color.White
	}
}
