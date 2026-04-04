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

// ─── SPRITES ──────────────────────────────────────────────────────

var spr = make(map[string]*ebiten.Image)

func load(path string) *ebiten.Image {
	if s, ok := spr[path]; ok {
		return s
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
	e := ebiten.NewImageFromImage(img)
	spr[path] = e
	return e
}

// ─── STATE ────────────────────────────────────────────────────────

type State int

const (
	Title State = iota
	Playing
	Dead
	Win
)

// ─── PLAYER ───────────────────────────────────────────────────────

type Player struct {
	X, Y       float64
	W, H       float64
	VX, VY     float64
	OnGround   bool
	Jumps      int
	MaxJumps   int
	HP         int
	MaxHP      int
	Score      int
	Coins      int
	Facing     int  // 1=right, -1=left
	Attacking  bool
	AttackT    int
	Invincible int
	Dead       bool
	WalkFrame  int
	WalkTimer  int
}

func NewPlayer(x, y float64) *Player {
	return &Player{
		X: x, Y: y, W: 28, H: 44,
		Jumps: 2, MaxJumps: 2,
		HP: 5, MaxHP: 5,
		Facing: 1,
	}
}

func (p *Player) Update(keys map[ebiten.Key]bool, platforms []*Platform, enemies []*Enemy) {
	if p.Dead {
		return
	}

	// Input
	p.VX = 0
	if keys[ebiten.KeyA] || keys[ebiten.KeyLeft] {
		p.VX = -220
		p.Facing = -1
	}
	if keys[ebiten.KeyD] || keys[ebiten.KeyRight] {
		p.VX = 220
		p.Facing = 1
	}

	// Jump
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		if p.Jumps > 0 {
			p.VY = -460
			p.Jumps--
			p.OnGround = false
		}
	}

	// Attack
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) || inpututil.IsKeyJustPressed(ebiten.KeyX) {
		if !p.Attacking {
			p.Attacking = true
			p.AttackT = 12
		}
	}
	if p.AttackT > 0 {
		p.AttackT--
		if p.AttackT <= 0 {
			p.Attacking = false
		}
	}

	// Gravity
	p.VY += 1100
	if p.VY > 700 {
		p.VY = 700
	}

	// Move X
	p.X += p.VX
	for _, pl := range platforms {
		if p.overlaps(pl.X, pl.Y, pl.W, pl.H) {
			if p.VX > 0 {
				p.X = pl.X - p.W
			} else if p.VX < 0 {
				p.X = pl.X + pl.W
			}
			p.VX = 0
		}
	}

	// Move Y
	p.Y += p.VY
	p.OnGround = false
	for _, pl := range platforms {
		if p.overlaps(pl.X, pl.Y, pl.W, pl.H) {
			if p.VY > 0 {
				p.Y = pl.Y - p.H
				p.VY = 0
				p.OnGround = true
				p.Jumps = p.MaxJumps
			} else if p.VY < 0 {
				p.Y = pl.Y + pl.H
				p.VY = 0
			}
		}
	}

	// Enemy collision
	for _, e := range enemies {
		if !e.Alive {
			continue
		}
		if p.overlaps(e.X, e.Y, e.W, e.H) {
			// Stomp from above
			if p.VY > 0 && p.Y+p.H < e.Y+e.H/2+10 {
				e.TakeDamage(1)
				p.VY = -350
				p.Score += 50
			} else {
				p.TakeDamage(1)
			}
		}
	}

	// Fall death
	if p.Y > 900 {
		p.HP = 0
		p.Dead = true
	}

	// Invincibility
	if p.Invincible > 0 {
		p.Invincible--
	}

	// Walk animation
	if p.OnGround && p.VX != 0 {
		p.WalkTimer++
		if p.WalkTimer > 6 {
			p.WalkTimer = 0
			p.WalkFrame = (p.WalkFrame + 1) % 11
		}
	} else {
		p.WalkFrame = 0
	}
}

func (p *Player) overlaps(ox, oy, ow, oh float64) bool {
	return p.X < ox+ow && p.X+p.W > ox && p.Y < oy+oh && p.Y+p.H > oy
}

func (p *Player) TakeDamage(dmg int) {
	if p.Invincible > 0 || p.Dead {
		return
	}
	p.HP -= dmg
	p.Invincible = 50
	if p.HP <= 0 {
		p.HP = 0
		p.Dead = true
	}
}

func (p *Player) AttackBox() (float64, float64, float64, float64) {
	if !p.Attacking {
		return 0, 0, 0, 0
	}
	ax := p.X + p.W
	ay := p.Y + 5
	aw := 35.0
	ah := 30.0
	if p.Facing < 0 {
		ax = p.X - aw
	}
	return ax, ay, aw, ah
}

func (p *Player) Draw(screen *ebiten.Image) {
	if p.Dead {
		return
	}
	if p.Invincible > 0 && p.Invincible%4 >= 2 {
		return
	}

	// Try walk sprite
	var s *ebiten.Image
	if p.OnGround && p.VX != 0 {
		s = load(fmt.Sprintf("assets/sprites/PlatformerComplete/Base pack/Player/p1_walk/PNG/p1_walk%02d.png", p.WalkFrame+1))
	} else if !p.OnGround {
		s = load("assets/sprites/PlatformerComplete/Base pack/Player/p1_jump.png")
	} else {
		s = load("assets/sprites/PlatformerComplete/Base pack/Player/p1_stand.png")
	}

	if s != nil {
		op := &ebiten.DrawImageOptions{}
		if p.Facing < 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(p.X+p.W, p.Y)
		} else {
			op.GeoM.Translate(p.X, p.Y)
		}
		screen.DrawImage(s, op)
	} else {
		// Fallback
		img := ebiten.NewImage(int(p.W), int(p.H))
		img.Fill(color.RGBA{0, 150, 255, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		screen.DrawImage(img, op)
	}

	// Attack visual
	if p.Attacking {
		ax, ay, aw, ah := p.AttackBox()
		img := ebiten.NewImage(int(aw), int(ah))
		img.Fill(color.RGBA{255, 255, 200, 180})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(ax, ay)
		screen.DrawImage(img, op)
	}
}

// ─── PLATFORM ─────────────────────────────────────────────────────

type Platform struct {
	X, Y, W, H float64
	Sprite     *ebiten.Image
	Ground     bool
}

func (p *Platform) Draw(screen *ebiten.Image) {
	if p.Ground {
		// Tile grass on top, dirt below
		tile := load("assets/sprites/PlatformerComplete/Base pack/Tiles/grass.png")
		if tile != nil {
			tw := tile.Bounds().Dx()
			th := tile.Bounds().Dy()
			for tx := 0; tx < int(p.W); tx += tw {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(p.X+float64(tx), p.Y)
				screen.DrawImage(tile, op)
			}
			// Dirt below
			dirt := load("assets/sprites/PlatformerComplete/Base pack/Tiles/dirt.png")
			if dirt != nil {
				for tx := 0; tx < int(p.W); tx += tw {
					for ty := th; ty < int(p.H); ty += th {
						op := &ebiten.DrawImageOptions{}
						op.GeoM.Translate(p.X+float64(tx), p.Y+float64(ty))
						screen.DrawImage(dirt, op)
					}
				}
			}
		} else {
			img := ebiten.NewImage(int(p.W), int(p.H))
			img.Fill(color.RGBA{100, 180, 80, 255})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.X, p.Y)
			screen.DrawImage(img, op)
		}
	} else {
		// Floating platform
		tile := load("assets/sprites/PlatformerComplete/Base pack/Tiles/stone.png")
		if tile != nil {
			tw := tile.Bounds().Dx()
			for tx := 0; tx < int(p.W); tx += tw {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(p.X+float64(tx), p.Y)
				screen.DrawImage(tile, op)
			}
		} else {
			img := ebiten.NewImage(int(p.W), int(p.H))
			img.Fill(color.RGBA{120, 120, 140, 255})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.X, p.Y)
			screen.DrawImage(img, op)
		}
	}
}

// ─── ENEMY ────────────────────────────────────────────────────────

type Enemy struct {
	X, Y, W, H   float64
	VX, VY       float64
	HP           int
	MaxHP        int
	Alive        bool
	Type         int // 0=slime, 1=bird, 2=snail
	StartX       float64
	PatrolDist   float64
	AnimFrame    int
	AnimTimer    int
}

func NewSlime(x, y float64, dist float64) *Enemy {
	return &Enemy{
		X: x, Y: y, W: 32, H: 28,
		HP: 2, MaxHP: 2, Alive: true,
		Type: 0, StartX: x, PatrolDist: dist,
		VX: 50,
	}
}

func NewBird(x, y float64, dist float64) *Enemy {
	return &Enemy{
		X: x, Y: y, W: 36, H: 28,
		HP: 1, MaxHP: 1, Alive: true,
		Type: 1, StartX: x, PatrolDist: dist,
		VX: 80,
	}
}

func NewSnail(x, y float64, dist float64) *Enemy {
	return &Enemy{
		X: x, Y: y, W: 32, H: 32,
		HP: 3, MaxHP: 3, Alive: true,
		Type: 2, StartX: x, PatrolDist: dist,
		VX: 30,
	}
}

func (e *Enemy) Update(dt float64, platforms []*Platform) {
	if !e.Alive {
		return
	}

	e.VY += 800
	if e.VY > 600 {
		e.VY = 600
	}

	// Patrol
	e.X += e.VX
	if e.X > e.StartX+e.PatrolDist {
		e.VX = -math.Abs(e.VX)
	}
	if e.X < e.StartX-e.PatrolDist {
		e.VX = math.Abs(e.VX)
	}

	// Bird floats
	if e.Type == 1 {
		e.VY = 0
		e.Y += math.Sin(e.X*0.02) * 0.5
	} else {
		e.Y += e.VY * dt
	}

	// Platform collision
	for _, p := range platforms {
		if e.X < p.X+p.W && e.X+e.W > p.X && e.Y < p.Y+p.H && e.Y+e.H > p.Y {
			if e.VY > 0 {
				e.Y = p.Y - e.H
				e.VY = 0
			}
		}
	}

	// Animation
	e.AnimTimer++
	if e.AnimTimer > 12 {
		e.AnimTimer = 0
		e.AnimFrame = 1 - e.AnimFrame
	}
}

func (e *Enemy) TakeDamage(n int) {
	e.HP -= n
	if e.HP <= 0 {
		e.Alive = false
	}
}

func (e *Enemy) Draw(screen *ebiten.Image) {
	if !e.Alive {
		return
	}

	var s *ebiten.Image
	switch e.Type {
	case 0:
		if e.AnimFrame == 0 {
			s = load("assets/sprites/PlatformerComplete/Base pack/Enemies/slimeWalk1.png")
		} else {
			s = load("assets/sprites/PlatformerComplete/Base pack/Enemies/slimeWalk2.png")
		}
	case 1:
		if e.AnimFrame == 0 {
			s = load("assets/sprites/PlatformerComplete/Base pack/Enemies/flyFly1.png")
		} else {
			s = load("assets/sprites/PlatformerComplete/Base pack/Enemies/flyFly2.png")
		}
	case 2:
		if e.AnimFrame == 0 {
			s = load("assets/sprites/PlatformerComplete/Base pack/Enemies/snailWalk1.png")
		} else {
			s = load("assets/sprites/PlatformerComplete/Base pack/Enemies/snailWalk2.png")
		}
	}

	if s != nil {
		op := &ebiten.DrawImageOptions{}
		if e.VX > 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(e.X+e.W, e.Y)
		} else {
			op.GeoM.Translate(e.X, e.Y)
		}
		screen.DrawImage(s, op)
	} else {
		c := color.RGBA{100, 200, 100, 255}
		if e.Type == 1 {
			c = color.RGBA{150, 100, 200, 255}
		} else if e.Type == 2 {
			c = color.RGBA{200, 150, 100, 255}
		}
		img := ebiten.NewImage(int(e.W), int(e.H))
		img.Fill(c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X, e.Y)
		screen.DrawImage(img, op)
	}
}

// ─── COIN ─────────────────────────────────────────────────────────

type Coin struct {
	X, Y     float64
	W, H     float64
	Collected bool
	Bob      float64
	Sprite   *ebiten.Image
	Value    int
}

func NewCoin(x, y float64) *Coin {
	return &Coin{
		X: x, Y: y, W: 16, H: 16,
		Sprite: load("assets/sprites/PlatformerComplete/Base pack/Items/coinGold.png"),
		Value:  10,
	}
}

func (c *Coin) Update(dt float64, camX float64) {
	c.Bob += dt * 4
}

func (c *Coin) Draw(screen *ebiten.Image) {
	if c.Collected {
		return
	}
	bob := math.Sin(c.Bob) * 4
	if c.Sprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(c.X, c.Y+bob)
		screen.DrawImage(c.Sprite, op)
	} else {
		img := ebiten.NewImage(16, 16)
		img.Fill(color.RGBA{255, 215, 0, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(c.X, c.Y+bob)
		screen.DrawImage(img, op)
	}
}

// ─── PARTICLE ─────────────────────────────────────────────────────

type Particle struct {
	X, Y, VX, VY float64
	Life, Max    int
	R, G, B      uint8
}

// ─── LEVEL ────────────────────────────────────────────────────────

type Level struct {
	Platforms []*Platform
	Enemies   []*Enemy
	Coins     []*Coin
	FlagX     float64
	Width     float64
}

func GenerateLevel(lvl int) *Level {
	l := &Level{}
	l.Width = 2000 + float64(lvl)*600
	groundY := 550.0

	// Ground segments with gaps
	x := 0.0
	for x < l.Width {
		segLen := 200.0 + rand.Float64()*300
		gap := 80.0 + rand.Float64()*60*float64(lvl)

		l.Platforms = append(l.Platforms, &Platform{
			X: x, Y: groundY, W: segLen, H: 200, Ground: true,
		})

		// Floating platforms
		numFloat := rand.Intn(3) + 1
		for i := 0; i < numFloat; i++ {
			fx := x + rand.Float64()*segLen
			fy := groundY - 80 - rand.Float64()*180
			fw := 64 + rand.Float64()*100
			l.Platforms = append(l.Platforms, &Platform{
				X: fx, Y: fy, W: fw, H: 16,
			})

			// Coins on floating platforms
			l.Coins = append(l.Coins, NewCoin(fx+fw/2-8, fy-30))
		}

		// Enemies
		if rand.Float64() < 0.4+float64(lvl)*0.1 {
			et := rand.Intn(3)
			ex := x + 50 + rand.Float64()*(segLen-100)
			switch et {
			case 0:
				l.Enemies = append(l.Enemies, NewSlime(ex, groundY-28, 80))
			case 1:
				l.Enemies = append(l.Enemies, NewBird(ex, groundY-120-rand.Float64()*80, 100))
			case 2:
				l.Enemies = append(l.Enemies, NewSnail(ex, groundY-32, 60))
			}
		}

		// Ground coins
		for i := 0; i < 3; i++ {
			l.Coins = append(l.Coins, NewCoin(x+rand.Float64()*segLen, groundY-30))
		}

		x += segLen + gap
	}

	// Flag at end
	l.FlagX = l.Width - 100

	return l
}

// ─── GAME ─────────────────────────────────────────────────────────

type Game struct {
	State     State
	Player    *Player
	Level     *Level
	CamX      float64
	Particles []Particle
	CurrentLvl int
	MaxLevels  int
}

func NewGame() *Game {
	return &Game{
		State:    Title,
		CamX:     0,
		MaxLevels: 3,
	}
}

func (g *Game) StartGame() {
	g.Player = NewPlayer(100, 400)
	g.CurrentLvl = 0
	g.CamX = 0
	g.Particles = nil
	g.loadLevel(g.CurrentLvl)
}

func (g *Game) loadLevel(n int) {
	g.Level = GenerateLevel(n + 1)
	g.Player.X = 100
	g.Player.Y = 400
	g.Player.VX = 0
	g.Player.VY = 0
	g.CamX = 0
}

func (g *Game) Update() error {
	switch g.State {
	case Title:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.State = Playing
			g.StartGame()
		}
	case Playing:
		g.updatePlaying()
	case Dead:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.State = Playing
			g.StartGame()
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

	// Keys
	keys := make(map[ebiten.Key]bool)
	for _, k := range []ebiten.Key{ebiten.KeyA, ebiten.KeyD, ebiten.KeyW, ebiten.KeyS, ebiten.KeyLeft, ebiten.KeyRight, ebiten.KeyUp, ebiten.KeyDown} {
		keys[k] = ebiten.IsKeyPressed(k)
	}

	g.Player.Update(keys, g.Level.Platforms, g.Level.Enemies)

	if g.Player.Dead {
		g.State = Dead
		return
	}

	// Update enemies
	for _, e := range g.Level.Enemies {
		e.Update(dt, g.Level.Platforms)
	}

	// Player attack enemies
	if g.Player.Attacking {
		ax, ay, aw, ah := g.Player.AttackBox()
		for _, e := range g.Level.Enemies {
			if !e.Alive {
				continue
			}
			if ax < e.X+e.W && ax+aw > e.X && ay < e.Y+e.H && ay+ah > e.Y {
				e.TakeDamage(1)
				g.spawnParticles(e.X+e.W/2, e.Y+e.H/2, 10, 255, 255, 0)
				if !e.Alive {
					g.Player.Score += 50
				}
			}
		}
	}

	// Coins
	for _, c := range g.Level.Coins {
		if c.Collected {
			continue
		}
		c.Update(dt, g.CamX)
		if g.Player.X < c.X+c.W && g.Player.X+g.Player.W > c.X &&
			g.Player.Y < c.Y+c.H && g.Player.Y+g.Player.H > c.Y {
			c.Collected = true
			g.Player.Coins++
			g.Player.Score += c.Value
			g.spawnParticles(c.X+8, c.Y+8, 6, 255, 215, 0)
		}
	}

	// Flag (reach end)
	if g.Player.X > g.Level.FlagX {
		g.CurrentLvl++
		if g.CurrentLvl >= g.MaxLevels {
			g.State = Win
			return
		}
		g.loadLevel(g.CurrentLvl)
	}

	// Particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := &g.Particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 300 * dt
		p.Life--
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}

	// Camera
	target := g.Player.X - 400
	g.CamX += (target - g.CamX) * 0.08
	if g.CamX < 0 {
		g.CamX = 0
	}
}

func (g *Game) spawnParticles(x, y float64, n int, r, gv, b uint8) {
	for i := 0; i < n; i++ {
		g.Particles = append(g.Particles, Particle{
			X: x, Y: y,
			VX: (rand.Float64() - 0.5) * 250,
			VY: -rand.Float64() * 200,
			Life: 15 + rand.Intn(20),
			Max: 35,
			R: r, G: gv, B: b,
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{135, 206, 235, 255})

	switch g.State {
	case Title:
		drawCentered(screen, "CITY PLATFORMER", 200, color.RGBA{0, 100, 200, 255}, 20)
		drawCentered(screen, "Press ENTER or SPACE", 300, color.White, 14)
		drawCentered(screen, "A/D or Arrows = Move", 380, color.RGBA{200, 200, 200, 255}, 12)
		drawCentered(screen, "SPACE/W/Up = Jump (Double!)", 410, color.RGBA{200, 200, 200, 255}, 12)
		drawCentered(screen, "J/X = Attack", 440, color.RGBA{200, 200, 200, 255}, 12)
		drawCentered(screen, "Stomp enemies from above!", 470, color.RGBA{200, 200, 200, 255}, 12)
	case Playing:
		g.drawGame(screen)
	case Dead:
		g.drawGame(screen)
		screen.Fill(color.RGBA{0, 0, 0, 180})
		drawCentered(screen, "GAME OVER", 300, color.RGBA{255, 50, 50, 255}, 20)
		drawCentered(screen, fmt.Sprintf("Score: %d", g.Player.Score), 360, color.White, 14)
		drawCentered(screen, "Press ENTER", 420, color.RGBA{150, 200, 255, 255}, 14)
	case Win:
		g.drawGame(screen)
		screen.Fill(color.RGBA{0, 0, 0, 180})
		drawCentered(screen, "YOU WIN!", 300, color.RGBA{100, 255, 100, 255}, 20)
		drawCentered(screen, fmt.Sprintf("Final Score: %d", g.Player.Score), 360, color.White, 14)
		drawCentered(screen, "Press ENTER", 420, color.RGBA{150, 200, 255, 255}, 14)
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Background
	bg := load("assets/sprites/PlatformerComplete/Base pack/bg.png")
	if bg != nil {
		screen.DrawImage(bg, nil)
	}

	// Camera transform
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-g.CamX, 0)

	// Platforms
	for _, p := range g.Level.Platforms {
		if p.X+p.W < g.CamX-50 || p.X > g.CamX+1330 {
			continue
		}
		p.Draw(screen)
	}

	// Coins
	for _, c := range g.Level.Coins {
		c.Draw(screen)
	}

	// Enemies
	for _, e := range g.Level.Enemies {
		e.Draw(screen)
	}

	// Flag at end
	flagImg := ebiten.NewImage(8, 80)
	flagImg.Fill(color.RGBA{255, 50, 50, 255})
	fop := &ebiten.DrawImageOptions{}
	fop.GeoM.Translate(g.Level.FlagX, 470)
	screen.DrawImage(flagImg, fop)

	// Player
	g.Player.Draw(screen)

	// Particles
	for _, p := range g.Particles {
		alpha := uint8(p.Life * 255 / p.Max)
		img := ebiten.NewImage(4, 4)
		img.Fill(color.RGBA{p.R, p.G, p.B, alpha})
		pop := &ebiten.DrawImageOptions{}
		pop.GeoM.Translate(p.X, p.Y)
		screen.DrawImage(img, pop)
	}

	// HUD
	drawText(screen, fmt.Sprintf("HP: %d/%d", g.Player.HP, g.Player.MaxHP), 20, 30, color.RGBA{255, 80, 80, 255})
	drawText(screen, fmt.Sprintf("Score: %d", g.Player.Score), 20, 55, color.White)
	drawText(screen, fmt.Sprintf("Coins: %d", g.Player.Coins), 20, 80, color.RGBA{255, 215, 0, 255})
	drawText(screen, fmt.Sprintf("Level: %d/%d", g.CurrentLvl+1, g.MaxLevels), 20, 105, color.RGBA{150, 200, 255, 255})

	ebitenutil.DebugPrint(screen, fmt.Sprintf("Pos: (%.0f, %.0f)", g.Player.X, g.Player.Y))
}

func drawCentered(screen *ebiten.Image, msg string, y int, c color.Color, scale int) {
	x := (1280 - len(msg)*7) / 2
	text.Draw(screen, msg, basicfont.Face7x13, x, y, c)
}

func drawText(screen *ebiten.Image, msg string, x, y int, c color.Color) {
	text.Draw(screen, msg, basicfont.Face7x13, x, y, c)
}

func (g *Game) Layout(w, h int) (int, int) {
	return 1280, 720
}

// ─── MAIN ─────────────────────────────────────────────────────────

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("City Platformer - Go365 Day 94")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
