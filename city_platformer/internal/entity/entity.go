// Package entity - игровые сущности для Village Platformer
// Go365 Day 93 - Деревенский платформер: Домики, деревья, холмы
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

// Player - главный герой (путник)
type Player struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	Physics   *Physics
	Health    *Health

	State     PlayerState
	JumpCount int
	MaxJumps  int

	Score      int
	ShootTimer float64
	AnimTimer  float64
}

// PlayerState - состояние игрока
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerHurt
)

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64, ss *sprite.SpriteSheet) *Player {
	p := &Player{
		Transform: NewTransform(x, y, 32, 48),
		Physics:   NewPhysics(),
		Health:    NewHealth(100),
		State:     PlayerIdle,
		MaxJumps:  2,
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
			// Анимация ходьбы
		}
	}

	if !p.Renderer.UseSprite {
		p.Renderer.SpriteColor = color.RGBA{100, 149, 237, 255} // Васильковый
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

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if p.Health.Dead {
		return
	}

	// Мигание при получении урона
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
			if hurtSprite := p.Renderer.SpriteSheet.GetPlayerSprite("jump"); hurtSprite != nil {
				p.Renderer.SetSprite(hurtSprite)
			}
		}
	}

	p.Renderer.Draw(screen, p.Transform, cameraX, cameraY)
}

// House - деревенский домик с двускатной крышей и трубой 🏡
type House struct {
	Transform   *Transform
	HouseType   string // beige, dark, gray
	SpriteParts map[string]*ebiten.Image
	SmokeTimer  float64
	SmokeParts  []SmokeParticle
	Color       color.Color
}

// SmokeParticle - частица дыма из трубы
type SmokeParticle struct {
	X, Y   float64
	VY     float64
	Life   float64
	Size   float64
}

// NewHouse создаёт новый домик
func NewHouse(x, y float64, houseType string, ss *sprite.SpriteSheet) *House {
	h := &House{
		Transform:   NewTransform(x, y-96, 96, 96), // Домик 96x96
		HouseType:   houseType,
		SpriteParts: make(map[string]*ebiten.Image),
		SmokeParts:  make([]SmokeParticle, 0),
	}

	// Загрузка частей домика
	if ss != nil {
		prefix := "house" + houseType
		// Стены
		h.SpriteParts["bottomLeft"] = ss.GetTile(prefix + "BottomLeft")
		h.SpriteParts["bottomMid"] = ss.GetTile(prefix + "BottomMid")
		h.SpriteParts["bottomRight"] = ss.GetTile(prefix + "BottomRight")
		h.SpriteParts["midLeft"] = ss.GetTile(prefix + "MidLeft")
		h.SpriteParts["midRight"] = ss.GetTile(prefix + "MidRight")
		// Крыша
		h.SpriteParts["topLeft"] = ss.GetTile(prefix + "TopLeft")
		h.SpriteParts["topMid"] = ss.GetTile(prefix + "TopMid")
		h.SpriteParts["topRight"] = ss.GetTile(prefix + "TopRight")
	}

	// Если нет спрайтов - используем цвет
	if h.SpriteParts["bottomLeft"] == nil {
		switch houseType {
		case "beige":
			h.Color = color.RGBA{210, 180, 140, 255}
		case "dark":
			h.Color = color.RGBA{101, 67, 33, 255}
		case "gray":
			h.Color = color.RGBA{128, 128, 128, 255}
		}
	}

	return h
}

// Update обновляет домик (дым из трубы)
func (h *House) Update(dt float64) {
	h.SmokeTimer += dt

	// Создаём новые частицы дыма каждые 0.5 сек
	if h.SmokeTimer >= 0.5 {
		h.SmokeTimer = 0
		// Дым идёт из трубы (середина крыши)
		h.SmokeParts = append(h.SmokeParts, SmokeParticle{
			X:    h.Transform.X + h.Transform.Width/2 - 5,
			Y:    h.Transform.Y - 10,
			VY:   -20,
			Life: 2.0,
			Size: 8,
		})
	}

	// Обновляем частицы дыма
	active := make([]SmokeParticle, 0)
	for i := range h.SmokeParts {
		s := &h.SmokeParts[i]
		s.Y += s.VY * dt
		s.X += math.Sin(s.Life*2) * 5 * dt // Дым колеблется
		s.Life -= dt
		s.Size += 2 * dt // Дым расширяется

		if s.Life > 0 {
			active = append(active, *s)
		}
	}
	h.SmokeParts = active
}

// Draw отрисовывает домик
func (h *House) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	// Рисуем частицы дыма
	for _, s := range h.SmokeParts {
		alpha := uint8(s.Life * 100)
		smokeColor := color.RGBA{200, 200, 200, alpha}
		vector.DrawFilledRect(screen, float32(s.X-cameraX), float32(s.Y-cameraY), float32(s.Size), float32(s.Size), smokeColor, false)
	}

	// Рисуем домик по частям
	if h.SpriteParts["bottomLeft"] != nil {
		// Нижний ряд (стены)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(h.Transform.X-cameraX, h.Transform.Y-cameraY+64)
		screen.DrawImage(h.SpriteParts["bottomLeft"], opts)

		opts = &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(h.Transform.X-cameraX+32, h.Transform.Y-cameraY+64)
		screen.DrawImage(h.SpriteParts["bottomMid"], opts)

		opts = &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(h.Transform.X-cameraX+64, h.Transform.Y-cameraY+64)
		screen.DrawImage(h.SpriteParts["bottomRight"], opts)

		// Средний ряд (стены)
		opts = &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(h.Transform.X-cameraX, h.Transform.Y-cameraY+32)
		screen.DrawImage(h.SpriteParts["midLeft"], opts)

		opts = &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(h.Transform.X-cameraX+64, h.Transform.Y-cameraY+32)
		screen.DrawImage(h.SpriteParts["midRight"], opts)

		// Верхний ряд (крыша с трубой)
		opts = &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(h.Transform.X-cameraX, h.Transform.Y-cameraY)
		screen.DrawImage(h.SpriteParts["topLeft"], opts)

		opts = &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(h.Transform.X-cameraX+32, h.Transform.Y-cameraY)
		screen.DrawImage(h.SpriteParts["topMid"], opts)

		opts = &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(h.Transform.X-cameraX+64, h.Transform.Y-cameraY)
		screen.DrawImage(h.SpriteParts["topRight"], opts)
	} else {
		// Векторный домик
		baseX := float32(h.Transform.X - cameraX)
		baseY := float32(h.Transform.Y - cameraY)

		// Стены
		vector.DrawFilledRect(screen, baseX, baseY+32, 96, 64, h.Color, false)
		// Крыша (треугольник)
		roofColor := color.RGBA{139, 69, 19, 255} // Коричневая крыша
		vector.DrawFilledRect(screen, baseX, baseY+32, 96, 30, roofColor, false)
		vector.DrawFilledRect(screen, baseX+10, baseY+15, 76, 20, roofColor, false)
		vector.DrawFilledRect(screen, baseX+25, baseY+5, 46, 15, roofColor, false)
		// Труба
		chimneyColor := color.RGBA{178, 34, 34, 255}
		vector.DrawFilledRect(screen, baseX+42, baseY-10, 12, 25, chimneyColor, false)
		// Окно
		windowColor := color.RGBA{100, 149, 237, 200}
		vector.DrawFilledRect(screen, baseX+20, baseY+45, 20, 20, windowColor, false)
		vector.DrawFilledRect(screen, baseX+56, baseY+45, 20, 20, windowColor, false)
		// Дверь
		doorColor := color.RGBA{101, 67, 33, 255}
		vector.DrawFilledRect(screen, baseX+38, baseY+55, 20, 41, doorColor, false)
	}
}

// Tree - дерево или ёлка 🌳🌲
type Tree struct {
	Transform *Transform
	TreeType  string // pine (ёлка), oak (дуб)
	Sprite    *ebiten.Image
	Color     color.Color
}

// NewTree создаёт новое дерево
func NewTree(x, y float64, treeType string, ss *sprite.SpriteSheet) *Tree {
	t := &Tree{
		Transform: NewTransform(x, y-120, 64, 120), // Дерево 64x120
		TreeType:  treeType,
	}

	// Загрузка спрайта
	if ss != nil {
		if treeType == "pine" {
			// Ёлка - используем сосну
			t.Sprite = ss.GetTree("pine0")
			if t.Sprite == nil {
				t.Sprite = ss.GetTree("pine1")
			}
		} else {
			// Дуб - используем обычное дерево
			t.Sprite = ss.GetTree("tree")
		}
	}

	if t.Sprite != nil {
		t.Transform.Width = float64(t.Sprite.Bounds().Dx())
		t.Transform.Height = float64(t.Sprite.Bounds().Dy())
	}

	if t.Sprite == nil {
		if treeType == "pine" {
			t.Color = color.RGBA{34, 139, 34, 255} // Тёмно-зелёная ёлка
		} else {
			t.Color = color.RGBA{60, 179, 113, 255} // Светло-зелёный дуб
		}
	}

	return t
}

// Draw отрисовывает дерево
func (t *Tree) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if t.Sprite != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(t.Transform.X-cameraX, t.Transform.Y-cameraY)
		screen.DrawImage(t.Sprite, opts)
	} else {
		// Векторное дерево
		baseX := float32(t.Transform.X - cameraX)
		baseY := float32(t.Transform.Y - cameraY)

		if t.TreeType == "pine" {
			// Ёлка (треугольники)
			firColor := color.RGBA{34, 139, 34, 255}
			// Нижний ярус
			vector.DrawFilledRect(screen, baseX, baseY+60, 64, 60, firColor, false)
			// Средний ярус
			vector.DrawFilledRect(screen, baseX+8, baseY+35, 48, 40, firColor, false)
			// Верхний ярус
			vector.DrawFilledRect(screen, baseX+16, baseY+15, 32, 30, firColor, false)
			// Ствол
			trunkColor := color.RGBA{101, 67, 33, 255}
			vector.DrawFilledRect(screen, baseX+27, baseY+115, 10, 25, trunkColor, false)
		} else {
			// Дуб (крона + ствол)
			leafColor := color.RGBA{60, 179, 113, 255}
			// Крона (круг/овал)
			vector.DrawFilledRect(screen, baseX+10, baseY+20, 44, 50, leafColor, false)
			vector.DrawFilledRect(screen, baseX, baseY+35, 64, 40, leafColor, false)
			vector.DrawFilledRect(screen, baseX+5, baseY+10, 54, 35, leafColor, false)
			// Ствол
			trunkColor := color.RGBA{101, 67, 33, 255}
			vector.DrawFilledRect(screen, baseX+24, baseY+65, 16, 55, trunkColor, false)
		}
	}
}

// Enemy - враг (простое существо)
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
	AnimTimer   float64
}

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType string, ss *sprite.SpriteSheet) *Enemy {
	e := &Enemy{
		Transform: NewTransform(x, y, 32, 32),
		Physics:   NewPhysics(),
		Health:    NewHealth(30),
		EnemyType: enemyType,
		Damage:    10,
		Speed:     50,
	}

	e.Renderer = NewSpriteRenderer(ss)

	// Загрузка спрайта
	if ss != nil {
		var spriteName string
		switch enemyType {
		case "slime":
			spriteName = "slimeWalk1"
		case "snake":
			spriteName = "snakeWalk"
		case "spider":
			spriteName = "spider_walk1"
		}
		if img := ss.GetEnemySprite(spriteName); img != nil {
			e.Renderer.SetSprite(img)
			e.Transform.Width = float64(img.Bounds().Dx())
			e.Transform.Height = float64(img.Bounds().Dy())
		}
	}

	if !e.Renderer.UseSprite {
		e.Renderer.SpriteColor = color.RGBA{255, 68, 68, 255}
	}

	e.Health.Current = e.Health.Max
	if e.PatrolStart == 0 {
		e.PatrolStart = x - 60
		e.PatrolEnd = x + 60
	}

	return e
}

// Update обновляет врага
func (e *Enemy) Update(dt float64, playerX, playerY float64) {
	e.AnimTimer += dt

	if !e.Health.Dead {
		e.Renderer.Update(dt)
		e.updateAI(dt)
	}

	if e.Health.Invincible > 0 {
		e.Health.Invincible -= dt
	}
}

// updateAI обновляет ИИ врага
func (e *Enemy) updateAI(dt float64) {
	if e.Transform.X <= e.PatrolStart {
		e.Physics.VelocityX = e.Speed
	} else if e.Transform.X >= e.PatrolEnd {
		e.Physics.VelocityX = -e.Speed
	} else {
		e.Physics.VelocityX = e.Speed
	}

	e.Transform.X += e.Physics.VelocityX * dt
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if e.Health.Dead {
		return
	}

	if e.Health.Invincible > 0 && int(e.Health.Invincible*10)%2 == 0 {
		return
	}

	e.Renderer.Draw(screen, e.Transform, cameraX, cameraY)
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

	if ss != nil {
		p.Sprite = ss.GetTile(tileType)
	}

	if p.Sprite == nil {
		switch tileType {
		case "grass":
			p.Color = color.RGBA{100, 180, 80, 255}
		case "grassHalf":
			p.Color = color.RGBA{120, 200, 100, 255}
		case "dirt":
			p.Color = color.RGBA{139, 90, 43, 255}
		default:
			p.Color = color.RGBA{100, 100, 100, 255}
		}
	}

	return p
}

// Draw отрисовывает платформу
func (p *Platform) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if p.Sprite != nil {
		tileSize := float64(p.Sprite.Bounds().Dx())
		for x := p.Transform.X; x < p.Transform.X+p.Transform.Width; x += tileSize {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(x-cameraX, p.Transform.Y-cameraY)
			screen.DrawImage(p.Sprite, opts)
		}
	} else if p.Color != nil {
		vector.DrawFilledRect(screen, float32(p.Transform.X-cameraX), float32(p.Transform.Y-cameraY), float32(p.Transform.Width), float32(p.Transform.Height), p.Color, false)
	}
}

// Item - предмет (монетка)
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
	ItemStar       = "star"
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

	var spriteName string
	switch itemType {
	case ItemCoinGold:
		spriteName = "coinGold"
	case ItemCoinSilver:
		spriteName = "coinSilver"
	case ItemStar:
		spriteName = "star"
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
		case ItemStar:
			i.Renderer.SpriteColor = color.RGBA{255, 255, 255, 255}
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

// Collectible - собираемый предмет (звезда, кристалл)
type Collectible struct {
	Transform   *Transform
	Renderer    *SpriteRenderer
	ItemType    string
	Value       int
	Collected   bool
	AnimTimer   float64
	FloatOffset float64
}

// NewCollectible создаёт новый собираемый предмет
func NewCollectible(x, y float64, itemType string, value int, ss *sprite.SpriteSheet) *Collectible {
	c := &Collectible{
		Transform: &Transform{
			X: x, Y: y, Width: 32, Height: 32,
			ScaleX: 1, ScaleY: 1, Facing: 1,
		},
		ItemType:  itemType,
		Value:     value,
	}

	c.Renderer = NewSpriteRenderer(ss)

	if ss != nil {
		if img := ss.GetItem(itemType); img != nil {
			c.Renderer.SetSprite(img)
			c.Transform.Width = float64(img.Bounds().Dx())
			c.Transform.Height = float64(img.Bounds().Dy())
		}
	}

	if !c.Renderer.UseSprite {
		c.Renderer.SpriteColor = color.RGBA{255, 255, 255, 255}
	}

	return c
}

// Update обновляет предмет
func (c *Collectible) Update(dt float64) {
	c.AnimTimer += dt
	c.FloatOffset = math.Sin(c.AnimTimer*3) * 5
	c.Renderer.Update(dt)
}

// Draw отрисовывает предмет
func (c *Collectible) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if c.Collected {
		return
	}

	origY := c.Transform.Y
	c.Transform.Y += c.FloatOffset
	c.Renderer.Draw(screen, c.Transform, cameraX, cameraY)
	c.Transform.Y = origY
}

// Projectile - снаряд
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
