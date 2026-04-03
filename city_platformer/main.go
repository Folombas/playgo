package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// ─── SPRITE LOADER ────────────────────────────────────────────────

var sprites = make(map[string]*ebiten.Image)

func loadSprite(path string) *ebiten.Image {
	if img, ok := sprites[path]; ok {
		return img
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil
	}
	eimg := ebiten.NewImageFromImage(img)
	sprites[path] = eimg
	return eimg
}

// ─── GAME STATE ───────────────────────────────────────────────────

type State int

const (
	Title State = iota
	Playing
	Dead
	Win
)

// ─── ENTITY ───────────────────────────────────────────────────────

type Entity struct {
	X, Y       float64
	W, H       float64
	VX, VY     float64
	HP         int
	MaxHP      int
	Sprite     *ebiten.Image
	AnimFrame  int
	AnimTimer  int
	Alive      bool
}

func (e *Entity) Draw(screen *ebiten.Image, fallback color.Color) {
	if e.Sprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X, e.Y)
		screen.DrawImage(e.Sprite, op)
	} else if fallback != nil {
		img := ebiten.NewImage(int(e.W), int(e.H))
		img.Fill(fallback)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X, e.Y)
		screen.DrawImage(img, op)
	}
}

func (a *Entity) Collides(b *Entity) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

// ─── PLAYER ───────────────────────────────────────────────────────

type Player struct {
	Entity
	Speed      float64
	AttackCD   int
	Attacking  bool
	AttackDir  int // 0=right,1=left,2=up,3=down
	Facing     int // same as AttackDir
	Keys       int
	Score      int
	Invincible int
}

func NewPlayer(x, y float64) *Player {
	p := &Player{
		Entity: Entity{X: x, Y: y, W: 32, H: 32, HP: 5, MaxHP: 5, Alive: true},
		Speed:  200,
	}
	p.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Player/p1_stand.png")
	return p
}

func (p *Player) Update(keys []ebiten.Key, dt float64, room *Room) {
	p.VX = 0
	p.VY = 0
	if p.HP <= 0 {
		p.Alive = false
		return
	}

	for _, k := range keys {
		switch k {
		case ebiten.KeyA, ebiten.KeyLeft:
			p.VX = -p.Speed
			p.Facing = 1
		case ebiten.KeyD, ebiten.KeyRight:
			p.VX = p.Speed
			p.Facing = 0
		case ebiten.KeyW, ebiten.KeyUp:
			p.VY = -p.Speed
			p.Facing = 2
		case ebiten.KeyS, ebiten.KeyDown:
			p.VY = p.Speed
			p.Facing = 3
		}
	}

	// Normalize diagonal
	if p.VX != 0 && p.VY != 0 {
		p.VX *= 0.707
		p.VY *= 0.707
	}

	p.X += p.VX * dt
	p.Y += p.VY * dt

	// Room bounds
	p.X = math.Max(0, math.Min(float64(room.W)-p.W, p.X))
	p.Y = math.Max(0, math.Min(float64(room.H)-p.H, p.Y))

	// Wall collision
	for _, w := range room.Walls {
		if p.Collides(&w.Entity) {
			// Push back
			p.X -= p.VX * dt
			p.Y -= p.VY * dt
		}
	}

	// Attack
	if p.AttackCD > 0 {
		p.AttackCD--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		if p.AttackCD <= 0 {
			p.Attacking = true
			p.AttackDir = p.Facing
			p.AttackCD = 20
		}
	}
	if p.Attacking {
		p.AttackCD--
		if p.AttackCD < 10 {
			p.Attacking = false
		}
	}

	if p.Invincible > 0 {
		p.Invincible--
	}
}

func (p *Player) AttackBox() *Entity {
	if !p.Attacking {
		return nil
	}
	bx, by, bw, bh := p.X, p.Y, p.W, p.H
	switch p.AttackDir {
	case 0: // right
		bx = p.X + p.W
		bw = 24
	case 1: // left
		bx = p.X - 24
		bw = 24
	case 2: // up
		by = p.Y - 24
		bh = 24
	case 3: // down
		by = p.Y + p.H
		bh = 24
	}
	return &Entity{X: bx, Y: by, W: bw, H: bh}
}

func (p *Player) TakeDamage(dmg int) {
	if p.Invincible > 0 {
		return
	}
	p.HP -= dmg
	p.Invincible = 40
	if p.HP < 0 {
		p.HP = 0
	}
}

// ─── ENEMY ────────────────────────────────────────────────────────

type EnemyType int

const (
	Slime EnemyType = iota
	Bat
	Fish
	Poker
)

type Enemy struct {
	Entity
	Type     EnemyType
	Speed    float64
	Damage   int
	ChaseRange float64
}

func NewEnemy(x, y float64, et EnemyType) *Enemy {
	e := &Enemy{
		Entity:     Entity{X: x, Y: y, W: 32, H: 32, HP: 2, MaxHP: 2, Alive: true},
		Type:       et,
		Speed:      60,
		Damage:     1,
		ChaseRange: 250,
	}
	switch et {
	case Slime:
		e.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Enemies/slimeWalk1.png")
		e.Speed = 40
		e.HP = 2
	case Bat:
		e.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Enemies/flyFly1.png")
		e.Speed = 80
		e.HP = 1
		e.ChaseRange = 350
	case Fish:
		e.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Enemies/fishSwim1.png")
		e.Speed = 50
		e.HP = 3
	case Poker:
		e.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Enemies/pokerMad.png")
		e.Speed = 70
		e.HP = 4
		e.Damage = 2
		e.W, e.H = 40, 40
	}
	e.MaxHP = e.HP
	return e
}

func (e *Enemy) Update(player *Player, dt float64) {
	if !e.Alive {
		return
	}

	dx := player.X - e.X
	dy := player.Y - e.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < e.ChaseRange {
		// Chase player
		nx := dx / dist
		ny := dy / dist
		e.X += nx * e.Speed * dt
		e.Y += ny * e.Speed * dt
	}

	// Animation
	e.AnimTimer++
	if e.AnimTimer > 15 {
		e.AnimTimer = 0
		e.AnimFrame = 1 - e.AnimFrame
	}
}

func (e *Enemy) TakeDamage(dmg int) {
	e.HP -= dmg
	if e.HP <= 0 {
		e.Alive = false
	}
}

// ─── ITEM ─────────────────────────────────────────────────────────

type ItemType int

const (
	Coin ItemType = iota
	Heart
	Key
	Gem
)

type Item struct {
	X, Y     float64
	W, H     float64
	Type     ItemType
	Value    int
	Collected bool
	BobTimer float64
	Sprite   *ebiten.Image
}

func NewItem(x, y float64, it ItemType) *Item {
	i := &Item{X: x, Y: y, W: 16, H: 16, Type: it}
	switch it {
	case Coin:
		i.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Items/coinGold.png")
		i.Value = 10
	case Heart:
		i.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Items/heart.png")
		i.Value = 1
	case Key:
		i.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Items/key.png")
		i.Value = 1
	case Gem:
		i.Sprite = loadSprite("assets/sprites/PlatformerComplete/Base pack/Items/gemBlue.png")
		i.Value = 50
		i.W, i.H = 20, 20
	}
	return i
}

func (i *Item) Update(dt float64) {
	i.BobTimer += dt * 3
}

// ─── ROOM ─────────────────────────────────────────────────────────

type Wall struct {
	Entity
}

type Door struct {
	X, Y    float64
	W, H    float64
	Open    bool
	Dir     int // 0=right, 1=bottom
}

type Room struct {
	X, Y, W, H int
	Enemies    []*Enemy
	Items      []*Item
	Walls      []*Wall
	Doors      []*Door
	Cleared    bool
	BG         *ebiten.Image
}

func NewRoom(x, y, w, h int, difficulty int) *Room {
	r := &Room{X: x, Y: y, W: w, H: h}
	r.BG = loadSprite("assets/sprites/PlatformerComplete/Base pack/bg_castle.png")

	// Generate walls (pillars)
	numWalls := rand.Intn(3) + difficulty
	for i := 0; i < numWalls; i++ {
		wx := 80 + rand.Intn(w-160)
		wy := 80 + rand.Intn(h-160)
		r.Walls = append(r.Walls, &Wall{
			Entity: Entity{X: float64(wx), Y: float64(wy), W: 32, H: 32},
		})
	}

	// Generate enemies
	numEnemies := rand.Intn(2) + difficulty
	for i := 0; i < numEnemies; i++ {
		ex := 100 + rand.Intn(w-200)
		ey := 100 + rand.Intn(h-200)
		et := EnemyType(rand.Intn(4))
		r.Enemies = append(r.Enemies, NewEnemy(float64(ex), float64(ey), et))
	}

	// Generate items
	numCoins := rand.Intn(4) + 2
	for i := 0; i < numCoins; i++ {
		ix := 50 + rand.Intn(w-100)
		iy := 50 + rand.Intn(h-100)
		r.Items = append(r.Items, NewItem(float64(ix), float64(iy), Coin))
	}

	// Gems (rare)
	if rand.Float64() < 0.3 {
		gx := 100 + rand.Intn(w-200)
		gy := 100 + rand.Intn(h-200)
		r.Items = append(r.Items, NewItem(float64(gx), float64(gy), Gem))
	}

	// Heart (if low HP chance)
	if rand.Float64() < 0.25 {
		hx := 100 + rand.Intn(w-200)
		hy := 100 + rand.Intn(h-200)
		r.Items = append(r.Items, NewItem(float64(hx), float64(hy), Heart))
	}

	// Key (needed for next room)
	if difficulty > 0 {
		kx := 100 + rand.Intn(w-200)
		ky := 100 + rand.Intn(h-200)
		r.Items = append(r.Items, NewItem(float64(kx), float64(ky), Key))
	}

	// Doors
	if difficulty < 5 {
		r.Doors = append(r.Doors, &Door{X: float64(w - 20), Y: float64(h/2 - 30), W: 20, H: 60, Dir: 0})
	}

	return r
}

func (r *Room) Update(dt float64, player *Player) {
	// Check if room cleared
	allDead := true
	for _, e := range r.Enemies {
		if e.Alive {
			allDead = false
			break
		}
	}
	r.Cleared = allDead

	// Update enemies
	for _, e := range r.Enemies {
		e.Update(player, dt)
		// Wall collision
		for _, w := range r.Walls {
			if e.Collides(&w.Entity) {
				// Push back
				dx := e.X - w.X
				dy := e.Y - w.Y
				if math.Abs(dx) > math.Abs(dy) {
					if dx > 0 {
						e.X = w.X + w.W
					} else {
						e.X = w.X - e.W
					}
				} else {
					if dy > 0 {
						e.Y = w.Y + w.H
					} else {
						e.Y = w.Y - e.H
					}
				}
			}
		}
		// Player collision
		if e.Alive && player.Collides(&e.Entity) {
			player.TakeDamage(e.Damage)
		}
	}

	// Update items
	for _, item := range r.Items {
		if item.Collected {
			continue
		}
		item.Update(dt)
		// Player pickup
		if player.X < item.X+item.W && player.X+player.W > item.X &&
			player.Y < item.Y+item.H && player.Y+player.H > item.Y {
			item.Collected = true
			player.Score += item.Value
			switch item.Type {
			case Heart:
				if player.HP < player.MaxHP {
					player.HP += item.Value
				}
			case Key:
				player.Keys += item.Value
			}
		}
	}
}

func (r *Room) Draw(screen *ebiten.Image) {
	// Background
	if r.BG != nil {
		// Tile the background
		for bx := 0; bx < r.W; bx += r.BG.Bounds().Dx() {
			for by := 0; by < r.H; by += r.BG.Bounds().Dy() {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(bx), float64(by))
				screen.DrawImage(r.BG, op)
			}
		}
	} else {
		bg := ebiten.NewImage(r.W, r.H)
		bg.Fill(color.RGBA{40, 40, 60, 255})
		screen.DrawImage(bg, nil)
	}

	// Walls
	for _, w := range r.Walls {
		w.Draw(screen, color.RGBA{100, 100, 120, 255})
	}

	// Items
	for _, item := range r.Items {
		if item.Collected {
			continue
		}
		bob := math.Sin(item.BobTimer) * 3
		if item.Sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(item.X, item.Y+bob)
			screen.DrawImage(item.Sprite, op)
		} else {
			c := color.RGBA{255, 215, 0, 255}
			if item.Type == Heart {
				c = color.RGBA{255, 50, 50, 255}
			} else if item.Type == Key {
				c = color.RGBA{255, 255, 0, 255}
			} else if item.Type == Gem {
				c = color.RGBA{0, 150, 255, 255}
			}
			img := ebiten.NewImage(int(item.W), int(item.H))
			img.Fill(c)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(item.X, item.Y+bob)
			screen.DrawImage(img, op)
		}
	}

	// Enemies
	for _, e := range r.Enemies {
		if !e.Alive {
			continue
		}
		c := color.RGBA{100, 200, 100, 255}
		if e.Type == Bat {
			c = color.RGBA{150, 100, 200, 255}
		} else if e.Type == Fish {
			c = color.RGBA{100, 150, 255, 255}
		} else if e.Type == Poker {
			c = color.RGBA{255, 100, 50, 255}
		}
		e.Draw(screen, c)
	}

	// Doors
	for _, d := range r.Doors {
		c := color.RGBA{150, 100, 50, 255}
		if d.Open {
			c = color.RGBA{50, 150, 50, 255}
		}
		img := ebiten.NewImage(int(d.W), int(d.H))
		img.Fill(c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(d.X, d.Y)
		screen.DrawImage(img, op)
	}
}

// ─── PARTICLE ─────────────────────────────────────────────────────

type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife int
	R, G, B      uint8
}

// ─── GAME ─────────────────────────────────────────────────────────

type Game struct {
	State     State
	Player    *Player
	Rooms     []*Room
	CurrentRoom int
	Particles []Particle
	ShakeTimer int
}

func NewGame() *Game {
	g := &Game{
		State: Title,
	}
	return g
}

func (g *Game) startGame() {
	g.Player = NewPlayer(64, 340)
	g.Rooms = nil
	g.CurrentRoom = 0
	g.Particles = nil
	g.ShakeTimer = 0

	// Generate 5 rooms
	for i := 0; i < 5; i++ {
		r := NewRoom(0, 0, 1280, 720, i+1)
		g.Rooms = append(g.Rooms, r)
	}
}

func (g *Game) Update() error {
	switch g.State {
	case Title:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.State = Playing
			g.startGame()
		}
	case Playing:
		g.updatePlaying()
	case Dead:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.State = Playing
			g.startGame()
		}
	case Win:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.State = Title
		}
	}
	return nil
}

func (g *Game) updatePlaying() {
	dt := 1.0 / 60.0
	room := g.Rooms[g.CurrentRoom]

	// Get pressed keys
	var keys []ebiten.Key
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		keys = append(keys, ebiten.KeyA)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		keys = append(keys, ebiten.KeyD)
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		keys = append(keys, ebiten.KeyW)
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		keys = append(keys, ebiten.KeyS)
	}

	g.Player.Update(keys, dt, room)

	if !g.Player.Alive {
		g.State = Dead
		return
	}

	room.Update(dt, g.Player)

	// Player attack
	atkBox := g.Player.AttackBox()
	if atkBox != nil {
		for _, e := range room.Enemies {
			if !e.Alive {
				continue
			}
			if atkBox.Collides(&e.Entity) {
				e.TakeDamage(1)
				g.spawnParticles(e.X+e.W/2, e.Y+e.H/2, 8, 255, 255, 0)
				if !e.Alive {
					g.Player.Score += 50
					g.ShakeTimer = 5
				}
			}
		}
	}

	// Check door interaction
	for _, d := range room.Doors {
		if d.Open {
			continue
		}
		if g.Player.X+g.Player.W > d.X && g.Player.X < d.X+d.W &&
			g.Player.Y+g.Player.H > d.Y && g.Player.Y < d.Y+d.H {
			if g.Player.Keys > 0 {
				d.Open = true
				g.Player.Keys--
				g.spawnParticles(d.X, d.Y+d.H/2, 15, 100, 255, 100)
			}
		}
		// Move to next room if door open and player near it
		if d.Open && g.Player.X > d.X {
			g.CurrentRoom++
			if g.CurrentRoom >= len(g.Rooms) {
				g.State = Win
				return
			}
			g.Player.X = 20
			g.Player.Y = float64(g.Rooms[g.CurrentRoom].H/2 - 16)
		}
	}

	// Update particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := &g.Particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 200 * dt
		p.Life--
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}

	if g.ShakeTimer > 0 {
		g.ShakeTimer--
	}
}

func (g *Game) spawnParticles(x, y float64, count int, r, gv, b uint8) {
	for i := 0; i < count; i++ {
		g.Particles = append(g.Particles, Particle{
			X: x, Y: y,
			VX: (rand.Float64() - 0.5) * 300,
			VY: (rand.Float64() - 0.5) * 300,
			Life: 20 + rand.Intn(20),
			MaxLife: 40,
			R: r, G: gv, B: b,
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 30, 255})

	switch g.State {
	case Title:
		g.drawTitle(screen)
	case Playing:
		g.drawGame(screen)
	case Dead:
		g.drawGame(screen)
		g.drawOverlay(screen, "YOU DIED", color.RGBA{255, 50, 50, 255})
	case Win:
		g.drawGame(screen)
		g.drawOverlay(screen, "YOU WIN!", color.RGBA{100, 255, 100, 255})
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 15, 30, 255})
	drawText(screen, "DUNGEON CRAWLER", 640, 200, color.RGBA{200, 150, 50, 255}, 24)
	drawText(screen, "Press ENTER or SPACE to Start", 640, 320, color.White, 14)
	drawText(screen, "WASD / Arrows = Move", 640, 400, color.RGBA{180, 180, 180, 255}, 12)
	drawText(screen, "SPACE / J = Attack", 640, 430, color.RGBA{180, 180, 180, 255}, 12)
	drawText(screen, "Kill enemies, collect keys, open doors!", 640, 480, color.RGBA{150, 150, 200, 255}, 12)
	drawText(screen, "5 rooms to clear", 640, 510, color.RGBA{150, 150, 200, 255}, 12)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	room := g.Rooms[g.CurrentRoom]

	// Screen shake
	op := &ebiten.DrawImageOptions{}
	if g.ShakeTimer > 0 {
		sx := (rand.Float64() - 0.5) * 6
		sy := (rand.Float64() - 0.5) * 6
		op.GeoM.Translate(sx, sy)
	}

	// Draw room
	room.Draw(screen)

	// Draw particles
	for _, p := range g.Particles {
		alpha := uint8(p.Life * 255 / p.MaxLife)
		img := ebiten.NewImage(5, 5)
		img.Fill(color.RGBA{p.R, p.G, p.B, alpha})
		pop := &ebiten.DrawImageOptions{}
		pop.GeoM.Translate(p.X, p.Y)
		screen.DrawImage(img, pop)
	}

	// Draw player
	if g.Player.Invincible <= 0 || g.Player.Invincible%4 < 2 {
		g.Player.Draw(screen, color.RGBA{0, 150, 255, 255})

		// Attack visual
		if g.Player.Attacking {
			ab := g.Player.AttackBox()
			if ab != nil {
				img := ebiten.NewImage(int(ab.W), int(ab.H))
				img.Fill(color.RGBA{255, 255, 200, 180})
				aop := &ebiten.DrawImageOptions{}
				aop.GeoM.Translate(ab.X, ab.Y)
				screen.DrawImage(img, aop)
			}
		}
	}

	// HUD
	drawText(screen, fmt.Sprintf("HP: %d/%d", g.Player.HP, g.Player.MaxHP), 20, 30, color.RGBA{255, 80, 80, 255}, 14)
	drawText(screen, fmt.Sprintf("Score: %d", g.Player.Score), 20, 55, color.White, 14)
	drawText(screen, fmt.Sprintf("Keys: %d", g.Player.Keys), 20, 80, color.RGBA{255, 255, 0, 255}, 14)
	drawText(screen, fmt.Sprintf("Room: %d/5", g.CurrentRoom+1), 20, 105, color.RGBA{150, 200, 255, 255}, 14)

	if room.Cleared {
		drawText(screen, "ROOM CLEARED!", 640, 100, color.RGBA{100, 255, 100, 255}, 16)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("Room %d | Enemies: %d | Keys: %d", g.CurrentRoom+1, countAlive(room.Enemies), g.Player.Keys))
}

func (g *Game) drawOverlay(screen *ebiten.Image, title string, c color.Color) {
	screen.Fill(color.RGBA{0, 0, 0, 180})
	drawText(screen, title, 640, 280, c, 24)
	drawText(screen, fmt.Sprintf("Final Score: %d", g.Player.Score), 640, 340, color.White, 16)
	drawText(screen, "Press ENTER", 640, 400, color.RGBA{150, 200, 255, 255}, 14)
}

func drawText(screen *ebiten.Image, msg string, x, y int, c color.Color, scale int) {
	text.Draw(screen, msg, basicfont.Face7x13, x-len(msg)*7/2, y, c)
}

func countAlive(enemies []*Enemy) int {
	n := 0
	for _, e := range enemies {
		if e.Alive {
			n++
		}
	}
	return n
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 720
}

// ─── MAIN ─────────────────────────────────────────────────────────

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Dungeon Crawler - Go365 Day 94")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
