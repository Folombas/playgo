// Package entity - пиксельные сущности для Pixel Platformer
// Go365 Day 93 - Полностью пиксельная игра!
package entity

import (
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"city_platformer/internal/sprite"
)

// Transform - компонент позиции
type Transform struct {
	X, Y     float64
	Width    float64
	Height   float64
	VX, VY   float64
	Facing   int // 1 = вправо, -1 = влево
}

// NewTransform создаёт Transform
func NewTransform(x, y, w, h float64) *Transform {
	return &Transform{
		X: x, Y: y, Width: w, Height: h,
		Facing: 1,
	}
}

// SpriteRenderer - рендерер пиксельных спрайтов
type SpriteRenderer struct {
	SpriteSheet *sprite.SpriteSheet
	CurrentImg  *ebiten.Image
	AnimFrames  []*ebiten.Image
	AnimFrame   int
	AnimTimer   float64
	AnimFPS     float64
	AnimPlaying bool
	AnimLoop    bool
}

// NewSpriteRenderer создаёт рендерер
func NewSpriteRenderer(ss *sprite.SpriteSheet) *SpriteRenderer {
	return &SpriteRenderer{
		SpriteSheet: ss,
		AnimFPS:     8,
		AnimLoop:    true,
	}
}

// SetSprite устанавливает спрайт
func (sr *SpriteRenderer) SetSprite(img *ebiten.Image) {
	sr.CurrentImg = img
	sr.AnimPlaying = false
}

// SetAnim устанавливает анимацию
func (sr *SpriteRenderer) SetAnim(frames []*ebiten.Image) {
	if len(frames) > 0 {
		validFrames := make([]*ebiten.Image, 0)
		for _, f := range frames {
			if f != nil {
				validFrames = append(validFrames, f)
			}
		}
		if len(validFrames) > 0 {
			sr.AnimFrames = validFrames
			sr.CurrentImg = validFrames[0]
			sr.AnimFrame = 0
			sr.AnimTimer = 0
			sr.AnimPlaying = true
		}
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

// Draw отрисовывает пиксельный спрайт
func (sr *SpriteRenderer) Draw(screen *ebiten.Image, transform *Transform, cameraX, cameraY float64) {
	if sr.CurrentImg == nil {
		// Пиксельная заглушка вместо векторов
		x := transform.X - cameraX
		y := transform.Y - cameraY
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(transform.Width), float32(transform.Height), color.RGBA{255, 0, 255, 255}, false)
		return
	}

	opts := &ebiten.DrawImageOptions{}
	// Масштабирование для пиксель-арта (без сглаживания)
	opts.GeoM.Scale(transform.Width/float64(sr.CurrentImg.Bounds().Dx()), transform.Height/float64(sr.CurrentImg.Bounds().Dy()))

	if transform.Facing == -1 {
		opts.GeoM.Scale(-1, 1)
		opts.GeoM.Translate(float64(transform.Width), 0)
	}

	opts.GeoM.Translate(transform.X-cameraX, transform.Y-cameraY)
	screen.DrawImage(sr.CurrentImg, opts)
}

// Physics - физика
type Physics struct {
	VelocityX    float64
	VelocityY    float64
	Gravity      float64
	Friction     float64
	OnGround     bool
}

// NewPhysics создаёт физику
func NewPhysics() *Physics {
	return &Physics{
		Gravity:  1000,
		Friction: 0.85,
	}
}

// Health - здоровье
type Health struct {
	Current    int
	Max        int
	Invincible float64
	Dead       bool
}

// NewHealth создаёт здоровье
func NewHealth(max int) *Health {
	return &Health{Current: max, Max: max}
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

// Player - игрок
type Player struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	Physics   *Physics
	Health    *Health
	State     PlayerState
	JumpCount int
	MaxJumps  int
	AnimTimer float64
}

type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
)

// NewPlayer создаёт игрока
func NewPlayer(x, y float64, ss *sprite.SpriteSheet) *Player {
	p := &Player{
		Transform: NewTransform(x, y, 32, 32),
		Physics:   NewPhysics(),
		Health:    NewHealth(100),
		MaxJumps:  2,
	}

	p.Renderer = NewSpriteRenderer(ss)

	// Загрузка пиксельных спрайтов
	if ss != nil {
		if stand := ss.GetPlayerSprite("stand"); stand != nil {
			p.Renderer.SetSprite(stand)
			p.Transform.Width = float64(stand.Bounds().Dx())
			p.Transform.Height = float64(stand.Bounds().Dy())
		}
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

	if p.Health.Invincible > 0 {
		p.Health.Invincible -= dt
	}

	p.Renderer.Update(dt)
	p.updateState()
}

func (p *Player) updateState() {
	if !p.Physics.OnGround {
		p.State = PlayerJumping
	} else if p.Physics.VelocityX != 0 {
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

// Stop останавливается
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

// ResetJump сбрасывает прыжок
func (p *Player) ResetJump() {
	p.JumpCount = 0
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if p.Health.Dead {
		return
	}

	// Мигание при уроне
	if p.Health.Invincible > 0 && int(p.Health.Invincible*10)%2 == 0 {
		return
	}

	// Выбор спрайта
	if p.Renderer.SpriteSheet != nil {
		switch p.State {
		case PlayerRunning:
			if len(p.Renderer.SpriteSheet.PlayerWalk) > 0 {
				p.Renderer.SetAnim(p.Renderer.SpriteSheet.PlayerWalk)
			}
		case PlayerJumping:
			if jump := p.Renderer.SpriteSheet.GetPlayerSprite("jump"); jump != nil {
				p.Renderer.SetSprite(jump)
			}
		case PlayerIdle:
			if stand := p.Renderer.SpriteSheet.GetPlayerSprite("stand"); stand != nil {
				p.Renderer.SetSprite(stand)
			}
		}
	}

	p.Renderer.Draw(screen, p.Transform, cameraX, cameraY)
}

// House - пиксельный домик
type House struct {
	Transform  *Transform
	HouseType  string
	SpriteParts map[string]*ebiten.Image
	SmokeTimer float64
	SmokeParts []SmokeParticle
}

type SmokeParticle struct {
	X, Y float64
	VY   float64
	Life float64
	Size float64
}

// NewHouse создаёт домик
func NewHouse(x, y float64, houseType string, ss *sprite.SpriteSheet) *House {
	h := &House{
		Transform:   NewTransform(x, y-96, 96, 96),
		HouseType:   houseType,
		SpriteParts: make(map[string]*ebiten.Image),
	}

	// Загрузка частей домика
	if ss != nil {
		parts := []string{"TopLeft", "TopMid", "TopRight", "MidLeft", "MidRight", "BottomLeft", "BottomMid", "BottomRight"}
		for _, part := range parts {
			h.SpriteParts[part] = ss.GetHousePart(houseType, part)
		}
	}

	return h
}

// Update обновляет домик (дым)
func (h *House) Update(dt float64) {
	h.SmokeTimer += dt
	if h.SmokeTimer >= 0.5 {
		h.SmokeTimer = 0
		h.SmokeParts = append(h.SmokeParts, SmokeParticle{
			X: h.Transform.X + h.Transform.Width/2 - 5,
			Y: h.Transform.Y - 10,
			VY: -20,
			Life: 2.0,
			Size: 8,
		})
	}

	active := make([]SmokeParticle, 0)
	for i := range h.SmokeParts {
		s := &h.SmokeParts[i]
		s.Y += s.VY * dt
		s.X += math.Sin(s.Life*2) * 5 * dt
		s.Life -= dt
		s.Size += 2 * dt
		if s.Life > 0 {
			active = append(active, *s)
		}
	}
	h.SmokeParts = active
}

// Draw отрисовывает домик
func (h *House) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	// Дым
	for _, s := range h.SmokeParts {
		alpha := uint8(s.Life * 150)
		c := color.RGBA{200, 200, 200, alpha}
		vector.DrawFilledRect(screen, float32(s.X-cameraX), float32(s.Y-cameraY), float32(s.Size), float32(s.Size), c, false)
	}

	// Домик по частям
	if h.SpriteParts["BottomLeft"] != nil {
		tileSize := 32.0
		parts := [][3]string{
			{"BottomLeft", "BottomMid", "BottomRight"},
			{"MidLeft", "", "MidRight"},
			{"TopLeft", "TopMid", "TopRight"},
		}

		for row, partRow := range parts {
			for col, part := range partRow {
				if part != "" && h.SpriteParts[part] != nil {
					opts := &ebiten.DrawImageOptions{}
					opts.GeoM.Translate(h.Transform.X-cameraX+float64(col)*tileSize, h.Transform.Y-cameraY+float64(row)*tileSize)
					screen.DrawImage(h.SpriteParts[part], opts)
				}
			}
		}
	}
}

// Tree - пиксельное дерево
type Tree struct {
	Transform *Transform
	TreeType  string
	Sprite    *ebiten.Image
}

// NewTree создаёт дерево
func NewTree(x, y float64, treeType string, ss *sprite.SpriteSheet) *Tree {
	t := &Tree{
		Transform: NewTransform(x, y-120, 64, 120),
		TreeType:  treeType,
	}

	if ss != nil {
		if treeType == "pine" {
			// Сосна
			t.Sprite = ss.GetTree("pine_0")
			if t.Sprite != nil {
				t.Transform.Width = float64(t.Sprite.Bounds().Dx())
				t.Transform.Height = float64(t.Sprite.Bounds().Dy())
			}
		} else {
			// Дуб
			t.Sprite = ss.GetTree("tree")
			if t.Sprite != nil {
				t.Transform.Width = float64(t.Sprite.Bounds().Dx())
				t.Transform.Height = float64(t.Sprite.Bounds().Dy())
			}
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
	}
}

// Enemy - пиксельный враг
type Enemy struct {
	Transform   *Transform
	Renderer    *SpriteRenderer
	Physics     *Physics
	Health      *Health
	EnemyType   string
	Damage      int
	Speed       float64
	PatrolStart float64
	PatrolEnd   float64
}

// NewEnemy создаёт врага
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
		case "bat":
			spriteName = "bat_fly"
		}
		if img := ss.GetEnemySprite(spriteName); img != nil {
			e.Renderer.SetSprite(img)
			e.Transform.Width = float64(img.Bounds().Dx())
			e.Transform.Height = float64(img.Bounds().Dy())
		}
	}

	e.Health.Current = e.Health.Max
	e.PatrolStart = x - 60
	e.PatrolEnd = x + 60

	return e
}

// Update обновляет врага
func (e *Enemy) Update(dt float64, playerX, playerY float64) {
	if e.Transform.X <= e.PatrolStart {
		e.Physics.VelocityX = e.Speed
	} else if e.Transform.X >= e.PatrolEnd {
		e.Physics.VelocityX = -e.Speed
	}

	e.Transform.X += e.Physics.VelocityX * dt
	e.Renderer.Update(dt)

	if e.Health.Invincible > 0 {
		e.Health.Invincible -= dt
	}
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
}

// NewPlatform создаёт платформу
func NewPlatform(x, y, width, height float64, tileType string, ss *sprite.SpriteSheet) *Platform {
	p := &Platform{
		Transform: NewTransform(x, y, width, height),
		TileType:  tileType,
	}

	if ss != nil {
		p.Sprite = ss.GetTile(tileType)
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
	ItemKeyBlue    = "keyBlue"
	ItemKeyGreen   = "keyGreen"
	ItemKeyRed     = "keyRed"
	ItemKeyYellow  = "keyYellow"
	ItemBomb       = "bomb"
	ItemFlag       = "flagBlue"
	ItemPlant      = "plant"
	ItemSpikes     = "spikes"
	ItemRock       = "rock"
)

// NewItem создаёт предмет
func NewItem(x, y float64, itemType string, value int, ss *sprite.SpriteSheet) *Item {
	i := &Item{
		Transform: NewTransform(x, y, 32, 32),
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
	case ItemKeyBlue, ItemKeyGreen, ItemKeyRed, ItemKeyYellow:
		spriteName = itemType
	case ItemBomb:
		spriteName = "bomb"
	case ItemFlag:
		spriteName = "flagBlue"
	case ItemPlant:
		spriteName = "plant"
	case ItemSpikes:
		spriteName = "spikes"
	case ItemRock:
		spriteName = "rock"
	default:
		spriteName = itemType
	}

	// Пробуем загрузить спрайт
	if ss != nil {
		img := ss.GetItem(spriteName)
		if img != nil {
			i.Renderer.SetSprite(img)
			i.Transform.Width = float64(img.Bounds().Dx())
			i.Transform.Height = float64(img.Bounds().Dy())
		} else {
			// Пробуем альтернативное имя
			altName := strings.ToLower(spriteName[:1]) + spriteName[1:]
			img = ss.GetItem(altName)
			if img != nil {
				i.Renderer.SetSprite(img)
				i.Transform.Width = float64(img.Bounds().Dx())
				i.Transform.Height = float64(img.Bounds().Dy())
			}
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

// Collectible - собираемый предмет
type Collectible struct {
	Transform   *Transform
	Renderer    *SpriteRenderer
	ItemType    string
	Value       int
	Collected   bool
	AnimTimer   float64
	FloatOffset float64
}

// Mushroom - гриб (декорация или предмет)
type Mushroom struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	MushType  string // red, brown, tan, tall
	AnimTimer float64
}

// Frog - лягушка (декорация)
type Frog struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	AnimTimer float64
	JumpY     float64
}

// Butterfly - бабочка (декорация, летает)
type Butterfly struct {
	Transform *Transform
	Renderer  *SpriteRenderer
	AnimTimer float64
	FloatY    float64
	FloatX    float64
}

// Cactus - кактус (декорация, препятствие)
type Cactus struct {
	Transform *Transform
	Renderer  *SpriteRenderer
}

// NewCollectible создаёт предмет
func NewCollectible(x, y float64, itemType string, value int, ss *sprite.SpriteSheet) *Collectible {
	c := &Collectible{
		Transform: NewTransform(x, y, 32, 32),
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

// NewMushroom создаёт гриб
func NewMushroom(x, y float64, mushType string, ss *sprite.SpriteSheet) *Mushroom {
	m := &Mushroom{
		Transform: NewTransform(x, y, 32, 32),
		MushType:  mushType,
	}

	m.Renderer = NewSpriteRenderer(ss)

	if ss != nil {
		var spriteName string
		switch mushType {
		case "red":
			spriteName = "shroomRedMid"
		case "brown":
			spriteName = "shroomBrownMid"
		case "tan":
			spriteName = "shroomTanMid"
		case "tall_red":
			spriteName = "tallShroom_red"
		case "tall_brown":
			spriteName = "tallShroom_brown"
		case "tiny_red":
			spriteName = "tinyShroom_red"
		}
		if img := ss.GetMushroom(spriteName); img != nil {
			m.Renderer.SetSprite(img)
			m.Transform.Width = float64(img.Bounds().Dx())
			m.Transform.Height = float64(img.Bounds().Dy())
		}
	}

	return m
}

// Update обновляет гриб
func (m *Mushroom) Update(dt float64) {
	m.AnimTimer += dt
}

// Draw отрисовывает гриб
func (m *Mushroom) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	m.Renderer.Draw(screen, m.Transform, cameraX, cameraY)
}

// NewFrog создаёт лягушку
func NewFrog(x, y float64, ss *sprite.SpriteSheet) *Frog {
	f := &Frog{
		Transform: NewTransform(x, y, 32, 32),
	}

	f.Renderer = NewSpriteRenderer(ss)

	if ss != nil {
		if img := ss.GetDetail("GrassLand_Frog.png"); img != nil {
			f.Renderer.SetSprite(img)
			f.Transform.Width = float64(img.Bounds().Dx())
			f.Transform.Height = float64(img.Bounds().Dy())
		}
	}

	return f
}

// Update обновляет лягушку
func (f *Frog) Update(dt float64) {
	f.AnimTimer += dt
	f.JumpY = math.Sin(f.AnimTimer*2) * 3
}

// Draw отрисовывает лягушку
func (f *Frog) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	origY := f.Transform.Y
	f.Transform.Y += f.JumpY
	f.Renderer.Draw(screen, f.Transform, cameraX, cameraY)
	f.Transform.Y = origY
}

// NewButterfly создаёт бабочку
func NewButterfly(x, y float64, ss *sprite.SpriteSheet) *Butterfly {
	b := &Butterfly{
		Transform: NewTransform(x, y, 24, 24),
	}

	b.Renderer = NewSpriteRenderer(ss)

	if ss != nil {
		if img := ss.GetDetail("GrassLand_Butterfly.png"); img != nil {
			b.Renderer.SetSprite(img)
			b.Transform.Width = float64(img.Bounds().Dx())
			b.Transform.Height = float64(img.Bounds().Dy())
		}
	}

	return b
}

// Update обновляет бабочку
func (b *Butterfly) Update(dt float64) {
	b.AnimTimer += dt
	b.FloatY = math.Sin(b.AnimTimer*3) * 20
	b.FloatX = math.Cos(b.AnimTimer*2) * 10
	b.Transform.X += b.FloatX * dt
	b.Transform.Y += b.FloatY * dt
}

// Draw отрисовывает бабочку
func (b *Butterfly) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	b.Renderer.Draw(screen, b.Transform, cameraX, cameraY)
}

// NewCactus создаёт кактус
func NewCactus(x, y float64, ss *sprite.SpriteSheet) *Cactus {
	c := &Cactus{
		Transform: NewTransform(x, y-40, 24, 40),
	}

	c.Renderer = NewSpriteRenderer(ss)

	if ss != nil {
		if img := ss.GetItem("cactus"); img != nil {
			c.Renderer.SetSprite(img)
			c.Transform.Width = float64(img.Bounds().Dx())
			c.Transform.Height = float64(img.Bounds().Dy())
		}
	}

	return c
}

// Draw отрисовывает кактус
func (c *Cactus) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	c.Renderer.Draw(screen, c.Transform, cameraX, cameraY)
}

// CheckCollision проверяет коллизию
func CheckCollision(a, b *Transform) bool {
	return a.X < b.X+b.Width && a.X+a.Width > b.X && a.Y < b.Y+b.Height && a.Y+a.Height > b.Y
}
