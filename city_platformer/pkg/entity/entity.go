// Package entity - игровые сущности
// Go365 Day 88 - переписано с нуля
package entity

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Player - игровой персонаж
type Player struct {
	X, Y         float64
	Width        float64
	Height       float64
	VX, VY       float64
	Speed        float64
	JumpForce    float64
	OnGround     bool
	Facing       int // 1 = вправо, -1 = влево
	AnimFrame    float64
	AnimTimer    float64
	Invincible   int   // Кадры неуязвимости
	IsCrouching  bool
	IsMoving     bool
	ShootCooldown int
}

// NewPlayer - создание нового игрока
func NewPlayer(x, y float64) *Player {
	return &Player{
		X: x, Y: y, Width: 32, Height: 32,
		Speed: 6.0, JumpForce: -14.0,
		Facing: 1,
	}
}

// Update - обновление состояния игрока
func (p *Player) Update() {
	// Анимация
	if p.IsMoving {
		p.AnimTimer += 0.2
		if p.AnimTimer > 4 {
			p.AnimTimer = 0
		}
		p.AnimFrame = math.Floor(p.AnimTimer)
	} else {
		p.AnimFrame = 0
	}
	
	// Неуязвимость
	if p.Invincible > 0 {
		p.Invincible--
	}
	
	// Кулдаун стрельбы
	if p.ShootCooldown > 0 {
		p.ShootCooldown--
	}
	
	// Сброс флага движения
	p.IsMoving = false
}

// MoveLeft - движение влево
func (p *Player) MoveLeft() {
	p.VX = -p.Speed
	p.Facing = -1
	p.IsMoving = true
}

// MoveRight - движение вправо
func (p *Player) MoveRight() {
	p.VX = p.Speed
	p.Facing = 1
	p.IsMoving = true
}

// Jump - прыжок
func (p *Player) Jump() {
	p.VY = p.JumpForce
	p.OnGround = false
}

// CanJump - можно ли прыгать
func (p *Player) CanJump() bool {
	return p.OnGround && !p.IsCrouching
}

// Crouch - присесть
func (p *Player) Crouch() {
	p.IsCrouching = true
	p.Height = 20
	p.VX = 0
}

// Stand - встать
func (p *Player) Stand() {
	p.IsCrouching = false
	p.Height = 32
}

// Shoot - выстрел
func (p *Player) Shoot() {
	p.ShootCooldown = 15 // 15 кадров между выстрелами
}

// CanShoot - можно ли стрелять
func (p *Player) CanShoot() bool {
	return p.ShootCooldown <= 0
}

// Draw - отрисовка игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	// Мигание если неуязвим
	if p.Invincible > 0 && (p.Invincible%6 < 3) {
		return
	}
	
	// Временная отрисовка (пока без спрайтов)
	if p.IsCrouching {
		// Приседание
		ebitenutil.DrawRect(screen, screenX, screenY+12, 32, 20, color.RGBA{50, 150, 255, 255})
	} else {
		// Стоя/бег
		ebitenutil.DrawRect(screen, screenX, screenY, 32, 32, color.RGBA{50, 150, 255, 255})
	}
	
	// Направление взгляда (глаза)
	eyeX := screenX + 20
	if p.Facing == -1 {
		eyeX = screenX + 8
	}
	ebitenutil.DrawRect(screen, eyeX, screenY+8, 6, 6, color.RGBA{255, 255, 255, 255})
	
	// Оружие
	gunX := screenX + 20
	if p.Facing == -1 {
		gunX = screenX - 4
	}
	ebitenutil.DrawRect(screen, gunX, screenY+18, 16, 6, color.RGBA{80, 80, 80, 255})
}

// Enemy - враг
type Enemy struct {
	X, Y       float64
	Width      float64
	Height     float64
	VX, VY     float64
	Type       string // "slime", "fly", "snail", "fish"
	AnimFrame  float64
	AnimTimer  float64
	Health     int
}

// NewEnemy - создание врага
func NewEnemy(x, y float64, enemyType string) *Enemy {
	e := &Enemy{
		X: x, Y: y, Width: 32, Height: 32,
		Type: enemyType,
		Health: 1,
	}
	
	// Настройка параметров по типу
	switch enemyType {
	case "fly":
		e.Y = y - 50 // Летает выше
		e.Height = 24
	case "snail":
		e.Width = 36
		e.Height = 28
	case "fish":
		e.Height = 20
	}
	
	return e
}

// Update - обновление врага
func (e *Enemy) Update() {
	e.AnimTimer += 0.15
	if e.AnimTimer > 4 {
		e.AnimTimer = 0
	}
	e.AnimFrame = math.Floor(e.AnimTimer)
}

// Draw - отрисовка врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := e.X - cameraX
	screenY := e.Y - cameraY
	
	var bodyColor color.RGBA
	
	// Цвет по типу врага
	switch e.Type {
	case "slime":
		bodyColor = color.RGBA{50, 200, 50, 255}
	case "fly":
		bodyColor = color.RGBA{200, 200, 50, 255}
	case "snail":
		bodyColor = color.RGBA{200, 100, 50, 255}
	case "fish":
		bodyColor = color.RGBA{50, 100, 200, 255}
	default:
		bodyColor = color.RGBA{150, 150, 150, 255}
	}
	
	// Тело врага
	ebitenutil.DrawRect(screen, screenX, screenY, e.Width, e.Height, bodyColor)
	
	// Глаза
	eyeY := screenY + 8
	ebitenutil.DrawRect(screen, screenX+8, eyeY, 6, 6, color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(screen, screenX+e.Width-14, eyeY, 6, 6, color.RGBA{255, 255, 255, 255})
}

// Boss - босс
type Boss struct {
	X, Y        float64
	Width       float64
	Height      float64
	VX, VY      float64
	Health      int
	MaxHealth   int
	AnimFrame   float64
	AnimTimer   float64
	AttackPhase int
}

// NewBoss - создание босса
func NewBoss(x, y float64) *Boss {
	return &Boss{
		X: x, Y: y, Width: 80, Height: 60,
		Health: 50,
		MaxHealth: 50,
	}
}

// Update - обновление босса
func (b *Boss) Update() {
	b.AnimTimer += 0.1
	if b.AnimTimer > 8 {
		b.AnimTimer = 0
	}
	b.AnimFrame = math.Floor(b.AnimTimer / 2)
	
	// Лёгкое парение
	b.Y += math.Sin(float64(time.Now().UnixNano())/3e8) * 0.5
}

// TakeDamage - получение урона
func (b *Boss) TakeDamage(damage int) {
	b.Health -= damage
	if b.Health < 0 {
		b.Health = 0
	}
}

// Draw - отрисовка босса
func (b *Boss) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := b.X - cameraX
	screenY := b.Y - cameraY
	
	// Тело босса (красный огромный враг)
	ebitenutil.DrawRect(screen, screenX, screenY, b.Width, b.Height, color.RGBA{200, 30, 30, 255})
	
	// Броня/панцирь
	ebitenutil.DrawRect(screen, screenX+10, screenY+10, b.Width-20, b.Height-30, color.RGBA{100, 100, 100, 255})
	
	// Глаза (огромные)
	ebitenutil.DrawRect(screen, screenX+20, screenY+15, 15, 15, color.RGBA{255, 150, 0, 255})
	ebitenutil.DrawRect(screen, screenX+45, screenY+15, 15, 15, color.RGBA{255, 150, 0, 255})
	
	// Зрачки
	ebitenutil.DrawRect(screen, screenX+25, screenY+20, 6, 6, color.RGBA{0, 0, 0, 255})
	ebitenutil.DrawRect(screen, screenX+50, screenY+20, 6, 6, color.RGBA{0, 0, 0, 255})
}

// Projectile - снаряд
type Projectile struct {
	X, Y    float64
	VX, VY  float64
	Width   float64
	Height  float64
	Life    int
	Active  bool
	IsEnemy bool // true = вражеский снаряд
}
