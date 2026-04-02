// Package entity - игровые сущности для City Platformer
// Go365 Day 93 - Neon Runner: Cyber Escape
package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"city_platformer/internal/sprite"
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

// Center возвращает центр сущности
func (t *Transform) Center() (float64, float64) {
	return t.X+t.Width/2, t.Y+t.Height/2
}

// SpriteRenderer - компонент отрисовки спрайта
type SpriteRenderer struct {
	SpriteSheet *sprite.SpriteSheet
	CurrentImg  *ebiten.Image
	AnimFrames  []*ebiten.Image
	AnimFrame   int
	AnimTimer   float64
	AnimFPS     float64
	AnimPlaying bool
	AnimLoop    bool
	UseSprite   bool
	SpriteColor color.Color
}

// NewSpriteRenderer создаёт рендерер спрайтов
func NewSpriteRenderer(ss *sprite.SpriteSheet) *SpriteRenderer {
	return &SpriteRenderer{
		SpriteSheet: ss,
		AnimFPS:     8,
		AnimLoop:    true,
	}
}

// SetSprite устанавливает статичный спрайт
func (sr *SpriteRenderer) SetSprite(img *ebiten.Image) {
	sr.CurrentImg = img
	sr.UseSprite = img != nil
	sr.AnimPlaying = false
}

// SetAnim устанавливает анимацию
func (sr *SpriteRenderer) SetAnim(frames []*ebiten.Image) {
	if len(frames) > 0 {
		sr.AnimFrames = frames
		sr.CurrentImg = frames[0]
		sr.AnimFrame = 0
		sr.AnimTimer = 0
		sr.AnimPlaying = true
		sr.UseSprite = true
	}
}

// Update обновляет анимацию
func (sr *SpriteRenderer) Update(dt float64) {
	if !sr.AnimPlaying || len(sr.AnimFrames) == 0 {
		return
	}

	sr.AnimTimer += dt
	frameDuration := 1.0 / sr.AnimFPS

	if sr.AnimTimer >= frameDuration {
		sr.AnimTimer = 0
		sr.AnimFrame++
		if sr.AnimFrame >= len(sr.AnimFrames) {
			if sr.AnimLoop {
				sr.AnimFrame = 0
			} else {
				sr.AnimFrame = len(sr.AnimFrames) - 1
				sr.AnimPlaying = false
			}
		}
		sr.CurrentImg = sr.AnimFrames[sr.AnimFrame]
	}
}

// Draw отрисовывает спрайт
func (sr *SpriteRenderer) Draw(screen *ebiten.Image, transform *Transform, cameraX, cameraY float64) {
	if sr.UseSprite && sr.CurrentImg != nil {
		sr.drawSprite(screen, transform, cameraX, cameraY)
	} else if sr.SpriteColor != nil {
		sr.drawVector(screen, transform, cameraX, cameraY)
	}
}

// drawSprite отрисовывает спрайт
func (sr *SpriteRenderer) drawSprite(screen *ebiten.Image, transform *Transform, cameraX, cameraY float64) {
	if sr.CurrentImg == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(transform.ScaleX, transform.ScaleY)

	if transform.Facing == -1 {
		opts.GeoM.Scale(-1, 1)
		opts.GeoM.Translate(float64(transform.Width), 0)
	}

	screenX := transform.X - cameraX
	screenY := transform.Y - cameraY
	opts.GeoM.Translate(screenX, screenY)

	screen.DrawImage(sr.CurrentImg, opts)
}

// drawVector отрисовывает векторную графику
func (sr *SpriteRenderer) drawVector(screen *ebiten.Image, transform *Transform, cameraX, cameraY float64) {
	x := transform.X - cameraX
	y := transform.Y - cameraY
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(transform.Width), float32(transform.Height), sr.SpriteColor, false)
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
	h.Invincible = 2.0
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
		Regen:   10,
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

// Player - главный герой (KAI - хакер-беглец)
type Player struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	Physics   *Physics
	Health    *Health
	Energy    *Energy

	State     PlayerState
	JumpCount int
	MaxJumps  int

	Score       int
	DataCollected int
	ShootTimer  float64
	AnimTimer   float64
}

// PlayerState - состояние игрока
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerHurt
	PlayerDead
)

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64, ss *sprite.SpriteSheet) *Player {
	p := &Player{
		Transform: NewTransform(x, y, 32, 48),
		Physics:   NewPhysics(),
		Health:    NewHealth(100),
		Energy:    NewEnergy(100),
		State:     PlayerIdle,
		MaxJumps:  2, // Двойной прыжок
	}

	p.Renderer = NewSpriteRenderer(ss)

	// Загрузка спрайтов
	if ss != nil {
		if standSprite := ss.GetPlayerSprite("stand"); standSprite != nil {
			p.Renderer.SetSprite(standSprite)
			p.Transform.Width = float64(standSprite.Bounds().Dx())
			p.Transform.Height = float64(standSprite.Bounds().Dy())
		}
		if walkAnim := ss.PlayerWalk; len(walkAnim) > 0 {
			// Анимация будет установлена при движении
		}
		if jumpSprite := ss.GetPlayerSprite("jump"); jumpSprite != nil {
			// Спрайт прыжка
		}
	}

	if !p.Renderer.UseSprite {
		p.Renderer.SpriteColor = color.RGBA{0, 240, 255, 255} // Неоновый голубой
	}

	return p
}

// Update обновляет игрока
func (p *Player) Update(dt float64) {
	p.Physics.VelocityX *= p.Physics.Friction
	p.Physics.VelocityY += p.Physics.Gravity * dt

	if p.Physics.VelocityY > 800 {
		p.Physics.VelocityY = 800
	}

	p.Transform.X += p.Physics.VelocityX * dt
	p.Transform.Y += p.Physics.VelocityY * dt

	p.Physics.IsMoving = math.Abs(p.Physics.VelocityX) > 10

	p.Energy.Update(dt)

	if p.Health.Invincible > 0 {
		p.Health.Invincible -= dt
	}

	if p.ShootTimer > 0 {
		p.ShootTimer -= dt
	}

	p.AnimTimer += dt
	p.Renderer.Update(dt)
	p.updateState()
}

// updateState обновляет состояние
func (p *Player) updateState() {
	if p.Health.Dead {
		p.State = PlayerDead
		return
	}

	if p.Health.Invincible > 0 {
		p.State = PlayerHurt
		return
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
}

// MoveRight движется вправо
func (p *Player) MoveRight() {
	p.Physics.VelocityX = 250
	p.Transform.Facing = 1
}

// Stop останавливает движение
func (p *Player) Stop() {
	p.Physics.VelocityX = 0
}

// Jump прыгает
func (p *Player) Jump() bool {
	if p.JumpCount < p.MaxJumps {
		jumpForce := 500.0
		if p.JumpCount == 1 {
			jumpForce = 400
		}
		p.Physics.VelocityY = -jumpForce
		p.JumpCount++
		p.Physics.OnGround = false
		return true
	}
	return false
}

// ResetJump сбрасывает счётчик прыжков
func (p *Player) ResetJump() {
	p.JumpCount = 0
}

// CanShoot может ли стрелять
func (p *Player) CanShoot() bool {
	return p.ShootTimer <= 0
}

// Shoot стреляет
func (p *Player) Shoot() {
	p.ShootTimer = 0.3
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

	// Установка правильного спрайта
	if p.Renderer.SpriteSheet != nil {
		switch p.State {
		case PlayerRunning:
			if len(p.Renderer.SpriteSheet.PlayerWalk) > 0 {
				p.Renderer.SetAnim(p.Renderer.SpriteSheet.PlayerWalk)
			}
		case PlayerJumping:
			if jumpSprite := p.Renderer.SpriteSheet.GetPlayerSprite("jump"); jumpSprite != nil {
				p.Renderer.SetSprite(jumpSprite)
			}
		case PlayerIdle:
			if standSprite := p.Renderer.SpriteSheet.GetPlayerSprite("stand"); standSprite != nil {
				p.Renderer.SetSprite(standSprite)
			}
		case PlayerHurt:
			if hurtSprite := p.Renderer.SpriteSheet.GetPlayerSprite("hurt"); hurtSprite != nil {
				p.Renderer.SetSprite(hurtSprite)
			}
		}
	}

	p.Renderer.Draw(screen, p.Transform, cameraX, cameraY)
}

// Enemy - враг
type Enemy struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	Physics   *Physics
	Health    *Health

	EnemyType string
	Damage    int
	Speed     float64

	PatrolStart float64
	PatrolEnd   float64
	Behavior    int

	AnimTimer float64
}

const (
	EnemyPatrol = iota
	EnemyChase
	EnemyFlying
	EnemyStationary
)

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType string, ss *sprite.SpriteSheet) *Enemy {
	e := &Enemy{
		Transform: NewTransform(x, y, 32, 32),
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
	case "fly":
		e.Behavior = EnemyFlying
		e.Health.Max = 20
		e.Speed = 80
		e.Transform.Height = 24
	case "bat":
		e.Behavior = EnemyChase
		e.Health.Max = 25
		e.Speed = 100
		e.Transform.Height = 24
	case "slime":
		e.Health.Max = 30
		e.Speed = 40
	case "snake":
		e.Health.Max = 30
		e.Speed = 70
	case "spider":
		e.Health.Max = 20
		e.Speed = 50
	case "ghost":
		e.Behavior = EnemyChase
		e.Health.Max = 40
		e.Speed = 60
	}

	// Загрузка спрайта
	if ss != nil {
		var spriteName string
		var animFrames []*ebiten.Image

		switch enemyType {
		case "fly":
			animFrames = []*ebiten.Image{
				ss.GetEnemySprite("flyFly1"),
				ss.GetEnemySprite("flyFly2"),
			}
		case "bat":
			spriteName = "bat_fly"
		case "slime":
			animFrames = []*ebiten.Image{
				ss.GetEnemySprite("slimeWalk1"),
				ss.GetEnemySprite("slimeWalk2"),
			}
		case "snake":
			spriteName = "snakeWalk"
		case "spider":
			animFrames = []*ebiten.Image{
				ss.GetEnemySprite("spider_walk1"),
				ss.GetEnemySprite("spider_walk2"),
			}
		case "ghost":
			spriteName = "ghost_normal"
		}

		if animFrames != nil && len(animFrames) > 0 && animFrames[0] != nil {
			validFrames := make([]*ebiten.Image, 0)
			for _, f := range animFrames {
				if f != nil {
					validFrames = append(validFrames, f)
				}
			}
			if len(validFrames) > 0 {
				e.Renderer.SetAnim(validFrames)
				e.Transform.Width = float64(validFrames[0].Bounds().Dx())
				e.Transform.Height = float64(validFrames[0].Bounds().Dy())
			}
		} else if spriteName != "" {
			if img := ss.GetEnemySprite(spriteName); img != nil {
				e.Renderer.SetSprite(img)
				e.Transform.Width = float64(img.Bounds().Dx())
				e.Transform.Height = float64(img.Bounds().Dy())
			}
		}
	}

	if !e.Renderer.UseSprite {
		e.Renderer.SpriteColor = color.RGBA{255, 68, 68, 255} // Красный
	}

	e.Health.Current = e.Health.Max
	if e.PatrolStart == 0 {
		e.PatrolStart = x - 100
		e.PatrolEnd = x + 100
	}

	return e
}

// Update обновляет врага
func (e *Enemy) Update(dt float64, playerX, playerY float64) {
	e.AnimTimer += dt

	if !e.Health.Dead {
		e.Renderer.Update(dt)
		e.updateAI(dt, playerX, playerY)
	}

	if e.Health.Invincible > 0 {
		e.Health.Invincible -= dt
	}
}

// updateAI обновляет ИИ врага
func (e *Enemy) updateAI(dt float64, playerX, playerY float64) {
	distX := playerX - e.Transform.X
	distance := math.Sqrt(distX * distX)

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
		} else {
			e.Physics.VelocityX = e.Speed * float64(e.Transform.Facing)
		}

	case EnemyChase:
		if distance < 300 {
			if distX > 0 {
				e.Physics.VelocityX = e.Speed
			} else {
				e.Physics.VelocityX = -e.Speed
			}
		} else {
			// Патрулирование если игрок далеко
			if e.Transform.X <= e.PatrolStart {
				e.Physics.VelocityX = e.Speed
			} else if e.Transform.X >= e.PatrolEnd {
				e.Physics.VelocityX = -e.Speed
			} else {
				e.Physics.VelocityX = e.Speed * float64(e.Transform.Facing) * 0.5
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

	e.Transform.X += e.Physics.VelocityX * dt
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if e.Health.Dead {
		// Крест вместо мёртвого врага
		x := e.Transform.X - cameraX
		y := e.Transform.Y - cameraY
		vector.StrokeLine(screen, float32(x), float32(y), float32(x+e.Transform.Width), float32(y+e.Transform.Height), 2, color.RGBA{255, 0, 0, 255}, false)
		vector.StrokeLine(screen, float32(x+e.Transform.Width), float32(y), float32(x), float32(y+e.Transform.Height), 2, color.RGBA{255, 0, 0, 255}, false)
		return
	}

	if e.Health.Invincible > 0 && int(e.Health.Invincible*10)%2 == 0 {
		return
	}

	e.Renderer.Draw(screen, e.Transform, cameraX, cameraY)

	// Полоска здоровья
	if e.Health.Current < e.Health.Max {
		hpPercent := float32(e.Health.Current) / float32(e.Health.Max)
		vector.DrawFilledRect(screen, float32(e.Transform.X-cameraX), float32(e.Transform.Y-cameraY-6), float32(e.Transform.Width), 4, color.RGBA{100, 100, 100, 255}, false)
		vector.DrawFilledRect(screen, float32(e.Transform.X-cameraX), float32(e.Transform.Y-cameraY-6), float32(e.Transform.Width)*hpPercent, 4, color.RGBA{0, 255, 0, 255}, false)
	}
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

const (
	ItemCoinGold   = "coinGold"
	ItemCoinSilver = "coinSilver"
	ItemCoinBronze = "coinBronze"
	ItemGemRed     = "gemRed"
	ItemGemBlue    = "gemBlue"
	ItemGemGreen   = "gemGreen"
	ItemGemYellow  = "gemYellow"
	ItemStar       = "star"
	ItemMushroom   = "mushroomRed"
	ItemData       = "data"
)

// NewItem создаёт новый предмет
func NewItem(x, y float64, itemType string, value int, ss *sprite.SpriteSheet) *Item {
	i := &Item{
		Transform: &Transform{
			X: x, Y: y, Width: 32, Height: 32,
			ScaleX: 1, ScaleY: 1, Facing: 1,
		},
		ItemType:  itemType,
		Value:     value,
	}

	i.Renderer = NewSpriteRenderer(ss)

	// Загрузка спрайта
	var spriteName string
	switch itemType {
	case ItemCoinGold, ItemCoinSilver, ItemCoinBronze:
		spriteName = itemType
	case ItemGemRed, ItemGemBlue, ItemGemGreen, ItemGemYellow:
		spriteName = itemType
	case ItemStar:
		spriteName = "star"
	case ItemMushroom:
		spriteName = "mushroomRed"
	case ItemData:
		spriteName = "keyBlue" // Временный спрайт для данных
	}

	if ss != nil {
		if img := ss.GetItem(spriteName); img != nil {
			i.Renderer.SetSprite(img)
			i.Transform.Width = float64(img.Bounds().Dx())
			i.Transform.Height = float64(img.Bounds().Dy())
		}
	}

	if !i.Renderer.UseSprite {
		switch itemType {
		case ItemCoinGold:
			i.Renderer.SpriteColor = color.RGBA{255, 215, 0, 255}
		case ItemCoinSilver:
			i.Renderer.SpriteColor = color.RGBA{192, 192, 192, 255}
		case ItemCoinBronze:
			i.Renderer.SpriteColor = color.RGBA{205, 127, 50, 255}
		case ItemGemRed:
			i.Renderer.SpriteColor = color.RGBA{255, 50, 50, 255}
		case ItemGemBlue:
			i.Renderer.SpriteColor = color.RGBA{50, 100, 255, 255}
		case ItemGemGreen:
			i.Renderer.SpriteColor = color.RGBA{50, 255, 50, 255}
		case ItemGemYellow:
			i.Renderer.SpriteColor = color.RGBA{255, 255, 50, 255}
		case ItemStar:
			i.Renderer.SpriteColor = color.RGBA{255, 255, 255, 255}
		case ItemMushroom:
			i.Renderer.SpriteColor = color.RGBA{255, 100, 100, 255}
		case ItemData:
			i.Renderer.SpriteColor = color.RGBA{0, 240, 255, 255}
		}
	}

	return i
}

// Update обновляет предмет
func (i *Item) Update(dt float64) {
	i.AnimTimer += dt
	i.FloatOffset = math.Sin(i.AnimTimer*3) * 5
	i.Renderer.Update(dt)
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

// Platform - платформа
type Platform struct {
	Transform *Transform
	TileType  string
	Sprite    *ebiten.Image
	Color     color.Color
}

// NewPlatform создаёт новую платформу
func NewPlatform(x, y, width, height float64, tileType string, ss *sprite.SpriteSheet) *Platform {
	p := &Platform{
		Transform: NewTransform(x, y, width, height),
		TileType:  tileType,
	}

	// Загрузка спрайта тайла
	if ss != nil {
		p.Sprite = ss.GetTile(tileType)
	}

	if p.Sprite == nil {
		switch tileType {
		case "grass":
			p.Color = color.RGBA{100, 180, 80, 255}
		case "dirt":
			p.Color = color.RGBA{139, 90, 43, 255}
		case "stone":
			p.Color = color.RGBA{128, 128, 128, 255}
		case "brickWall":
			p.Color = color.RGBA{178, 34, 34, 255}
		case "castle":
			p.Color = color.RGBA{100, 100, 150, 255}
		default:
			p.Color = color.RGBA{100, 100, 100, 255}
		}
	}

	return p
}

// Draw отрисовывает платформу
func (p *Platform) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if p.Sprite != nil {
		// Отрисовка тайлами
		tileSize := float64(p.Sprite.Bounds().Dx())
		for x := p.Transform.X; x < p.Transform.X+p.Transform.Width; x += tileSize {
			for y := p.Transform.Y; y < p.Transform.Y+p.Transform.Height; y += tileSize {
				if x >= cameraX && x < cameraX+screenWidth && y >= cameraY && y < cameraY+screenHeight {
					opts := &ebiten.DrawImageOptions{}
					opts.GeoM.Translate(x-cameraX, y-cameraY)
					screen.DrawImage(p.Sprite, opts)
				}
			}
		}
	} else if p.Color != nil {
		vector.DrawFilledRect(screen, float32(p.Transform.X-cameraX), float32(p.Transform.Y-cameraY), float32(p.Transform.Width), float32(p.Transform.Height), p.Color, false)
	}
}

// Projectile - снаряд (луч)
type Projectile struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	VelocityX float64
	VelocityY float64
	LifeTime  float64
	Damage    int
	Active    bool
}

// NewProjectile создаёт новый снаряд
func NewProjectile(x, y, vx, vy float64, damage int, ss *sprite.SpriteSheet) *Projectile {
	p := &Projectile{
		Transform: NewTransform(x, y, 16, 8),
		VelocityX: vx,
		VelocityY: vy,
		LifeTime:  1.5,
		Damage:    damage,
		Active:    true,
	}

	p.Renderer = NewSpriteRenderer(ss)

	// Жёлтый луч
	if !p.Renderer.UseSprite {
		p.Renderer.SpriteColor = color.RGBA{255, 255, 100, 255}
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

// Particle - частица
type Particle struct {
	X, Y, VX, VY float64
	Life         float64
	Color        color.Color
	Size         float64
}

// CheckCollision проверяет коллизию AABB
func CheckCollision(a, b *Transform) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

const (
	screenWidth  = 1280
	screenHeight = 720
)
