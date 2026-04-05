package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// ======================== КОНСТАНТЫ ========================
const (
	Tile   = 48
	GW     = 15
	GH     = 13
	HUD    = 40
	WinW   = GW * Tile
	WinH   = GH*Tile + HUD
	CD     = 8
	ECMD   = 15
	BOMB_T = 180
	EXPL_T = 30
)

// ======================== ТАЙЛЫ ========================
const (
	T_EMPTY = iota
	T_STONE
	T_BRICK
)

// ======================== НАПРАВЛЕНИЯ ========================
const (
	D_DOWN = iota
	D_UP
	D_LEFT
	D_RIGHT
)

// ======================== ЦВЕТА ========================
var (
	C_BG        = color.RGBA{50, 50, 80, 255}
	C_HUD_BG    = color.RGBA{20, 20, 40, 255}
	C_STONE     = color.RGBA{120, 120, 140, 255}
	C_BRICK     = color.RGBA{180, 100, 60, 255}
	C_GRASS     = color.RGBA{60, 140, 60, 255}
	C_PLAYER    = color.RGBA{255, 150, 200, 255}
	C_PLAYER_H  = color.RGBA{255, 200, 230, 255}
	C_ENEMY_BAL = color.RGBA{255, 120, 120, 255}
	C_ENEMY_CHA = color.RGBA{100, 100, 255, 255}
	C_ENEMY_SPD = color.RGBA{100, 255, 100, 255}
	C_ENEMY_E   = color.RGBA{255, 255, 100, 255}
	C_ENEMY_TEL = color.RGBA{200, 100, 255, 255}
	C_BOMB      = color.RGBA{40, 40, 40, 255}
	C_BOMB_H    = color.RGBA{80, 80, 80, 255}
	C_BOMB_KICK = color.RGBA{100, 100, 100, 255}
	C_EXPL      = color.RGBA{255, 200, 50, 255}
	C_EXPL_C    = color.RGBA{255, 255, 200, 255}
	C_WHITE     = color.RGBA{255, 255, 255, 255}
	C_GREEN     = color.RGBA{100, 255, 100, 255}
	C_RED       = color.RGBA{255, 80, 80, 255}
	C_YELLOW    = color.RGBA{255, 255, 80, 255}
	C_KEY_DEBUG = color.RGBA{0, 255, 150, 255}
	C_DOOR      = color.RGBA{200, 180, 100, 255}
	C_SHIELD    = color.RGBA{100, 200, 255, 200}

	// Power-up цвета
	C_PU_FIRE  = color.RGBA{255, 100, 0, 255}
	C_PU_BOMB  = color.RGBA{80, 80, 80, 255}
	C_PU_SPEED = color.RGBA{255, 255, 0, 255}
	C_PU_HEART = color.RGBA{255, 50, 100, 255}
	C_PU_SHIELD = color.RGBA{100, 200, 255, 255}
	C_PU_KICK  = color.RGBA{200, 150, 100, 255}
	C_GOLD     = color.RGBA{255, 220, 100, 255}

	// Menu цвета
	C_MENU_BG     = color.RGBA{10, 10, 30, 255}
	C_MENU_TITLE  = color.RGBA{255, 220, 100, 255}
	C_MENU_BTN_BG = color.RGBA{40, 40, 80, 200}
	C_MENU_BTN_H  = color.RGBA{60, 60, 120, 230}
)

// ======================== СОСТОЯНИЕ ========================
type State int

const (
	S_MENU State = iota
	S_PLAY
	S_PAUSE
	S_OPTIONS
	S_DEAD
	S_WIN
)

// ======================== POWER-UPS ========================
type PowerUp struct {
	gx, gy int
	kind   int
	anim   int
	frame  int
}

const (
	PU_FIRE = iota
	PU_BOMB
	PU_SPEED
	PU_HEART
	PU_SHIELD
	PU_KICK
)

func puColor(kind int) color.RGBA {
	switch kind {
	case PU_FIRE:
		return C_PU_FIRE
	case PU_BOMB:
		return C_PU_BOMB
	case PU_SPEED:
		return C_PU_SPEED
	case PU_HEART:
		return C_PU_HEART
	case PU_SHIELD:
		return C_PU_SHIELD
	case PU_KICK:
		return C_PU_KICK
	default:
		return C_WHITE
	}
}

func puSymbol(kind int) string {
	switch kind {
	case PU_FIRE:
		return "F"
	case PU_BOMB:
		return "B"
	case PU_SPEED:
		return "S"
	case PU_HEART:
		return "H"
	case PU_SHIELD:
		return "D"
	case PU_KICK:
		return "K"
	default:
		return "?"
	}
}

func puName(kind int) string {
	switch kind {
	case PU_FIRE:
		return "Fire+1"
	case PU_BOMB:
		return "Bomb+1"
	case PU_SPEED:
		return "Speed+"
	case PU_HEART:
		return "Life+1"
	case PU_SHIELD:
		return "Shield"
	case PU_KICK:
		return "Kick"
	default:
		return "?"
	}
}

// ======================== ЧАСТИЦЫ ========================
type Particle struct {
	x, y    float64
	vx, vy  float64
	life    int
	maxLife int
	clr     color.RGBA
	size    int
}

func spawnParticles(gx, gy int, clr color.RGBA, count int) []Particle {
	ps := make([]Particle, 0, count)
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 1.0 + rand.Float64()*3.0
		ps = append(ps, Particle{
			x:       float64(gx*Tile + Tile/2),
			y:       float64(gy*Tile + HUD + Tile/2),
			vx:      math.Cos(angle) * speed,
			vy:      math.Sin(angle) * speed - 1,
			life:    20 + rand.Intn(20),
			maxLife: 40,
			clr:     clr,
			size:    2 + rand.Intn(4),
		})
	}
	return ps
}

// ======================== ТРЯСКА ЭКРАНА ========================
type ScreenShake struct {
	intensity float64
	timer     int
}

func (s *ScreenShake) shake(intensity float64, duration int) {
	s.intensity = intensity
	s.timer = duration
}

func (s *ScreenShake) update() {
	if s.timer > 0 {
		s.timer--
	} else {
		s.intensity = 0
	}
}

func (s *ScreenShake) offset() (float64, float64) {
	if s.intensity == 0 {
		return 0, 0
	}
	return (rand.Float64() - 0.5) * s.intensity * 2,
		(rand.Float64() - 0.5) * s.intensity * 2
}

// ======================== КНОПКА МЕНЮ ========================
type MenuButton struct {
	x, y, w, h int
	label      string
	spr        *ebiten.Image
	hover      bool
}

func (b *MenuButton) contains(mx, my int) bool {
	return mx >= b.x && mx < b.x+b.w && my >= b.y && my < b.y+b.h
}

func (b *MenuButton) Draw(s *ebiten.Image) {
	if b.spr != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(b.w)/float64(b.spr.Bounds().Dx()), float64(b.h)/float64(b.spr.Bounds().Dy()))
		op.GeoM.Translate(float64(b.x), float64(b.y))
		if b.hover {
			op.ColorM.Scale(1.2, 1.2, 1.2, 1)
		}
		s.DrawImage(b.spr, op)
	} else {
		bg := C_MENU_BTN_BG
		if b.hover {
			bg = C_MENU_BTN_H
		}
		rectCached(s, b.x, b.y, b.w, b.h, bg)
		bw := text.BoundString(basicfont.Face7x13, b.label)
		text.Draw(s, b.label, basicfont.Face7x13, b.x+b.w/2-bw.Dx()/2, b.y+b.h/2+5, C_WHITE)
	}
}

// ======================== ИГРОК ========================
type Player struct {
	gx, gy       int
	dir          int
	lives        int
	bombs        int
	active       int
	radius       int
	cd           int
	anim         int
	frame        int
	inv          int
	shield       bool
	speedBoost   int
	kick         bool
	score        int
	combos       int
	comboTimer   int
}

func NewPlayer() *Player {
	return &Player{gx: 1, gy: 1, dir: D_DOWN, lives: 3, bombs: 1, radius: 2, inv: 90, cd: CD}
}

func (p *Player) Update(g *Game) {
	if p.inv > 0 {
		p.inv--
	}
	if p.cd > 0 {
		p.cd--
	}
	if p.comboTimer > 0 {
		p.comboTimer--
		if p.comboTimer == 0 {
			p.combos = 0
		}
	}

	if p.cd > 0 {
		return
	}

	dx, dy := 0, 0
	pressed := false

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		dy, dx, pressed = -1, 0, true
		p.dir = D_UP
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		dy, dx, pressed = 1, 0, true
		p.dir = D_DOWN
	} else if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		dy, dx, pressed = 0, -1, true
		p.dir = D_LEFT
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		dy, dx, pressed = 0, 1, true
		p.dir = D_RIGHT
	}

	if pressed {
		g.keys = ""
		if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.keys += "W"
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
			g.keys += "A"
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.keys += "S"
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
			g.keys += "D"
		}

		p.frame++
		p.anim = 1 + (p.frame/6)%2
	} else {
		p.anim = 0
		p.frame = 0
		g.keys = ""
	}

	if dx == 0 && dy == 0 {
		return
	}

	// Пытаемся пнуть бомбу
	if p.kick {
		for i, b := range g.bombs {
			if (b.gx == p.gx+dx && b.gy == p.gy) || (b.gx == p.gx && b.gy == p.gy+dy) {
				if b.timer > BOMB_T - 60 {
					newX, newY := b.gx+dx, b.gy+dy
					if g.walkable(newX, newY, p.gx, p.gy) {
						g.bombs[i].gx = newX
						g.bombs[i].gy = newY
						p.cd = CD
						return
					}
				}
			}
		}
	}

	nx, ny := p.gx+dx, p.gy+dy
	if g.walkable(nx, ny, p.gx, p.gy) {
		p.gx, p.gy = nx, ny
		p.cd = CD
		if p.speedBoost > 0 {
			p.cd = CD / 2
		}
	} else if g.walkable(nx, p.gy, p.gx, p.gy) {
		p.gx = nx
		p.cd = CD
	} else if g.walkable(p.gx, ny, p.gx, p.gy) {
		p.gy = ny
		p.cd = CD
	}
}

func (p *Player) Draw(s *ebiten.Image, sprites map[string]*ebiten.Image) {
	if p.inv > 0 && p.inv%8 < 4 {
		return
	}

	px, py := p.gx*Tile, p.gy*Tile+HUD

	// Shield эффект
	if p.shield {
		rectCached(s, px+2, py+2, Tile-4, Tile-4, C_SHIELD)
	}

	// Спрайт игрока
	if spr := sprites["player"]; spr != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(Tile)/float64(spr.Bounds().Dx()), float64(Tile)/float64(spr.Bounds().Dy()))
		op.GeoM.Translate(float64(px), float64(py))
		s.DrawImage(spr, op)
	} else {
		// Fallback на примитивы
		rectCached(s, px+4, py+4, Tile-8, Tile-8, C_PLAYER)
		rectCached(s, px+12, py+2, Tile-24, 16, C_PLAYER_H)
	}

	// Глаза
	ex, ey := px+Tile/2, py+Tile/3
	switch p.dir {
	case D_UP:
		ey -= 4
	case D_DOWN:
		ey += 4
	case D_LEFT:
		ex -= 6
	case D_RIGHT:
		ex += 6
	}
	rectCached(s, ex-3, ey-3, 3, 3, color.Black)
	rectCached(s, ex+3, ey-3, 3, 3, color.Black)

	if p.anim > 0 {
		rectCached(s, px+8, py+Tile-6, 8, 6, C_PLAYER)
		rectCached(s, px+Tile-16, py+Tile-6, 8, 6, C_PLAYER)
	}
}

// ======================== ВРАГИ ========================
const (
	E_BALLOON = iota
	E_CHASER
	E_SPLITTER
	E_TELEPORTER
)

type Enemy struct {
	gx, gy int
	kind   int
	dir    int
	alive  bool
	cd     int
	anim   int
	frame  int
	hp     int
}

func NewEnemy(gx, gy int, kind int) *Enemy {
	hp := 1
	if kind == E_SPLITTER {
		hp = 2
	}
	return &Enemy{gx: gx, gy: gy, kind: kind, dir: rand.Intn(4), alive: true, cd: ECMD/2, hp: hp}
}

func (e *Enemy) Update(g *Game) {
	if !e.alive {
		return
	}
	if e.cd > 0 {
		e.cd--
		return
	}

	e.frame++
	e.anim = (e.frame / 8) % 2

	switch e.kind {
	case E_BALLOON:
		e.updateBalloon(g)
	case E_CHASER:
		e.updateChaser(g)
	case E_SPLITTER:
		e.updateSplitter(g)
	case E_TELEPORTER:
		e.updateTeleporter(g)
	}
}

func (e *Enemy) updateBalloon(g *Game) {
	e.cd = ECMD + rand.Intn(20)
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	for _, d := range dirs {
		nx, ny := e.gx+d[0], e.gy+d[1]
		if g.walkableEnemy(nx, ny) {
			e.gx, e.gy = nx, ny
			return
		}
	}
}

func (e *Enemy) updateChaser(g *Game) {
	e.cd = ECMD - 3 + rand.Intn(10)

	dx := g.player.gx - e.gx
	dy := g.player.gy - e.gy

	moves := [][2]int{}
	if rand.Float64() < 0.7 {
		if dx > 0 {
			moves = append(moves, [2]int{1, 0})
		} else if dx < 0 {
			moves = append(moves, [2]int{-1, 0})
		}
		if dy > 0 {
			moves = append(moves, [2]int{0, 1})
		} else if dy < 0 {
			moves = append(moves, [2]int{0, -1})
		}
	}
	moves = append(moves, [2]int{0, -1}, [2]int{0, 1}, [2]int{-1, 0}, [2]int{1, 0})

	rand.Shuffle(len(moves), func(i, j int) { moves[i], moves[j] = moves[j], moves[i] })

	for _, d := range moves {
		nx, ny := e.gx+d[0], e.gy+d[1]
		if g.walkableEnemy(nx, ny) {
			e.gx, e.gy = nx, ny
			return
		}
	}
}

func (e *Enemy) updateSplitter(g *Game) {
	e.cd = ECMD + rand.Intn(15)
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	for _, d := range dirs {
		nx, ny := e.gx+d[0], e.gy+d[1]
		if g.walkableEnemy(nx, ny) {
			e.gx, e.gy = nx, ny
			return
		}
	}
}

func (e *Enemy) updateTeleporter(g *Game) {
	e.cd = 60 + rand.Intn(40)

	for attempts := 0; attempts < 20; attempts++ {
		nx := 1 + rand.Intn(GW-2)
		ny := 1 + rand.Intn(GH-2)
		if g.grid[ny][nx] == T_EMPTY {
			dist := abs(nx-g.player.gx) + abs(ny-g.player.gy)
			if dist > 3 {
				e.gx, e.gy = nx, ny
				g.particles = append(g.particles, spawnParticles(e.gx, e.gy, C_ENEMY_TEL, 8)...)
				return
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (e *Enemy) Draw(s *ebiten.Image, sprites map[string]*ebiten.Image) {
	if !e.alive {
		return
	}
	px, py := e.gx*Tile, e.gy*Tile+HUD

	clr := C_ENEMY_BAL
	switch e.kind {
	case E_CHASER:
		clr = C_ENEMY_CHA
	case E_SPLITTER:
		clr = C_ENEMY_SPD
	case E_TELEPORTER:
		clr = C_ENEMY_TEL
	}

	if spr := sprites["enemy"]; spr != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(Tile)/float64(spr.Bounds().Dx()), float64(Tile)/float64(spr.Bounds().Dy()))
		op.GeoM.Translate(float64(px), float64(py))
		s.DrawImage(spr, op)
	} else {
		switch e.kind {
		case E_BALLOON:
			rectCached(s, px+4, py+4, Tile-8, Tile-8, clr)
			rectCached(s, px+8, py+2, Tile-16, 4, clr)
		case E_CHASER:
			rectCached(s, px+4, py+4, Tile-8, Tile-8, clr)
			if g_player != nil {
				if e.gx < g_player.gx {
					rectCached(s, px+Tile-6, py+Tile/2-4, 6, 8, C_ENEMY_E)
				} else if e.gx > g_player.gx {
					rectCached(s, px, py+Tile/2-4, 6, 8, C_ENEMY_E)
				}
			}
		case E_SPLITTER:
			rectCached(s, px+2, py+2, Tile-4, Tile-4, clr)
			if e.hp > 1 {
				rectCached(s, px+4, py+Tile-8, (Tile-8)*e.hp/2, 4, C_ENEMY_E)
			}
		case E_TELEPORTER:
			rectCached(s, px+Tile/2-4, py+4, 8, Tile-8, clr)
			rectCached(s, px+4, py+Tile/2-4, Tile-8, 8, clr)
		}
	}

	rectCached(s, px+10, py+14, 8, 8, C_ENEMY_E)
	rectCached(s, px+Tile-18, py+14, 8, 8, C_ENEMY_E)
}

var g_player *Player

// ======================== БОМБА ========================
type Bomb struct {
	gx, gy int
	timer  int
	radius int
	kicked bool
	owner  int
}

// ======================== ВЗРЫВ ========================
type Cell struct {
	gx, gy int
	t      int
}

// ======================== КЭШИРОВАННЫЕ ИЗОБРАЖЕНИЯ ========================
var cachedRects = make(map[string]*ebiten.Image)

func rectCached(s *ebiten.Image, x, y, w, h int, c color.Color) {
	key := fmt.Sprintf("%d_%d_%d_%d_%d_%d_%d", c.(color.RGBA).R, c.(color.RGBA).G, c.(color.RGBA).B, c.(color.RGBA).A, w, h, 0)
	if img, ok := cachedRects[key]; ok {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		s.DrawImage(img, op)
		return
	}
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	cachedRects[key] = img
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	s.DrawImage(img, op)
}

// ======================== ИГРА ========================
type Game struct {
	grid    [][]int
	player  *Player
	enemies []*Enemy
	bombs   []Bomb
	explos  []Cell
	pus     []PowerUp
	particles []Particle
	state   State
	keys    string
	sprites map[string]*ebiten.Image
	menuSprites map[string]*ebiten.Image
	level   int
	spacePrev bool
	enterPrev bool
	shake   ScreenShake
	score   int
	buttons []*MenuButton
	highScore int
	menuAnim int
}

func NewGame() *Game {
	g := &Game{
		state:       S_MENU,
		sprites:     make(map[string]*ebiten.Image),
		menuSprites: make(map[string]*ebiten.Image),
		level:       1,
		highScore:   0,
	}
	g.loadSprites()
	g.loadMenuSprites()
	g.initButtons()
	initAudio()
	return g
}

func (g *Game) loadSprites() {
	tryLoad := func(name, file string, target map[string]*ebiten.Image) {
		img, _, err := ebitenutil.NewImageFromFile("assets/sprites/" + file)
		if err == nil {
			target[name] = img
		}
	}
	
	// Основные спрайты
	tryLoad("player", "player_stand.png", g.sprites)
	tryLoad("enemy", "enemy1.png", g.sprites)
	tryLoad("bomb", "bomb.png", g.sprites)
	tryLoad("brick", "brick.png", g.sprites)
	tryLoad("stone", "stone.png", g.sprites)
	tryLoad("grass", "grass.png", g.sprites)
	tryLoad("explosion", "explosion.png", g.sprites)

	if len(g.sprites) > 0 {
		fmt.Printf("✓ Loaded %d game sprites\n", len(g.sprites))
	}
}

func (g *Game) loadMenuSprites() {
	tryLoad := func(name, file string) {
		img, _, err := ebitenutil.NewImageFromFile("assets/sprites/menu/" + file)
		if err == nil {
			g.menuSprites[name] = img
		}
	}

	tryLoad("play", "play button.png")
	tryLoad("options", "Options Button.png")
	tryLoad("exit", "Exit Button.png")
	tryLoad("back", "Back Button.png")
	tryLoad("stars", "stars.png")
	tryLoad("stars_bg", "stars back.png")
	tryLoad("particles1", "particles 1.png")

	if len(g.menuSprites) > 0 {
		fmt.Printf("✓ Loaded %d menu sprites\n", len(g.menuSprites))
	}
}

func (g *Game) initButtons() {
	g.buttons = nil
	g.buttons = append(g.buttons, &MenuButton{
		x: WinW/2 - 90, y: WinH/2 - 60, w: 180, h: 44,
		label: "PLAY", spr: g.menuSprites["play"],
	})
	g.buttons = append(g.buttons, &MenuButton{
		x: WinW/2 - 90, y: WinH/2, w: 180, h: 44,
		label: "OPTIONS", spr: g.menuSprites["options"],
	})
	g.buttons = append(g.buttons, &MenuButton{
		x: WinW/2 - 90, y: WinH/2 + 60, w: 180, h: 44,
		label: "EXIT", spr: g.menuSprites["exit"],
	})
}

func (g *Game) initLevel() {
	g.grid = make([][]int, GH)
	for y := range g.grid {
		g.grid[y] = make([]int, GW)
	}

	for y := 0; y < GH; y += 2 {
		for x := 0; x < GW; x += 2 {
			g.grid[y][x] = T_STONE
		}
	}

	r := rand.New(rand.NewSource(42 + int64(g.level)*7))
	g.pus = nil

	for y := 1; y < GH-1; y++ {
		for x := 1; x < GW-1; x++ {
			if x <= 2 && y <= 2 {
				continue
			}
			if r.Float32() < 0.30 {
				g.grid[y][x] = T_BRICK
				if r.Float32() < 0.25 {
					kind := r.Intn(6)
					g.pus = append(g.pus, PowerUp{gx: x, gy: y, kind: kind})
				}
			}
		}
	}

	g.player = NewPlayer()
	g.bombs = nil
	g.explos = nil
	g.particles = nil

	g.enemies = nil
	count := 3 + g.level
	positions := [][2]int{}
	for y := 1; y < GH-1; y++ {
		for x := 4; x < GW-1; x++ {
			if g.grid[y][x] == T_EMPTY {
				positions = append(positions, [2]int{x, y})
			}
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	for i := 0; i < count && i < len(positions); i++ {
		kind := E_BALLOON
		if g.level >= 2 && i < count/3 {
			kind = E_CHASER
		}
		if g.level >= 3 && i < count/4 {
			kind = E_SPLITTER
		}
		if g.level >= 4 && i < count/5 {
			kind = E_TELEPORTER
		}
		g.enemies = append(g.enemies, NewEnemy(positions[i][0], positions[i][1], kind))
	}

	g_player = g.player
}

func (g *Game) walkable(nx, ny, fromX, fromY int) bool {
	if nx < 0 || nx >= GW || ny < 0 || ny >= GH {
		return false
	}
	if g.grid[ny][nx] != T_EMPTY {
		return false
	}
	for _, b := range g.bombs {
		if b.gx == nx && b.gy == ny && !(b.gx == fromX && b.gy == fromY) {
			return false
		}
	}
	return true
}

func (g *Game) walkableEnemy(nx, ny int) bool {
	if nx < 0 || nx >= GW || ny < 0 || ny >= GH {
		return false
	}
	if g.grid[ny][nx] != T_EMPTY {
		return false
	}
	for _, b := range g.bombs {
		if b.gx == nx && b.gy == ny {
			return false
		}
	}
	return true
}

func (g *Game) Update() error {
	frameCount++

	// ===== МЕНЮ =====
	if g.state == S_MENU {
		g.menuAnim++

		mx, my := ebiten.CursorPosition()
		for _, btn := range g.buttons {
			btn.hover = btn.contains(mx, my)
			if btn.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if btn.label == "PLAY" {
					g.state = S_PLAY
					g.initLevel()
					g.score = 0
					playSound(soundMenu)
					fmt.Println("▶ Game started! Level", g.level)
				} else if btn.label == "OPTIONS" {
					g.state = S_OPTIONS
					g.initOptionsButtons()
					playSound(soundMenu)
				} else if btn.label == "EXIT" {
					os.Exit(0)
				}
			}
		}

		enter := ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace)
		if enter && !g.enterPrev {
			g.state = S_PLAY
			g.initLevel()
			g.score = 0
			playSound(soundMenu)
		}
		g.enterPrev = enter
		return nil
	}

	// ===== OPTIONS =====
	if g.state == S_OPTIONS {
		esc := ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyBackspace)
		if esc && !g.enterPrev {
			g.state = S_MENU
			g.initButtons()
			playSound(soundMenu)
		}
		mx, my := ebiten.CursorPosition()
		for _, btn := range g.buttons {
			btn.hover = btn.contains(mx, my)
			if btn.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if btn.label == "▶ BACK" {
					g.state = S_MENU
					g.initButtons()
					playSound(soundMenu)
				}
			}
		}
		g.enterPrev = esc
		return nil
	}

	// ===== PAUSE =====
	if g.state == S_PAUSE {
		esc := ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyP)
		if esc && !g.enterPrev {
			g.state = S_PLAY
			playSound(soundMenu)
		}
		
		mx, my := ebiten.CursorPosition()
		for _, btn := range g.buttons {
			btn.hover = btn.contains(mx, my)
			if btn.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if btn.label == "▶ RESUME" {
					g.state = S_PLAY
					playSound(soundMenu)
				} else if btn.label == "▶ RESTART" {
					g.level = 1
					g.initLevel()
					g.state = S_PLAY
					g.score = 0
					playSound(soundMenu)
				}
			}
		}
		
		g.enterPrev = esc
		return nil
	}

	// ===== GAME OVER / WIN =====
	if g.state == S_DEAD || g.state == S_WIN {
		enter := ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace)
		if enter && !g.enterPrev {
			if g.state == S_WIN {
				g.level++
				g.initLevel()
				g.state = S_PLAY
				playSound(soundMenu)
				fmt.Println("▶ Level", g.level)
			} else {
				g.level = 1
				g.initLevel()
				g.state = S_PLAY
				g.score = 0
				playSound(soundMenu)
				fmt.Println("▶ Restart")
			}
		}
		g.enterPrev = enter
		return nil
	}

	// ===== ИГРА =====
	space := ebiten.IsKeyPressed(ebiten.KeySpace)
	g_player = g.player

	// Пауза
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.state = S_PAUSE
		g.initPauseButtons()
		playSound(soundMenu)
		return nil
	}

	g.player.Update(g)
	g.shake.update()

	// Бомба
	if space && !g.spacePrev && g.player.active < g.player.bombs {
		hasBomb := false
		for _, b := range g.bombs {
			if b.gx == g.player.gx && b.gy == g.player.gy {
				hasBomb = true
				break
			}
		}
		if !hasBomb {
			g.bombs = append(g.bombs, Bomb{
				gx: g.player.gx, gy: g.player.gy,
				timer: BOMB_T, radius: g.player.radius,
			})
			g.player.active++
			playSound(soundPlace)
		}
	}
	g.spacePrev = space

	// Рестарт
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.level = 1
		g.initLevel()
		g.state = S_PLAY
		g.score = 0
	}

	// Бомбы
	for i := len(g.bombs) - 1; i >= 0; i-- {
		g.bombs[i].timer--
		if g.bombs[i].timer <= 0 {
			g.doExplosion(i)
		}
	}

	// Взрывы
	for i := len(g.explos) - 1; i >= 0; i-- {
		g.explos[i].t--
		if g.explos[i].t <= 0 {
			g.explos = append(g.explos[:i], g.explos[i+1:]...)
		}
	}

	// Частицы
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.15
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Враги
	for _, e := range g.enemies {
		e.Update(g)
	}

	// Коллизия: игрок ↔ враг
	if g.player.inv <= 0 {
		for _, e := range g.enemies {
			if e.alive && e.gx == g.player.gx && e.gy == g.player.gy {
				g.playerHit()
			}
		}
	}

	// Коллизия: игрок ↔ взрыв
	for _, c := range g.explos {
		if c.gx == g.player.gx && c.gy == g.player.gy && g.player.inv <= 0 {
			g.playerHit()
		}
	}

	// Коллизия: враг ↔ взрыв
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		for _, c := range g.explos {
			if c.gx == e.gx && c.gy == e.gy {
				e.hp--
				if e.hp <= 0 {
					e.alive = false
					g.player.combos++
					g.player.comboTimer = 60
					bonus := 100 * g.player.combos
					g.player.score += bonus
					g.particles = append(g.particles, spawnParticles(e.gx, e.gy, C_EXPL, 12)...)

					if e.kind == E_SPLITTER {
						for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
							nx, ny := e.gx+d[0], e.gy+d[1]
							if g.walkableEnemy(nx, ny) {
								child := NewEnemy(nx, ny, E_BALLOON)
								child.cd = 30
								g.enemies = append(g.enemies, child)
								break
							}
						}
					}

					playSound(soundKill)
				} else {
					g.particles = append(g.particles, spawnParticles(e.gx, e.gy, C_ENEMY_E, 6)...)
					playSound(soundPlace)
				}
			}
		}
	}

	// Проверка победы
	alive := 0
	for _, e := range g.enemies {
		if e.alive {
			alive++
		}
	}
	if alive == 0 && len(g.enemies) > 0 {
		g.state = S_WIN
		g.player.score += 500 * g.level
		if g.player.score > g.highScore {
			g.highScore = g.player.score
		}
		playSound(soundWin)
		g.shake.shake(5, 20)
		fmt.Println("🏆 Level complete! Score:", g.player.score)
	}

	return nil
}

func (g *Game) initPauseButtons() {
	g.buttons = nil
	
	if spr, ok := g.menuSprites["back"]; ok {
		g.buttons = append(g.buttons, &MenuButton{
			x: WinW/2 - 80, y: WinH/2 - 40, w: 160, h: 50,
			label: "▶ RESUME", spr: spr,
		})
	} else {
		g.buttons = append(g.buttons, &MenuButton{
			x: WinW/2 - 60, y: WinH/2 - 40, w: 120, h: 40,
			label: "▶ RESUME",
		})
	}

	if spr, ok := g.menuSprites["play"]; ok {
		g.buttons = append(g.buttons, &MenuButton{
			x: WinW/2 - 80, y: WinH/2 + 30, w: 160, h: 50,
			label: "▶ RESTART", spr: spr,
		})
	} else {
		g.buttons = append(g.buttons, &MenuButton{
			x: WinW/2 - 60, y: WinH/2 + 30, w: 120, h: 40,
			label: "▶ RESTART",
		})
	}
}

func (g *Game) initOptionsButtons() {
	g.buttons = nil
	if spr, ok := g.menuSprites["back"]; ok {
		g.buttons = append(g.buttons, &MenuButton{
			x: WinW/2 - 80, y: WinH/2 + 80, w: 160, h: 44,
			label: "▶ BACK", spr: spr,
		})
	} else {
		g.buttons = append(g.buttons, &MenuButton{
			x: WinW/2 - 60, y: WinH/2 + 80, w: 120, h: 40,
			label: "▶ BACK",
		})
	}
}

func (g *Game) playerHit() {
	if g.player.shield {
		g.player.shield = false
		g.player.inv = 60
		g.particles = append(g.particles, spawnParticles(g.player.gx, g.player.gy, C_SHIELD, 16)...)
		playSound(soundPlace)
		return
	}

	g.player.lives--
	g.player.inv = 90
	g.shake.shake(4, 15)
	g.particles = append(g.particles, spawnParticles(g.player.gx, g.player.gy, C_RED, 20)...)
	playSound(soundDie)

	if g.player.lives <= 0 {
		g.state = S_DEAD
		if g.player.score > g.highScore {
			g.highScore = g.player.score
		}
		fmt.Println("💀 Game Over! Score:", g.player.score)
	}
}

func (g *Game) doExplosion(idx int) {
	b := g.bombs[idx]
	g.bombs = append(g.bombs[:idx], g.bombs[idx+1:]...)
	g.player.active--
	playSound(soundExpl)
	g.shake.shake(3, 10)

	g.explos = append(g.explos, Cell{b.gx, b.gy, EXPL_T})
	g.particles = append(g.particles, spawnParticles(b.gx, b.gy, C_EXPL, 8)...)

	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for _, d := range dirs {
		for i := 1; i <= b.radius; i++ {
			nx, ny := b.gx+d[0]*i, b.gy+d[1]*i
			if nx < 0 || nx >= GW || ny < 0 || ny >= GH {
				break
			}
			if g.grid[ny][nx] == T_STONE {
				g.particles = append(g.particles, spawnParticles(nx, ny, C_STONE, 4)...)
				break
			}
			g.explos = append(g.explos, Cell{nx, ny, EXPL_T})

			if g.grid[ny][nx] == T_BRICK {
				g.grid[ny][nx] = T_EMPTY
				g.particles = append(g.particles, spawnParticles(nx, ny, C_BRICK, 10)...)

				for j := len(g.pus) - 1; j >= 0; j-- {
					if g.pus[j].gx == nx && g.pus[j].gy == ny {
						g.pus = append(g.pus[:j], g.pus[j+1:]...)
						break
					}
				}
				break
			}

			for j := range g.bombs {
				if g.bombs[j].gx == nx && g.bombs[j].gy == ny {
					g.bombs[j].timer = 1
				}
			}
		}
	}
}

func (g *Game) applyPowerUp(pu *PowerUp) {
	playSound(soundMenu)
	g.particles = append(g.particles, spawnParticles(pu.gx, pu.gy, puColor(pu.kind), 12)...)

	switch pu.kind {
	case PU_FIRE:
		g.player.radius++
		fmt.Println("🔥 Fire +1 → radius", g.player.radius)
	case PU_BOMB:
		g.player.bombs++
		fmt.Println("💣 Bomb +1 → max", g.player.bombs)
	case PU_SPEED:
		g.player.speedBoost++
		g.player.cd = CD / 2
		fmt.Println("⚡ Speed Up!")
	case PU_HEART:
		g.player.lives++
		fmt.Println("❤ Life +1 →", g.player.lives)
	case PU_SHIELD:
		g.player.shield = true
		fmt.Println("🛡 Shield ON!")
	case PU_KICK:
		g.player.kick = true
		fmt.Println("👟 Kick ON!")
	}
}

func (g *Game) Draw(s *ebiten.Image) {
	s.Fill(C_BG)

	if g.state == S_MENU {
		g.drawMenu(s)
		return
	}

	if g.state == S_OPTIONS {
		g.drawOptions(s)
		return
	}

	if g.state == S_PAUSE {
		g.drawPause(s)
		return
	}

	// Тряска экрана
	sx, sy := g.shake.offset()

	// Поле
	for y := 0; y < GH; y++ {
		for x := 0; x < GW; x++ {
			px, py := x*Tile+int(sx), y*Tile+HUD+int(sy)

			if spr := g.sprites["grass"]; spr != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(float64(Tile)/float64(spr.Bounds().Dx()), float64(Tile)/float64(spr.Bounds().Dy()))
				op.GeoM.Translate(float64(px), float64(py))
				s.DrawImage(spr, op)
			} else {
				rectCached(s, px, py, Tile, Tile, C_GRASS)
			}

			switch g.grid[y][x] {
			case T_STONE:
				if spr := g.sprites["stone"]; spr != nil {
					drawTile(spr, s, px, py)
				} else {
					rectCached(s, px+1, py+1, Tile-2, Tile-2, C_STONE)
					rectCached(s, px+4, py+4, Tile-8, Tile-8, color.RGBA{140, 140, 160, 255})
				}
			case T_BRICK:
				if spr := g.sprites["brick"]; spr != nil {
					drawTile(spr, s, px, py)
				} else {
					rectCached(s, px+1, py+1, Tile-2, Tile-2, C_BRICK)
					rectCached(s, px, py+Tile/2, Tile, 2, color.RGBA{140, 70, 40, 255})
					rectCached(s, px+Tile/2, py, 2, Tile/2, color.RGBA{140, 70, 40, 255})
					rectCached(s, px+Tile/4, py+Tile/2, 2, Tile/2, color.RGBA{140, 70, 40, 255})
					rectCached(s, px+Tile*3/4, py+Tile/2, 2, Tile/2, color.RGBA{140, 70, 40, 255})
				}

				for _, pu := range g.pus {
					if pu.gx == x && pu.gy == y {
						pu.frame++
						if pu.frame%20 < 10 {
							rectCached(s, px+Tile-8, py+2, 6, 6, puColor(pu.kind))
						}
					}
				}
			}
		}
	}

	// Power-ups на поле
	for i := len(g.pus) - 1; i >= 0; i-- {
		pu := &g.pus[i]
		pu.anim = (pu.frame / 10) % 2

		px, py := pu.gx*Tile, pu.gy*Tile+HUD

		if g.player.gx == pu.gx && g.player.gy == pu.gy {
			g.applyPowerUp(pu)
			g.pus = append(g.pus[:i], g.pus[i+1:]...)
			continue
		}

		clr := puColor(pu.kind)
		bobY := py
		if pu.anim == 1 {
			bobY -= 4
		}
		rectCached(s, px+8, bobY+8, Tile-16, Tile-16, clr)
		rectCached(s, px+12, bobY+4, Tile-24, 6, puColor(pu.kind))

		sym := puSymbol(pu.kind)
		text.Draw(s, sym, basicfont.Face7x13, px+16, bobY+Tile-12, C_WHITE)
	}

	// Взрывы
	for _, c := range g.explos {
		px, py := c.gx*Tile, c.gy*Tile+HUD
		a := float64(c.t) / EXPL_T
		if spr := g.sprites["explosion"]; spr != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(Tile)/float64(spr.Bounds().Dx()), float64(Tile)/float64(spr.Bounds().Dy()))
			op.GeoM.Translate(float64(px), float64(py))
			op.ColorM.Scale(1, 1, 1, a)
			s.DrawImage(spr, op)
		} else {
			rectAlpha(s, px, py, Tile, Tile, C_EXPL, a)
		}
	}

	// Частицы
	for _, p := range g.particles {
		a := float64(p.life) / float64(p.maxLife)
		rectAlpha(s, int(p.x)-p.size/2, int(p.y)-p.size/2, p.size, p.size, p.clr, a)
	}

	// Бомбы
	for _, b := range g.bombs {
		px, py := b.gx*Tile, b.gy*Tile+HUD
		pulse := 1.0
		if b.timer < 60 {
			pulse = 1.0 + 0.15*(float64(b.timer%12)/12)
		}
		if spr := g.sprites["bomb"]; spr != nil {
			op := &ebiten.DrawImageOptions{}
			sc := float64(Tile) / float64(spr.Bounds().Dx()) * pulse
			op.GeoM.Scale(sc, sc)
			off := (float64(Tile) - float64(spr.Bounds().Dx())*sc) / 2
			op.GeoM.Translate(float64(px)+off, float64(py)+off)
			s.DrawImage(spr, op)
		} else {
			sz := int(float64(Tile-8) * pulse)
			off := (Tile - sz) / 2
			rectCached(s, px+off, py+off, sz, sz, C_BOMB)
			rectCached(s, px+off+4, py+off+4, sz-8, sz-8, C_BOMB_H)
			rectCached(s, px+Tile/2-1, py+2, 3, 8, C_BRICK)

			if b.timer < 30 && b.timer%4 < 2 {
				rectCached(s, px+off, py+off, sz, sz, C_EXPL)
			}
		}
	}

	// Враги
	for _, e := range g.enemies {
		e.Draw(s, g.sprites)
	}

	// Игрок
	g.player.Draw(s, g.sprites)

	// HUD (без тряски)
	g.drawHUD(s)

	// Debug: клавиши
	if g.keys != "" {
		text.Draw(s, "KEYS:"+g.keys, basicfont.Face7x13, WinW-70, WinH-8, C_KEY_DEBUG)
	}

	// Game Over / Win overlay
	if g.state == S_DEAD {
		g.drawOverlay(s, "GAME OVER", C_RED)
	} else if g.state == S_WIN {
		g.drawOverlay(s, fmt.Sprintf("LEVEL %d COMPLETE! Score: %d", g.level, g.player.score), C_GREEN)
	}
}

func (g *Game) drawHUD(s *ebiten.Image) {
	rectCached(s, 0, 0, WinW, HUD, C_HUD_BG)

	alive := 0
	for _, e := range g.enemies {
		if e.alive {
			alive++
		}
	}

	hud := fmt.Sprintf("♥%d  💣%d/%d  🔥%d  👾%d  Score:%d  [R]Restart  [P]Pause",
		g.player.lives, g.player.active, g.player.bombs, g.player.radius, alive, g.player.score)
	text.Draw(s, hud, basicfont.Face7x13, 8, 25, C_WHITE)
	text.Draw(s, fmt.Sprintf("Lv.%d", g.level), basicfont.Face7x13, WinW-40, 25, C_GREEN)

	x := 250
	if g.player.speedBoost > 0 {
		text.Draw(s, "⚡", basicfont.Face7x13, x, 25, C_PU_SPEED)
		x += 15
	}
	if g.player.shield {
		text.Draw(s, "🛡", basicfont.Face7x13, x, 25, C_PU_SHIELD)
		x += 15
	}
	if g.player.kick {
		text.Draw(s, "👟", basicfont.Face7x13, x, 25, C_PU_KICK)
		x += 15
	}
	if g.player.combos > 1 {
		text.Draw(s, fmt.Sprintf("x%d COMBO!", g.player.combos), basicfont.Face7x13, x, 25, C_YELLOW)
	}
}

func (g *Game) drawMenu(s *ebiten.Image) {
	s.Fill(C_MENU_BG)

	// Фон со звёздами
	if sprBg := g.menuSprites["stars_bg"]; sprBg != nil {
		for i := 0; i < 5; i++ {
			for j := 0; j < 4; j++ {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(i*200), float64(j*200))
				s.DrawImage(sprBg, op)
			}
		}
	}

	// Анимированные звёзды
	if sprStars := g.menuSprites["stars"]; sprStars != nil {
		t := frameCount / 60
		for i := 0; i < 8; i++ {
			x := (i*120 + int(t*20)) % (WinW + 40) - 20
			y := 50 + int(math.Sin(float64(frameCount)/30+float64(i))*20)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(0.5, 0.5)
			op.GeoM.Translate(float64(x), float64(y))
			s.DrawImage(sprStars, op)
		}
	}

	// Меню-частицы
	if sprPart := g.menuSprites["particles1"]; sprPart != nil {
		t := frameCount / 40
		for i := 0; i < 12; i++ {
			x := (i*80 + int(t*10)) % (WinW + 30) - 15
			y := 30 + int(math.Sin(float64(frameCount)/25+float64(i)*1.5)*30) + 100
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(0.3, 0.3)
			op.ColorM.Scale(1, 0.8, 0.5, 0.6)
			op.GeoM.Translate(float64(x), float64(y))
			s.DrawImage(sprPart, op)
		}
	}

	// Заголовок
	title := "BOMBERMAN GO"
	bw := text.BoundString(basicfont.Face7x13, title)
	
	// Тень заголовка
	text.Draw(s, title, basicfont.Face7x13, WinW/2-bw.Dx()/2+2, WinH/2-98, color.RGBA{0, 0, 0, 150})
	text.Draw(s, title, basicfont.Face7x13, WinW/2-bw.Dx()/2, WinH/2-100, C_MENU_TITLE)

	// Подзаголовок
	sub := "Go365 Challenge — Day 96"
	bwSub := text.BoundString(basicfont.Face7x13, sub)
	text.Draw(s, sub, basicfont.Face7x13, WinW/2-bwSub.Dx()/2, WinH/2-75, color.RGBA{150, 150, 150, 255})

	// Кнопки
	for _, btn := range g.buttons {
		btn.Draw(s)
	}

	// Управление
	text.Draw(s, "WASD / Arrows — Move", basicfont.Face7x13, WinW/2-80, WinH/2+80, C_WHITE)
	text.Draw(s, "SPACE — Place Bomb", basicfont.Face7x13, WinW/2-70, WinH/2+100, C_WHITE)
	text.Draw(s, "ESC / P — Pause", basicfont.Face7x13, WinW/2-70, WinH/2+120, C_WHITE)
	text.Draw(s, "R — Restart", basicfont.Face7x13, WinW/2-50, WinH/2+140, C_WHITE)

	// High Score
	if g.highScore > 0 {
		hs := fmt.Sprintf("High Score: %d", g.highScore)
		bwHS := text.BoundString(basicfont.Face7x13, hs)
		text.Draw(s, hs, basicfont.Face7x13, WinW/2-bwHS.Dx()/2, WinH/2+200, C_YELLOW)
	}
}

func (g *Game) drawOptions(s *ebiten.Image) {
	s.Fill(C_MENU_BG)

	// Фон со звёздами
	if sprBg := g.menuSprites["stars_bg"]; sprBg != nil {
		for i := 0; i < 5; i++ {
			for j := 0; j < 4; j++ {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(i*200), float64(j*200))
				s.DrawImage(sprBg, op)
			}
		}
	}
	if sprStars := g.menuSprites["stars"]; sprStars != nil {
		t := frameCount / 60
		for i := 0; i < 8; i++ {
			x := (i*120 + int(t*20)) % (WinW + 40) - 20
			y := 50 + int(math.Sin(float64(frameCount)/30+float64(i))*20)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(0.5, 0.5)
			op.GeoM.Translate(float64(x), float64(y))
			s.DrawImage(sprStars, op)
		}
	}

	// Диалоговая рамка
	frameW, frameH := 300, 180
	frameX, frameY := WinW/2-frameW/2, WinH/2-frameH/2-30
	rectCached(s, frameX, frameY, frameW, frameH, color.RGBA{30, 30, 60, 220})
	rectCached(s, frameX+2, frameY+2, frameW-4, frameH-4, color.RGBA{20, 20, 50, 200})

	// Заголовок
	title := "CONTROLS"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WinW/2-bw.Dx()/2, frameY+20, C_GOLD)

	// Управление
	controls := []string{
		"WASD / Arrows — Move",
		"SPACE — Place Bomb",
		"ESC / P — Pause",
		"R — Restart Level",
	}
	for i, line := range controls {
		text.Draw(s, line, basicfont.Face7x13, WinW/2-80, frameY+50+i*20, C_WHITE)
	}

	// Power-ups legend
	text.Draw(s, "Power-ups:", basicfont.Face7x13, WinW/2-45, frameY+135, C_YELLOW)
	text.Draw(s, "F=Fire  B=Bomb  S=Speed  H=Life  D=Shield  K=Kick", basicfont.Face7x13, WinW/2-165, frameY+155, C_WHITE)

	// Кнопка Back
	for _, btn := range g.buttons {
		btn.Draw(s)
	}
}

func (g *Game) drawPause(s *ebiten.Image) {
	// Затемнение
	rectAlpha(s, 0, 0, WinW, WinH, color.RGBA{0, 0, 0, 180}, 0.7)

	// Заголовок
	title := "PAUSED"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WinW/2-bw.Dx()/2, WinH/2-100, C_YELLOW)

	// Кнопки
	for _, btn := range g.buttons {
		btn.Draw(s)
	}
}

func (g *Game) drawOverlay(s *ebiten.Image, msg string, clr color.Color) {
	rectAlpha(s, 0, 0, WinW, WinH, color.RGBA{0, 0, 0, 180}, 0.7)

	bw := text.BoundString(basicfont.Face7x13, msg)
	text.Draw(s, msg, basicfont.Face7x13, WinW/2-bw.Dx()/2, WinH/2-20, clr)

	sub := "Press ENTER to continue"
	bw2 := text.BoundString(basicfont.Face7x13, sub)
	text.Draw(s, sub, basicfont.Face7x13, WinW/2-bw2.Dx()/2, WinH/2+30, C_WHITE)

	text.Draw(s, fmt.Sprintf("Final Score: %d", g.player.score), basicfont.Face7x13, WinW/2-60, WinH/2+60, C_YELLOW)

	if g.highScore > 0 {
		hs := fmt.Sprintf("High Score: %d", g.highScore)
		bwHS := text.BoundString(basicfont.Face7x13, hs)
		text.Draw(s, hs, basicfont.Face7x13, WinW/2-bwHS.Dx()/2, WinH/2+90, C_YELLOW)
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	return WinW, WinH
}

// ======================== УТИЛИТЫ ========================
func drawTile(spr *ebiten.Image, dst *ebiten.Image, px, py int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(Tile)/float64(spr.Bounds().Dx()), float64(Tile)/float64(spr.Bounds().Dy()))
	op.GeoM.Translate(float64(px), float64(py))
	dst.DrawImage(spr, op)
}

func rectAlpha(s *ebiten.Image, x, y, w, h int, c color.Color, a float64) {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorM.Scale(1, 1, 1, a)
	s.DrawImage(img, op)
}

// ======================== ЗВУК ========================
var (
	audioCtx   *audio.Context
	soundPlace *audio.Player
	soundExpl  *audio.Player
	soundKill  *audio.Player
	soundDie   *audio.Player
	soundWin   *audio.Player
	soundMenu  *audio.Player
)

func generateWAV(samples []float64, sampleRate int) []byte {
	buf := new(bytes.Buffer)
	numCh := uint16(1)
	bitsPS := uint16(16)
	byteRate := sampleRate * int(numCh) * int(bitsPS) / 8
	blockAlign := numCh * bitsPS / 8
	dataSize := len(samples) * 2

	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, numCh)
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(byteRate))
	binary.Write(buf, binary.LittleEndian, blockAlign)
	binary.Write(buf, binary.LittleEndian, bitsPS)
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))

	for _, s := range samples {
		if s > 1 {
			s = 1
		}
		if s < -1 {
			s = -1
		}
		binary.Write(buf, binary.LittleEndian, int16(s*32767))
	}
	return buf.Bytes()
}

func makeBeep(duration float64, freq float64, sr int) []float64 {
	n := int(float64(sr) * duration)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		out[i] = math.Sin(2*math.Pi*freq*t) * math.Exp(-t*8) * 0.5
	}
	return out
}

func makeNoise(duration float64, sr int) []float64 {
	n := int(float64(sr) * duration)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		out[i] = (rand.Float64()*2 - 1) * math.Exp(-t*6) * 0.6
	}
	return out
}

func makeSweep(duration float64, f1, f2 float64, sr int) []float64 {
	n := int(float64(sr) * duration)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		p := float64(i) / float64(n)
		freq := f1 + (f2-f1)*p
		out[i] = math.Sin(2*math.Pi*freq*t) * math.Exp(-t*4) * 0.5
	}
	return out
}

func makeArp(duration float64, freqs []float64, sr int) []float64 {
	n := int(float64(sr) * duration)
	out := make([]float64, n)
	stepDur := duration / float64(len(freqs))
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		idx := int(t / stepDur)
		if idx >= len(freqs) {
			idx = len(freqs) - 1
		}
		out[i] = math.Sin(2*math.Pi*freqs[idx]*t) * math.Exp(-t*5) * 0.4
	}
	return out
}

func initAudio() {
	audioCtx = audio.NewContext(44100)

	wavPlace := generateWAV(makeBeep(0.08, 880, 44100), 44100)
	soundPlace, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavPlace))

	wavExpl := generateWAV(makeNoise(0.4, 44100), 44100)
	soundExpl, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavExpl))

	wavKill := generateWAV(makeSweep(0.2, 600, 200, 44100), 44100)
	soundKill, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavKill))

	wavDie := generateWAV(makeSweep(0.5, 400, 80, 44100), 44100)
	soundDie, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavDie))

	wavWin := generateWAV(makeArp(0.5, []float64{523, 659, 784, 1047}, 44100), 44100)
	soundWin, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavWin))

	wavMenu := generateWAV(makeBeep(0.05, 660, 44100), 44100)
	soundMenu, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavMenu))
}

func playSound(player *audio.Player) {
	if player == nil {
		return
	}
	player.Rewind()
	player.Play()
}

// ======================== ГЛОБАЛЬНЫЕ ========================
var frameCount int64

// ======================== MAIN ========================
func main() {
	fmt.Println("═══════════════════════════════════")
	fmt.Println("  BOMBERMAN GO — Go365 Day 96")
	fmt.Println("  4 Enemy Types | 6 Power-Ups")
	fmt.Println("  New: Menu | Pause | High Score")
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	ebiten.SetWindowSize(WinW, WinH)
	ebiten.SetWindowTitle("Bomberman Go — Go365 Day 96")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
