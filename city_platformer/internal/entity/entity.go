// Package entity - игровые сущности City Platformer
// Go365 Day 91 - Постапокалиптический город
package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"city_platformer/internal/sprite"
)

// Player - игрок (выживший)
type Player struct {
	X, Y           float64
	Width          float64
	Height         float64
	VX, VY         float64
	Speed          float64
	JumpForce      float64
	OnGround       bool
	Facing         int // 1 = вправо, -1 = влево
	AnimFrame      float64
	AnimTimer      float64
	Invincible     int
	IsCrouching    bool
	IsMoving       bool
	ShootCooldown  int
	Health         int
	MaxHealth      int
	Ammo           int
	MaxAmmo        int
	State          string // stand, run, jump, crouch, shoot
	spriteSheet    *sprite.SpriteSheet
}

// NewPlayer - создание нового игрока
func NewPlayer(x, y float64, ss *sprite.SpriteSheet) *Player {
	return &Player{
		X: x, Y: y, Width: 32, Height: 48,
		Speed: 5.0, JumpForce: -15.0,
		Facing:    1,
		Health:    100,
		MaxHealth: 100,
		Ammo:      30,
		MaxAmmo:   50,
		State:     "stand",
		spriteSheet: ss,
	}
}

// Update - обновление состояния игрока
func (p *Player) Update() {
	// Анимация
	if p.IsMoving && p.OnGround {
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

	// Перезарядка
	if p.ShootCooldown > 0 {
		p.ShootCooldown--
	}

	// Определение состояния
	if !p.OnGround {
		p.State = "jump"
	} else if p.IsCrouching {
		p.State = "crouch"
	} else if p.IsMoving {
		p.State = "run"
	} else {
		p.State = "stand"
	}

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
	p.Height = 32
}

// Stand - встать
func (p *Player) Stand() {
	p.IsCrouching = false
	p.Height = 48
}

// CanShoot - можно ли стрелять
func (p *Player) CanShoot() bool {
	return p.ShootCooldown <= 0 && p.Ammo > 0
}

// Shoot - выстрел
func (p *Player) Shoot() {
	p.ShootCooldown = 15
	p.Ammo--
	p.State = "shoot"
}

// Reload - перезарядка
func (p *Player) Reload() {
	if p.Ammo < p.MaxAmmo {
		p.Ammo = p.MaxAmmo
	}
}

// TakeDamage - получение урона
func (p *Player) TakeDamage(damage int) {
	if p.Invincible > 0 {
		return
	}
	p.Health -= damage
	p.Invincible = 120 // 2 секунды при 60 FPS
	p.VY = -8
	p.VX = float64(-p.Facing) * 5
}

// Heal - лечение
func (p *Player) Heal(amount int) {
	p.Health += amount
	if p.Health > p.MaxHealth {
		p.Health = p.MaxHealth
	}
}

// AddAmmo - добавить патроны
func (p *Player) AddAmmo(amount int) {
	p.Ammo += amount
	if p.Ammo > p.MaxAmmo {
		p.Ammo = p.MaxAmmo
	}
}

// GetSprite - получение спрайта игрока
func (p *Player) GetSprite() *ebiten.Image {
	frame := int(p.AnimFrame)
	return p.spriteSheet.GetPlayerSprite(p.State, frame)
}

// Draw - отрисовка игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY

	// Мигание если неуязвим
	if p.Invincible > 0 && (p.Invincible%6 < 3) {
		return
	}

	spriteImg := p.GetSprite()
	if spriteImg != nil {
		opts := &ebiten.DrawImageOptions{}

		// Отражение по горизонтали если смотрит влево
		if p.Facing == -1 {
			opts.GeoM.Scale(-1, 1)
			opts.GeoM.Translate(float64(p.Width), 0)
		}

		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(spriteImg, opts)
	}
}

// Enemy - враг
type Enemy struct {
	X, Y        float64
	Width       float64
	Height      float64
	VX, VY      float64
	Speed       float64
	Type        string // mutant, robot, zombie
	Health      int
	MaxHealth   int
	Damage      int
	Facing      int
	AnimFrame   float64
	AnimTimer   float64
	AttackRange float64
	ShootCooldown int
	spriteSheet *sprite.SpriteSheet
}

// NewEnemy - создание врага
func NewEnemy(x, y float64, enemyType string, ss *sprite.SpriteSheet) *Enemy {
	e := &Enemy{
		X: x, Y: y,
		Type:        enemyType,
		Facing:      -1,
		spriteSheet: ss,
	}

	// Настройка параметров по типу
	switch enemyType {
	case "mutant":
		e.Width, e.Height = 40, 40
		e.Health, e.MaxHealth = 30, 30
		e.Damage = 15
		e.Speed = 1.5
		e.AttackRange = 50
	case "robot":
		e.Width, e.Height = 36, 44
		e.Health, e.MaxHealth = 50, 50
		e.Damage = 20
		e.Speed = 1.0
		e.AttackRange = 300
		e.ShootCooldown = 120
	case "zombie":
		e.Width, e.Height = 32, 44
		e.Health, e.MaxHealth = 20, 20
		e.Damage = 10
		e.Speed = 0.8
		e.AttackRange = 40
	}

	return e
}

// Update - обновление врага
func (e *Enemy) Update(playerX, playerY float64) {
	e.AnimTimer += 0.1
	if e.AnimTimer > 4 {
		e.AnimTimer = 0
	}
	e.AnimFrame = math.Floor(e.AnimTimer)

	// Перезарядка
	if e.ShootCooldown > 0 {
		e.ShootCooldown--
	}

	// Определение направления к игроку
	if playerX > e.X {
		e.Facing = 1
	} else {
		e.Facing = -1
	}
}

// GetSprite - получение спрайта врага
func (e *Enemy) GetSprite() *ebiten.Image {
	frame := int(e.AnimFrame)
	return e.spriteSheet.GetEnemySprite(e.Type, frame)
}

// Draw - отрисовка врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := e.X - cameraX
	screenY := e.Y - cameraY

	spriteImg := e.GetSprite()
	if spriteImg != nil {
		opts := &ebiten.DrawImageOptions{}

		if e.Facing == -1 {
			opts.GeoM.Scale(-1, 1)
			opts.GeoM.Translate(float64(e.Width), 0)
		}

		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(spriteImg, opts)
	}
}

// CanShoot - может ли враг стрелять
func (e *Enemy) CanShoot() bool {
	return e.Type == "robot" && e.ShootCooldown <= 0
}

// Shoot - выстрел врага
func (e *Enemy) Shoot() {
	e.ShootCooldown = 120
}

// TakeDamage - получение урона
func (e *Enemy) TakeDamage(damage int) {
	e.Health -= damage
}

// IsAlive - жив ли враг
func (e *Enemy) IsAlive() bool {
	return e.Health > 0
}

// Projectile - снаряд (пуля)
type Projectile struct {
	X, Y      float64
	Width     float64
	Height    float64
	VX, VY    float64
	Life      int
	Active    bool
	IsEnemy   bool
	Damage    int
	spriteSheet *sprite.SpriteSheet
}

// NewProjectile - создание снаряда
func NewProjectile(x, y, vx, vy float64, isEnemy bool, damage int, ss *sprite.SpriteSheet) *Projectile {
	return &Projectile{
		X: x, Y: y,
		VX: vx, VY: vy,
		Width: 12, Height: 6,
		Life:      120,
		Active:    true,
		IsEnemy:   isEnemy,
		Damage:    damage,
		spriteSheet: ss,
	}
}

// Update - обновление снаряда
func (p *Projectile) Update() {
	p.X += p.VX
	p.Y += p.VY
	p.VY += 0.1 // Гравитация
	p.Life--

	if p.Life <= 0 {
		p.Active = false
	}
}

// GetSprite - получение спрайта пули
func (p *Projectile) GetSprite() *ebiten.Image {
	return p.spriteSheet.GetBullet()
}

// Draw - отрисовка снаряда
func (p *Projectile) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if !p.Active {
		return
	}

	screenX := p.X - cameraX
	screenY := p.Y - cameraY

	spriteImg := p.GetSprite()
	if spriteImg != nil {
		opts := &ebiten.DrawImageOptions{}

		// Отражение если летит влево
		if p.VX < 0 {
			opts.GeoM.Scale(-1, 1)
			opts.GeoM.Translate(float64(p.Width), 0)
		}

		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(spriteImg, opts)
	}
}

// Item - предмет
type Item struct {
	X, Y        float64
	Width       float64
	Height      float64
	Type        string // medkit, ammo, food, parts
	Value       int
	Collected   bool
	AnimFrame   float64
	AnimTimer   float64
	spriteSheet *sprite.SpriteSheet
}

// NewItem - создание предмета
func NewItem(x, y float64, itemType string, value int, ss *sprite.SpriteSheet) *Item {
	item := &Item{
		X: x, Y: y,
		Type:        itemType,
		Value:       value,
		spriteSheet: ss,
	}

	// Размеры по типу
	switch itemType {
	case "medkit":
		item.Width, item.Height = 24, 20
	case "ammo":
		item.Width, item.Height = 20, 16
	case "food":
		item.Width, item.Height = 20, 20
	case "parts":
		item.Width, item.Height = 24, 24
	}

	return item
}

// Update - обновление предмета
func (i *Item) Update() {
	i.AnimTimer += 0.1
	if i.AnimTimer > 4 {
		i.AnimTimer = 0
	}
	i.AnimFrame = math.Floor(i.AnimTimer)
}

// GetSprite - получение спрайта предмета
func (i *Item) GetSprite() *ebiten.Image {
	return i.spriteSheet.GetItemSprite(i.Type)
}

// Draw - отрисовка предмета
func (i *Item) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if i.Collected {
		return
	}

	screenX := i.X - cameraX
	screenY := i.Y - cameraY

	// Анимация парения
	offsetY := math.Sin(i.AnimTimer*0.5) * 3

	spriteImg := i.GetSprite()
	if spriteImg != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(screenX, screenY+offsetY)
		screen.DrawImage(spriteImg, opts)
	}
}
