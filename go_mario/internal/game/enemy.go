package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// EnemyType - тип врага
type EnemyType int

const (
	// Slimes
	SlimeGreen EnemyType = iota
	SlimeBlue
	SlimePurple
	SlimeBlock

	// Fish
	FishBlue
	FishGreen
	FishPink

	// Worms
	WormGreen
	WormPink

	// Others
	Snail
	Mouse
	Frog
	Fly
	Ladybug
	Saw
	SawHalf
	Barnacle
	Bee

	// Alien (player-like)
	Alien
)

// Enemy - враг
type Enemy struct {
	x           float64
	y           float64
	vx          float64
	vy          float64
	width       float32
	height      float32
	onGround    bool
	alive       bool
	animFrame   int
	enemyType   EnemyType
	color       string // For alien variants
	facing      int
	hp          int
	maxHP       int
	damage      int
	score       int
}

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType EnemyType) *Enemy {
	e := &Enemy{
		x:         x,
		y:         y,
		alive:     true,
		enemyType: enemyType,
		facing:    1,
		hp:        1,
		damage:    10,
	}

	// Set size and HP based on type
	switch enemyType {
	case SlimeGreen, SlimeBlue, SlimePurple:
		e.width, e.height = 30, 25
		e.hp = 1
	case SlimeBlock:
		e.width, e.height = 40, 40
		e.hp = 3
	case FishBlue, FishGreen, FishPink:
		e.width, e.height = 35, 20
		e.hp = 1
	case WormGreen, WormPink:
		e.width, e.height = 40, 15
		e.hp = 2
	case Snail:
		e.width, e.height = 35, 25
		e.hp = 2
	case Mouse, Frog:
		e.width, e.height = 30, 25
		e.hp = 1
	case Fly, Ladybug:
		e.width, e.height = 25, 20
		e.hp = 1
		e.onGround = false // Flying
	case Saw, SawHalf:
		e.width, e.height = 40, 40
		e.hp = 999 // Invincible
		e.damage = 20
	case Barnacle, Bee:
		e.width, e.height = 30, 30
		e.hp = 1
	case Alien:
		e.width, e.height = 40, 50
		e.hp = 3
		e.damage = 15
		e.color = "Blue"
	}

	e.maxHP = e.hp
	e.score = e.hp * 50

	return e
}

// Update обновляет врага
func (e *Enemy) Update(player *Player, worldWidth int) {
	if !e.alive {
		return
	}

	e.animFrame++

	// Different behavior for different enemy types
	switch e.enemyType {
	case SlimeGreen, SlimeBlue, SlimePurple, SlimeBlock:
		e.updateSlime()
	case FishBlue, FishGreen, FishPink:
		e.updateFish()
	case WormGreen, WormPink:
		e.updateWorm()
	case Saw, SawHalf:
		e.updateSaw()
	case Fly, Ladybug:
		e.updateFlying()
	case Alien:
		e.updateAlien(player)
	default:
		e.updateBasic()
	}

	// World bounds
	if e.x < 0 {
		e.x = 0
		e.vx = -e.vx
	}
	if e.x > float64(worldWidth)-float64(e.width) {
		e.x = float64(worldWidth) - float64(e.width)
		e.vx = -e.vx
	}
}

// updateSlime - логика слизня
func (e *Enemy) updateSlime() {
	if e.animFrame%60 == 0 {
		e.vx = float64(e.facing) * 0.5
	}
	if e.animFrame%120 == 0 {
		e.facing = -e.facing
	}
	e.x += e.vx
}

// updateFish - логика рыбы
func (e *Enemy) updateFish() {
	e.x += float64(e.facing) * 1.5
	if e.animFrame%180 == 0 {
		e.facing = -e.facing
	}
}

// updateWorm - логика червя
func (e *Enemy) updateWorm() {
	if e.animFrame%90 == 0 {
		e.facing = -e.facing
	}
	e.x += float64(e.facing) * 0.8
}

// updateSaw - логика пилы
func (e *Enemy) updateSaw() {
	// Saws move in patterns
	e.x += math.Sin(float64(e.animFrame)*0.05) * 2
}

// updateFlying - логика летающих врагов
func (e *Enemy) updateFlying() {
	e.x += float64(e.facing) * 1.0
	e.y += math.Sin(float64(e.animFrame)*0.1) * 0.5
	if e.animFrame%200 == 0 {
		e.facing = -e.facing
	}
}

// updateAlien - логика инопланетянина (как игрок)
func (e *Enemy) updateAlien(player *Player) {
	// Simple AI: move towards player
	if player.x > e.x {
		e.facing = 1
		e.x += 0.3
	} else {
		e.facing = -1
		e.x -= 0.3
	}
}

// updateBasic - базовая логика
func (e *Enemy) updateBasic() {
	e.x += float64(e.facing) * 0.5
	if e.animFrame%100 == 0 {
		e.facing = -e.facing
	}
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64, frames []*ebiten.Image) {
	if !e.alive {
		return
	}

	x := float32(e.x - cameraX)
	y := float32(e.y - cameraY)

	// Off-screen check
	if x < -50 || x > screenWidth+50 || y < -50 || y > screenHeight+50 {
		return
	}

	// Use sprite if available
	if len(frames) > 0 && frames[0] != nil {
		frame := e.getFrame(frames)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		if e.facing < 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(e.width), 0)
		}
		screen.DrawImage(frame, op)
		return
	}

	// Fallback to procedural
	e.drawProcedural(screen, x, y)
}

// getFrame возвращает текущий кадр анимации
func (e *Enemy) getFrame(frames []*ebiten.Image) *ebiten.Image {
	if len(frames) == 0 {
		return nil
	}

	// Cycle through frames
	frameIdx := (e.animFrame / 10) % len(frames)
	return frames[frameIdx]
}

// drawProcedural рисует врага процедурно
func (e *Enemy) drawProcedural(screen *ebiten.Image, x, y float32) {
	var c color.RGBA

	switch e.enemyType {
	case SlimeGreen:
		c = color.RGBA{50, 200, 50, 255}
	case SlimeBlue:
		c = color.RGBA{50, 50, 200, 255}
	case SlimePurple:
		c = color.RGBA{150, 50, 200, 255}
	case FishBlue:
		c = color.RGBA{50, 100, 200, 255}
	case WormGreen:
		c = color.RGBA{50, 180, 50, 255}
	case Saw:
		c = color.RGBA{150, 150, 150, 255}
	default:
		c = color.RGBA{150, 50, 50, 255}
	}

	// Draw enemy body
	vector.DrawFilledRect(screen, x, y, e.width, e.height, c, false)

	// Eyes
	eyeX := x + e.width/2
	eyeY := y + e.height/3
	vector.DrawFilledCircle(screen, eyeX-float32(e.facing)*5, eyeY, 5, color.RGBA{255, 255, 255, 255})
	vector.DrawFilledCircle(screen, eyeX-float32(e.facing)*5, eyeY, 3, color.RGBA{0, 0, 0, 255})
}

// TakeDamage получает урон
func (e *Enemy) TakeDamage(amount int) {
	e.hp -= amount
	if e.hp <= 0 {
		e.alive = false
	}
}

// IsDead проверяет, мёртв ли враг
func (e *Enemy) IsDead() bool {
	return !e.alive
}

// GetScore возвращает очки за врага
func (e *Enemy) GetScore() int {
	return e.score
}

// CreateEnemies создаёт набор врагов для уровня
func CreateEnemies(count int, worldWidth int) []*Enemy {
	enemies := make([]*Enemy, 0, count)

	enemyTypes := []EnemyType{
		SlimeGreen, SlimeBlue, SlimePurple,
		FishBlue, FishGreen,
		WormGreen, WormPink,
		Snail, Mouse, Frog,
		Saw,
		Alien,
	}

	for i := 0; i < count; i++ {
		x := float64(rand.Intn(worldWidth - 100) + 50)
		y := float64(screenHeight - 150 - rand.Intn(200))

		enemyType := enemyTypes[rand.Intn(len(enemyTypes))]
		enemy := NewEnemy(x, y, enemyType)

		enemies = append(enemies, enemy)
	}

	return enemies
}
