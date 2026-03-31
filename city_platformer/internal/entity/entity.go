// Package entity - игровые сущности для Sunny Adventure
// Go365 Day 91 - Доброе сказочное приключение
package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"sunny_adventure/internal/sprite"
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

// SpriteRenderer - компонент отрисовки спрайта
type SpriteRenderer struct {
	SpriteSheet *sprite.SpriteSheet
	CurrentImg  *ebiten.Image
	AnimName    string
	Anim        *sprite.Animation
	AnimFrame   int
	AnimTimer   float64
	AnimPlaying bool
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

	screen.DrawImage(sr.CurrentImg, opts)
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
		Acceleration: 600,
		Friction:     0.85,
		Gravity:      1000,
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

// Light - солнечная энергия
type Light struct {
	Current  float64
	Max      float64
	Regen    float64
	Dimming  float64 // Скорость потери света
}

// NewLight создаёт компонент света
func NewLight(max float64) *Light {
	return &Light{
		Current: max,
		Max:     max,
		Regen:   10,
		Dimming: 2,
	}
}

// Update обновляет свет
func (l *Light) Update(dt float64) {
	if l.Current < l.Max {
		l.Current += l.Regen * dt
		if l.Current > l.Max {
			l.Current = l.Max
		}
	}
}

// UseLight тратит свет
func (l *Light) UseLight(amount float64) bool {
	if l.Current >= amount {
		l.Current -= amount
		return true
	}
	return false
}

// Player - Зайчик (игрок)
type Player struct {
	Transform   *Transform
	Renderer    *SpriteRenderer
	Physics     *Physics
	Health      *Health
	Light       *Light
	State       PlayerState
	JumpCount   int
	MaxJumps    int
	Size        PlayerSize
	FriendCount int // Количество друзей
	Score       int
	ShootTimer  float64
	CarrotCount int // Собранные морковки
}

// PlayerSize - размер солнышка
type PlayerSize int

const (
	SizeSmall PlayerSize = iota
	SizeNormal
	SizeBig
)

// PlayerState - состояние игрока
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerHappy
	PlayerHurt
)

// NewPlayer создаёт нового игрока (Зайчика)
func NewPlayer(x, y float64, ss *sprite.SpriteSheet) *Player {
	p := &Player{
		Transform: NewTransform(x, y, 48, 48),
		Physics:   NewPhysics(),
		Health:    NewHealth(100),
		Light:     NewLight(100),
		State:     PlayerIdle,
		MaxJumps:  2, // Двойной прыжок!
		Size:      SizeNormal,
	}

	p.Renderer = NewSpriteRenderer(ss)

	// Загрузка спрайта зайчика
	if bunnySprite := ss.GetPlayerSprite("bunny"); bunnySprite != nil {
		p.Renderer.SetSprite(bunnySprite)
	}

	return p
}

// Update обновляет игрока
func (p *Player) Update(dt float64) {
	p.Transform.Update(dt)
	p.Physics.Update(dt, p.Transform)
	p.Health.Update(dt)
	p.Light.Update(dt)
	p.Renderer.Update(dt)

	if p.ShootTimer > 0 {
		p.ShootTimer -= dt
	}

	p.updateAnimation()
}

// updateAnimation обновляет анимацию
func (p *Player) updateAnimation() {
	if p.Health.Invincible > 0 && int(p.Health.Invincible*10)%2 == 0 {
		return // Мигание
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
	p.Physics.VelocityX = -250
	p.Transform.Facing = -1
	p.Physics.IsMoving = true
}

// MoveRight движется вправо
func (p *Player) MoveRight() {
	p.Physics.VelocityX = 250
	p.Transform.Facing = 1
	p.Physics.IsMoving = true
}

// Jump прыгает
func (p *Player) Jump() {
	if p.JumpCount < p.MaxJumps {
		p.Physics.ApplyJump(500)
		p.JumpCount++
		p.State = PlayerJumping
	}
}

// ResetJump сбрасывает счётчик прыжков
func (p *Player) ResetJump() {
	p.JumpCount = 0
}

// CanShoot может ли стрелять лучиком
func (p *Player) CanShoot() bool {
	return p.ShootTimer <= 0 && p.Light.Current >= 10
}

// Shoot стреляет лучиком света
func (p *Player) Shoot() {
	p.ShootTimer = 0.3
	p.Light.UseLight(10)
}

// Grow увеличивается в размере
func (p *Player) Grow() {
	if p.Size == SizeSmall {
		p.Size = SizeNormal
		p.Transform.Width = 48
		p.Transform.Height = 48
		p.Health.Max = 100
		p.Health.Heal(50)
	} else if p.Size == SizeNormal {
		p.Size = SizeBig
		p.Transform.Width = 64
		p.Transform.Height = 64
		p.Health.Max = 150
		p.Health.Heal(50)
	}
}

// Shrink уменьшается
func (p *Player) Shrink() {
	if p.Size == SizeBig {
		p.Size = SizeNormal
		p.Transform.Width = 48
		p.Transform.Height = 48
		p.Health.Max = 100
		if p.Health.Current > 100 {
			p.Health.Current = 100
		}
	} else if p.Size == SizeNormal {
		p.Size = SizeSmall
		p.Transform.Width = 32
		p.Transform.Height = 32
		p.Health.Max = 50
		if p.Health.Current > 50 {
			p.Health.Current = 50
		}
	}
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	p.Renderer.Draw(screen, p.Transform, cameraX, cameraY)
}

// FriendType - тип друга
type FriendType string

const (
	FriendBee      FriendType = "bee"      // Пчёлка 🐝
	FriendLadybug  FriendType = "ladybug"  // Божья коровка 🐞
	FriendFrog     FriendType = "frog"     // Лягушонок 🐸
	FriendSnail    FriendType = "snail"    // Улитка 🐌
	FriendGhost    FriendType = "ghost"    // Призрачок 👻
)

// Friend - друг (маленькое существо)
type Friend struct {
	Transform  *Transform
	Renderer   *SpriteRenderer
	FriendType FriendType
	Collected  bool
	AnimTimer  float64
	FloatY     float64
}

// NewFriend создаёт нового друга
func NewFriend(x, y float64, friendType FriendType, ss *sprite.SpriteSheet) *Friend {
	f := &Friend{
		Transform:  NewTransform(x, y, 32, 32),
		FriendType: friendType,
	}

	f.Renderer = NewSpriteRenderer(ss)

	// Загрузка спрайта по типу
	var spriteName string
	switch friendType {
	case FriendBee:
		spriteName = "bee_fly"
	case FriendLadybug:
		spriteName = "ladyBug_walk"
	case FriendFrog:
		spriteName = "frog"
	case FriendSnail:
		spriteName = "snail_walk"
	case FriendGhost:
		spriteName = "ghost_normal"
	}

	if img := ss.GetEnemySprite(spriteName); img != nil {
		f.Renderer.SetSprite(img)
	}

	return f
}

// Update обновляет друга
func (f *Friend) Update(dt float64) {
	f.AnimTimer += dt
	f.FloatY = math.Sin(f.AnimTimer*2) * 3
}

// Draw отрисовывает друга
func (f *Friend) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if f.Collected {
		return
	}

	origY := f.Transform.Y
	f.Transform.Y += f.FloatY
	f.Renderer.Draw(screen, f.Transform, cameraX, cameraY)
	f.Transform.Y = origY
}

// EnemyType - тип врага
type EnemyType string

const (
	EnemyWind     EnemyType = "wind"     // Ветерок 🌬️
	EnemyStorm    EnemyType = "storm"    // Тучка ⛈️
	EnemyBat      EnemyType = "bat"      // Летучая мышь 🦇
	EnemySpider   EnemyType = "spider"   // Паучок 🕷️
	EnemySnake    EnemyType = "snake"    // Змейка 🐍
)

// Enemy - враг
type Enemy struct {
	Transform     *Transform
	Renderer      *SpriteRenderer
	Physics       *Physics
	EnemyType     EnemyType
	Health        *Health
	Damage        int
	Speed         float64
	Behavior      EnemyBehavior
	PatrolStart   float64
	PatrolEnd     float64
	Converted     bool // Превращён в друга
	ShootCooldown float64
}

// EnemyBehavior - поведение врага
type EnemyBehavior int

const (
	EnemyPatrol EnemyBehavior = iota
	EnemyChase
	EnemyFlying
	EnemyStationary
)

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType EnemyType, ss *sprite.SpriteSheet) *Enemy {
	e := &Enemy{
		Transform: NewTransform(x, y, 40, 40),
		Physics:   NewPhysics(),
		Health:    NewHealth(30),
		EnemyType: enemyType,
		Damage:    10,
		Speed:     60,
		Behavior:  EnemyPatrol,
	}

	e.Renderer = NewSpriteRenderer(ss)

	// Настройка по типу врага
	switch enemyType {
	case EnemyWind:
		e.Behavior = EnemyFlying
		e.Health.Max = 20
		e.Speed = 80
	case EnemyStorm:
		e.Behavior = EnemyFlying
		e.Health.Max = 40
		e.Damage = 20
		e.Speed = 50
	case EnemyBat:
		e.Behavior = EnemyChase
		e.Health.Max = 25
		e.Speed = 100
	case EnemySpider:
		e.Behavior = EnemyStationary
		e.Health.Max = 20
		e.Damage = 15
	case EnemySnake:
		e.Behavior = EnemyPatrol
		e.Health.Max = 30
		e.Speed = 90
		e.PatrolStart = x - 80
		e.PatrolEnd = x + 80
	}

	// Загрузка спрайта
	var spriteName string
	switch enemyType {
	case EnemyWind:
		spriteName = "fly_fly"
	case EnemyStorm:
		spriteName = "ghost"
	case EnemyBat:
		spriteName = "bat_fly"
	case EnemySpider:
		spriteName = "spider_walk1"
	case EnemySnake:
		spriteName = "snake_walk"
	}

	if img := ss.GetEnemySprite(spriteName); img != nil {
		e.Renderer.SetSprite(img)
	}

	return e
}

// Update обновляет врага
func (e *Enemy) Update(dt float64, playerX, playerY float64) {
	e.Transform.Update(dt)
	e.Health.Update(dt)
	e.Renderer.Update(dt)

	if e.ShootCooldown > 0 {
		e.ShootCooldown -= dt
	}
}

// AI обновляет ИИ врага
func (e *Enemy) AI(dt float64, playerX, playerY float64) {
	if !e.Health.IsAlive() || e.Converted {
		return
	}

	distX := playerX - e.Transform.X
	distance := math.Sqrt(distX*distX)

	if distX > 0 {
		e.Transform.Facing = 1
	} else {
		e.Transform.Facing = -1
	}

	switch e.Behavior {
	case EnemyPatrol:
		if e.Transform.X <= e.PatrolStart {
			e.Physics.VelocityX = e.Speed
		} else if e.Transform.X >= e.PatrolEnd {
			e.Physics.VelocityX = -e.Speed
		}

	case EnemyChase:
		if distance < 300 {
			if distX > 0 {
				e.Physics.VelocityX = e.Speed
			} else {
				e.Physics.VelocityX = -e.Speed
			}
		}

	case EnemyFlying:
		if distance < 400 {
			if distX > 0 {
				e.Physics.VelocityX = e.Speed * 0.7
			} else {
				e.Physics.VelocityX = -e.Speed * 0.7
			}
		}
	}
}

// Convert превращает врага в друга
func (e *Enemy) Convert() {
	e.Converted = true
	e.Health.Current = e.Health.Max
	e.Behavior = EnemyPatrol
	e.Damage = 0
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	e.Renderer.Draw(screen, e.Transform, cameraX, cameraY)
}

// Projectile - лучик света
type Projectile struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	VelocityX float64
	VelocityY float64
	LifeTime  float64
	Damage    int
	Active    bool
	IsFriend  bool // Лучик добра
}

// NewProjectile создаёт новый снаряд (лучик)
func NewProjectile(x, y, vx, vy float64, damage int, isFriend bool, ss *sprite.SpriteSheet) *Projectile {
	p := &Projectile{
		Transform: NewTransform(x, y, 24, 8),
		VelocityX: vx,
		VelocityY: vy,
		LifeTime:  1.5,
		Damage:    damage,
		IsFriend:  isFriend,
		Active:    true,
	}

	p.Renderer = NewSpriteRenderer(ss)

	// Жёлтый лучик света
	if bulletSprite := ss.GetItemSprite("coinGold"); bulletSprite != nil {
		p.Renderer.SetSprite(bulletSprite)
	}

	return p
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

// Cloud - облачко (коллекционный предмет)
type Cloud struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	Collected bool
	AnimTimer float64
	FloatY    float64
	CloudNum  int // 1, 2, или 3
}

// NewCloud создаёт новое облачко
func NewCloud(x, y float64, cloudNum int, ss *sprite.SpriteSheet) *Cloud {
	c := &Cloud{
		Transform: NewTransform(x, y, 48, 32),
		CloudNum:  cloudNum,
	}

	c.Renderer = NewSpriteRenderer(ss)

	// Загрузка спрайта облака
	var spriteName string
	switch cloudNum {
	case 1:
		spriteName = "cloud1"
	case 2:
		spriteName = "cloud2"
	case 3:
		spriteName = "cloud3"
	}

	if img := ss.GetItemSprite(spriteName); img != nil {
		c.Renderer.SetSprite(img)
	}

	return c
}

// Update обновляет облачко
func (c *Cloud) Update(dt float64) {
	c.AnimTimer += dt
	c.FloatY = math.Sin(c.AnimTimer) * 2
}

// Draw отрисовывает облачко
func (c *Cloud) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if c.Collected {
		return
	}

	origY := c.Transform.Y
	c.Transform.Y += c.FloatY
	c.Renderer.Draw(screen, c.Transform, cameraX, cameraY)
	c.Transform.Y = origY
}

// CheckCollision проверяет коллизию AABB
func CheckCollision(a, b *Transform) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}
