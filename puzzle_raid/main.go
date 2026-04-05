package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// ======================== КОНСТАНТЫ ========================
const (
	COLS      = 8
	ROWS      = 8
	TILE      = 56
	BOARD_OFFX = 48
	BOARD_OFFY = 80
	HUD       = 50
	COLORS    = 4 // blue, green, red, yellow

	WIN_W = COLS*TILE + BOARD_OFFX*2
	WIN_H = ROWS*TILE + BOARD_OFFY + HUD + 20
)

// ======================== СОСТОЯНИЯ ========================
type State int

const (
	S_MENU State = iota
	S_PLAY
	S_PAUSE
	S_DEAD
	S_WIN
)

// ======================== ЦВЕТА ========================
var (
	C_BG       = color.RGBA{15, 15, 35, 255}
	C_HUD_BG   = color.RGBA{10, 10, 25, 255}
	C_WHITE    = color.RGBA{255, 255, 255, 255}
	C_GREEN    = color.RGBA{100, 255, 100, 255}
	C_RED      = color.RGBA{255, 80, 80, 255}
	C_YELLOW   = color.RGBA{255, 255, 80, 255}
	C_GOLD     = color.RGBA{255, 220, 100, 255}
	C_TILE_BG  = color.RGBA{30, 30, 50, 255}
)

// ======================== ЧАСТИЦЫ ========================
type Particle struct {
	x, y    float64
	vx, vy  float64
	life    int
	maxLife int
	clr     color.RGBA
	size    int
}

func spawnParticles(px, py int, clr color.RGBA, count int) []Particle {
	ps := make([]Particle, 0, count)
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 1.0 + rand.Float64()*3.0
		ps = append(ps, Particle{
			x: float64(px), y: float64(py),
			vx: math.Cos(angle) * speed, vy: math.Sin(angle)*speed - 1,
			life: 20 + rand.Intn(20), maxLife: 40,
			clr: clr, size: 2 + rand.Intn(4),
		})
	}
	return ps
}

// ======================== МОНЕТА ========================
type Coin struct {
	gx, gy int
	value  int
	anim   int
	frame  int
}

// ======================== ВРАГ ========================
type Enemy struct {
	gx, gy  int
	dir     int
	cd      int
	alive   bool
	frame   int
	hp      int
	kind    int // 0=slow, 1=fast, 2=smart
}

func NewEnemy(gx, gy, kind int) *Enemy {
	hp := 1
	if kind == 2 {
		hp = 2
	}
	return &Enemy{gx: gx, gy: gy, kind: kind, dir: rand.Intn(4), alive: true, cd: 20 + rand.Intn(20), hp: hp}
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

	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	if e.kind == 2 && g.player != nil {
		// Smart: движется к игроку
		dx := g.player.gx - e.gx
		dy := g.player.gy - e.gy
		moves := [][2]int{}
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
		rand.Shuffle(len(moves), func(i, j int) { moves[i], moves[j] = moves[j], moves[i] })
		moves = append(moves, dirs...)

		for _, d := range moves {
			nx, ny := e.gx+d[0], e.gy+d[1]
			if g.validEnemyPos(nx, ny) {
				e.gx, e.gy = nx, ny
				e.cd = 25 + rand.Intn(15)
				return
			}
		}
	} else {
		// Random
		rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
		for _, d := range dirs {
			nx, ny := e.gx+d[0], e.gy+d[1]
			if g.validEnemyPos(nx, ny) {
				e.gx, e.gy = nx, ny
				e.cd = 30 + rand.Intn(20)
				if e.kind == 1 {
					e.cd -= 10 // fast
				}
				return
			}
		}
	}
}

// ======================== ИГРОК ========================
type Player struct {
	gx, gy  int
	inv     int
	lives   int
	cd      int
	frame   int
	score   int
	combo   int
	comboT  int
	strikeR int // радиус удара
	speed   int // frames cooldown
}

func NewPlayer() *Player {
	return &Player{gx: COLS / 2, gy: ROWS / 2, lives: 3, inv: 90, cd: 15, strikeR: 1, speed: 15}
}

func (p *Player) Update(g *Game) {
	if p.inv > 0 {
		p.inv--
	}
	if p.cd > 0 {
		p.cd--
	}
	if p.comboT > 0 {
		p.comboT--
		if p.comboT == 0 {
			p.combo = 0
		}
	}

	// Движение
	dx, dy := 0, 0
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		dy = -1
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		dy = 1
	} else if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		dx = -1
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		dx = 1
	}

	if (dx != 0 || dy != 0) && p.cd <= 0 {
		nx, ny := p.gx+dx, p.gy+dy
		if nx >= 0 && nx < COLS && ny >= 0 && ny < ROWS {
			p.gx, p.gy = nx, ny
			p.cd = p.speed
			p.frame++

			// Проверяем монеты
			for i := len(g.coins) - 1; i >= 0; i-- {
				if g.coins[i].gx == p.gx && g.coins[i].gy == p.gy {
					g.coins = append(g.coins[:i], g.coins[i+1:]...)
					p.score += 50
					playSoundP(sndCoin)
				}
			}
		}
	}

	// Удар
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.doStrike(p)
	}
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
		bg := color.RGBA{40, 40, 80, 200}
		if b.hover {
			bg = color.RGBA{60, 60, 120, 230}
		}
		rectC(s, b.x, b.y, b.w, b.h, bg)
		bw := text.BoundString(basicfont.Face7x13, b.label)
		text.Draw(s, b.label, basicfont.Face7x13, b.x+b.w/2-bw.Dx()/2, b.y+b.h/2+5, C_WHITE)
	}
}

// ======================== ИГРА ========================
type Game struct {
	board  [ROWS][COLS]int // 0=empty, 1-4=colors
	state  State
	player *Player
	enemies []*Enemy
	coins  []Coin
	particles []Particle

	level    int
	timeLeft int
	maxTime  int
	target   int // очков для победы

	buttons     []*MenuButton
	menuSprites map[string]*ebiten.Image
	tileSprites [COLORS][]*ebiten.Image
	coinSprites []*ebiten.Image
	menuAnim    int
	highScore   int
	enterPrev   bool

	shakeIntensity float64
	shakeTimer     int
}

func NewGame() *Game {
	g := &Game{
		state:       S_MENU,
		menuSprites: make(map[string]*ebiten.Image),
		tileSprites: [COLORS][]*ebiten.Image{},
	}
	g.loadSprites()
	g.initMenuButtons()
	return g
}

func (g *Game) loadSprites() {
	// Tile sprites: 4 colors
	colors := []string{"blue", "green", "red", "yellow"}
	for ci, c := range colors {
		for i := 1; i <= 10; i++ {
			path := fmt.Sprintf("assets/sprites/tiles_%s/tile%s_%02d.png", c, capitalize(c), i)
			img, _, err := ebitenutil.NewImageFromFile(path)
			if err == nil {
				scaled := scaleImg(img, TILE, TILE)
				g.tileSprites[ci] = append(g.tileSprites[ci], scaled)
			}
		}
	}

	// Coin sprites
	for i := 1; i <= 10; i++ {
		path := fmt.Sprintf("assets/sprites/coins/coin_%02d.png", i)
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err == nil {
			g.coinSprites = append(g.coinSprites, scaleImg(img, 32, 32))
		}
	}

	// Menu sprites
	menuFiles := map[string]string{
		"play":    "play button.png",
		"back":    "Back Button.png",
		"stars":   "stars.png",
		"stars_bg": "stars back.png",
	}
	for name, file := range menuFiles {
		path := "assets/sprites/menu/" + file
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err == nil {
			g.menuSprites[name] = img
		}
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return s[:1] + s[1:]
}

func scaleImg(src *ebiten.Image, w, h int) *ebiten.Image {
	b := src.Bounds()
	dst := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(b.Dx()), float64(h)/float64(b.Dy()))
	dst.DrawImage(src, op)
	return dst
}

func (g *Game) initMenuButtons() {
	g.buttons = nil
	if spr, ok := g.menuSprites["play"]; ok {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 80, y: WIN_H/2 - 20, w: 160, h: 50,
			label: "PLAY", spr: spr,
		})
	} else {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 60, y: WIN_H/2 - 20, w: 120, h: 40,
			label: "▶ PLAY",
		})
	}
}

func (g *Game) initPauseButtons() {
	g.buttons = nil
	if spr, ok := g.menuSprites["back"]; ok {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 80, y: WIN_H/2 - 40, w: 160, h: 50,
			label: "▶ RESUME", spr: spr,
		})
	} else {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 60, y: WIN_H/2 - 40, w: 120, h: 40,
			label: "▶ RESUME",
		})
	}
	if spr, ok := g.menuSprites["play"]; ok {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 80, y: WIN_H/2 + 30, w: 160, h: 50,
			label: "▶ RESTART", spr: spr,
		})
	} else {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 60, y: WIN_H/2 + 30, w: 120, h: 40,
			label: "▶ RESTART",
		})
	}
}

func (g *Game) startLevel() {
	g.board = [ROWS][COLS]int{}
	g.coins = nil
	g.enemies = nil
	g.particles = nil

	// Заполняем поле тайлами
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			g.board[r][c] = 1 + rand.Intn(COLORS)
		}
	}

	// Удаляем начальные match-3
	for {
		matches := g.findMatches()
		if len(matches) == 0 {
			break
		}
		for _, m := range matches {
			g.board[m[0]][m[1]] = 1 + rand.Intn(COLORS)
		}
	}

	g.player = NewPlayer()
	g.maxTime = 60 + g.level*10
	g.timeLeft = g.maxTime
	g.target = 500 * g.level

	// Враги
	enemyCount := 2 + g.level
	for i := 0; i < enemyCount; i++ {
		for attempts := 0; attempts < 30; attempts++ {
			ex := rand.Intn(COLS)
			ey := rand.Intn(ROWS)
			dist := abs(ex-g.player.gx) + abs(ey-g.player.gy)
			if dist > 3 {
				kind := 0
				if g.level >= 2 && i < enemyCount/3 {
					kind = 1
				}
				if g.level >= 3 && i < enemyCount/4 {
					kind = 2
				}
				g.enemies = append(g.enemies, NewEnemy(ex, ey, kind))
				break
			}
		}
	}

	// Монеты
	coinCount := 3 + g.level
	for i := 0; i < coinCount; i++ {
		cx, cy := rand.Intn(COLS), rand.Intn(ROWS)
		g.coins = append(g.coins, Coin{gx: cx, gy: cy, value: 50})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *Game) validEnemyPos(nx, ny int) bool {
	if nx < 0 || nx >= COLS || ny < 0 || ny >= ROWS {
		return false
	}
	if g.board[ny][nx] == 0 {
		return false
	}
	for _, e := range g.enemies {
		if e.alive && e.gx == nx && e.gy == ny {
			return false
		}
	}
	return true
}

func (g *Game) findMatches() [][2]int {
	matched := make(map[[2]int]bool)

	// Horizontal
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS-2; c++ {
			v := g.board[r][c]
			if v == 0 {
				continue
			}
			if g.board[r][c+1] == v && g.board[r][c+2] == v {
				matched[[2]int{r, c}] = true
				matched[[2]int{r, c + 1}] = true
				matched[[2]int{r, c + 2}] = true
			}
		}
	}

	// Vertical
	for c := 0; c < COLS; c++ {
		for r := 0; r < ROWS-2; r++ {
			v := g.board[r][c]
			if v == 0 {
				continue
			}
			if g.board[r+1][c] == v && g.board[r+2][c] == v {
				matched[[2]int{r, c}] = true
				matched[[2]int{r + 1, c}] = true
				matched[[2]int{r + 2, c}] = true
			}
		}
	}

	result := make([][2]int, 0, len(matched))
	for k := range matched {
		result = append(result, k)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i][0] == result[j][0] {
			return result[i][1] < result[j][1]
		}
		return result[i][0] < result[j][0]
	})
	return result
}

func (g *Game) doStrike(p *Player) {
	r := p.strikeR
	colorsHit := make(map[int]int)
	destroyed := 0

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			nx, ny := p.gx+dx, p.gy+dy
			if nx >= 0 && nx < COLS && ny >= 0 && ny < ROWS {
				if g.board[ny][nx] != 0 {
					clr := g.board[ny][nx]
					colorsHit[clr]++
					px := nx*TILE + BOARD_OFFX + TILE/2
					py := ny*TILE + BOARD_OFFY + TILE/2
					g.particles = append(g.particles, spawnParticles(px, py, tileColor(clr), 6)...)
					g.board[ny][nx] = 0
					destroyed++
				}
			}
		}
	}

	if destroyed > 0 {
		g.shake(3, 8)
		playSoundP(sndStrike)

		// Подсчёт очков: комбо за одинаковый цвет
		for clr, count := range colorsHit {
			base := count * 10
			if count >= 3 {
				p.combo++
				p.comboT = 60
				base *= p.combo
			}
			p.score += base

			// Спавн монет
			if count >= 2 {
				for i := 0; i < count/2; i++ {
					g.coins = append(g.coins, Coin{gx: p.gx, gy: p.gy, value: 25})
				}
			}
			_ = clr
		}
	}
}

func tileColor(clr int) color.RGBA {
	switch clr {
	case 1:
		return color.RGBA{80, 120, 255, 255}
	case 2:
		return color.RGBA{80, 255, 80, 255}
	case 3:
		return color.RGBA{255, 80, 80, 255}
	case 4:
		return color.RGBA{255, 255, 80, 255}
	default:
		return C_WHITE
	}
}

func (g *Game) shake(intensity float64, timer int) {
	g.shakeIntensity = intensity
	g.shakeTimer = timer
}

func (g *Game) Update() error {
	frameCount++

	// ===== MENU =====
	if g.state == S_MENU {
		g.menuAnim++
		mx, my := ebiten.CursorPosition()
		for _, btn := range g.buttons {
			btn.hover = btn.contains(mx, my)
			if btn.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.state = S_PLAY
				g.level = 1
				g.startLevel()
				g.highScore = 0
				return nil
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = S_PLAY
			g.level = 1
			g.startLevel()
			g.highScore = 0
		}
		return nil
	}

	// ===== PAUSE =====
	if g.state == S_PAUSE {
		esc := ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyP)
		if esc && !g.enterPrev {
			g.state = S_PLAY
		}
		mx, my := ebiten.CursorPosition()
		for _, btn := range g.buttons {
			btn.hover = btn.contains(mx, my)
			if btn.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if btn.label == "▶ RESUME" {
					g.state = S_PLAY
				} else if btn.label == "▶ RESTART" {
					g.state = S_PLAY
					g.level = 1
					g.startLevel()
				}
			}
		}
		g.enterPrev = esc
		return nil
	}

	// ===== DEAD/WIN =====
	if g.state == S_DEAD || g.state == S_WIN {
		enter := ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace)
		if enter && !g.enterPrev {
			g.state = S_MENU
			g.initMenuButtons()
		}
		g.enterPrev = enter
		return nil
	}

	// ===== PLAY =====
	if g.shakeTimer > 0 {
		g.shakeTimer--
	}

	// Timer
	if frameCount%60 == 0 {
		g.timeLeft--
		if g.timeLeft <= 0 {
			if g.player.score >= g.target {
				g.state = S_WIN
				if g.player.score > g.highScore {
					g.highScore = g.player.score
				}
			} else {
				g.state = S_DEAD
				if g.player.score > g.highScore {
					g.highScore = g.player.score
				}
			}
			return nil
		}
	}

	g.player.Update(g)

	// Enemies
	for _, e := range g.enemies {
		e.Update(g)
		if e.alive && g.player.inv <= 0 && e.gx == g.player.gx && e.gy == g.player.gy {
			g.playerHit()
		}
	}

	// Particles
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

	// Coins anim
	for i := range g.coins {
		g.coins[i].frame++
	}

	// Check win
	if g.player.score >= g.target {
		g.state = S_WIN
		if g.player.score > g.highScore {
			g.highScore = g.player.score
		}
	}

	// Pause
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.state = S_PAUSE
		g.initPauseButtons()
	}

	return nil
}

func (g *Game) playerHit() {
	g.player.lives--
	g.player.inv = 90
	g.shake(5, 15)
	px := g.player.gx*TILE + BOARD_OFFX + TILE/2
	py := g.player.gy*TILE + BOARD_OFFY + TILE/2
	g.particles = append(g.particles, spawnParticles(px, py, C_RED, 20)...)

	if g.player.lives <= 0 {
		g.state = S_DEAD
		if g.player.score > g.highScore {
			g.highScore = g.player.score
		}
	}
}

func (g *Game) Draw(s *ebiten.Image) {
	s.Fill(C_BG)

	if g.state == S_MENU {
		g.drawMenu(s)
		return
	}
	if g.state == S_PAUSE {
		g.drawPause(s)
		return
	}

	// Shake offset
	sx, sy := 0.0, 0.0
	if g.shakeTimer > 0 {
		sx = (rand.Float64() - 0.5) * g.shakeIntensity * 2
		sy = (rand.Float64() - 0.5) * g.shakeIntensity * 2
	}

	// Board background
	rectC(s, BOARD_OFFX-4+int(sx), BOARD_OFFY-4+int(sy), COLS*TILE+8, ROWS*TILE+8, C_TILE_BG)

	// Tiles
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			v := g.board[r][c]
			if v == 0 {
				continue
			}
			x := c*TILE + BOARD_OFFX + int(sx)
			y := r*TILE + BOARD_OFFY + int(sy)

			// Sprite or fallback
			sprites := g.tileSprites[v-1]
			if len(sprites) > 0 {
				spr := sprites[int(frameCount/30)%len(sprites)]
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x), float64(y))
				s.DrawImage(spr, op)
			} else {
				rectC(s, x+2, y+2, TILE-4, TILE-4, tileColor(v))
			}
		}
	}

	// Coins
	for _, coin := range g.coins {
		x := coin.gx*TILE + BOARD_OFFX + TILE/2 - 16 + int(sx)
		y := coin.gy*TILE + BOARD_OFFY + TILE/2 - 16 + int(sy) + int(math.Sin(float64(coin.frame)/10)*4)

		if len(g.coinSprites) > 0 {
			spr := g.coinSprites[int(coin.frame/8)%len(g.coinSprites)]
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			s.DrawImage(spr, op)
		} else {
			rectC(s, x+8, y+8, 16, 16, C_GOLD)
		}
	}

	// Enemies
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		x := e.gx*TILE + BOARD_OFFX + int(sx)
		y := e.gy*TILE + BOARD_OFFY + int(sy)
		clr := color.RGBA{200, 80, 80, 255}
		if e.kind == 1 {
			clr = color.RGBA{255, 150, 50, 255}
		} else if e.kind == 2 {
			clr = color.RGBA{180, 80, 255, 255}
		}
		rectC(s, x+4, y+4, TILE-8, TILE-8, clr)
		rectC(s, x+12, y+10, 8, 8, C_WHITE)
		rectC(s, x+TILE-20, y+10, 8, 8, C_WHITE)
	}

	// Particles
	for _, p := range g.particles {
		a := float64(p.life) / float64(p.maxLife)
		rectAlphaC(s, int(p.x)-p.size/2+int(sx), int(p.y)-p.size/2+int(sy), p.size, p.size, p.clr, a)
	}

	// Player
	if g.player != nil {
		if g.player.inv <= 0 || g.player.inv%6 < 3 {
			x := g.player.gx*TILE + BOARD_OFFX + int(sx)
			y := g.player.gy*TILE + BOARD_OFFY + int(sy)
			rectC(s, x+6, y+6, TILE-12, TILE-12, color.RGBA{100, 180, 255, 255})
			rectC(s, x+14, y+2, TILE-28, 14, color.RGBA{150, 220, 255, 255})
		}
	}

	// HUD
	g.drawHUD(s)

	// Overlays
	if g.state == S_DEAD {
		g.drawOverlay(s, "GAME OVER", C_RED)
	} else if g.state == S_WIN {
		g.drawOverlay(s, fmt.Sprintf("LEVEL %d COMPLETE!", g.level), C_GREEN)
	}
}

func (g *Game) drawHUD(s *ebiten.Image) {
	rectC(s, 0, 0, WIN_W, HUD, C_HUD_BG)

	if g.player == nil {
		return
	}

	hud := fmt.Sprintf("♥%d  Score:%d  Lv.%d  Time:%ds  Target:%d",
		g.player.lives, g.player.score, g.level, g.timeLeft, g.target)
	text.Draw(s, hud, basicfont.Face7x13, 8, 28, C_WHITE)

	if g.player.combo > 1 {
		text.Draw(s, fmt.Sprintf("x%d COMBO!", g.player.combo), basicfont.Face7x13, WIN_W-120, 28, C_YELLOW)
	}
}

func (g *Game) drawMenu(s *ebiten.Image) {
	s.Fill(C_BG)

	// Stars bg
	if sprBg := g.menuSprites["stars_bg"]; sprBg != nil {
		for i := 0; i < 4; i++ {
			for j := 0; j < 3; j++ {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(i*200), float64(j*200))
				s.DrawImage(sprBg, op)
			}
		}
	}
	if sprStars := g.menuSprites["stars"]; sprStars != nil {
		t := frameCount / 60
		for i := 0; i < 8; i++ {
			x := (i*90 + int(t*15)) % (WIN_W + 40) - 20
			y := 40 + int(math.Sin(float64(frameCount)/30+float64(i))*15)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(0.4, 0.4)
			op.GeoM.Translate(float64(x), float64(y))
			s.DrawImage(sprStars, op)
		}
	}

	title := "PUZZLE RAID"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2+2, WIN_H/2-98, color.RGBA{0, 0, 0, 150})
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-100, C_GOLD)

	sub := "Go365 Challenge — Day 97"
	bwSub := text.BoundString(basicfont.Face7x13, sub)
	text.Draw(s, sub, basicfont.Face7x13, WIN_W/2-bwSub.Dx()/2, WIN_H/2-75, color.RGBA{150, 150, 150, 255})

	for _, btn := range g.buttons {
		btn.Draw(s)
	}

	text.Draw(s, "WASD/Arrows — Move", basicfont.Face7x13, WIN_W/2-80, WIN_H/2+80, C_WHITE)
	text.Draw(s, "SPACE — Strike tiles", basicfont.Face7x13, WIN_W/2-75, WIN_H/2+100, C_WHITE)
	text.Draw(s, "ESC/P — Pause", basicfont.Face7x13, WIN_W/2-60, WIN_H/2+120, C_WHITE)

	text.Draw(s, "Match 3+ same color = COMBO!", basicfont.Face7x13, WIN_W/2-115, WIN_H/2+160, C_YELLOW)

	if g.highScore > 0 {
		hs := fmt.Sprintf("High Score: %d", g.highScore)
		bwHS := text.BoundString(basicfont.Face7x13, hs)
		text.Draw(s, hs, basicfont.Face7x13, WIN_W/2-bwHS.Dx()/2, WIN_H/2+200, C_GOLD)
	}
}

func (g *Game) drawPause(s *ebiten.Image) {
	rectAlphaC(s, 0, 0, WIN_W, WIN_H, color.RGBA{0, 0, 0, 180}, 0.7)
	title := "PAUSED"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-100, C_YELLOW)
	for _, btn := range g.buttons {
		btn.Draw(s)
	}
}

func (g *Game) drawOverlay(s *ebiten.Image, msg string, clr color.Color) {
	rectAlphaC(s, 0, 0, WIN_W, WIN_H, color.RGBA{0, 0, 0, 180}, 0.7)
	bw := text.BoundString(basicfont.Face7x13, msg)
	text.Draw(s, msg, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-20, clr)
	sub := "Press ENTER for menu"
	bw2 := text.BoundString(basicfont.Face7x13, sub)
	text.Draw(s, sub, basicfont.Face7x13, WIN_W/2-bw2.Dx()/2, WIN_H/2+30, C_WHITE)

	if g.player != nil {
		text.Draw(s, fmt.Sprintf("Score: %d", g.player.score), basicfont.Face7x13, WIN_W/2-50, WIN_H/2+60, C_GOLD)
	}
	if g.highScore > 0 {
		hs := fmt.Sprintf("High Score: %d", g.highScore)
		bwHS := text.BoundString(basicfont.Face7x13, hs)
		text.Draw(s, hs, basicfont.Face7x13, WIN_W/2-bwHS.Dx()/2, WIN_H/2+90, C_GOLD)
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	return WIN_W, WIN_H
}

// ======================== УТИЛИТЫ ========================
var cachedRects = make(map[string]*ebiten.Image)

func rectC(s *ebiten.Image, x, y, w, h int, c color.Color) {
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	key := fmt.Sprintf("%d_%d_%d_%d_%d_%d", rgba.R, rgba.G, rgba.B, rgba.A, w, h)
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

func rectAlphaC(s *ebiten.Image, x, y, w, h int, c color.Color, a float64) {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorM.Scale(1, 1, 1, a)
	s.DrawImage(img, op)
}

// ======================== ЗВУК ========================
var sndStrike, sndCoin *PlayerSound

type PlayerSound struct {
	data []byte
}

func generateWAV(samples []float64, sampleRate int) []byte {
	buf := make([]byte, 44+len(samples)*2)
	copy(buf[:4], []byte("RIFF"))
	dataSize := uint32(36 + len(samples)*2)
	buf[4] = byte(dataSize)
	buf[5] = byte(dataSize >> 8)
	buf[6] = byte(dataSize >> 16)
	buf[7] = byte(dataSize >> 24)
	copy(buf[8:12], []byte("WAVE"))
	copy(buf[12:16], []byte("fmt "))
	buf[16] = 16
	buf[20] = 1
	buf[22] = 1
	buf[24] = 68
	buf[25] = 172
	buf[26] = 0
	buf[27] = 0
	buf[28] = 136
	buf[29] = 172
	buf[30] = 0
	buf[31] = 0
	buf[32] = 2
	buf[33] = 0
	buf[34] = 16
	buf[35] = 0
	copy(buf[36:40], []byte("data"))
	ds := uint32(len(samples) * 2)
	buf[40] = byte(ds)
	buf[41] = byte(ds >> 8)
	buf[42] = byte(ds >> 16)
	buf[43] = byte(ds >> 24)

	for i, v := range samples {
		if v > 1 {
			v = 1
		}
		if v < -1 {
			v = -1
		}
		s := int16(v * 32767)
		buf[44+i*2] = byte(s)
		buf[44+i*2+1] = byte(s >> 8)
	}
	return buf
}

func makeBeep(dur float64, freq float64, sr int) []float64 {
	n := int(float64(sr) * dur)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		out[i] = math.Sin(6.2832*freq*t) * math.Exp(-t*8) * 0.5
	}
	return out
}

func makeNoise(dur float64, sr int) []float64 {
	n := int(float64(sr) * dur)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		out[i] = (rand.Float64()*2 - 1) * math.Exp(-t*6) * 0.6
	}
	return out
}

func playSoundP(snd *PlayerSound) {
	_ = snd
}

// ======================== MAIN ========================
var frameCount int64

func main() {
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  PUZZLE RAID — Go365 Day 97")
	fmt.Println("  Match-3 + Action + Enemies + Coins")
	fmt.Println("═══════════════════════════════════════")

	ebiten.SetWindowSize(WIN_W, WIN_H)
	ebiten.SetWindowTitle("Puzzle Raid — Go365 Day 97")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
