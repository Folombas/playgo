package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// PlayerState represents the current state of the player.
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerFalling
	PlayerHurt
	PlayerDead
)

// Player represents the main character.
type Player struct {
	X, Y           float64
	VX, VY         float64
	Width, Height  float64
	State          PlayerState
	FacingRight    bool
	OnGround       bool
	JumpsLeft      int
	MaxJumps       int
	HP             int
	MaxHP          int
	InvincibleTime int
	Score          int
	Coins          int
	AnimFrame      int
	AnimTimer      int
	Sprite         *ebiten.Image
	WalkSprites    []*ebiten.Image
	IdleSprite     *ebiten.Image
	JumpSprite     *ebiten.Image
	HurtSprite     *ebiten.Image
}

// NewPlayer creates a new player entity.
func NewPlayer(x, y float64) *Player {
	return &Player{
		X:          x,
		Y:          y,
		Width:      32,
		Height:     48,
		State:      PlayerIdle,
		FacingRight: true,
		OnGround:   false,
		JumpsLeft:  2,
		MaxJumps:   2,
		HP:         5,
		MaxHP:      5,
		Score:      0,
		Coins:      0,
	}
}

// Update updates player physics and animation.
func (p *Player) Update(gravity, friction float64) {
	// Apply gravity only when airborne
	if !p.OnGround {
		p.VY += gravity
	}

	// Apply friction to horizontal velocity
	p.VX *= friction

	// Clamp horizontal velocity
	maxSpeed := 300.0
	if p.VX > maxSpeed {
		p.VX = maxSpeed
	}
	if p.VX < -maxSpeed {
		p.VX = -maxSpeed
	}

	// Clamp terminal velocity
	maxFall := 600.0
	if p.VY > maxFall {
		p.VY = maxFall
	}

	// Update position
	p.X += p.VX
	p.Y += p.VY

	// Update state
	if !p.OnGround {
		if p.VY < 0 {
			p.State = PlayerJumping
		} else {
			p.State = PlayerFalling
		}
	} else if p.VX != 0 {
		p.State = PlayerRunning
	} else {
		p.State = PlayerIdle
	}

	// Invincibility timer
	if p.InvincibleTime > 0 {
		p.InvincibleTime--
	}

	// Animation
	p.AnimTimer++
	if p.AnimTimer >= 8 {
		p.AnimTimer = 0
		if len(p.WalkSprites) > 0 {
			p.AnimFrame = (p.AnimFrame + 1) % len(p.WalkSprites)
		}
	}
}

// Jump makes the player jump if jumps are available.
func (p *Player) Jump(force float64) bool {
	if p.JumpsLeft > 0 {
		p.VY = -force
		p.JumpsLeft--
		p.OnGround = false
		return true
	}
	return false
}

// TakeDamage reduces HP and sets invincibility.
func (p *Player) TakeDamage(amount int) {
	if p.InvincibleTime > 0 {
		return
	}
	p.HP -= amount
	p.InvincibleTime = 60 // 1 second at 60fps
	p.State = PlayerHurt
	if p.HP <= 0 {
		p.HP = 0
		p.State = PlayerDead
	}
}

// ResetGround resets jump counter when landing.
func (p *Player) ResetGround() {
	p.OnGround = true
	p.JumpsLeft = p.MaxJumps
	if p.State == PlayerJumping || p.State == PlayerFalling {
		p.State = PlayerIdle
	}
}

// CurrentSprite returns the current sprite based on state.
func (p *Player) CurrentSprite() *ebiten.Image {
	switch p.State {
	case PlayerHurt, PlayerDead:
		if p.HurtSprite != nil {
			return p.HurtSprite
		}
	case PlayerJumping, PlayerFalling:
		if p.JumpSprite != nil {
			return p.JumpSprite
		}
	case PlayerRunning:
		if len(p.WalkSprites) > 0 {
			return p.WalkSprites[p.AnimFrame]
		}
	}
	if p.IdleSprite != nil {
		return p.IdleSprite
	}
	return p.Sprite
}

// Platform represents a solid platform.
type Platform struct {
	X, Y          float64
	Width, Height float64
	Sprite        *ebiten.Image
	TileWidth     int
	TileHeight    int
}

// NewPlatform creates a new platform.
func NewPlatform(x, y, w, h float64) *Platform {
	return &Platform{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
	}
}

// EnemyType represents different enemy types.
type EnemyType int

const (
	EnemySlime EnemyType = iota
	EnemyBat
	EnemySnail
	EnemyFish
)

// Enemy represents an enemy entity.
type Enemy struct {
	X, Y           float64
	VX, VY         float64
	Width, Height  float64
	Type           EnemyType
	HP             int
	MaxHP          int
	PatrolStart    float64
	PatrolEnd      float64
	Alive          bool
	AnimFrame      int
	AnimTimer      int
	Sprite         *ebiten.Image
	WalkSprites    []*ebiten.Image
	DeadSprite     *ebiten.Image
	FacingRight    bool
}

// NewEnemy creates a new enemy.
func NewEnemy(x, y float64, et EnemyType, patrolStart, patrolEnd float64) *Enemy {
	hp := 1
	speed := 50.0
	if et == EnemyBat {
		hp = 1
		speed = 80.0
	}

	e := &Enemy{
		X:           x,
		Y:           y,
		Width:       32,
		Height:      32,
		Type:        et,
		HP:          hp,
		MaxHP:       hp,
		PatrolStart: patrolStart,
		PatrolEnd:   patrolEnd,
		Alive:       true,
		FacingRight: true,
		VX:          speed,
	}
	return e
}

// Update updates enemy AI and physics.
func (e *Enemy) Update(gravity float64) {
	if !e.Alive {
		return
	}

	e.VY += gravity

	// Patrol movement
	e.X += e.VX
	e.Y += e.VY

	// Reverse at patrol bounds
	if e.X <= e.PatrolStart {
		e.VX = -e.VX
		e.FacingRight = true
	}
	if e.X >= e.PatrolEnd {
		e.VX = -e.VX
		e.FacingRight = false
	}

	// Animation
	e.AnimTimer++
	if e.AnimTimer >= 10 {
		e.AnimTimer = 0
		e.AnimFrame = (e.AnimFrame + 1) % 2
	}
}

// TakeDamage damages the enemy.
func (e *Enemy) TakeDamage(amount int) {
	e.HP -= amount
	if e.HP <= 0 {
		e.Alive = false
	}
}

// CurrentSprite returns current enemy sprite.
func (e *Enemy) CurrentSprite() *ebiten.Image {
	if !e.Alive {
		if e.DeadSprite != nil {
			return e.DeadSprite
		}
		return e.Sprite
	}
	if len(e.WalkSprites) > 0 {
		return e.WalkSprites[e.AnimFrame]
	}
	return e.Sprite
}

// ItemType represents different collectible types.
type ItemType int

const (
	ItemCoin ItemType = iota
	ItemGem
	ItemStar
	ItemHeart
	ItemKey
)

// Item represents a collectible item.
type Item struct {
	X, Y          float64
	Width, Height float64
	Type          ItemType
	Collected     bool
	Value         int
	FloatOffset   float64
	FloatTimer    float64
	Sprite        *ebiten.Image
}

// NewItem creates a new item.
func NewItem(x, y float64, it ItemType, value int) *Item {
	w, h := 16.0, 16.0
	if it == ItemStar {
		w, h = 24.0, 24.0
	}
	return &Item{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
		Type:   it,
		Value:  value,
	}
}

// Update updates item animation.
func (i *Item) Update(dt float64) {
	if i.Collected {
		return
	}
	i.FloatTimer += dt
	i.FloatOffset = math.Sin(i.FloatTimer*3) * 3
}

// CheckAABB checks axis-aligned bounding box collision.
func CheckAABB(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah >= by
}
