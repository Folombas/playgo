package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ProjectileType определяет тип снаряда
type ProjectileType int

const (
	ProjectileBullet ProjectileType = iota
	ProjectileFireball
	ProjectileMagic
)

// Projectile представляет снаряд
type Projectile struct {
	*Entity
	VX         float64
	VY         float64
	Damage     float64
	Lifetime   float64
	MaxLifetime float64
	Type       ProjectileType
	Source     *Player
}

// NewProjectile создаёт новый снаряд
func NewProjectile(x, y, angle, speed, damage float64, lifetime float64, projType ProjectileType, source *Player) *Projectile {
	vx := math.Cos(angle) * speed
	vy := math.Sin(angle) * speed

	var projColor color.Color
	var size float64

	switch projType {
	case ProjectileFireball:
		projColor = color.RGBA{255, 100, 0, 255}
		size = 10
	case ProjectileMagic:
		projColor = color.RGBA{150, 50, 255, 255}
		size = 8
	default:
		projColor = color.RGBA{255, 255, 0, 255}
		size = 6
	}

	return &Projectile{
		Entity:      NewEntity(x, y, size, size, 0, 1, damage, projColor),
		VX:          vx,
		VY:          vy,
		Damage:      damage,
		Lifetime:    lifetime,
		MaxLifetime: lifetime,
		Type:        projType,
		Source:      source,
	}
}

// Update обновляет снаряд
func (p *Projectile) Update() {
	if !p.IsActive {
		return
	}

	p.X += p.VX
	p.Y += p.VY
	p.Lifetime -= 1.0 / 60.0

	if p.Lifetime <= 0 {
		p.IsActive = false
	}
}

// Draw отрисовывает снаряд
func (p *Projectile) Draw(screen *ebiten.Image) {
	if !p.IsActive {
		return
	}

	switch p.Type {
	case ProjectileFireball:
		// Огненный шар с эффектом
		vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), float32(p.Width/2), color.RGBA{255, 100, 0, 255}, true)
		vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), float32(p.Width/3), color.RGBA{255, 200, 50, 255}, true)

	case ProjectileMagic:
		// Магический снаряд
		vector.DrawFilledRect(screen, float32(p.X-p.Width/2), float32(p.Y-p.Height/2), float32(p.Width), float32(p.Height), p.Color, true)

	default:
		// Обычная пуля
		vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), float32(p.Width/2), p.Color, true)
	}
}

// FromPlayerToEnemy создаёт снаряд от игрока к ближайшему врагу
func FromPlayerToEnemy(player *Player, enemies []*Enemy, projType ProjectileType, speed, damage float64, lifetime float64) *Projectile {
	if len(enemies) == 0 {
		return nil
	}

	// Найти ближайшего врага в радиусе атаки
	var nearest *Enemy
	nearestDist := player.Range

	for _, enemy := range enemies {
		if !enemy.IsActive {
			continue
		}

		dist := player.DistanceTo(enemy.X, enemy.Y)
		if dist < nearestDist {
			nearestDist = dist
			nearest = enemy
		}
	}

	if nearest == nil {
		return nil
	}

	// Вычислить угол к врагу
	angle := math.Atan2(nearest.Y-player.Y, nearest.X-player.X)

	return NewProjectile(player.X, player.Y, angle, speed, damage, lifetime, projType, player)
}
