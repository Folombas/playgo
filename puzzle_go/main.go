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
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// ======================== КОНСТАНТЫ ========================
const (
	COLS       = 8
	ROWS       = 8
	TILE       = 64
	HUD        = 60
	GEM_TYPES  = 6
	MATCH_MIN  = 3
	BOARD_OFFX = 32
	BOARD_OFFY = 80
	WIN_W      = COLS*TILE + BOARD_OFFX*2
	WIN_H      = ROWS*TILE + BOARD_OFFY + HUD
	TARGET_SCORE = 5000 // Цель для победы
)

// ======================== СОСТОЯНИЯ ========================
type State int

const (
	S_MENU State = iota
	S_PLAY
	S_PAUSE
	S_OPTIONS
	S_WIN
	S_LOSE
	S_NO_MOVES // Нет ходов — перемешивание
)

// ======================== КЭШ ========================
var (
	// 1x1 белый пиксель для рисования прямоугольников
	whitePixel *ebiten.Image

	// Гемы
	gemSprites [GEM_TYPES]*ebiten.Image

	// UI спрайты
	coinSprite    *ebiten.Image
	selectorSpr   *ebiten.Image
	bgTileSpr     *ebiten.Image
	hudBgSpr      *ebiten.Image

	// Звуки
	actx     *audio.Context
	sndMatch *audio.Player
	sndSwap  *audio.Player
	sndBad   *audio.Player
	sndCombo *audio.Player
	sndWin   *audio.Player

	loaded bool
)

func ensureWhitePixel() {
	if whitePixel == nil {
		whitePixel = ebiten.NewImage(1, 1)
		whitePixel.Fill(color.White)
	}
}

// ======================== ЗАГРУЗКА СПРАЙТОВ ========================
func loadSprites() {
	if loaded {
		return
	}
	loaded = true

	ensureWhitePixel()

	// Гемы — пробуем новые jewel спрайты, fallback на gem0-5
	jewelNames := []string{
		"jewelblue_0",   // 0 — синий
		"jewelred",      // 1 — красный
		"jewelgreen",    // 2 — зелёный
		"jewelyellow",   // 3 — жёлтый
		"jewelviolet",   // 4 — фиолетовый
		"jewelorange",   // 5 — оранжевый
	}
	fallbackNames := []string{"gem0", "gem1", "gem2", "gem3", "gem4", "gem5"}

	for i := 0; i < GEM_TYPES; i++ {
		// Пробуем jewel
		img, _, _ := ebitenutil.NewImageFromFile(fmt.Sprintf("assets/sprites/%s.png", jewelNames[i]))
		if img == nil {
			// Fallback на gem
			img, _, _ = ebitenutil.NewImageFromFile(fmt.Sprintf("assets/sprites/%s.png", fallbackNames[i]))
		}
		if img != nil {
			// Масштабируем к TILE
			w, h := img.Size()
			if w != TILE || h != TILE {
				gemSprites[i] = scaleImg(img, TILE, TILE)
			} else {
				gemSprites[i] = img
			}
		}
	}

	// Coin
	coinSprite, _, _ = ebitenutil.NewImageFromFile("assets/sprites/coin.png")

	// Selector
	selectorSpr, _, _ = ebitenutil.NewImageFromFile("assets/sprites/selector.png")

	// Background tile
	bgTileSpr, _, _ = ebitenutil.NewImageFromFile("assets/sprites/ground.png")
	if bgTileSpr == nil {
		bgTileSpr, _, _ = ebitenutil.NewImageFromFile("assets/sprites/backtiles/BackTile_01.png")
	}
	if bgTileSpr == nil {
		bgTileSpr = mkImg(TILE, TILE, color.RGBA{30, 30, 50, 255})
	}

	// HUD background
	hudBgSpr = mkImg(WIN_W, HUD, color.RGBA{20, 20, 40, 255})
}

func scaleImg(src *ebiten.Image, w, h int) *ebiten.Image {
	dst := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	sw, sh := src.Size()
	op.GeoM.Scale(float64(w)/float64(sw), float64(h)/float64(sh))
	dst.DrawImage(src, op)
	return dst
}

func mkImg(w, h int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	return img
}

// ======================== АУДИО ========================
func initAudio() {
	actx = audio.NewContext(44100)
	sndMatch, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(arp(0.2, []float64{523, 659, 784}, 44100))))
	sndSwap, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(tone(0.1, 440, 44100))))
	sndBad, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(noise(0.15, 44100))))
	sndCombo, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(arp(0.3, []float64{523, 659, 784, 1047}, 44100))))
	sndWin, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(arp(0.5, []float64{523, 659, 784, 1047, 1319}, 44100))))
}

func play(p *audio.Player) {
	if p == nil {
		return
	}
	p.Rewind()
	p.Play()
}

func mkWAV(s []int16) []byte {
	buf := &bytes.Buffer{}
	// WAV header
	write := func(data interface{}) { binary.Write(buf, binary.LittleEndian, data) }
	write([]byte("RIFF"))
	write(uint32(36 + len(s)*2))
	write([]byte("WAVE"))
	write([]byte("fmt "))
	write(uint32(16))
	write(uint16(1))
	write(uint16(1))
	write(uint32(44100))
	write(uint32(88200))
	write(uint16(2))
	write(uint16(16))
	write([]byte("data"))
	write(uint32(len(s) * 2))
	for _, v := range s {
		write(v)
	}
	return buf.Bytes()
}

func tone(dur float64, freq float64, sr int) []int16 {
	n := int(float64(sr) * dur)
	s := make([]int16, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		s[i] = int16(math.Sin(2*math.Pi*freq*t) * 32767 * 0.5)
	}
	return s
}

func noise(dur float64, sr int) []int16 {
	n := int(float64(sr) * dur)
	s := make([]int16, n)
	for i := 0; i < n; i++ {
		s[i] = int16((rand.Float64()*2 - 1) * 32767 * 0.3)
	}
	return s
}

func arp(dur float64, fs []float64, sr int) []int16 {
	n := int(float64(sr) * dur)
	s := make([]int16, n)
	seg := n / len(fs)
	for idx, f := range fs {
		start := idx * seg
		end := start + seg
		if idx == len(fs)-1 {
			end = n
		}
		for i := start; i < end; i++ {
			t := float64(i-start) / float64(seg)
			env := 1.0 - t
			s[i] = int16(math.Sin(2*math.Pi*f*float64(i)/float64(sr)) * 32767 * 0.3 * env)
		}
	}
	return s
}

// ======================== ЧАСТИЦЫ ========================
type Particle struct {
	x, y, vx, vy   float64
	life, maxLife  int
	clr            color.Color
	sz             int
	rotation       float64
	rotSpeed       float64
}

func spawnParts(x, y float64, clr color.Color, n int, parts *[]Particle) {
	for i := 0; i < n; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 1 + rand.Float64()*3
		*parts = append(*parts, Particle{
			x: x, y: y,
			vx: math.Cos(angle) * speed,
			vy: math.Sin(angle) * speed - 2,
			life: 30 + rand.Intn(30),
			maxLife: 60,
			clr: clr,
			sz: 2 + rand.Intn(5),
			rotSpeed: (rand.Float64() - 0.5) * 0.2,
		})
	}
}

// ======================== АНИМАЦИИ ========================
type SelectAnim struct {
	r, c int
	t    int
}

func (s *SelectAnim) Update() { s.t++ }

func (s *SelectAnim) Draw(screen *ebiten.Image, ox, oy float64) {
	px := float64(s.c*TILE) + ox
	py := float64(s.r*TILE) + oy
	pulse := 0.5 + 0.5*math.Sin(float64(s.t)*0.1)

	// Пульсирующая рамка
	c := color.RGBA{255, 255, 255, uint8(150 + 105*pulse)}
	rect(screen, int(px)+2, int(py)+2, TILE-4, TILE-4, c)

	// Искры по углам
	if s.t%10 < 5 {
		cornerClr := color.RGBA{255, 255, 100, 200}
		for _, off := range [][2]int{{4, 4}, {TILE - 6, 4}, {4, TILE - 6}, {TILE - 6, TILE - 6}} {
			rect(screen, int(px)+off[0], int(py)+off[1], 4, 4, cornerClr)
		}
	}
}

type RemoveAnim struct {
	r, c   int
	x, y   float64
	t, total int
	typeId int
}

func (a *RemoveAnim) Update() bool {
	a.t++
	if a.t == 1 {
		spawnParts(a.x+TILE/2, a.y+TILE/2, gemColor(a.typeId), 12, nil)
	}
	return a.t >= a.total
}

func (a *RemoveAnim) Draw(screen *ebiten.Image) {
	progress := float64(a.t) / float64(a.total)
	scale := 1.0 + progress*0.5
	alpha := 1.0 - progress

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(a.x, a.y)
	op.GeoM.Translate(float64(TILE)/2, float64(TILE)/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(-float64(TILE)/2, -float64(TILE)/2)
	op.ColorScale.ScaleAlpha(float32(alpha))

	if gemSprites[a.typeId] != nil {
		screen.DrawImage(gemSprites[a.typeId], op)
	} else {
		c := gemColor(a.typeId)
		op.ColorScale.SetR(float32(c.(color.RGBA).R) / 255)
		op.ColorScale.SetG(float32(c.(color.RGBA).G) / 255)
		op.ColorScale.SetB(float32(c.(color.RGBA).B) / 255)
		screen.DrawImage(mkImg(TILE, TILE, c), op)
	}

	// Вспышка
	if progress < 0.3 {
		flashAlpha := 1.0 - progress/0.3
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(a.x, a.y)
		op2.ColorScale.ScaleAlpha(float32(flashAlpha * 0.5))
		screen.DrawImage(mkImg(TILE, TILE, color.White), op2)
	}
}

// SlideAnim — АНИМАЦИЯ ПЕРЕМЕЩЕНИЯ (теперь используется!)
type SlideAnim struct {
	r, c   int
	sx, sy float64 // start
	tx, ty float64 // target
	t      int
	total  int
	spr    *ebiten.Image
	typeId int
}

func (a *SlideAnim) Update() bool {
	a.t++
	return a.t >= a.total
}

func (a *SlideAnim) Draw(screen *ebiten.Image) {
	progress := float64(a.t) / float64(a.total)
	// Ease out cubic
	eased := 1.0 - math.Pow(1.0-progress, 3)
	x := a.sx + (a.tx-a.sx)*eased
	y := a.sy + (a.ty-a.sy)*eased

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)

	if a.spr != nil {
		screen.DrawImage(a.spr, op)
	} else if gemSprites[a.typeId] != nil {
		screen.DrawImage(gemSprites[a.typeId], op)
	}
}

// ======================== УТИЛИТЫ ========================
func gemColor(typeId int) color.Color {
	clrs := []color.Color{
		color.RGBA{200, 220, 255, 255}, // blue
		color.RGBA{255, 80, 80, 255},   // red
		color.RGBA{80, 220, 80, 255},   // green
		color.RGBA{255, 220, 80, 255},  // yellow
		color.RGBA{180, 80, 255, 255},  // violet
		color.RGBA{255, 160, 40, 255},  // orange
	}
	if typeId >= 0 && typeId < len(clrs) {
		return clrs[typeId]
	}
	return color.White
}

func rect(s *ebiten.Image, x, y, w, h int, c color.Color) {
	ensureWhitePixel()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.GeoM.Scale(float64(w), float64(h))
	if col, ok := c.(color.RGBA); ok {
		op.ColorScale.SetR(float32(col.R) / 255)
		op.ColorScale.SetG(float32(col.G) / 255)
		op.ColorScale.SetB(float32(col.B) / 255)
		op.ColorScale.SetA(float32(col.A) / 255)
	}
	s.DrawImage(whitePixel, op)
}

func rectAlpha(s *ebiten.Image, x, y, w, h int, c color.Color, a float32) {
	ensureWhitePixel()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.GeoM.Scale(float64(w), float64(h))
	if col, ok := c.(color.RGBA); ok {
		op.ColorScale.SetR(float32(col.R) / 255)
		op.ColorScale.SetG(float32(col.G) / 255)
		op.ColorScale.SetB(float32(col.B) / 255)
		op.ColorScale.SetA(float32(col.A) / 255 * a)
	}
	s.DrawImage(whitePixel, op)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
		op.GeoM.Translate(float64(b.x), float64(b.y))
		if b.hover {
			op.ColorScale.SetR(1.2)
			op.ColorScale.SetG(1.2)
			op.ColorScale.SetB(1.2)
		}
		s.DrawImage(b.spr, op)
	} else {
		c := color.RGBA{60, 60, 120, 200}
		if b.hover {
			c = color.RGBA{80, 80, 160, 230}
		}
		rect(s, b.x, b.y, b.w, b.h, c)
		bw := text.BoundString(basicfont.Face7x13, b.label)
		text.Draw(s, b.label, basicfont.Face7x13, b.x+b.w/2-bw.Dx()/2, b.y+b.h/2+5, color.White)
	}
}

// ======================== GAME STRUCT ========================
type Game struct {
	board   [ROWS][COLS]int
	score   int
	combo   int
	moves   int
	state   State
	selR    int
	selC    int
	hovR    int
	hovC    int
	particles []Particle
	selectAnim   *SelectAnim
	removeAnims  []RemoveAnim
	slideAnims   []SlideAnim
	busy        bool
	busyTimer   int
	flash       int
	msg         string
	msgTimer    int
	menuButtons  []*MenuButton
	menuAnim     int
	highScore    int
	cascadePending bool
}

func (g *Game) loadMenuSprites() {
	g.menuButtons = nil
	sprPlay, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/play button.png")
	sprOpt, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/Options Button.png")
	sprExit, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/Exit Button.png")
	g.menuButtons = append(g.menuButtons, &MenuButton{x: WIN_W/2 - 60, y: 300, w: 120, h: 40, label: "PLAY", spr: sprPlay})
	g.menuButtons = append(g.menuButtons, &MenuButton{x: WIN_W/2 - 60, y: 360, w: 120, h: 40, label: "OPTIONS", spr: sprOpt})
	g.menuButtons = append(g.menuButtons, &MenuButton{x: WIN_W/2 - 60, y: 420, w: 120, h: 40, label: "EXIT", spr: sprExit})
}

func (g *Game) initPauseButtons() {
	g.menuButtons = nil
	sprResume, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/play button.png")
	sprBack, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/Back Button.png")
	g.menuButtons = append(g.menuButtons, &MenuButton{x: WIN_W/2 - 60, y: 280, w: 120, h: 40, label: "RESUME", spr: sprResume})
	g.menuButtons = append(g.menuButtons, &MenuButton{x: WIN_W/2 - 60, y: 340, w: 120, h: 40, label: "RESTART", spr: sprBack})
}

func (g *Game) initOptionsButtons() {
	g.menuButtons = nil
	sprBack, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/Back Button.png")
	g.menuButtons = append(g.menuButtons, &MenuButton{x: WIN_W/2 - 60, y: 400, w: 120, h: 40, label: "BACK", spr: sprBack})
}

func (g *Game) loadHighScore() {
	data, err := os.ReadFile("highscore.txt")
	if err == nil { fmt.Sscanf(string(data), "%d", &g.highScore) }
}
func (g *Game) saveHighScore() {
	if g.score > g.highScore {
		g.highScore = g.score
		os.WriteFile("highscore.txt", []byte(fmt.Sprintf("%d", g.highScore)), 0644)
	}
}

func NewGame() *Game {
	g := &Game{selR: -1, selC: -1, hovR: -1, hovC: -1}
	g.loadHighScore()
	g.loadMenuSprites()
	initAudio()
	loadSprites()
	return g
}

func (g *Game) start() {
	g.score, g.combo, g.moves = 0, 0, 0
	g.state = S_PLAY
	g.selR, g.selC = -1, -1
	g.busy, g.busyTimer, g.cascadePending = false, 0, false
	g.slideAnims, g.removeAnims, g.particles = nil, nil, nil
	g.selectAnim, g.msg = nil, ""
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			for {
				g.board[r][c] = rand.Intn(GEM_TYPES)
				if !wouldMatchAt(g.board, r, c, g.board[r][c]) { break }
			}
		}
	}
}

func wouldMatchAt(board [ROWS][COLS]int, r, c, t int) bool {
	cnt := 1
	for i := c - 1; i >= 0 && board[r][i] == t; i-- { cnt++ }
	for i := c + 1; i < COLS && board[r][i] == t; i++ { cnt++ }
	if cnt >= MATCH_MIN { return true }
	cnt = 1
	for i := r - 1; i >= 0 && board[i][c] == t; i-- { cnt++ }
	for i := r + 1; i < ROWS && board[i][c] == t; i++ { cnt++ }
	return cnt >= MATCH_MIN
}

func (g *Game) click(r, c int) {
	if g.busy || g.state != S_PLAY { return }
	if g.selR < 0 {
		g.selR, g.selC = r, c
		g.selectAnim = &SelectAnim{r: r, c: c, t: 0}
		return
	}
	if g.selR == r && g.selC == c {
		g.selR, g.selC = -1, -1; g.selectAnim = nil; return
	}
	if abs(g.selR-r)+abs(g.selC-c) == 1 {
		g.moves++; g.trySwap(g.selR, g.selC, r, c)
		g.selR, g.selC = -1, -1; g.selectAnim = nil
	} else {
		g.selR, g.selC = r, c; g.selectAnim = &SelectAnim{r: r, c: c, t: 0}
	}
}

func (g *Game) trySwap(r1, c1, r2, c2 int) {
	g.board[r1][c1], g.board[r2][c2] = g.board[r2][c2], g.board[r1][c1]
	sx1, sy1 := float64(c1*TILE)+BOARD_OFFX, float64(r1*TILE)+BOARD_OFFY
	sx2, sy2 := float64(c2*TILE)+BOARD_OFFX, float64(r2*TILE)+BOARD_OFFY
	g.slideAnims = append(g.slideAnims, SlideAnim{r: r1, c: c1, sx: sx1, sy: sy1, tx: sx2, ty: sy2, t: 0, total: 10, typeId: g.board[r1][c1], spr: gemSprites[g.board[r1][c1]]})
	g.slideAnims = append(g.slideAnims, SlideAnim{r: r2, c: c2, sx: sx2, sy: sy2, tx: sx1, ty: sy1, t: 0, total: 10, typeId: g.board[r2][c2], spr: gemSprites[g.board[r2][c2]]})
	matches := g.findMatches()
	if len(matches) > 0 { play(sndSwap); g.combo = 1; g.busy = true; g.removeMatches(matches) } else {
		g.board[r1][c1], g.board[r2][c2] = g.board[r2][c2], g.board[r1][c1]; play(sndBad); g.flash = 10
	}
}

func (g *Game) findMatches() map[[2]int]bool {
	matched := make(map[[2]int]bool)
	for r := 0; r < ROWS; r++ {
		for c := 0; c <= COLS-MATCH_MIN; c++ {
			t := g.board[r][c]; if t < 0 { continue }; match := true
			for i := 1; i < MATCH_MIN; i++ { if g.board[r][c+i] != t { match = false; break } }
			if match { for i := 0; i < MATCH_MIN; i++ { matched[[2]int{r, c+i}] = true } }
		}
	}
	for c := 0; c < COLS; c++ {
		for r := 0; r <= ROWS-MATCH_MIN; r++ {
			t := g.board[r][c]; if t < 0 { continue }; match := true
			for i := 1; i < MATCH_MIN; i++ { if g.board[r+i][c] != t { match = false; break } }
			if match { for i := 0; i < MATCH_MIN; i++ { matched[[2]int{r+i, c}] = true } }
		}
	}
	return matched
}

func (g *Game) removeMatches(matches map[[2]int]bool) {
	pts := 0
	for pos := range matches {
		r, c := pos[0], pos[1]
		px, py := float64(c*TILE)+BOARD_OFFX+TILE/2, float64(r*TILE)+BOARD_OFFY+TILE/2
		spawnParts(px, py, gemColor(g.board[r][c]), 8, &g.particles)
		g.removeAnims = append(g.removeAnims, RemoveAnim{r: r, c: c, x: float64(c*TILE)+BOARD_OFFX, y: float64(r*TILE)+BOARD_OFFY, t: 0, total: 20, typeId: g.board[r][c]})
		pts += 10; g.board[r][c] = -1
	}
	g.score += pts * g.combo
	if g.combo > 1 { play(sndCombo); g.msg = fmt.Sprintf("COMBO x%d!", g.combo); g.msgTimer = 60 } else { play(sndMatch) }
	g.busyTimer = 25; g.cascadePending = true
}

func (g *Game) applyGravity() {
	for c := 0; c < COLS; c++ {
		wr := ROWS - 1
		for r := ROWS - 1; r >= 0; r-- {
			if g.board[r][c] >= 0 {
				if r != wr {
					sy, ty := float64(r*TILE)+BOARD_OFFY, float64(wr*TILE)+BOARD_OFFY
					g.board[wr][c] = g.board[r][c]; g.board[r][c] = -1
					g.slideAnims = append(g.slideAnims, SlideAnim{r: wr, c: c, sx: float64(c*TILE)+BOARD_OFFX, sy: sy, tx: float64(c*TILE)+BOARD_OFFX, ty: ty, t: 0, total: 12, typeId: g.board[wr][c], spr: gemSprites[g.board[wr][c]]})
				}
				wr--
			}
		}
	}
}

func (g *Game) fillEmpty() {
	for c := 0; c < COLS; c++ {
		for r := 0; r < ROWS; r++ {
			if g.board[r][c] < 0 {
				g.board[r][c] = rand.Intn(GEM_TYPES)
				sy := -float64((ROWS - r) * TILE)
				ty := float64(r*TILE) + BOARD_OFFY
				g.slideAnims = append(g.slideAnims, SlideAnim{r: r, c: c, sx: float64(c*TILE)+BOARD_OFFX, sy: sy, tx: float64(c*TILE)+BOARD_OFFX, ty: ty, t: 0, total: 15, typeId: g.board[r][c], spr: gemSprites[g.board[r][c]]})
			}
		}
	}
}

func (g *Game) hasValidMoves() bool {
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			if c+1 < COLS {
				g.board[r][c], g.board[r][c+1] = g.board[r][c+1], g.board[r][c]
				if len(g.findMatches()) > 0 { g.board[r][c], g.board[r][c+1] = g.board[r][c+1], g.board[r][c]; return true }
				g.board[r][c], g.board[r][c+1] = g.board[r][c+1], g.board[r][c]
			}
			if r+1 < ROWS {
				g.board[r][c], g.board[r+1][c] = g.board[r+1][c], g.board[r][c]
				if len(g.findMatches()) > 0 { g.board[r][c], g.board[r+1][c] = g.board[r+1][c], g.board[r][c]; return true }
				g.board[r][c], g.board[r+1][c] = g.board[r+1][c], g.board[r][c]
			}
		}
	}
	return false
}

func (g *Game) shuffleBoard() {
	flat := make([]int, 0, ROWS*COLS)
	for r := 0; r < ROWS; r++ { for c := 0; c < COLS; c++ { flat = append(flat, g.board[r][c]) } }
	rand.Shuffle(len(flat), func(i, j int) { flat[i], flat[j] = flat[j], flat[i] })
	idx := 0; for r := 0; r < ROWS; r++ { for c := 0; c < COLS; c++ { g.board[r][c] = flat[idx]; idx++ } }
	m := g.findMatches()
	for len(m) > 0 { for pos := range m { g.board[pos[0]][pos[1]] = rand.Intn(GEM_TYPES) }; m = g.findMatches() }
	if !g.hasValidMoves() { g.shuffleBoard() }
}

// ======================== UPDATE ========================
func (g *Game) Update() error {
	g.menuAnim++

	// Update particles
	active := make([]Particle, 0)
	for _, p := range g.particles {
		p.x += p.vx; p.y += p.vy; p.vy += 0.1; p.life--; p.rotation += p.rotSpeed
		if p.life > 0 { active = append(active, p) }
	}
	g.particles = active

	// Update remove animations
	ra := make([]RemoveAnim, 0)
	for _, a := range g.removeAnims { if !a.Update() { ra = append(ra, a) } }
	g.removeAnims = ra

	// Update slide animations
	sa := make([]SlideAnim, 0)
	for _, a := range g.slideAnims { if !a.Update() { sa = append(sa, a) } }
	g.slideAnims = sa

	// Update select animation
	if g.selectAnim != nil { g.selectAnim.Update() }

	// Message timer
	if g.msgTimer > 0 { g.msgTimer-- }
	if g.flash > 0 { g.flash-- }

	// Cascade check — ЗАМЕНА goroutine!
	if g.cascadePending && !g.busy {
		g.busyTimer--
		if g.busyTimer <= 0 {
			g.cascadePending = false
			g.applyGravity()
			g.fillEmpty()
			g.busy = len(g.slideAnims) > 0
			if !g.busy {
				matches := g.findMatches()
				if len(matches) > 0 {
					g.combo++; g.busy = true; g.removeMatches(matches)
				} else {
					g.combo = 0
					// Check valid moves
					if !g.hasValidMoves() {
						g.shuffleBoard()
						g.msg = "SHUFFLED!"
						g.msgTimer = 90
					}
					// Check win
					if g.score >= TARGET_SCORE {
						g.state = S_WIN; g.saveHighScore(); play(sndWin)
					}
				}
			}
		}
	}

	// If slide anims running, wait
	if len(g.slideAnims) > 0 && g.cascadePending { return nil }

	switch g.state {
	case S_MENU:
		mx, my := ebiten.CursorPosition()
		for _, b := range g.menuButtons { b.hover = b.contains(mx, my) }
		if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) { /* nav */ }
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			for _, b := range g.menuButtons {
				if b.hover {
					if b.label == "PLAY" { g.start() }
					if b.label == "OPTIONS" { g.state = S_OPTIONS; g.initOptionsButtons() }
					if b.label == "EXIT" { os.Exit(0) }
				}
			}
		}
	case S_PLAY:
		mx, my := ebiten.CursorPosition()
		g.hovR, g.hovC = px2rc(mx, my)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.hovR >= 0 { g.click(g.hovR, g.hovC) }
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = S_PAUSE; g.initPauseButtons()
		}
	case S_PAUSE:
		mx, my := ebiten.CursorPosition()
		for _, b := range g.menuButtons { b.hover = b.contains(mx, my) }
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			for _, b := range g.menuButtons {
				if b.hover {
					if b.label == "RESUME" { g.state = S_PLAY }
					if b.label == "RESTART" { g.start() }
				}
			}
		}
	case S_OPTIONS:
		mx, my := ebiten.CursorPosition()
		for _, b := range g.menuButtons { b.hover = b.contains(mx, my) }
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			for _, b := range g.menuButtons {
				if b.hover { if b.label == "BACK" { g.state = S_MENU } }
			}
		}
	case S_WIN:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) { g.state = S_MENU }
	}
	return nil
}

func px2rc(px, py int) (int, int) {
	r, c := (py-BOARD_OFFY)/TILE, (px-BOARD_OFFX)/TILE
	if r < 0 || r >= ROWS || c < 0 || c >= COLS { return -1, -1 }
	return r, c
}

// ======================== DRAW ========================
func (g *Game) Draw(s *ebiten.Image) {
	s.Fill(color.RGBA{10, 10, 30, 255})
	switch g.state {
	case S_MENU: g.drawMenu(s)
	case S_PLAY: g.drawPlay(s)
	case S_PAUSE: g.drawPlay(s); g.drawPause(s)
	case S_OPTIONS: g.drawOptions(s)
	case S_WIN: g.drawWin(s)
	}
}

func (g *Game) drawMenu(s *ebiten.Image) {
	// Background tiles
	for y := 0; y < WIN_H/TILE+1; y++ {
		for x := 0; x < WIN_W/TILE+1; x++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*TILE), float64(y*TILE))
			s.DrawImage(bgTileSpr, op)
		}
	}
	rectAlpha(s, 0, 0, WIN_W, WIN_H, color.RGBA{0, 0, 0, 0}, 0.6)

	// Title
	title := "PUZZLE GO"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, 120, color.RGBA{255, 220, 100, 255})
	sub := "Match-3 Gem Crusher"
	bw2 := text.BoundString(basicfont.Face7x13, sub)
	text.Draw(s, sub, basicfont.Face7x13, WIN_W/2-bw2.Dx()/2, 145, color.RGBA{200, 200, 200, 255})

	// Buttons
	for _, b := range g.menuButtons { b.Draw(s) }

	// High Score
	hs := fmt.Sprintf("Best: %d", g.highScore)
	bw3 := text.BoundString(basicfont.Face7x13, hs)
	text.Draw(s, hs, basicfont.Face7x13, WIN_W/2-bw3.Dx()/2, 520, color.RGBA{255, 215, 0, 255})
}

func (g *Game) drawPlay(s *ebiten.Image) {
	// Background
	for y := 0; y < ROWS; y++ {
		for x := 0; x < COLS; x++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(BOARD_OFFX+x*TILE), float64(BOARD_OFFY+y*TILE))
			s.DrawImage(bgTileSpr, op)
		}
	}

	// Board gems (skip cells being animated)
	animCells := make(map[[2]int]bool)
	for _, a := range g.slideAnims { animCells[[2]int{a.r, a.c}] = true }

	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			if g.board[r][c] < 0 || animCells[[2]int{r, c}] { continue }
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(BOARD_OFFX+c*TILE), float64(BOARD_OFFY+r*TILE))
			if gemSprites[g.board[r][c]] != nil {
				s.DrawImage(gemSprites[g.board[r][c]], op)
			}
		}
	}

	// Slide animations
	for _, a := range g.slideAnims { a.Draw(s) }

	// Remove animations
	for _, a := range g.removeAnims { a.Draw(s) }

	// Hover highlight
	if g.hovR >= 0 && !g.busy {
		rectAlpha(s, BOARD_OFFX+g.hovC*TILE, BOARD_OFFY+g.hovR*TILE, TILE, TILE, color.White, 0.2)
	}

	// Selection
	if g.selR >= 0 && g.selectAnim != nil {
		g.selectAnim.Draw(s, BOARD_OFFX, BOARD_OFFY)
	}

	// Flash effect
	if g.flash > 0 { rectAlpha(s, 0, 0, WIN_W, WIN_H, color.White, float32(g.flash)/20.0) }

	// Particles
	for _, p := range g.particles {
		alpha := float32(p.life) / float32(p.maxLife)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x, p.y); op.GeoM.Rotate(p.rotation)
		op.GeoM.Translate(-float64(p.sz)/2, -float64(p.sz)/2)
		op.GeoM.Scale(float64(p.sz), float64(p.sz))
		if col, ok := p.clr.(color.RGBA); ok {
			op.ColorScale.SetR(float32(col.R)/255); op.ColorScale.SetG(float32(col.G)/255)
			op.ColorScale.SetB(float32(col.B)/255); op.ColorScale.SetA(alpha * float32(col.A)/255)
		}
		s.DrawImage(whitePixel, op)
	}

	// HUD
	if hudBgSpr != nil {
		op := &ebiten.DrawImageOptions{}; op.GeoM.Translate(0, 0); s.DrawImage(hudBgSpr, op)
	}
	scoreTxt := fmt.Sprintf("Score: %d", g.score)
	text.Draw(s, scoreTxt, basicfont.Face7x13, 15, 18, color.RGBA{255, 255, 255, 255})
	movesTxt := fmt.Sprintf("Moves: %d", g.moves)
	text.Draw(s, movesTxt, basicfont.Face7x13, 200, 18, color.RGBA{200, 200, 200, 255})
	targetTxt := fmt.Sprintf("Target: %d", TARGET_SCORE)
	text.Draw(s, targetTxt, basicfont.Face7x13, 350, 18, color.RGBA{150, 200, 255, 255})

	// Progress bar
	progress := float64(g.score) / float64(TARGET_SCORE)
	if progress > 1 { progress = 1 }
	rect(s, 15, 30, 400, 8, color.RGBA{50, 50, 50, 255})
	rect(s, 15, 30, int(400*progress), 8, color.RGBA{0, 200, 100, 255})

	// Combo message
	if g.msgTimer > 0 {
		alpha := float32(g.msgTimer) / 60.0
		bw := text.BoundString(basicfont.Face7x13, g.msg)
		rectAlpha(s, WIN_W/2-bw.Dx()/2-10, WIN_H/2-60, bw.Dx()+20, 30, color.RGBA{0, 0, 0, 0}, alpha*0.7)
		text.Draw(s, g.msg, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-42, color.RGBA{255, 255, 100, 255})
	}
}

func (g *Game) drawPause(s *ebiten.Image) {
	rectAlpha(s, 0, 0, WIN_W, WIN_H, color.RGBA{0, 0, 0, 0}, 0.7)
	for _, b := range g.menuButtons { b.Draw(s) }
}

func (g *Game) drawOptions(s *ebiten.Image) {
	s.Fill(color.RGBA{20, 20, 40, 255})
	title := "CONTROLS"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, 100, color.White)
	text.Draw(s, "Click: Select/Swap gems", basicfont.Face7x13, WIN_W/2-90, 160, color.RGBA{200, 200, 200, 255})
	text.Draw(s, "ESC/P: Pause", basicfont.Face7x13, WIN_W/2-55, 190, color.RGBA{200, 200, 200, 255})
	text.Draw(s, "Match 3+ gems to score!", basicfont.Face7x13, WIN_W/2-85, 250, color.RGBA{255, 255, 100, 255})
	text.Draw(s, fmt.Sprintf("Target: %d points to win", TARGET_SCORE), basicfont.Face7x13, WIN_W/2-100, 280, color.RGBA{100, 200, 255, 255})
	for _, b := range g.menuButtons { b.Draw(s) }
}

func (g *Game) drawWin(s *ebiten.Image) {
	g.drawPlay(s)
	rectAlpha(s, 0, 0, WIN_W, WIN_H, color.RGBA{0, 0, 0, 0}, 0.8)
	title := "YOU WIN!"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-40, color.RGBA{255, 215, 0, 255})
	sc := fmt.Sprintf("Score: %d  Moves: %d", g.score, g.moves)
	bw2 := text.BoundString(basicfont.Face7x13, sc)
	text.Draw(s, sc, basicfont.Face7x13, WIN_W/2-bw2.Dx()/2, WIN_H/2+10, color.White)
	text.Draw(s, "Press ENTER for menu", basicfont.Face7x13, WIN_W/2-70, WIN_H/2+50, color.RGBA{200, 200, 200, 255})
}

func (g *Game) Layout(w, h int) (int, int) { return WIN_W, WIN_H }

// ======================== MAIN ========================
func main() {
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(WIN_W, WIN_H)
	ebiten.SetWindowTitle("Puzzle GO - Match-3")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewGame()); err != nil { log.Fatal(err) }
}
