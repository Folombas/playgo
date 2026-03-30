// Package entity - игровые сущности с компонентной архитектурой
// Go365 Day 90 - City Survivor
package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"city_platformer/internal/sprite"
)

// Component - базовый интерфейс компонента
type Component interface {
	Update(dt float64)
}

// Transform - компонент позиции и размера
type Transform struct {
	X, Y     float64
	Width    float64
	Height   float64
	VX, VY   float64
	Rotation float64
	ScaleX   float64
	ScaleY   float64
	Facing   int // 1 = вправо, -1 = влево
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

// Update обновляет позицию
func (t *Transform) Update(dt float64) {
	t.X += t.VX * dt
	t.Y += t.VY * dt
}

// Center возвращает центр сущности
func (t *Transform) Center() (float64, float64) {
	return t.X + t.Width/2, t.Y + t.Height/2
}

// Bounds возвращает прямоугольник коллизии
func (t *Transform) Bounds() (float64, float64, float64, float64) {
	return t.X, t.Y, t.Width, t.Height
}

// SpriteRenderer - компонент отрисовки спрайта
type SpriteRenderer struct {
	SpriteSheet *sprite.SpriteSheet
	CurrentImg  *ebiten.Image
	AnimName    string
	Anim        *sprite.Animation
	AnimFrame   int
	AnimTimer   float64
	AnimPlaying bool
	Color       *ebiten.ColorM // Цветовой модификатор
	ScaleX      float64
	ScaleY      float64
}

// NewSpriteRenderer создаёт рендерер спрайтов
func NewSpriteRenderer(ss *sprite.SpriteSheet) *SpriteRenderer {
	return &SpriteRenderer{
		SpriteSheet: ss,
		ScaleX:      1,
		ScaleY:      1,
	}
}

// SetAnim устанавливает анимацию
func (sr *SpriteRenderer) SetAnim(anim *sprite.Animation) {
	sr.Anim = anim
	sr.AnimPlaying = true
	sr.AnimFrame = 0
	sr.AnimTimer = 0
	if len(anim.Frames) > 0 {
		sr.CurrentImg = anim.Frames[0]
	}
}

// SetSprite устанавливает статичный спрайт
func (sr *SpriteRenderer) SetSprite(img *ebiten.Image) {
	sr.AnimPlaying = false
	sr.CurrentImg = img
}

// Update обновляет анимацию
func (sr *SpriteRenderer) Update(dt float64) {
	if !sr.AnimPlaying || sr.Anim == nil {
		return
	}

	sr.AnimTimer += dt
	if sr.AnimTimer >= sr.Anim.FrameTime {
		sr.AnimTimer = 0
		sr.AnimFrame++
		if sr.AnimFrame >= len(sr.Anim.Frames) {
			if sr.Anim.Loop {
				sr.AnimFrame = 0
			} else {
				sr.AnimFrame = len(sr.Anim.Frames) - 1
				sr.AnimPlaying = false
			}
		}
		sr.CurrentImg = sr.Anim.Frames[sr.AnimFrame]
	}
}

// Draw отрисовывает спрайт
func (sr *SpriteRenderer) Draw(screen *ebiten.Image, transform *Transform, cameraX, cameraY float64) {
	if sr.CurrentImg == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{}

	// Масштабирование
	opts.GeoM.Scale(transform.ScaleX*sr.ScaleX, transform.ScaleY*sr.ScaleY)

	// Отражение по горизонтали
	if transform.Facing == -1 {
		opts.GeoM.Scale(-1, 1)
		opts.GeoM.Translate(float64(transform.Width), 0)
	}

	// Позиция - спрайт рисуется от ног (bottom anchor)
	// Сдвигаем вверх на высоту хитбокса чтобы ноги были на Y позиции
	screenX := transform.X - cameraX
	screenY := transform.Y - cameraY - transform.Height
	opts.GeoM.Translate(screenX, screenY)

	// Цвет
	if sr.Color != nil {
		opts.ColorM = *sr.Color
	}

	screen.DrawImage(sr.CurrentImg, opts)
}

// Animator - компонент управления анимациями
type Animator struct {
	Animations  map[string]*sprite.Animation
	CurrentAnim string
	BlendTime   float64
	BlendTimer  float64
}

// NewAnimator создаёт аниматор
func NewAnimator() *Animator {
	return &Animator{
		Animations: make(map[string]*sprite.Animation),
	}
}

// AddAnim добавляет анимацию
func (a *Animator) AddAnim(name string, anim *sprite.Animation) {
	a.Animations[name] = anim
}

// Play запускает анимацию
func (a *Animator) Play(name string) {
	if _, ok := a.Animations[name]; ok {
		a.CurrentAnim = name
	}
}

// GetCurrentAnim возвращает текущую анимацию
func (a *Animator) GetCurrentAnim() *sprite.Animation {
	return a.Animations[a.CurrentAnim]
}

// Physics - компонент физики
type Physics struct {
	VelocityX   float64
	VelocityY   float64
	Acceleration float64
	Friction    float64
	Gravity     float64
	OnGround    bool
	IsMoving    bool
}

// NewPhysics создаёт компонент физики
func NewPhysics() *Physics {
	return &Physics{
		Acceleration: 500,
		Friction:     0.85,
		Gravity:      800,  // Уменьшенная гравитация (было 1500)
	}
}

// Update обновляет физику
func (p *Physics) Update(dt float64, transform *Transform) {
	// Гравитация
	p.VelocityY += p.Gravity * dt

	// Трение
	if !p.IsMoving {
		p.VelocityX *= p.Friction
	}

	// Применяем скорость
	transform.VX = p.VelocityX
	transform.VY = p.VelocityY

	p.IsMoving = false
}

// ApplyJump применяет силу прыжка
func (p *Physics) ApplyJump(force float64) {
	p.VelocityY = -force
	p.OnGround = false
}

// Health - компонент здоровья
type Health struct {
	Current    int
	Max        int
	Invincible float64 // Таймер неуязвимости
	Dead       bool
}

// NewHealth создаёт компонент здоровья
func NewHealth(max int) *Health {
	return &Health{
		Current: max,
		Max:     max,
	}
}

// Update обновляет таймеры
func (h *Health) Update(dt float64) {
	if h.Invincible > 0 {
		h.Invincible -= dt
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
	h.Invincible = 2.0 // 2 секунды неуязвимости
}

// Heal лечит
func (h *Health) Heal(amount int) {
	h.Current += amount
	if h.Current > h.Max {
		h.Current = h.Max
	}
}

// IsAlive проверяет жив ли объект
func (h *Health) IsAlive() bool {
	return !h.Dead
}

// Player - сущность игрока
type Player struct {
	Transform   *Transform
	Renderer    *SpriteRenderer
	Physics     *Physics
	Health      *Health
	Animator    *Animator
	State       PlayerState
	Ammo        int
	MaxAmmo     int
	Speed       float64
	JumpForce   float64
	ShootCooldown float64
}

// PlayerState - состояние игрока
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerCrouching
	PlayerShooting
	PlayerHurt
)

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64, ss *sprite.SpriteSheet) *Player {
	// Размер хитбокса для Элисы - воительницы
	p := &Player{
		Transform: NewTransform(x, y, 40, 50), // Хитбокс Элисы
		Physics:   NewPhysics(),
		Health:    NewHealth(100),
		Animator:  NewAnimator(),
		State:     PlayerIdle,
		Ammo:      30,
		MaxAmmo:   50,
		Speed:     300,
		JumpForce: 550,
	}

	p.Renderer = NewSpriteRenderer(ss)
	// Спрайты Элисы не масштабируем - они нужного размера
	p.Renderer.ScaleX = 1.0
	p.Renderer.ScaleY = 1.0

	// Загрузка анимаций
	if walkAnim := ss.GetPlayerAnim("walk"); walkAnim != nil {
		p.Animator.AddAnim("walk", walkAnim)
	}
	if hurtAnim := ss.GetPlayerAnim("hurt"); hurtAnim != nil {
		p.Animator.AddAnim("hurt", hurtAnim)
	}
	if attackAnim := ss.GetPlayerAnim("attack"); attackAnim != nil {
		p.Animator.AddAnim("attack", attackAnim)
	}

	// Установка начального спрайта
	if standSprite := ss.GetPlayerSprite("stand"); standSprite != nil {
		p.Renderer.SetSprite(standSprite)
	}

	return p
}

// Update обновляет игрока
func (p *Player) Update(dt float64) {
	p.Transform.Update(dt)
	p.Physics.Update(dt, p.Transform)
	p.Health.Update(dt)
	p.Renderer.Update(dt)

	// Обновление анимации
	p.updateAnimation()

	// Перезарядка оружия
	if p.ShootCooldown > 0 {
		p.ShootCooldown -= dt
	}
}

// updateAnimation обновляет анимацию игрока
func (p *Player) updateAnimation() {
	// Анимация смерти
	if p.Health.Dead {
		if deathAnim := p.Renderer.SpriteSheet.GetPlayerAnim("death"); deathAnim != nil {
			p.Renderer.SetAnim(deathAnim)
		}
		return
	}

	// Анимация получения урона
	if p.Health.Invincible > 0 && p.Health.Invincible < 1.8 {
		if hurtAnim := p.Renderer.SpriteSheet.GetPlayerAnim("hurt"); hurtAnim != nil {
			p.Renderer.SetAnim(hurtAnim)
		}
		return
	}

	if p.Health.Invincible > 0 && int(p.Health.Invincible*10)%2 == 0 {
		return // Мигание при неуязвимости
	}

	if p.State == PlayerJumping {
		if jumpSprite := p.Renderer.SpriteSheet.GetPlayerSprite("jump"); jumpSprite != nil {
			p.Renderer.SetSprite(jumpSprite)
		}
	} else if p.State == PlayerCrouching {
		if duckSprite := p.Renderer.SpriteSheet.GetPlayerSprite("duck"); duckSprite != nil {
			p.Renderer.SetSprite(duckSprite)
		}
	} else if p.State == PlayerRunning && p.Physics.OnGround {
		if walkAnim := p.Animator.GetCurrentAnim(); walkAnim != nil {
			p.Renderer.SetAnim(walkAnim)
		}
	} else if p.State == PlayerIdle {
		if standSprite := p.Renderer.SpriteSheet.GetPlayerSprite("stand"); standSprite != nil {
			p.Renderer.SetSprite(standSprite)
		}
	}
}

// MoveLeft движется влево
func (p *Player) MoveLeft() {
	p.Physics.VelocityX = -p.Speed
	p.Transform.Facing = -1
	p.Physics.IsMoving = true
	if p.State != PlayerJumping && p.State != PlayerCrouching {
		p.State = PlayerRunning
	}
}

// MoveRight движется вправо
func (p *Player) MoveRight() {
	p.Physics.VelocityX = p.Speed
	p.Transform.Facing = 1
	p.Physics.IsMoving = true
	if p.State != PlayerJumping && p.State != PlayerCrouching {
		p.State = PlayerRunning
	}
}

// Jump прыгает
func (p *Player) Jump() {
	if p.Physics.OnGround && p.State != PlayerCrouching {
		p.Physics.ApplyJump(p.JumpForce)
		p.State = PlayerJumping
	}
}

// Crouch приседает
func (p *Player) Crouch() {
	p.State = PlayerCrouching
	p.Transform.Height = 40
}

// Stand встаёт
func (p *Player) Stand() {
	p.State = PlayerIdle
	p.Transform.Height = 60
}

// ClimbUp лазает вверх по лестнице
func (p *Player) ClimbUp() {
	p.Physics.VelocityY = -150
	p.Physics.OnGround = true // На лестнице считаем что на земле
	p.State = PlayerRunning
}

// ClimbDown лазает вниз по лестнице
func (p *Player) ClimbDown() {
	p.Physics.VelocityY = 150
	p.Physics.OnGround = true
	p.State = PlayerCrouching
}

// CanShoot может ли стрелять
func (p *Player) CanShoot() bool {
	return p.ShootCooldown <= 0 && p.Ammo > 0 && p.Health.IsAlive()
}

// Shoot стреляет
func (p *Player) Shoot() {
	p.ShootCooldown = 0.3
	p.Ammo--
	p.State = PlayerShooting
}

// Reload перезаряжается
func (p *Player) Reload() {
	p.Ammo = p.MaxAmmo
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	p.Renderer.Draw(screen, p.Transform, cameraX, cameraY)
}

// Enemy - сущность врага
type Enemy struct {
	Transform    *Transform
	Renderer     *SpriteRenderer
	Physics      *Physics
	Health       *Health
	Animator     *Animator
	EnemyType    string
	Damage       int
	Speed        float64
	AttackRange  float64
	ShootCooldown float64
	Behavior     EnemyBehavior
}

// EnemyBehavior - тип поведения врага
type EnemyBehavior int

const (
	EnemyMelee EnemyBehavior = iota
	EnemyRanged
	EnemyFlying
	EnemyPatrol
)

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType string, ss *sprite.SpriteSheet) *Enemy {
	e := &Enemy{
		Transform: NewTransform(x, y, 40, 40),
		Physics:   NewPhysics(),
		Health:    NewHealth(30),
		Animator:  NewAnimator(),
		EnemyType: enemyType,
		Damage:    10,
		Speed:     80,
		Behavior:  EnemyMelee,
	}

	e.Renderer = NewSpriteRenderer(ss)

	// Настройка параметров по типу врага
	switch enemyType {
	case "fish":
		e.Behavior = EnemyFlying
		e.Health.Max = 20
		e.Health.Current = 20
		e.Speed = 100
		if swimAnim := ss.GetEnemyAnim("fishSwim"); swimAnim != nil {
			e.Animator.AddAnim("swim", swimAnim)
			e.Renderer.SetAnim(swimAnim)
		}
	case "fly":
		e.Behavior = EnemyFlying
		e.Health.Max = 15
		e.Health.Current = 15
		e.Speed = 120
		if flyAnim := ss.GetEnemyAnim("flyFly"); flyAnim != nil {
			e.Animator.AddAnim("fly", flyAnim)
			e.Renderer.SetAnim(flyAnim)
		}
	case "slime":
		e.Behavior = EnemyMelee
		e.Health.Max = 25
		e.Health.Current = 25
		e.Damage = 15
		e.Speed = 50
		if walkAnim := ss.GetEnemyAnim("slimeWalk"); walkAnim != nil {
			e.Animator.AddAnim("walk", walkAnim)
			e.Renderer.SetAnim(walkAnim)
		}
	case "snail":
		e.Behavior = EnemyMelee
		e.Health.Max = 40
		e.Health.Current = 40
		e.Damage = 20
		e.Speed = 30
		if walkAnim := ss.GetEnemyAnim("snailWalk"); walkAnim != nil {
			e.Animator.AddAnim("walk", walkAnim)
			e.Renderer.SetAnim(walkAnim)
		}
	case "blocker":
		e.Behavior = EnemyPatrol
		e.Health.Max = 50
		e.Health.Current = 50
		e.Damage = 25
		e.Speed = 40
		e.Transform.Width = 50
		e.Transform.Height = 50
	default:
		// По умолчанию - слайм
		if walkAnim := ss.GetEnemyAnim("slimeWalk"); walkAnim != nil {
			e.Animator.AddAnim("walk", walkAnim)
			e.Renderer.SetAnim(walkAnim)
		}
	}

	return e
}

// Update обновляет врага
func (e *Enemy) Update(dt float64, playerX, playerY float64) {
	e.Transform.Update(dt)
	e.Physics.Update(dt, e.Transform)
	e.Health.Update(dt)
	e.Renderer.Update(dt)

	// Перезарядка
	if e.ShootCooldown > 0 {
		e.ShootCooldown -= dt
	}
}

// AI обновляет ИИ врага
func (e *Enemy) AI(dt float64, playerX, playerY float64) {
	if !e.Health.IsAlive() {
		return
	}

	distX := playerX - e.Transform.X
	distY := playerY - e.Transform.Y
	distance := math.Sqrt(distX*distX + distY*distY)

	// Определение направления
	if distX > 0 {
		e.Transform.Facing = 1
	} else {
		e.Transform.Facing = -1
	}

	switch e.Behavior {
	case EnemyMelee:
		// Преследование игрока
		if distance < 400 {
			if distX > 0 {
				e.Physics.VelocityX = e.Speed
			} else {
				e.Physics.VelocityX = -e.Speed
			}
			e.Physics.IsMoving = true
		}

	case EnemyFlying:
		// Летающие враги медленно двигаются к игроку
		if distance < 500 {
			if distX > 0 {
				e.Physics.VelocityX = e.Speed * 0.5
			} else {
				e.Physics.VelocityX = -e.Speed * 0.5
			}
			e.Physics.IsMoving = true
		}

	case EnemyPatrol:
		// Патрулирование
		e.Physics.VelocityX = e.Speed * float64(e.Transform.Facing)
		e.Physics.IsMoving = true
	}
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	e.Renderer.Draw(screen, e.Transform, cameraX, cameraY)
}

// Projectile - снаряд
type Projectile struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	VelocityX float64
	VelocityY float64
	LifeTime  float64
	IsEnemy   bool
	Damage    int
	Active    bool
}

// NewProjectile создаёт новый снаряд
func NewProjectile(x, y, vx, vy float64, isEnemy bool, damage int, ss *sprite.SpriteSheet) *Projectile {
	p := &Projectile{
		Transform: NewTransform(x, y, 16, 8),
		VelocityX: vx,
		VelocityY: vy,
		LifeTime:  2.0,
		IsEnemy:   isEnemy,
		Damage:    damage,
		Active:    true,
	}

	p.Renderer = NewSpriteRenderer(ss)

	// Спрайт пули
	if bulletSprite := ss.GetItemSprite("coinGold"); bulletSprite != nil {
		// Используем монету как заглушку для пули
		p.Renderer.SetSprite(bulletSprite)
	}

	return p
}

// Update обновляет снаряд
func (p *Projectile) Update(dt float64) {
	p.Transform.X += p.VelocityX * dt
	p.Transform.Y += p.VelocityY * dt
	p.VelocityY += 200 * dt // Гравитация
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
	p.Renderer.Draw(screen, p.Transform, cameraX, cameraY)
}

// Item - предмет
type Item struct {
	Transform  *Transform
	Renderer   *SpriteRenderer
	ItemType   string
	Value      int
	Collected  bool
	AnimTimer  float64
	FloatOffset float64
}

// NewItem создаёт новый предмет
func NewItem(x, y float64, itemType string, value int, ss *sprite.SpriteSheet) *Item {
	i := &Item{
		Transform: NewTransform(x, y, 32, 32),
		ItemType:  itemType,
		Value:     value,
	}

	i.Renderer = NewSpriteRenderer(ss)

	// Спрайт предмета
	if itemSprite := ss.GetItemSprite(itemType); itemSprite != nil {
		i.Renderer.SetSprite(itemSprite)
	}

	return i
}

// Update обновляет предмет
func (i *Item) Update(dt float64) {
	i.AnimTimer += dt
	i.FloatOffset = math.Sin(i.AnimTimer*3) * 5
}

// Draw отрисовывает предмет
func (i *Item) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if i.Collected {
		return
	}

	// Сохраняем оригинальную Y
	origY := i.Transform.Y
	i.Transform.Y += i.FloatOffset
	i.Renderer.Draw(screen, i.Transform, cameraX, cameraY)
	i.Transform.Y = origY
}

// CheckCollision проверяет коллизию AABB
func CheckCollision(a, b *Transform) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}
