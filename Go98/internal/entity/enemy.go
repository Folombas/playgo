package entity

import (
	"math"
	"math/rand"

	"dungeon_crawler/internal/config"
	"dungeon_crawler/internal/helper"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// EnemyType defines enemy behavior pattern
type EnemyType int

const (
	EnemySlime EnemyType = iota // Patrols back and forth
	EnemyBee                    // Flies toward player when close
	EnemyFly                    // Aggressive, always chases
	EnemySnail                  // Stationary, damages on contact
)

// Enemy represents an enemy
type Enemy struct {
	Base
	Type       EnemyType
	HP         int
	MaxHP      int
	Damage     int
	Speed      float64
	PatrolDir  int // -1 or 1 for slime
	PatrolDist int
	OriginX    float64
	OriginY    float64
	AlertRange float64
	HitFlash   int
	DeathTimer int
}

func NewEnemy(x, y int, enemyType EnemyType, floorNum int) *Enemy {
	var hp, dmg int
	var speed float64

	switch enemyType {
	case EnemySlime:
		hp = 20 + floorNum*5
		dmg = 8 + floorNum*2
		speed = config.EnemySlimeSpeed
	case EnemyBee:
		hp = 15 + floorNum*3
		dmg = 10 + floorNum*2
		speed = config.EnemyBeeSpeed
	case EnemyFly:
		hp = 25 + floorNum*5
		dmg = 12 + floorNum*3
		speed = config.EnemyFlySpeed
	case EnemySnail:
		hp = 30 + floorNum*5
		dmg = 15 + floorNum*3
		speed = 0.5
	default:
		hp = 20
		dmg = 10
		speed = 1
	}

	return &Enemy{
		Base: Base{
			X:      float64(x * config.TileSize),
			Y:      float64(y * config.TileSize),
			Width:  28,
			Height: 28,
			Active: true,
		},
		Type:       enemyType,
		HP:         hp,
		MaxHP:      hp,
		Damage:     dmg,
		Speed:      speed,
		OriginX:    float64(x * config.TileSize),
		OriginY:    float64(y * config.TileSize),
		PatrolDist: 64 + rand.Intn(64),
		AlertRange: 120,
	}
}

func (e *Enemy) Update(player *Player, dungeon interface{ IsWalkable(int, int) bool }) {
	if !e.Active {
		if e.DeathTimer > 0 {
			e.DeathTimer--
		}
		return
	}

	if e.HitFlash > 0 {
		e.HitFlash--
	}

	dx := player.X - e.X
	dy := player.Y - e.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	switch e.Type {
	case EnemySlime:
		e.updateSlime(dungeon)
	case EnemyBee:
		e.updateBee(player, dist, dungeon)
	case EnemyFly:
		e.updateFly(player, dist, dungeon)
	case EnemySnail:
		// Stationary, does nothing
	}
}

func (e *Enemy) updateSlime(dungeon interface{ IsWalkable(int, int) bool }) {
	// Patrol back and forth
	moveX := float64(e.PatrolDir) * e.Speed
	newX := e.X + moveX

	tileSize := float64(config.TileSize)
	if dungeon.IsWalkable(int(newX/tileSize), int(e.Y/tileSize)) &&
		dungeon.IsWalkable(int((newX+e.Width)/tileSize), int(e.Y/tileSize)) {
		e.X = newX
	} else {
		e.PatrolDir *= -1
	}

	// Check if too far from origin
	if math.Abs(e.X-e.OriginX) > float64(e.PatrolDist) {
		e.PatrolDir *= -1
	}
}

func (e *Enemy) updateBee(player *Player, dist float64, dungeon interface{ IsWalkable(int, int) bool }) {
	// Only alert when player is close
	if dist > e.AlertRange {
		return
	}

	dx := player.X - e.X
	dy := player.Y - e.Y
	length := math.Sqrt(dx*dx + dy*dy)

	if length > 0 {
		nx := (dx / length) * e.Speed
		ny := (dy / length) * e.Speed

		tileSize := float64(config.TileSize)
		newX := e.X + nx
		newY := e.Y + ny

		if dungeon.IsWalkable(int(newX/tileSize), int(e.Y/tileSize)) {
			e.X = newX
		}
		if dungeon.IsWalkable(int(e.X/tileSize), int(newY/tileSize)) {
			e.Y = newY
		}
	}
}

func (e *Enemy) updateFly(player *Player, dist float64, dungeon interface{ IsWalkable(int, int) bool }) {
	// Always chases player
	dx := player.X - e.X
	dy := player.Y - e.Y
	length := math.Sqrt(dx*dx + dy*dy)

	if length > 0 {
		nx := (dx / length) * e.Speed
		ny := (dy / length) * e.Speed

		tileSize := float64(config.TileSize)
		newX := e.X + nx
		newY := e.Y + ny

		if dungeon.IsWalkable(int(newX/tileSize), int(e.Y/tileSize)) {
			e.X = newX
		}
		if dungeon.IsWalkable(int(e.X/tileSize), int(newY/tileSize)) {
			e.Y = newY
		}
	}
}

func (e *Enemy) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if !e.Active {
		return
	}

	if e.Image == nil {
		// Fallback: colored rectangle based on type
		var c color.Color
		if e.HitFlash > 0 && e.HitFlash%4 < 2 {
			c = color.RGBA{255, 255, 255, 255} // Flash white
		} else {
			switch e.Type {
			case EnemySlime:
				c = color.RGBA{0, 200, 0, 255} // Green slime
			case EnemyBee:
				c = color.RGBA{255, 200, 0, 255} // Yellow bee
			case EnemyFly:
				c = color.RGBA{200, 0, 0, 255} // Red fly
			case EnemySnail:
				c = color.RGBA{128, 76, 25, 255} // Brown snail
			}
		}

		helper.DrawRect(screen, e.X-offsetX, e.Y-offsetY, e.Width, e.Height, c)
	} else {
		op := &ebiten.DrawImageOptions{}
		if e.HitFlash > 0 && e.HitFlash%4 < 2 {
			op.ColorScale.SetR(1.5)
			op.ColorScale.SetG(1.5)
			op.ColorScale.SetB(1.5)
		}
		op.GeoM.Translate(e.X-offsetX, e.Y-offsetY)
		screen.DrawImage(e.Image, op)
	}

	// Draw HP bar
	if e.HP < e.MaxHP {
		barWidth := e.Width
		hpRatio := float64(e.HP) / float64(e.MaxHP)

		// Background
		helper.DrawRect(screen, e.X-offsetX, e.Y-offsetY-8, barWidth, 4, color.RGBA{76, 76, 76, 255})

		// HP
		hpColor := color.RGBA{0, 255, 0, 255}
		if hpRatio < 0.3 {
			hpColor = color.RGBA{255, 0, 0, 255}
		}
		helper.DrawRect(screen, e.X-offsetX, e.Y-offsetY-8, barWidth*hpRatio, 4, hpColor)
	}
}

// TakeDamage reduces enemy HP
func (e *Enemy) TakeDamage(amount int) {
	e.HP -= amount
	e.HitFlash = 10
	if e.HP <= 0 {
		e.Active = false
		e.DeathTimer = 30
	}
}

// IsDead returns true if enemy is defeated
func (e *Enemy) IsDead() bool {
	return !e.Active && e.DeathTimer <= 0
}

// IsAlive returns true if enemy is still active
func (e *Enemy) IsAlive() bool {
	return e.Active
}
