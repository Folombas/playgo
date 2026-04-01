// Package entity - игровые сущности для Cyber City Runner
// Go365 Day 92 - Киберпанк-платформер
package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Transform - компонент позиции и размера
type Transform struct {
	X, Y     float64
	Width    float64
	Height   float64
	VX, VY   float64
	ScaleX   float64
	ScaleY   float64
	Facing   int // 1 = вправо, -1 = влево
	Rotation float64
}

// NewTransform создаёт новый компонент Transform
func NewTransform(x, y, w, h float64) *Transform {
	return &Transform{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
		ScaleX: 1,
		ScaleY: 1,
		Facing: 1,
	}
}

// Center возвращает центр сущности
func (t *Transform) Center() (float64, float64) {
	return t.X+t.Width/2, t.Y+t.Height/2
}

// Bottom возвращает позицию низа
func (t *Transform) Bottom() float64 {
	return t.Y + t.Height
}

// Right возвращает позицию правой границы
func (t *Transform) Right() float64 {
	return t.X + t.Width
}

// Physics - компонент физики
type Physics struct {
	VelocityX    float64
	VelocityY    float64
	Acceleration float64
	Friction     float64
	Gravity      float64
	OnGround     bool
	OnWall       bool
	WallSide     int // 1 = справа, -1 = слева
	IsMoving     bool
	CanDoubleJump bool
}

// NewPhysics создаёт компонент физики
func NewPhysics() *Physics {
	return &Physics{
		Acceleration: 800,
		Friction:     0.82,
		Gravity:      1200,
		CanDoubleJump: true,
	}
}

// Health - компонент здоровья
type Health struct {
	Current    int
	Max        int
	Armor      int
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
	// Броня поглощает 50% урона
	if h.Armor > 0 {
		armorAbsorb := int(float64(h.Armor) * 0.5)
		if armorAbsorb > amount/2 {
			armorAbsorb = amount / 2
		}
		h.Armor -= armorAbsorb
		amount -= armorAbsorb
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

// AddArmor добавляет броню
func (h *Health) AddArmor(amount int) {
	h.Armor += amount
	if h.Armor > 100 {
		h.Armor = 100
	}
}

// Energy - компонент энергии (для рывков и хакерства)
type Energy struct {
	Current  float64
	Max      float64
	Regen    float64
	Drain    float64
}

// NewEnergy создаёт компонент энергии
func NewEnergy(max float64) *Energy {
	return &Energy{
		Current: max,
		Max:     max,
		Regen:   15,
	}
}

// Use тратит энергию
func (e *Energy) Use(amount float64) bool {
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
	InShadow     bool    // в тени
	Alertness    float64 // уровень тревоги врагов
}

// NewStealth создаёт компонент стелса
func NewStealth() *Stealth {
	return &Stealth{
		Visibility: 0,
		Noise:      0,
		InShadow:   false,
	}
}

// Hacker - компонент хакерства
type Hacker struct {
	Level        int     // уровень навыка
	HackSpeed    float64 // скорость взлома
	HackRange    float64 // дальность взлома
	Hacking      bool    // сейчас взламывает
	HackProgress float64 // прогресс взлома
	HackTarget   interface{} // цель взлома
}

// NewHacker создаёт компонент хакерства
func NewHacker() *Hacker {
	return &Hacker{
		Level:     1,
		HackSpeed: 0.5,
		HackRange: 100,
	}
}

// Player - главный герой (KAI)
type Player struct {
	Transform *Transform
	Physics   *Physics
	Health    *Health
	Energy    *Energy
	Stealth   *Stealth
	Hacker    *Hacker
	
	State      PlayerState
	JumpCount  int
	MaxJumps   int
	Dashing    bool
	DashTimer  float64
	WallSliding bool
	
	// Инвентарь
	DataChips  int // собранные данные
	Grenades   int
	EMPCharges int
	
	// Анимация
	AnimFrame   float64
	AnimTimer   float64
}

// PlayerState - состояние игрока
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerFalling
	PlayerWallSlide
	PlayerWallJump
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
		Hacker:    NewHacker(),
		State:     PlayerIdle,
		MaxJumps:  2,
		Grenades:  3,
		EMPCharges: 2,
	}
}

// Update обновляет игрока
func (p *Player) Update(dt float64) {
	p.Physics.VelocityX *= p.Physics.Friction
	p.Physics.VelocityY += p.Physics.Gravity * dt
	
	// Ограничение скорости падения
	if p.Physics.VelocityY > 800 {
		p.Physics.VelocityY = 800
	}
	
	p.Transform.X += p.Physics.VelocityX * dt
	p.Transform.Y += p.Physics.VelocityY * dt
	
	p.Physics.IsMoving = math.Abs(p.Physics.VelocityX) > 10
	
	// Обновление энергии
	p.Energy.Update(dt)
	
	// Обновление неуязвимости
	if p.Health.Invincible > 0 {
		p.Health.Invincible -= dt
	}
	
	// Обновление рывка
	if p.Dashing {
		p.DashTimer -= dt
		if p.DashTimer <= 0 {
			p.Dashing = false
		}
	}
	
	// Обновление анимации
	p.AnimTimer += dt * 10
	p.AnimFrame = math.Mod(p.AnimTimer, 4)
	
	// Определение состояния
	p.updateState()
}

// updateState обновляет состояние
func (p *Player) updateState() {
	if p.Health.Dead {
		p.State = PlayerDead
		return
	}
	
	if p.Dashing {
		p.State = PlayerDashing
		return
	}
	
	if p.WallSliding {
		p.State = PlayerWallSlide
		return
	}
	
	if !p.Physics.OnGround {
		if p.Physics.VelocityY < 0 {
			p.State = PlayerJumping
		} else {
			p.State = PlayerFalling
		}
	} else if p.Physics.IsMoving {
		p.State = PlayerRunning
	} else {
		p.State = PlayerIdle
	}
}

// MoveLeft движется влево
func (p *Player) MoveLeft() {
	if !p.Dashing {
		p.Physics.VelocityX = -300
		p.Transform.Facing = -1
	}
}

// MoveRight движется вправо
func (p *Player) MoveRight() {
	if !p.Dashing {
		p.Physics.VelocityX = 300
		p.Transform.Facing = 1
	}
}

// Jump прыгает
func (p *Player) Jump() bool {
	if p.JumpCount < p.MaxJumps {
		jumpForce := 550.0
		// Двойной прыжок слабее
		if p.JumpCount == 1 {
			jumpForce = 450
		}
		p.Physics.VelocityY = -jumpForce
		p.JumpCount++
		p.Physics.OnGround = false
		return true
	}
	return false
}

// WallJump прыгает от стены
func (p *Player) WallJump() {
	p.JumpCount = 0
	jumpForce := 500.0
	p.Physics.VelocityY = -jumpForce
	// Отталкивание от стены
	p.Physics.VelocityX = float64(-p.Physics.WallSide) * 400
	p.WallSliding = false
}

// Dash выполняет рывок
func (p *Player) Dash() bool {
	if p.Energy.Use(25) && !p.Dashing {
		p.Dashing = true
		p.DashTimer = 0.15
		// Сохраняем направление
		if p.Physics.VelocityX == 0 {
			p.Physics.VelocityX = float64(p.Transform.Facing) * 600
		} else {
			p.Physics.VelocityX *= 2
		}
		p.Physics.VelocityY = 0
		return true
	}
	return false
}

// ResetJump сбрасывает счётчик прыжков
func (p *Player) ResetJump() {
	p.JumpCount = 0
}

// UseGrenade использует гранату
func (p *Player) UseGrenade() bool {
	if p.Grenades > 0 {
		p.Grenades--
		return true
	}
	return false
}

// UseEMP использует EMP-импульс
func (p *Player) UseEMP() bool {
	if p.EMPCharges > 0 {
		p.EMPCharges--
		return true
	}
	return false
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
	
	// Тело (тёмная куртка)
	bodyColor := color.RGBA{30, 30, 50, 255}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(p.Transform.Width), float32(p.Transform.Height), bodyColor, false)
	
	// Неоновая полоска (голубая)
	stripColor := color.RGBA{0, 255, 255, 255}
	stripHeight := float32(4)
	vector.DrawFilledRect(screen, float32(x), float32(y+10), float32(p.Transform.Width), stripHeight, stripColor, false)
	
	// Глаза (визор)
	visorColor := color.RGBA{0, 255, 255, 255}
	visorX := x + 8
	if p.Transform.Facing == -1 {
		visorX = x + 4
	}
	vector.DrawFilledRect(screen, float32(visorX), float32(y+8), 12, 6, visorColor, false)
	
	// Эффект рывка
	if p.Dashing {
		for i := 1; i <= 3; i++ {
			alpha := uint8(255 - i*60)
			trailColor := color.RGBA{0, 255, 255, alpha}
			trailX := x - float64(p.Transform.Facing)*float64(i)*8
			vector.DrawFilledRect(screen, float32(trailX), float32(y), float32(p.Transform.Width), float32(p.Transform.Height), trailColor, false)
		}
	}
	
	// Отладочная информация
	ebitenutil.DebugPrintAt(screen, "KAI", int(x), int(y-15))
}

// Enemy - враг
type Enemy struct {
	Transform *Transform
	Health    *Health
	State     EnemyState
	
	EnemyType  EnemyType
	Behavior   EnemyBehavior
	Damage     int
	
	// ИИ
	PatrolStart  float64
	PatrolEnd    float64
	DetectRange  float64
	AttackRange  float64
	AttackCooldown float64
	MoveSpeed    float64
	
	// Для отрисовки
	AnimTimer float64
	Color     color.Color
}

// EnemyState - состояние врага
type EnemyState int

const (
	EnemyPatrol EnemyState = iota
	EnemyChase
	EnemyAttack
	EnemyAlert
	EnemyHurt
	EnemyDead
)

// EnemyType - тип врага
type EnemyType string

const (
	EnemySoldier EnemyType = "soldier"
	EnemyDrone   EnemyType = "drone"
	EnemyRobot   EnemyType = "robot"
	EnemyElite   EnemyType = "elite"
)

// EnemyBehavior - поведение
type EnemyBehavior int

const (
	BehaviorPatrol EnemyBehavior = iota
	BehaviorChase
	BehaviorStationary
	BehaviorFlying
)

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType EnemyType) *Enemy {
	e := &Enemy{
		Transform: NewTransform(x, y, 32, 48),
		Health:    NewHealth(50),
		EnemyType: enemyType,
		AnimTimer: 0,
	}
	
	// Настройка по типу
	switch enemyType {
	case EnemySoldier:
		e.Behavior = BehaviorPatrol
		e.Damage = 15
		e.DetectRange = 200
		e.AttackRange = 150
		e.MoveSpeed = 80
		e.Color = color.RGBA{100, 100, 150, 255}
		e.Health.Max = 50
	case EnemyDrone:
		e.Behavior = BehaviorFlying
		e.Damage = 10
		e.DetectRange = 150
		e.MoveSpeed = 100
		e.Color = color.RGBA{255, 100, 100, 255}
		e.Health.Max = 30
		e.Transform.Height = 24
	case EnemyRobot:
		e.Behavior = BehaviorChase
		e.Damage = 20
		e.DetectRange = 100
		e.AttackRange = 40
		e.MoveSpeed = 60
		e.Color = color.RGBA{150, 100, 100, 255}
		e.Health.Max = 80
	case EnemyElite:
		e.Behavior = BehaviorPatrol
		e.Damage = 30
		e.DetectRange = 250
		e.AttackRange = 200
		e.MoveSpeed = 90
		e.Color = color.RGBA{200, 100, 200, 255}
		e.Health.Max = 150
	}
	
	e.Health.Current = e.Health.Max
	e.PatrolStart = x - 100
	e.PatrolEnd = x + 100
	
	return e
}

// Update обновляет врага
func (e *Enemy) Update(dt float64, playerX, playerY float64) {
	e.AnimTimer += dt
	
	if e.Health.Dead {
		return
	}
	
	if e.Health.Invincible > 0 {
		e.Health.Invincible -= dt
	}
	
	if e.AttackCooldown > 0 {
		e.AttackCooldown -= dt
	}
	
	// ИИ
	e.updateAI(dt, playerX, playerY)
}

// updateAI обновляет ИИ
func (e *Enemy) updateAI(dt float64, playerX, playerY float64) {
	distX := playerX - e.Transform.X
	distY := playerY - e.Transform.Y
	distance := math.Sqrt(distX*distX + distY*distY)
	
	// Направление
	if distX > 0 {
		e.Transform.Facing = 1
	} else {
		e.Transform.Facing = -1
	}
	
	// Проверка обнаружения
	if distance < e.DetectRange {
		e.State = EnemyChase
	} else {
		e.State = EnemyPatrol
	}
	
	switch e.State {
	case EnemyPatrol:
		// Патрулирование
		if e.Transform.X <= e.PatrolStart {
			e.Transform.VX = e.MoveSpeed
		} else if e.Transform.X >= e.PatrolEnd {
			e.Transform.VX = -e.MoveSpeed
		} else {
			e.Transform.VX = e.MoveSpeed * float64(e.Transform.Facing)
		}
		
	case EnemyChase:
		// Преследование
		if e.Behavior == BehaviorFlying {
			// Дрон летит к игроку
			if distX > 10 {
				e.Transform.VX = e.MoveSpeed * 0.7
			} else if distX < -10 {
				e.Transform.VX = -e.MoveSpeed * 0.7
			} else {
				e.Transform.VX = 0
			}
			
			if distY > 10 {
				e.Transform.VY = e.MoveSpeed * 0.5
			} else if distY < -10 {
				e.Transform.VY = -e.MoveSpeed * 0.5
			}
		} else {
			// Наземный враг
			if distX > 0 {
				e.Transform.VX = e.MoveSpeed
			} else {
				e.Transform.VX = -e.MoveSpeed
			}
		}
		
		// Атака
		if distance < e.AttackRange && e.AttackCooldown <= 0 {
			e.State = EnemyAttack
		}
	}
	
	// Применение скорости
	e.Transform.X += e.Transform.VX * dt
	if e.Behavior == BehaviorFlying {
		e.Transform.Y += e.Transform.VY * dt
	}
}

// TakeDamage получает урон
func (e *Enemy) TakeDamage(amount int) {
	e.Health.TakeDamage(amount)
	e.State = EnemyHurt
	e.Health.Invincible = 0.3
	
	if e.Health.Dead {
		e.State = EnemyDead
	}
}

// Attack атакует игрока
func (e *Enemy) Attack() (int, bool) {
	if e.AttackCooldown <= 0 && e.State == EnemyChase {
		e.AttackCooldown = 1.0
		return e.Damage, true
	}
	return 0, false
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if e.Health.Dead {
		// Крестик вместо мёртвого врага
		x := e.Transform.X - cameraX
		y := e.Transform.Y - cameraY
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
	
	// Тело
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(e.Transform.Width), float32(e.Transform.Height), e.Color, false)
	
	// Глаза/визор
	eyeColor := color.RGBA{255, 0, 0, 255}
	if e.State == EnemyAlert {
		eyeColor = color.RGBA{255, 255, 0, 255}
	}
	eyeX := x + 6
	if e.Transform.Facing == 1 {
		eyeX = x + e.Transform.Width - 14
	}
	vector.DrawFilledRect(screen, float32(eyeX), float32(y+10), 8, 4, eyeColor, false)
	
	// Полоска здоровья
	if e.Health.Current < e.Health.Max {
		hpPercent := float32(e.Health.Current) / float32(e.Health.Max)
		vector.DrawFilledRect(screen, float32(x), float32(y-8), float32(e.Transform.Width), 4, color.RGBA{255, 0, 0, 255}, false)
		vector.DrawFilledRect(screen, float32(x), float32(y-8), float32(e.Transform.Width)*hpPercent, 4, color.RGBA{0, 255, 0, 255}, false)
	}
}

// Terminal - терминал для взлома
type Terminal struct {
	Transform    *Transform
	Hacked       bool
	HackProgress float64
	HackTime     float64
	LinkedDoor   *Door
	Color        color.Color
}

// NewTerminal создаёт терминал
func NewTerminal(x, y float64) *Terminal {
	return &Terminal{
		Transform: NewTransform(x, y, 40, 50),
		Color:     color.RGBA{0, 255, 0, 255},
	}
}

// Hack начинает взлом
func (t *Terminal) Hack(dt float64, hackSpeed float64) bool {
	if t.Hacked {
		return true
	}
	
	t.HackProgress += hackSpeed * dt
	t.HackTime += dt
	
	if t.HackProgress >= 1 {
		t.Hacked = true
		t.Color = color.RGBA{0, 200, 255, 255}
		return true
	}
	return false
}

// Reset сбрасывает прогресс
func (t *Terminal) Reset() {
	if !t.Hacked {
		t.HackProgress = 0
		t.HackTime = 0
	}
}

// Draw отрисовывает терминал
func (t *Terminal) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	x := t.Transform.X - cameraX
	y := t.Transform.Y - cameraY
	
	// Корпус
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(t.Transform.Width), float32(t.Transform.Height), color.RGBA{50, 50, 80, 255}, false)
	
	// Экран
	screenColor := t.Color
	if !t.Hacked {
		// Мигание при взломе
		if int(t.HackProgress*10)%2 == 0 {
			screenColor = color.RGBA{255, 100, 100, 255}
		}
	}
	vector.DrawFilledRect(screen, float32(x+5), float32(y+10), float32(t.Transform.Width-10), float32(t.Transform.Height-20), screenColor, false)
	
	// Прогресс взлома
	if !t.Hacked && t.HackProgress > 0 {
		progressY := float32(y + t.Transform.Height - 15 - t.HackProgress*(float64(t.Transform.Height)-25))
		vector.DrawFilledRect(screen, float32(x+7), progressY, float32(t.Transform.Width-14), float32(t.HackProgress*(float64(t.Transform.Height)-25)), color.RGBA{0, 255, 0, 255}, false)
	}
}

// Door - дверь
type Door struct {
	Transform *Transform
	Open      bool
	Locked    bool
	Hacked    bool
	RequiresHack bool
	Color     color.Color
}

// NewDoor создаёт дверь
func NewDoor(x, y, height float64) *Door {
	return &Door{
		Transform: NewTransform(x, y, 50, height),
		Color:     color.RGBA{150, 150, 150, 255},
		Locked:    true,
		RequiresHack: true,
	}
}

// OpenDoor открывает дверь
func (d *Door) OpenDoor() {
	if !d.Locked || d.Hacked {
		d.Open = true
		d.Color = color.RGBA{100, 200, 100, 255}
	}
}

// Hack взламывает дверь
func (d *Door) Hack() {
	d.Hacked = true
	d.Locked = false
	d.Open = true
	d.Color = color.RGBA{0, 255, 255, 255}
}

// Draw отрисовывает дверь
func (d *Door) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	x := d.Transform.X - cameraX
	y := d.Transform.Y - cameraY
	
	if d.Open {
		// Открытая дверь - рама
		vector.StrokeRect(screen, float32(x), float32(y), float32(d.Transform.Width), float32(d.Transform.Height), 4, color.RGBA{100, 100, 100, 255}, false)
	} else {
		// Закрытая дверь
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(d.Transform.Width), float32(d.Transform.Height), d.Color, false)
		
		// Индикатор замка
		lockColor := color.RGBA{255, 0, 0, 255}
		if d.Hacked {
			lockColor = color.RGBA{0, 255, 255, 255}
		}
		vector.DrawFilledRect(screen, float32(x+d.Transform.Width-15), float32(y+d.Transform.Height/2-5), 8, 10, lockColor, false)
	}
}

// Item - предмет
type Item struct {
	Transform *Transform
	ItemType  string
	Value     int
	Collected bool
	
	// Анимация
	AnimTimer   float64
	FloatOffset float64
	Color       color.Color
}

// ItemType constants
const (
	ItemHealth    = "health"
	ItemEnergy    = "energy"
	ItemArmor     = "armor"
	ItemData      = "data"
	ItemGrenade   = "grenade"
	ItemEMP       = "emp"
)

// NewItem создаёт предмет
func NewItem(x, y float64, itemType string, value int) *Item {
	i := &Item{
		Transform: &Transform{
			X: x, Y: y, Width: 24, Height: 24,
			ScaleX: 1, ScaleY: 1, Facing: 1,
		},
		ItemType:  itemType,
		Value:     value,
	}
	
	// Цвет по типу
	switch itemType {
	case ItemHealth:
		i.Color = color.RGBA{255, 50, 50, 255}
	case ItemEnergy:
		i.Color = color.RGBA{50, 200, 255, 255}
	case ItemArmor:
		i.Color = color.RGBA{100, 100, 150, 255}
	case ItemData:
		i.Color = color.RGBA{0, 255, 0, 255}
	case ItemGrenade:
		i.Color = color.RGBA{255, 150, 50, 255}
	case ItemEMP:
		i.Color = color.RGBA{200, 100, 255, 255}
	}
	
	return i
}

// Update обновляет предмет
func (i *Item) Update(dt float64) {
	i.AnimTimer += dt
	i.FloatOffset = math.Sin(i.AnimTimer*3) * 3
}

// Draw отрисовывает предмет
func (i *Item) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if i.Collected {
		return
	}

	x := i.Transform.X - cameraX
	y := i.Transform.Y - cameraY + i.FloatOffset
	w := i.Transform.Width
	h := i.Transform.Height

	// Рисуем как квадрат с поворотом (ромб)
	centerX := float32(x + w/2)
	centerY := float32(y + h/2)
	
	// Простой ромб через 4 прямоугольника
	halfW := float32(w / 2)
	halfH := float32(h / 2)
	
	// Верхняя часть
	vector.DrawFilledRect(screen, centerX-halfW/1.5, centerY-halfH, halfW/0.75, halfH, i.Color, false)
	// Нижняя часть
	vector.DrawFilledRect(screen, centerX-halfW/1.5, centerY, halfW/0.75, halfH, i.Color, false)
	
	// Контур
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 2, color.RGBA{255, 255, 255, 150}, false)
}

// Camera - камера наблюдения
type Camera struct {
	Transform  *Transform
	DetectRange float64
	DetectionAngle float64 // угол обзора в радианах
	Rotating   bool
	Rotation   float64
	RotationSpeed float64
	Alert      bool
}

// NewCamera создаёт камеру
func NewCamera(x, y float64) *Camera {
	return &Camera{
		Transform:    NewTransform(x, y, 30, 20),
		DetectRange:  250,
		DetectionAngle: math.Pi / 3, // 60 градусов
		Rotating:     true,
		RotationSpeed: 1.5,
	}
}

// Update обновляет камеру
func (c *Camera) Update(dt float64) {
	if c.Rotating {
		c.Rotation += c.RotationSpeed * dt
		if c.Rotation > math.Pi*2 {
			c.Rotation -= math.Pi * 2
		}
	}
}

// CanDetect проверяет, может ли камера обнаружить игрока
func (c *Camera) CanDetect(playerX, playerY float64) bool {
	dx := playerX - (c.Transform.X + c.Transform.Width/2)
	dy := playerY - (c.Transform.Y + c.Transform.Height/2)
	distance := math.Sqrt(dx*dx + dy*dy)
	
	if distance > c.DetectRange {
		return false
	}
	
	angleToPlayer := math.Atan2(dy, dx)
	angleDiff := math.Abs(angleToPlayer - c.Rotation)
	if angleDiff > math.Pi {
		angleDiff = math.Pi*2 - angleDiff
	}
	
	return angleDiff < c.DetectionAngle/2
}

// Draw отрисовывает камеру
func (c *Camera) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	x := c.Transform.X - cameraX
	y := c.Transform.Y - cameraY
	
	// База
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(c.Transform.Width), float32(c.Transform.Height), color.RGBA{100, 100, 100, 255}, false)
	
	// Вращающаяся часть
	centerX := float32(x + c.Transform.Width/2)
	centerY := float32(y + c.Transform.Height/2)
	
	// Конус обзора
	coneColor := color.RGBA{255, 0, 0, 50}
	if c.Alert {
		coneColor = color.RGBA{255, 255, 0, 100}
	}
	
	// Рисуем луч обзора
	rayLength := float32(c.DetectRange)
	endX := centerX + rayLength*float32(math.Cos(c.Rotation))
	endY := centerY + rayLength*float32(math.Sin(c.Rotation))
	vector.StrokeLine(screen, centerX, centerY, endX, endY, 2, coneColor, false)
	
	// Глаз камеры
	eyeColor := color.RGBA{255, 0, 0, 255}
	if c.Alert {
		eyeColor = color.RGBA{255, 255, 0, 255}
	}
	vector.DrawFilledRect(screen, centerX-4, centerY-4, 8, 8, eyeColor, false)
}

// Turret - турель
type Turret struct {
	Transform    *Transform
	DetectRange  float64
	AttackRange  float64
	AttackCooldown float64
	Damage       int
	Rotation     float64
	Active       bool
	Hacked       bool
	Color        color.Color
}

// NewTurret создаёт турель
func NewTurret(x, y float64) *Turret {
	return &Turret{
		Transform:    NewTransform(x, y, 40, 40),
		DetectRange:  300,
		AttackRange:  280,
		AttackCooldown: 0,
		Damage:       20,
		Active:       true,
		Color:        color.RGBA{200, 100, 100, 255},
	}
}

// Update обновляет турель
func (t *Turret) Update(dt float64, playerX, playerY float64) {
	if !t.Active || t.Hacked {
		return
	}
	
	if t.AttackCooldown > 0 {
		t.AttackCooldown -= dt
	}
	
	// Поворот к игроку
	dx := playerX - (t.Transform.X + t.Transform.Width/2)
	dy := playerY - (t.Transform.Y + t.Transform.Height/2)
	distance := math.Sqrt(dx*dx + dy*dy)
	
	if distance < t.DetectRange {
		t.Rotation = math.Atan2(dy, dx)
	}
}

// CanAttack проверяет, может ли турель атаковать
func (t *Turret) CanAttack(playerX, playerY float64) bool {
	if !t.Active || t.Hacked {
		return false
	}
	
	dx := playerX - (t.Transform.X + t.Transform.Width/2)
	dy := playerY - (t.Transform.Y + t.Transform.Height/2)
	distance := math.Sqrt(dx*dx + dy*dy)
	
	return distance < t.AttackRange && t.AttackCooldown <= 0
}

// Attack атакует
func (t *Turret) Attack() int {
	t.AttackCooldown = 0.8
	return t.Damage
}

// Hack взламывает турель
func (t *Turret) Hack() {
	t.Hacked = true
	t.Active = false
	t.Color = color.RGBA{0, 255, 255, 255}
}

// Draw отрисовывает турель
func (t *Turret) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	x := t.Transform.X - cameraX
	y := t.Transform.Y - cameraY
	
	// База
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(t.Transform.Width), float32(t.Transform.Height), t.Color, false)
	
	// Пушка
	centerX := float32(x + t.Transform.Width/2)
	centerY := float32(y + t.Transform.Height/2)
	barrelLength := float32(25)
	endX := centerX + barrelLength*float32(math.Cos(t.Rotation))
	endY := centerY + barrelLength*float32(math.Sin(t.Rotation))
	vector.StrokeLine(screen, centerX, centerY, endX, endY, 6, color.RGBA{50, 50, 50, 255}, false)
	
	// Индикатор
	indicatorColor := color.RGBA{255, 0, 0, 255}
	if t.Hacked {
		indicatorColor = color.RGBA{0, 255, 255, 255}
	} else if !t.Active {
		indicatorColor = color.RGBA{100, 100, 100, 255}
	}
	vector.DrawFilledRect(screen, centerX-3, centerY-3, 6, 6, indicatorColor, false)
}

// CheckCollision проверяет коллизию AABB
func CheckCollision(a, b *Transform) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

// CheckCircleCollision проверяет коллизию круг-прямоугольник
func CheckCircleCollision(cx, cy, radius float64, rect *Transform) bool {
	closestX := cx
	closestY := cy
	
	if cx < rect.X {
		closestX = rect.X
	} else if cx > rect.X+rect.Width {
		closestX = rect.X + rect.Width
	}
	
	if cy < rect.Y {
		closestY = rect.Y
	} else if cy > rect.Y+rect.Height {
		closestY = rect.Y + rect.Height
	}
	
	dx := cx - closestX
	dy := cy - closestY
	
	return (dx*dx + dy*dy) < (radius * radius)
}
