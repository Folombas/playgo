package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// PlayerStats - характеристики игрока
type PlayerStats struct {
	level      int
	experience int
	maxExp     int
	strength   int
	defense    int
	vitality   int
	agility    int
	luck       int
	statPoints int
}

// Equipment - экипировка игрока
type Equipment struct {
	weapon *Item
	armor  *Item
}

// Player - игровой персонаж
type Player struct {
	x            float64
	y            float64
	vy           float64
	width        float32
	height       float32
	onGround     bool
	score        int
	coins        int
	lives        int
	facing       int
	animFrame    int
	invincible   int
	stats        PlayerStats
	equipment    Equipment
	maxHealth    int
	currentHealth int
}

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64) *Player {
	return &Player{
		x:          x,
		y:          y,
		width:      30,
		height:     40,
		onGround:   true,
		lives:      3,
		facing:     1,
		invincible: 0,
		stats: PlayerStats{
			level:      1,
			experience: 0,
			maxExp:     100,
			strength:   5,
			defense:    2,
			vitality:   10,
			agility:    3,
			luck:       1,
		},
		maxHealth:     100,
		currentHealth: 100,
	}
}

// Update обновляет состояние игрока
func (p *Player) Update(input *InputState, worldWidth int, groundLevel float64) {
	// Movement
	if input.Left {
		p.x -= moveSpeed
		p.facing = -1
		p.animFrame++
	}
	if input.Right {
		p.x += moveSpeed
		p.facing = 1
		p.animFrame++
	}

	// Jumping
	if input.Jump && p.onGround {
		p.vy = jumpForce
		p.onGround = false
	}

	// Variable jump height
	if !input.Jump && p.vy < -jumpForce/2 {
		p.vy *= 0.5
	}

	// Gravity
	p.vy += gravity
	if p.vy > 15 {
		p.vy = 15
	}
	p.y += p.vy

	// Ground collision
	if p.y >= groundLevel {
		p.y = groundLevel
		p.vy = 0
		p.onGround = true
	}

	// World bounds
	if p.x < 0 {
		p.x = 0
	}
	if p.x > float64(worldWidth)-float64(p.width) {
		p.x = float64(worldWidth) - float64(p.width)
	}

	// Invincibility timer
	if p.invincible > 0 {
		p.invincible--
	}
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64, playerSprite *Sprite) {
	x := float32(p.x - cameraX)
	y := float32(p.y - cameraY)

	// Off-screen check
	if x < -50 || x > screenWidth+50 || y < -50 || y > screenHeight+50 {
		return
	}

	// Use sprite if available
	if playerSprite != nil && playerSprite.image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		if p.facing < 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(p.width), 0)
		}
		screen.DrawImage(playerSprite.image, op)
		return
	}

	// Fallback to procedural drawing
	p.drawProcedural(screen, x, y)
}

// drawProcedural рисует игрока процедурно (Mario-style)
func (p *Player) drawProcedural(screen *ebiten.Image, x, y float32) {
	legOffset := float32(math.Sin(float64(p.animFrame)*0.5)) * 4
	if !p.onGround {
		legOffset = 2
	}

	dir := p.facing
	if dir == 0 {
		dir = 1
	}

	// Boots (brown)
	bootColor := color.RGBA{101, 67, 33, 255}
	vector.DrawFilledRect(screen, x+8-legOffset, y+p.height-8, 10, 8, bootColor, false)
	vector.DrawFilledRect(screen, x+p.width-18+legOffset, y+p.height-8, 10, 8, bootColor, false)

	// Pants (blue overalls)
	pantsColor := color.RGBA{50, 50, 180, 255}
	vector.DrawFilledRect(screen, x+6, y+p.height-22, p.width-12, 14, pantsColor, false)
	vector.DrawFilledRect(screen, x+8, y+p.height-28, 5, 10, pantsColor, false)
	vector.DrawFilledRect(screen, x+p.width-13, y+p.height-28, 5, 10, pantsColor, false)

	// Shirt (red)
	shirtColor := color.RGBA{200, 30, 30, 255}
	vector.DrawFilledRect(screen, x+5, y+p.height-35, p.width-10, 12, shirtColor, false)
	vector.DrawFilledRect(screen, x+2, y+p.height-32, 6, 8, shirtColor, false)
	vector.DrawFilledRect(screen, x+p.width-8, y+p.height-32, 6, 8, shirtColor, false)

	// Head (skin tone)
	skinColor := color.RGBA{255, 220, 180, 255}
	headX := x + p.width/2
	headY := y + p.height - 42
	vector.DrawFilledCircle(screen, headX, headY, 14, skinColor, false)

	// Hat (red cap)
	hatColor := color.RGBA{200, 30, 30, 255}
	vector.DrawFilledRect(screen, x+6, y+p.height-52, p.width-12, 8, hatColor, false)
	brimX := x + p.width/2 - 18 + float32(dir)*8
	vector.DrawFilledRect(screen, brimX, y+p.height-48, 24, 5, hatColor, false)

	// Eyes
	eyeX := headX + float32(dir)*4
	eyeY := headY + 2
	vector.DrawFilledCircle(screen, eyeX, eyeY, 5, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, eyeX+float32(dir), eyeY, 3, color.RGBA{0, 0, 0, 255}, false)

	// Nose
	noseX := headX + float32(dir)*8
	noseY := headY + 8
	vector.DrawFilledCircle(screen, noseX, noseY, 4, color.RGBA{255, 180, 160, 255}, false)

	// Mustache
	mustacheX := headX + float32(dir)*6
	mustacheY := headY + 12
	vector.DrawFilledRect(screen, mustacheX-6, mustacheY, 12, 4, color.RGBA{50, 30, 20, 255}, false)

	// Gloves (white)
	gloveColor := color.RGBA{255, 255, 255, 255}
	handX := x + p.width/2 + float32(dir)*12
	handY := y + p.height - 25 + legOffset
	vector.DrawFilledCircle(screen, handX, handY, 6, gloveColor, false)
}

// TakeDamage наносит урон игроку
func (p *Player) TakeDamage(amount int) {
	actualDamage := amount - p.stats.defense
	if actualDamage < 1 {
		actualDamage = 1
	}
	p.currentHealth -= actualDamage
	p.invincible = 120
}

// Heal лечит игрока
func (p *Player) Heal(amount int) {
	p.currentHealth += amount
	if p.currentHealth > p.maxHealth {
		p.currentHealth = p.maxHealth
	}
}

// IsDead проверяет, мёртв ли игрок
func (p *Player) IsDead() bool {
	return p.currentHealth <= 0 || p.lives <= 0
}

// Reset сбрасывает игрока
func (p *Player) Reset(x, y float64) {
	p.x = x
	p.y = y
	p.vy = 0
	p.invincible = 60
	p.currentHealth = p.maxHealth
}

// AddExperience добавляет опыт
func (p *Player) AddExperience(amount int) {
	p.stats.experience += amount
	for p.stats.experience >= p.stats.maxExp {
		p.stats.experience -= p.stats.maxExp
		p.stats.level++
		p.stats.maxExp = int(float64(p.stats.maxExp) * 1.5)
		p.stats.statPoints += 3
		p.maxHealth += 10
		p.currentHealth = p.maxHealth
	}
}
