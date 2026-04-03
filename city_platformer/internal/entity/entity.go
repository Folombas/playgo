package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Player is the main character.
type Player struct {
	X, Y         float64
	W, H         float64
	VY           float64
	OnGround     bool
	Jumps        int
	MaxJumps     int
	Score        int
	Coins        int
	HP           int
	MaxHP        int
	Invincible   int
	Sprite       *ebiten.Image
	WalkFrames   []*ebiten.Image
	Frame        int
	FrameTimer   int
}

// NewPlayer creates a player at the given position.
func NewPlayer(x, y float64) *Player {
	return &Player{
		X: x, Y: y, W: 32, H: 48,
		Jumps: 2, MaxJumps: 2,
		HP: 5, MaxHP: 5,
		OnGround: true,
	}
}

// Update updates player physics.
func (p *Player) Update() {
	if !p.OnGround {
		p.VY += 1000
	}
	p.Y += p.VY
	p.OnGround = false

	if p.Invincible > 0 {
		p.Invincible--
	}

	p.FrameTimer++
	if p.FrameTimer > 6 {
		p.FrameTimer = 0
		p.Frame = (p.Frame + 1) % len(p.WalkFrames)
	}
}

// Jump makes the player jump.
func (p *Player) Jump(force float64) bool {
	if p.Jumps > 0 {
		p.VY = -force
		p.Jumps--
		p.OnGround = false
		return true
	}
	return false
}

// Land resets jumps when landing.
func (p *Player) Land() {
	p.OnGround = true
	p.Jumps = p.MaxJumps
	p.VY = 0
}

// Hit damages the player.
func (p *Player) Hit(dmg int) {
	if p.Invincible > 0 {
		return
	}
	p.HP -= dmg
	p.Invincible = 60
	if p.HP < 0 {
		p.HP = 0
	}
}

// Obstacle is something to avoid.
type Obstacle struct {
	X, Y   float64
	W, H   float64
	Sprite *ebiten.Image
	Type   int // 0=ground, 1=flying
}

// Update moves the obstacle left.
func (o *Obstacle) Update(speed float64) {
	o.X -= speed
}

// Coin is a collectible.
type Coin struct {
	X, Y   float64
	W, H   float64
	Sprite *ebiten.Image
	Collected bool
	BobTimer float64
}

// Update animates the coin.
func (c *Coin) Update(speed float64) {
	c.X -= speed
	c.BobTimer += 0.05
}

// AABB checks overlap.
func AABB(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah >= by
}

// Lerp linear interpolation.
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// Clamp clamps value.
func Clamp(v, min, max float64) float64 {
	return math.Max(min, math.Min(max, v))
}
