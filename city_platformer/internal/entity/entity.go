// Package entity - игровые сущности с компонентной архитектурой (ECS)
// Go365 Day 91 - Cyber City Runner
package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"cyber_city_runner/internal/sprite"
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
	Color       *ebiten.ColorM
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
	opts.GeoM.Scale(transform.ScaleX*sr.ScaleX, transform.ScaleY*sr.ScaleY)

	if transform.Facing == -1 {
		opts.GeoM.Scale(-1, 1)
		opts.GeoM.Translate(float64(transform.Width), 0)
	}

	screenX := transform.X - cameraX
	screenY := transform.Y - cameraY - transform.Height
	opts.GeoM.Translate(screenX, screenY)

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
	VelocityX    float64
	VelocityY    float64
	Acceleration float64
	Friction     float64
	Gravity      float64
	OnGround     bool
	IsMoving     bool
}

// NewPhysics создаёт компонент физики
func NewPhysics() *Physics {
	return &Physics{
		Acceleration: 800,
		Friction:     0.85,
		Gravity:      1200,
	}
}

// Update обновляет физику
func (p *Physics) Update(dt float64, transform *Transform) {
	p.VelocityY += p.Gravity * dt

	if !p.IsMoving {
		p.VelocityX *= p.Friction
	}

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
	h.Invincible = 2.0
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

// CharacterType - тип персонажа
type CharacterType int

const (
	CharHacker CharacterType = iota // Хакер 👨‍💻
	CharRobot                       // Робот 🤖
	CharNinja                       // Ниндзя 🥷
)

// CharacterConfig - конфигурация персонажа
type CharacterConfig struct {
	Type        CharacterType
	Name        string
	Width       float64
	Height      float64
	Speed       float64
	JumpForce   float64
	MaxHealth   int
	DashCost    float64
	Special     string
	WalkAnim    string
	JumpAnim    string
	RunAnim    string
}

// GetCharacterConfig возвращает конфигурацию для типа персонажа
func GetCharacterConfig(charType CharacterType) *CharacterConfig {
	switch charType {
	case CharHacker:
		return &CharacterConfig{
			Type:        CharHacker,
			Name:        "Хакер",
			Width:       40,
			Height:      56,
			Speed:       300,
			JumpForce:   550,
			MaxHealth:   100,
			DashCost:    20,
			Special:     "Взлом",
			WalkAnim:    "walk",
			JumpAnim:    "jump",
			RunAnim:     "run",
		}
	case CharRobot:
		return &CharacterConfig{
			Type:        CharRobot,
			Name:        "Робот",
			Width:       44,
			Height:      52,
			Speed:       250,
			JumpForce:   480,
			MaxHealth:   150,
			DashCost:    25,
			Special:     "Щит",
			WalkAnim:    "walk",
			JumpAnim:    "jump",
			RunAnim:     "run",
		}
	case CharNinja:
		return &CharacterConfig{
			Type:        CharNinja,
			Name:        "Ниндзя",
			Width:       38,
			Height:      54,
			Speed:       380,
			JumpForce:   620,
			MaxHealth:   80,
			DashCost:    15,
			Special:     "Невидимость",
			WalkAnim:    "walk",
			JumpAnim:    "jump",
			RunAnim:     "run",
		}
	default:
		return GetCharacterConfig(CharHacker)
	}
}

// Player - сущность игрока
type Player struct {
	Transform     *Transform
	Renderer      *SpriteRenderer
	Physics       *Physics
	Health        *Health
	Animator      *Animator
	CharacterType CharacterType
	State         PlayerState
	Energy        float64
	MaxEnergy     float64
	Speed         float64
	JumpForce     float64
	DashCooldown  float64
	Dashing       bool
	DashTimer     float64
	JumpCount     int
	MaxJumps      int
	Invisible     float64
	Score         int
}

// PlayerState - состояние игрока
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerDashing
	PlayerCrouching
	PlayerHurt
)

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64, charType CharacterType, ss *sprite.SpriteSheet) *Player {
	config := GetCharacterConfig(charType)

	p := &Player{
		Transform:     NewTransform(x, y, config.Width, config.Height),
		Physics:       NewPhysics(),
		Health:        NewHealth(config.MaxHealth),
		Animator:      NewAnimator(),
		CharacterType: charType,
		State:         PlayerIdle,
		Energy:        100,
		MaxEnergy:     100,
		Speed:         config.Speed,
		JumpForce:     config.JumpForce,
		MaxJumps:      2, // Двойной прыжок
	}

	p.Renderer = NewSpriteRenderer(ss)

	// Загрузка анимаций
	if config.WalkAnim != "" {
		if walkAnim := ss.GetPlayerAnim(config.WalkAnim); walkAnim != nil {
			p.Animator.AddAnim("walk", walkAnim)
		}
	}
	if config.JumpAnim != "" {
		if jumpAnim := ss.GetPlayerAnim(config.JumpAnim); jumpAnim != nil {
			p.Animator.AddAnim("jump", jumpAnim)
		}
	}
	if config.RunAnim != "" {
		if runAnim := ss.GetPlayerAnim(config.RunAnim); runAnim != nil {
			p.Animator.AddAnim("run", runAnim)
		}
	}

	return p
}

// Update обновляет игрока
func (p *Player) Update(dt float64) {
	p.Transform.Update(dt)
	p.Physics.Update(dt, p.Transform)
	p.Health.Update(dt)
	p.Renderer.Update(dt)

	// Восстановление энергии
	if p.Energy < p.MaxEnergy && !p.Dashing {
		p.Energy += 15 * dt
		if p.Energy > p.MaxEnergy {
			p.Energy = p.MaxEnergy
		}
	}

	// Кулдаун рывка
	if p.DashCooldown > 0 {
		p.DashCooldown -= dt
	}

	// Таймер рывка
	if p.Dashing {
		p.DashTimer -= dt
		if p.DashTimer <= 0 {
			p.Dashing = false
			p.State = PlayerRunning
		}
	}

	// Невидимость
	if p.Invisible > 0 {
		p.Invisible -= dt
	}

	p.updateAnimation()
}

// updateAnimation обновляет анимацию игрока
func (p *Player) updateAnimation() {
	if p.Health.Dead {
		return
	}

	if p.Invisible > 0 && int(p.Invisible*10)%2 == 0 {
		return // Мигание
	}

	if p.Dashing {
		if runAnim := p.Animator.GetCurrentAnim(); runAnim != nil {
			p.Renderer.SetAnim(runAnim)
		}
	} else if !p.Physics.OnGround {
		if jumpAnim := p.Animator.GetCurrentAnim(); jumpAnim != nil {
			p.Renderer.SetAnim(jumpAnim)
		}
	} else if p.Physics.IsMoving {
		if p.Speed > 350 {
			if runAnim := p.Animator.GetCurrentAnim(); runAnim != nil {
				p.Renderer.SetAnim(runAnim)
			}
		} else {
			if walkAnim := p.Animator.GetCurrentAnim(); walkAnim != nil {
				p.Renderer.SetAnim(walkAnim)
			}
		}
	} else {
		if walkAnim := p.Animator.GetCurrentAnim(); walkAnim != nil {
			p.Renderer.SetAnim(walkAnim)
		}
	}
}

// MoveLeft движется влево
func (p *Player) MoveLeft() {
	p.Physics.VelocityX = -p.Speed
	p.Transform.Facing = -1
	p.Physics.IsMoving = true
	if p.State != PlayerJumping && p.State != PlayerDashing {
		p.State = PlayerRunning
	}
}

// MoveRight движется вправо
func (p *Player) MoveRight() {
	p.Physics.VelocityX = p.Speed
	p.Transform.Facing = 1
	p.Physics.IsMoving = true
	if p.State != PlayerJumping && p.State != PlayerDashing {
		p.State = PlayerRunning
	}
}

// Jump прыгает (поддерживает двойной прыжок)
func (p *Player) Jump() {
	if p.JumpCount < p.MaxJumps {
		p.Physics.ApplyJump(p.JumpForce)
		p.JumpCount++
		p.State = PlayerJumping
	}
}

// ResetJump сбрасывает счётчик прыжков
func (p *Player) ResetJump() {
	p.JumpCount = 0
}

// Dash выполняет рывок
func (p *Player) Dash() bool {
	config := GetCharacterConfig(p.CharacterType)
	if p.Energy >= config.DashCost && p.DashCooldown <= 0 {
		p.Energy -= config.DashCost
		p.Dashing = true
		p.DashTimer = 0.2
		p.DashCooldown = 1.0

		// Ускорение в направлении движения
		dashForce := p.Speed * 3
		if p.Transform.Facing == 1 {
			p.Physics.VelocityX = dashForce
		} else {
			p.Physics.VelocityX = -dashForce
		}
		p.State = PlayerDashing
		return true
	}
	return false
}

// Crouch приседает
func (p *Player) Crouch() {
	p.State = PlayerCrouching
	p.Transform.Height = 30
}

// Stand встаёт
func (p *Player) Stand() {
	config := GetCharacterConfig(p.CharacterType)
	p.State = PlayerIdle
	p.Transform.Height = config.Height
}

// CanShoot может ли стрелять
func (p *Player) CanShoot() bool {
	return p.Health.IsAlive()
}

// ActivateSpecial активирует специальную способность
func (p *Player) ActivateSpecial() {
	if p.CharacterType == CharNinja {
		p.Invisible = 3.0 // 3 секунды невидимости
	}
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	p.Renderer.Draw(screen, p.Transform, cameraX, cameraY)
}

// Enemy - сущность врага
type Enemy struct {
	Transform     *Transform
	Renderer      *SpriteRenderer
	Physics       *Physics
	Health        *Health
	Animator      *Animator
	EnemyType     string
	Damage        int
	Speed         float64
	AttackRange   float64
	ShootCooldown float64
	Behavior      EnemyBehavior
	PatrolStart   float64
	PatrolEnd     float64
	DetectionRange float64
	Alerted       bool
}

// EnemyBehavior - тип поведения врага
type EnemyBehavior int

const (
	EnemyMelee EnemyBehavior = iota
	EnemyRanged
	EnemyFlying
	EnemyPatrol
	EnemyTurret
	EnemyCamera
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

	switch enemyType {
	case "drone":
		e.Behavior = EnemyFlying
		e.Health.Max = 25
		e.Health.Current = 25
		e.Speed = 90
		e.Damage = 15
		e.DetectionRange = 300
	case "robodog":
		e.Behavior = EnemyPatrol
		e.Health.Max = 40
		e.Health.Current = 40
		e.Speed = 120
		e.Damage = 20
		e.PatrolStart = x - 100
		e.PatrolEnd = x + 100
	case "turret":
		e.Behavior = EnemyTurret
		e.Health.Max = 50
		e.Health.Current = 50
		e.Damage = 25
		e.DetectionRange = 400
		e.ShootCooldown = 1.5
	case "camera":
		e.Behavior = EnemyCamera
		e.Health.Max = 20
		e.Health.Current = 20
		e.DetectionRange = 350
		e.Transform.Width = 30
		e.Transform.Height = 30
	default:
		e.Behavior = EnemyMelee
	}

	return e
}

// Update обновляет врага
func (e *Enemy) Update(dt float64, playerX, playerY float64) {
	e.Transform.Update(dt)
	e.Physics.Update(dt, e.Transform)
	e.Health.Update(dt)
	e.Renderer.Update(dt)

	if e.ShootCooldown > 0 {
		e.ShootCooldown -= dt
	}

	// Проверка обнаружения игрока
	distX := playerX - e.Transform.X
	distY := playerY - e.Transform.Y
	distance := math.Sqrt(distX*distX + distY*distY)

	if distance < e.DetectionRange {
		e.Alerted = true
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

	if distX > 0 {
		e.Transform.Facing = 1
	} else {
		e.Transform.Facing = -1
	}

	switch e.Behavior {
	case EnemyMelee:
		if e.Alerted && distance < 400 {
			if distX > 0 {
				e.Physics.VelocityX = e.Speed
			} else {
				e.Physics.VelocityX = -e.Speed
			}
			e.Physics.IsMoving = true
		}

	case EnemyFlying:
		if e.Alerted && distance < 500 {
			if distX > 0 {
				e.Physics.VelocityX = e.Speed * 0.7
			} else {
				e.Physics.VelocityX = -e.Speed * 0.7
			}
			e.Physics.IsMoving = true
		}

	case EnemyPatrol:
		if e.Alerted && distance < 300 {
			// Преследование
			if distX > 0 {
				e.Physics.VelocityX = e.Speed * 1.2
			} else {
				e.Physics.VelocityX = -e.Speed * 1.2
			}
		} else {
			// Патрулирование
			if e.Transform.X <= e.PatrolStart {
				e.Physics.VelocityX = e.Speed
			} else if e.Transform.X >= e.PatrolEnd {
				e.Physics.VelocityX = -e.Speed
			}
		}
		e.Physics.IsMoving = true

	case EnemyTurret:
		// Турель не двигается, только стреляет
		break

	case EnemyCamera:
		// Камера только обнаруживает
		break
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

	if bulletSprite := ss.GetItemSprite("coinGold"); bulletSprite != nil {
		p.Renderer.SetSprite(bulletSprite)
	}

	return p
}

// Update обновляет снаряд
func (p *Projectile) Update(dt float64) {
	p.Transform.X += p.VelocityX * dt
	p.Transform.Y += p.VelocityY * dt
	p.VelocityY += 200 * dt
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
	Transform   *Transform
	Renderer    *SpriteRenderer
	ItemType    string
	Value       int
	Collected   bool
	AnimTimer   float64
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
