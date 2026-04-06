package entity

import (
	"image/color"

	"dungeon_crawler/internal/config"
	"dungeon_crawler/internal/helper"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Player represents the player character
type Player struct {
	Base
	HP           int
	MaxHP        int
	AttackDMG    int
	Speed        float64
	Keys         int
	Coins        int
	Gems         int
	Potions      int
	Floor        int
	Score        int
	Invincible   int
	AttackCooldown int
	Attacking    bool
	AttackFrame  int
	AttackDir    AttackDirection
	Facing       Direction
	WalkFrame    int
	WalkTimer    int
	HitFlash     int
}

type Direction int
const (
	DirDown Direction = iota
	DirUp
	DirLeft
	DirRight
)

type AttackDirection int
const (
	AttackNone AttackDirection = iota
	AttackUp
	AttackDown
	AttackLeft
	AttackRight
)

func NewPlayer(x, y int) *Player {
	return &Player{
		Base: Base{
			X: float64(x * config.TileSize),
			Y: float64(y * config.TileSize),
			Width:  28,
			Height: 28,
			Active: true,
		},
		HP:        config.PlayerMaxHP,
		MaxHP:     config.PlayerMaxHP,
		AttackDMG: config.PlayerAttackDMG,
		Speed:     config.PlayerSpeed,
		Invincible: 0,
	}
}

func (p *Player) Update(dungeon interface{ IsWalkable(int, int) bool }) error {
	// Update timers
	if p.Invincible > 0 {
		p.Invincible--
	}
	if p.AttackCooldown > 0 {
		p.AttackCooldown--
	}
	if p.HitFlash > 0 {
		p.HitFlash--
	}

	// Attack input
	if p.AttackCooldown <= 0 && !p.Attacking {
		if inpututil.IsKeyJustPressed(config.GameKeys.Attack) {
			p.Attacking = true
			p.AttackFrame = 15
			p.AttackCooldown = 20

			// Determine attack direction based on facing
			switch p.Facing {
			case DirUp:
				p.AttackDir = AttackUp
			case DirDown:
				p.AttackDir = AttackDown
			case DirLeft:
				p.AttackDir = AttackLeft
			case DirRight:
				p.AttackDir = AttackRight
			}
		}
	}

	// Update attack animation
	if p.Attacking {
		p.AttackFrame--
		if p.AttackFrame <= 0 {
			p.Attacking = false
			p.AttackDir = AttackNone
		}
	}

	// Movement
	if !p.Attacking {
		dx, dy := 0.0, 0.0

		if ebiten.IsKeyPressed(config.GameKeys.Up) {
			dy = -p.Speed
			p.Facing = DirUp
		}
		if ebiten.IsKeyPressed(config.GameKeys.Down) {
			dy = p.Speed
			p.Facing = DirDown
		}
		if ebiten.IsKeyPressed(config.GameKeys.Left) {
			dx = -p.Speed
			p.Facing = DirLeft
		}
		if ebiten.IsKeyPressed(config.GameKeys.Right) {
			dx = p.Speed
			p.Facing = DirRight
		}

		// Normalize diagonal movement
		if dx != 0 && dy != 0 {
			dx *= 0.707
			dy *= 0.707
		}

		// Try movement with collision
		newX := p.X + dx
		newY := p.Y + dy

		tileSize := float64(config.TileSize)
		margin := 4.0

		// Check X movement
		if dungeon.IsWalkable(int((newX+margin)/tileSize), int((p.Y+margin)/tileSize)) &&
			dungeon.IsWalkable(int((newX+p.Width-margin)/tileSize), int((p.Y+margin)/tileSize)) &&
			dungeon.IsWalkable(int((newX+margin)/tileSize), int((p.Y+p.Height-margin)/tileSize)) &&
			dungeon.IsWalkable(int((newX+p.Width-margin)/tileSize), int((p.Y+p.Height-margin)/tileSize)) {
			p.X = newX
		}

		// Check Y movement
		if dungeon.IsWalkable(int((p.X+margin)/tileSize), int((newY+margin)/tileSize)) &&
			dungeon.IsWalkable(int((p.X+p.Width-margin)/tileSize), int((newY+margin)/tileSize)) &&
			dungeon.IsWalkable(int((p.X+margin)/tileSize), int((newY+p.Height-margin)/tileSize)) &&
			dungeon.IsWalkable(int((p.X+p.Width-margin)/tileSize), int((newY+p.Height-margin)/tileSize)) {
			p.Y = newY
		}

		// Update walk animation
		if dx != 0 || dy != 0 {
			p.WalkTimer++
			if p.WalkTimer%8 == 0 {
				p.WalkFrame = (p.WalkFrame + 1) % 2
			}
		} else {
			p.WalkFrame = 0
			p.WalkTimer = 0
		}
	}

	return nil
}

func (p *Player) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if p.Image == nil {
		// Fallback: draw colored rectangle
		
		// Flash when hit
		c := color.RGBA{0, 150, 255, 255}
		if p.HitFlash > 0 && p.HitFlash%4 < 2 {
			c = color.RGBA{255, 0, 0, 255}
		}
		
		helper.DrawRect(screen, p.X-offsetX, p.Y-offsetY, p.Width, p.Height, c)
		return
	}

	op := &ebiten.DrawImageOptions{}
	
	// Invincibility flash
	if p.Invincible > 0 && p.Invincible%6 < 3 {
		op.ColorScale.SetA(0.5)
	}
	
	op.GeoM.Translate(p.X-offsetX, p.Y-offsetY)
	screen.DrawImage(p.Image, op)

	// Draw attack effect
	if p.Attacking && p.AttackFrame > 0 {
		var ax, ay float64
		attackSize := 20.0
		
		switch p.AttackDir {
		case AttackUp:
			ax = p.X + p.Width/2 - attackSize/2
			ay = p.Y - attackSize
		case AttackDown:
			ax = p.X + p.Width/2 - attackSize/2
			ay = p.Y + p.Height
		case AttackLeft:
			ax = p.X - attackSize
			ay = p.Y + p.Height/2 - attackSize/2
		case AttackRight:
			ax = p.X + p.Width
			ay = p.Y + p.Height/2 - attackSize/2
		}
		
		helper.DrawRect(screen, ax-offsetX, ay-offsetY, attackSize, attackSize, color.RGBA{255, 200, 0, 180})
	}
}

// Heal restores HP
func (p *Player) Heal(amount int) {
	p.HP += amount
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
}

// TakeDamage reduces HP and sets invincibility
func (p *Player) TakeDamage(amount int) {
	if p.Invincible > 0 {
		return
	}
	p.HP -= amount
	p.Invincible = config.InvincibleTime
	p.HitFlash = 20
}

// IsDead returns true if HP <= 0
func (p *Player) IsDead() bool {
	return p.HP <= 0
}

// GetAttackBox returns the attack hitbox
func (p *Player) GetAttackBox() (float64, float64, float64, float64) {
	size := 24.0
	switch p.AttackDir {
	case AttackUp:
		return p.X + p.Width/2 - size/2, p.Y - size, size, size
	case AttackDown:
		return p.X + p.Width/2 - size/2, p.Y + p.Height, size, size
	case AttackLeft:
		return p.X - size, p.Y + p.Height/2 - size/2, size, size
	case AttackRight:
		return p.X + p.Width, p.Y + p.Height/2 - size/2, size, size
	}
	return 0, 0, 0, 0
}
