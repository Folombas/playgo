// Package entity - компоненты и сущности для Cyber City Runner
// Go365 Day 92 - Киберпанк-платформер
package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Transform - компонент позиции и размера
type Transform struct {
	X, Y     float64
	Width    float64
	Height   float64
	VX, VY   float64
	Facing   int // 1 = вправо, -1 = влево
}

// NewTransform создаёт новый компонент Transform
func NewTransform(x, y, w, h float64) *Transform {
	return &Transform{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
		Facing: 1,
	}
}

// Center возвращает центр сущности
func (t *Transform) Center() (float64, float64) {
	return t.X + t.Width/2, t.Y + t.Height/2
}

// Physics - компонент физики
type Physics struct {
	VelocityX    float64
	VelocityY    float64
	Acceleration float64
	Friction     float64
	Gravity      float64
	OnGround     bool
	OnWall       bool // На стене (для стена-рана)
	IsMoving     bool
}

// NewPhysics создаёт компонент физики
func NewPhysics() *Physics {
	return &Physics{
		Acceleration: 800,
		Friction:     0.82,
		Gravity:      1200,
	}
}

// Health - компонент здоровья
type Health struct {
	Current    int
	Max        int
	Invincible float64
	Dead       bool
}

// NewHealth создаёт компонент здоровья
func NewHealth(max int) *Health {
	return &Health{
		Current: max,
		Max:     max,
	}
}

// TakeDamage получает урон
func (h *Health) TakeDamage(amount int) {
	if h.Invincible > 0 || h.Dead {
		return
	}
	h.Current -= amount
	if h.Current <= 0 {
		h.Current = 0
		h.Dead = true
	}
	h.Invincible = 1.5
}

// Heal лечит
func (h *Health) Heal(amount int) {
	h.Current += amount
	if h.Current > h.Max {
		h.Current = h.Max
	}
}

// Energy - энергия для способностей
type Energy struct {
	Current float64
	Max     float64
	Regen   float64
}

// NewEnergy создаёт компонент энергии
func NewEnergy(max float64) *Energy {
	return &Energy{
		Current: max,
		Max:     max,
		Regen:   15,
	}
}

// UseEnergy тратит энергию
func (e *Energy) UseEnergy(amount float64) bool {
	if e.Current >= amount {
		e.Current -= amount
		return true
	}
	return false
}

// Update обновляет регенерацию
func (e *Energy) Update(dt float64) {
	if e.Current < e.Max {
		e.Current += e.Regen * dt
		if e.Current > e.Max {
			e.Current = e.Max
		}
	}
}

// Stealth - компонент стелса
type Stealth struct {
	Visibility   float64 // 0-1, насколько виден
	Noise        float64 // 0-1, уровень шума
	InShadow     bool    // В тени (скрыт)
	DetectionRad float64 // Радиус обнаружения
}

// NewStealth создаёт компонент стелса
func NewStealth() *Stealth {
	return &Stealth{
		Visibility:   0,
		Noise:        0,
		InShadow:     false,
		DetectionRad: 200,
	}
}

// Update обновляет стелс
func (s *Stealth) Update(dt float64) {
	// Шум затухает
	s.Noise -= dt * 0.5
	if s.Noise < 0 {
		s.Noise = 0
	}

	// Видимость зависит от освещения
	if s.InShadow {
		s.Visibility -= dt * 0.3
	} else {
		s.Visibility += dt * 0.2
	}

	if s.Visibility < 0 {
		s.Visibility = 0
	}
	if s.Visibility > 1 {
		s.Visibility = 1
	}
}

// Hacker - компонент хакерства
type Hacker struct {
	Level       int     // Уровень взлома
	HackPower   float64 // Сила взлома
	HackTimer   float64 // Таймер взлома
	CanHack     bool    // Может ли взламывать
}

// NewHacker создаёт компонент хакерства
func NewHacker(level int) *Hacker {
	return &Hacker{
		Level:     level,
		HackPower: float64(level) * 10,
		CanHack:   true,
	}
}

// StartHack начинает взлом
func (h *Hacker) StartHack() {
	h.CanHack = false
	h.HackTimer = 2.0 // 2 секунды на взлом
}

// Update обновляет взлом
func (h *Hacker) Update(dt float64) {
	if !h.CanHack {
		h.HackTimer -= dt
		if h.HackTimer <= 0 {
			h.CanHack = true
		}
	}
}

// Player - главный герой KAI
type Player struct {
	Transform *Transform
	Physics   *Physics
	Health    *Health
	Energy    *Energy
	Stealth   *Stealth
	Hacker    *Hacker

	State        PlayerState
	JumpCount    int
	MaxJumps     int
	CanDash      bool
	DashCooldown float64
	WallSlide    bool

	Armor      int
	MaxArmor   int
	DataChunks int // Собранные данные

	AnimTimer  float64
}

// PlayerState - состояние игрока
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerWallSlide
	PlayerDashing
	PlayerHacking
	PlayerHurt
	PlayerDead
)

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64) *Player {
	return &Player{
		Transform: NewTransform(x, y, 32, 48),
		Physics:   NewPhysics(),
		Health:    NewHealth(100),
		Energy:    NewEnergy(100),
		Stealth:   NewStealth(),
		Hacker:    NewHacker(1),
		State:     PlayerIdle,
		MaxJumps:  2,
		MaxArmor:  100,
	}
}

// Update обновляет игрока
func (p *Player) Update(dt float64) {
	p.Physics.VelocityX *= p.Physics.Friction
	p.Physics.VelocityY += p.Physics.Gravity * dt

	if p.Physics.VelocityY > 1000 {
		p.Physics.VelocityY = 1000
	}

	p.Transform.X += p.Physics.VelocityX * dt
	p.Transform.Y += p.Physics.VelocityY * dt

	p.Physics.IsMoving = math.Abs(p.Physics.VelocityX) > 10

	p.Energy.Update(dt)
	p.Stealth.Update(dt)
	p.Hacker.Update(dt)

	if p.Health.Invincible > 0 {
		p.Health.Invincible -= dt
	}

	if p.DashCooldown > 0 {
		p.DashCooldown -= dt
	}

	p.AnimTimer += dt

	p.updateState()
}

// updateState обновляет состояние
func (p *Player) updateState() {
	if p.Health.Dead {
		p.State = PlayerDead
		return
	}

	if p.Hacker.HackTimer > 0 && !p.Hacker.CanHack {
		p.State = PlayerHacking
		return
	}

	if !p.Physics.OnGround && p.Physics.OnWall && p.Physics.VelocityY > 0 {
		p.State = PlayerWallSlide
		p.WallSlide = true
	} else {
		p.WallSlide = false
	}

	if !p.Physics.OnGround {
		p.State = PlayerJumping
	} else if p.Physics.IsMoving {
		p.State = PlayerRunning
	} else {
		p.State = PlayerIdle
	}
}

// MoveLeft движется влево
func (p *Player) MoveLeft() {
	p.Physics.VelocityX = -300
	p.Transform.Facing = -1
}

// MoveRight движется вправо
func (p *Player) MoveRight() {
	p.Physics.VelocityX = 300
	p.Transform.Facing = 1
}

// Jump прыгает
func (p *Player) Jump() bool {
	if p.JumpCount < p.MaxJumps {
		jumpForce := 550.0
		if p.JumpCount == 1 {
			jumpForce = 450
		}
		p.Physics.VelocityY = -jumpForce
		p.JumpCount++
		p.Physics.OnGround = false
		p.Stealth.Noise += 0.3 // Шум от прыжка
		return true
	}
	return false
}

// ResetJump сбрасывает счётчик прыжков
func (p *Player) ResetJump() {
	p.JumpCount = 0
}

// Dash делает рывок
func (p *Player) Dash() bool {
	if p.DashCooldown > 0 || p.Energy.Current < 20 {
		return false
	}

	p.Energy.UseEnergy(20)
	p.DashCooldown = 0.5

	dashForce := 600.0
	p.Physics.VelocityX = float64(p.Transform.Facing) * dashForce
	p.Physics.VelocityY = -100
	p.State = PlayerDashing
	p.Stealth.Noise += 0.2

	return true
}

// WallRun отталкивается от стены
func (p *Player) WallRun() {
	if p.Physics.OnWall && p.Physics.VelocityY > 0 {
		p.Physics.VelocityY = -400
		p.Physics.VelocityX = -float64(p.Transform.Facing) * 400
		p.Stealth.Noise += 0.1
	}
}

// CanShoot может ли стрелять
func (p *Player) CanShoot() bool {
	return true // Энерго-пистолет бесконечный
}

// Shoot стреляет
func (p *Player) Shoot() {
	p.Stealth.Noise += 0.15 // Шум выстрела
}

// Hack начинает взлом
func (p *Player) Hack() bool {
	if p.Hacker.CanHack {
		p.Hacker.StartHack()
		return true
	}
	return false
}

// UseEMP использует EMP-импульс
func (p *Player) UseEMP() bool {
	if p.Energy.Current >= 40 {
		p.Energy.UseEnergy(40)
		p.Stealth.Noise += 0.5 // Сильный шум
		return true
	}
	return false
}

// ThrowGrenade бросает гранату
func (p *Player) ThrowGrenade() bool {
	p.Stealth.Noise += 0.4 // Шум броска
	return true
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if p.Health.Dead {
		return
	}

	// Мигание при неуязвимости
	if p.Health.Invincible > 0 && int(p.Health.Invincible*10)%2 == 0 {
		return
	}

	x := p.Transform.X - cameraX
	y := p.Transform.Y - cameraY

	// Тело (неоновый голубой)
	bodyColor := color.RGBA{0, 255, 255, 255}
	if p.State == PlayerDashing {
		bodyColor = color.RGBA{100, 255, 255, 200}
	}

	vector.DrawFilledRect(screen, float32(x), float32(y), float32(p.Transform.Width), float32(p.Transform.Height), bodyColor, false)

	// Глаза/визор (неоновый розовый)
	eyeColor := color.RGBA{255, 0, 255, 255}
	eyeX := x + 8
	if p.Transform.Facing == -1 {
		eyeX = x + 4
	}
	vector.DrawFilledRect(screen, float32(eyeX), float32(y+8), 12, 6, eyeColor, false)

	// Эффект рывка
	if p.State == PlayerDashing {
		for i := 0; i < 3; i++ {
			alpha := uint8(150 - i*40)
			trailX := x - float64(i)*8*float64(p.Transform.Facing)
			vector.DrawFilledRect(screen, float32(trailX), float32(y), float32(p.Transform.Width), float32(p.Transform.Height), color.RGBA{0, 255, 255, alpha}, false)
		}
	}

	// Индикатор в тени
	if p.Stealth.InShadow {
		vector.StrokeRect(screen, float32(x-2), float32(y-2), float32(p.Transform.Width+4), float32(p.Transform.Height+4), 1, color.RGBA{100, 100, 255, 150}, false)
	}
}

// Enemy - враг
type Enemy struct {
	Transform *Transform
	Physics   *Physics
	Health    *Health

	EnemyType string
	Damage    int
	Speed     float64

	PatrolStart float64
	PatrolEnd   float64
	Behavior    int

	AlertLevel   float64 // Уровень тревоги
	DetectionRad float64 // Радиус обнаружения
	AnimTimer    float64
}

const (
	EnemyPatrol = iota
	EnemyChase
	EnemyAttack
	EnemyFlying
	EnemyTurret
)

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType string) *Enemy {
	e := &Enemy{
		Transform: NewTransform(x, y, 32, 48),
		Physics:   NewPhysics(),
		Health:    NewHealth(50),
		EnemyType: enemyType,
		Damage:    15,
		Speed:     80,
		Behavior:  EnemyPatrol,
		DetectionRad: 250,
	}

	// Настройка по типу
	switch enemyType {
	case "soldier":
		e.Health.Max = 60
		e.Speed = 100
		e.Behavior = EnemyPatrol
	case "drone":
		e.Health.Max = 30
		e.Speed = 120
		e.Behavior = EnemyFlying
		e.Transform.Height = 24
	case "turret":
		e.Health.Max = 80
		e.Damage = 25
		e.Behavior = EnemyTurret
		e.Transform.Width = 40
		e.Transform.Height = 32
	case "robot":
		e.Health.Max = 100
		e.Speed = 150
		e.Damage = 20
		e.Behavior = EnemyChase
	case "elite":
		e.Health.Max = 150
		e.Speed = 90
		e.Damage = 30
		e.Behavior = EnemyAttack
	}

	e.Health.Current = e.Health.Max
	if e.PatrolStart == 0 {
		e.PatrolStart = x - 150
		e.PatrolEnd = x + 150
	}

	return e
}

// Update обновляет врага
func (e *Enemy) Update(dt float64, playerX, playerY float64, playerVisible bool) {
	e.AnimTimer += dt

	// Проверка обнаружения игрока
	distX := playerX - e.Transform.X
	distY := playerY - e.Transform.Y
	distance := math.Sqrt(distX*distX + distY*distY)

	if playerVisible && distance < e.DetectionRad {
		e.AlertLevel += dt * 0.5
	} else {
		e.AlertLevel -= dt * 0.3
	}
	if e.AlertLevel < 0 {
		e.AlertLevel = 0
	}
	if e.AlertLevel > 1 {
		e.AlertLevel = 1
	}

	// ИИ в зависимости от поведения
	if !e.Health.Dead {
		e.updateAI(dt, playerX, playerY, distance, playerVisible)
	}

	if e.Health.Invincible > 0 {
		e.Health.Invincible -= dt
	}
}

// updateAI обновляет ИИ
func (e *Enemy) updateAI(dt float64, playerX, playerY, distance float64, playerVisible bool) {
	distX := playerX - e.Transform.X

	if distX > 0 {
		e.Transform.Facing = 1
	} else {
		e.Transform.Facing = -1
	}

	switch e.Behavior {
	case EnemyPatrol:
		if e.AlertLevel > 0.5 {
			// Преследование
			if distX > 0 {
				e.Transform.VX = e.Speed
			} else {
				e.Transform.VX = -e.Speed
			}
		} else {
			// Патрулирование
			if e.Transform.X <= e.PatrolStart {
				e.Transform.VX = e.Speed
			} else if e.Transform.X >= e.PatrolEnd {
				e.Transform.VX = -e.Speed
			} else {
				e.Transform.VX = e.Speed * float64(e.Transform.Facing)
			}
		}

	case EnemyChase:
		if distance < 400 || e.AlertLevel > 0.3 {
			if distX > 0 {
				e.Transform.VX = e.Speed
			} else {
				e.Transform.VX = -e.Speed
			}
		}

	case EnemyFlying:
		// Дрон летает вокруг
		e.Transform.X += math.Sin(e.AnimTimer*2) * 50 * dt
		if distance < 300 && e.AlertLevel > 0.3 {
			if distX > 0 {
				e.Transform.VX = e.Speed * 0.8
			} else {
				e.Transform.VX = -e.Speed * 0.8
			}
		}

	case EnemyTurret:
		// Турель стоит на месте
		e.Transform.VX = 0

	case EnemyAttack:
		if distance < 500 {
			if distX > 0 {
				e.Transform.VX = e.Speed * 0.9
			} else {
				e.Transform.VX = -e.Speed * 0.9
			}
		}
	}

	e.Transform.X += e.Transform.VX * dt
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if e.Health.Dead {
		x := e.Transform.X - cameraX
		y := e.Transform.Y - cameraY
		// Крест на мёртвом враге
		vector.StrokeLine(screen, float32(x), float32(y), float32(x+e.Transform.Width), float32(y+e.Transform.Height), 2, color.RGBA{255, 0, 0, 255}, false)
		vector.StrokeLine(screen, float32(x+e.Transform.Width), float32(y), float32(x), float32(y+e.Transform.Height), 2, color.RGBA{255, 0, 0, 255}, false)
		return
	}

	// Мигание при уроне
	if e.Health.Invincible > 0 && int(e.Health.Invincible*10)%2 == 0 {
		return
	}

	x := e.Transform.X - cameraX
	y := e.Transform.Y - cameraY

	// Цвет врага по типу
	enemyColor := color.RGBA{255, 0, 255, 255} // Розовый по умолчанию
	switch e.EnemyType {
	case "soldier":
		enemyColor = color.RGBA{200, 50, 50, 255} // Красный солдат
	case "drone":
		enemyColor = color.RGBA{150, 50, 200, 255} // Фиолетовый дрон
	case "turret":
		enemyColor = color.RGBA{100, 100, 100, 255} // Серая турель
	case "robot":
		enemyColor = color.RGBA{255, 100, 0, 255} // Оранжевый робот
	case "elite":
		enemyColor = color.RGBA{255, 255, 0, 255} // Жёлтый элитный
	}

	// Тело
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(e.Transform.Width), float32(e.Transform.Height), enemyColor, false)

	// Глаза/сенсоры
	eyeColor := color.RGBA{255, 0, 0, 255}
	if e.AlertLevel > 0.7 {
		eyeColor = color.RGBA{255, 255, 0, 255} // Жёлтый когда встревожен
	}
	eyeX := x + 6
	if e.Transform.Facing == -1 {
		eyeX = x + e.Transform.Width - 14
	}
	vector.DrawFilledRect(screen, float32(eyeX), float32(y+8), 8, 6, eyeColor, false)

	// Индикатор тревоги над головой
	if e.AlertLevel > 0.3 {
		alertWidth := float32(e.Transform.Width) * float32(e.AlertLevel)
		vector.DrawFilledRect(screen, float32(x), float32(y-6), alertWidth, 4, color.RGBA{255, 255, 0, 255}, false)
	}

	// Полоска здоровья
	if e.Health.Current < e.Health.Max {
		hpPercent := float32(e.Health.Current) / float32(e.Health.Max)
		vector.DrawFilledRect(screen, float32(x), float32(y-10), float32(e.Transform.Width), 3, color.RGBA{255, 0, 0, 255}, false)
		vector.DrawFilledRect(screen, float32(x), float32(y-10), float32(e.Transform.Width)*hpPercent, 3, color.RGBA{0, 255, 0, 255}, false)
	}
}

// Item - предмет
type Item struct {
	Transform *Transform
	ItemType  string
	Value     int
	Collected bool
	AnimTimer float64
}

const (
	ItemStimpack   = "stimpack"
	ItemEnergy     = "energy"
	ItemArmor      = "armor"
	ItemData       = "data"
	ItemKeycard    = "keycard"
	ItemGrenade    = "grenade"
	ItemEMP        = "emp"
)

// NewItem создаёт новый предмет
func NewItem(x, y float64, itemType string, value int) *Item {
	width, height := 24.0, 24.0
	if itemType == ItemKeycard {
		width, height = 28, 16
	}

	return &Item{
		Transform: NewTransform(x, y, width, height),
		ItemType:  itemType,
		Value:     value,
	}
}

// Update обновляет предмет
func (i *Item) Update(dt float64) {
	i.AnimTimer += dt
}

// Draw отрисовывает предмет
func (i *Item) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if i.Collected {
		return
	}

	x := i.Transform.X - cameraX
	y := i.Transform.Y - cameraY + math.Sin(i.AnimTimer*3)*5

	var itemColor color.Color
	switch i.ItemType {
	case ItemStimpack:
		itemColor = color.RGBA{255, 50, 50, 255} // Красный стимпак
	case ItemEnergy:
		itemColor = color.RGBA{0, 255, 255, 255} // Голубая энергия
	case ItemArmor:
		itemColor = color.RGBA{100, 100, 100, 255} // Серая броня
	case ItemData:
		itemColor = color.RGBA{0, 255, 0, 255} // Зелёные данные
	case ItemKeycard:
		itemColor = color.RGBA{255, 255, 0, 255} // Жёлтая ключ-карта
	case ItemGrenade:
		itemColor = color.RGBA{255, 165, 0, 255} // Оранжевая граната
	case ItemEMP:
		itemColor = color.RGBA{150, 50, 255, 255} // Фиолетовый EMP
	}

	// Рисуем предмет
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(i.Transform.Width), float32(i.Transform.Height), itemColor, false)

	// Неоновая обводка
	vector.StrokeRect(screen, float32(x-1), float32(y-1), float32(i.Transform.Width+2), float32(i.Transform.Height+2), 1, itemColor, false)
}

// Projectile - снаряд
type Projectile struct {
	Transform *Transform
	VelocityX float64
	VelocityY float64
	LifeTime  float64
	Damage    int
	Active    bool
	Friendly  bool
}

// NewProjectile создаёт новый снаряд
func NewProjectile(x, y, vx, vy float64, damage int, friendly bool) *Projectile {
	return &Projectile{
		Transform: NewTransform(x, y, 16, 8),
		VelocityX: vx,
		VelocityY: vy,
		LifeTime:  2.0,
		Damage:    damage,
		Friendly:  friendly,
		Active:    true,
	}
}

// Update обновляет снаряд
func (p *Projectile) Update(dt float64) {
	p.Transform.X += p.VelocityX * dt
	p.Transform.Y += p.VelocityY * dt
	p.LifeTime -= dt

	if p.LifeTime <= 0 {
		p.Active = false
	}
}

// Draw отрисовывает снаряд
func (p *Projectile) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if !p.Active {
		return
	}

	x := p.Transform.X - cameraX
	y := p.Transform.Y - cameraY

	var projColor color.Color
	if p.Friendly {
		projColor = color.RGBA{0, 255, 255, 255} // Голубой (игрок)
	} else {
		projColor = color.RGBA{255, 0, 0, 255} // Красный (враг)
	}

	vector.DrawFilledRect(screen, float32(x), float32(y), float32(p.Transform.Width), float32(p.Transform.Height), projColor, false)

	// След
	for i := 0; i < 3; i++ {
		alpha := uint8(150 - i*40)
		trailX := x - float64(i)*6
		vector.DrawFilledRect(screen, float32(trailX), float32(y), float32(p.Transform.Width), float32(p.Transform.Height), color.RGBA{projColor.(color.RGBA).R, projColor.(color.RGBA).G, projColor.(color.RGBA).B, alpha}, false)
	}
}

// Particle - частица
type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    float64
	MaxLife float64
	Color   color.Color
	Size    float64
}

// Update обновляет частицу
func (p *Particle) Update(dt float64) {
	p.X += p.VX * dt
	p.Y += p.VY * dt
	p.VY += 200 * dt // Гравитация
	p.Life -= dt
}

// Draw отрисовывает частицу
func (p *Particle) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	x := p.X - cameraX
	y := p.Y - cameraY
	alpha := uint8(p.Life / p.MaxLife * 255)

	if c, ok := p.Color.(color.RGBA); ok {
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(p.Size), float32(p.Size), color.RGBA{c.R, c.G, c.B, alpha}, false)
	}
}

// CheckCollision проверяет коллизию AABB
func CheckCollision(a, b *Transform) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}
