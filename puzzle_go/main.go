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
	BOARD_OFFX = 32
	BOARD_OFFY = 80
	MATCH_MIN  = 3
)

const (
	WIN_W = COLS*TILE + BOARD_OFFX*2
	WIN_H = ROWS*TILE + BOARD_OFFY + HUD
)

// ======================== СОСТОЯНИЯ ========================
type State int

const (
	S_MENU State = iota
	S_PLAY
	S_PAUSE
	S_OPTIONS
	S_WIN
)

// ======================== КЭШ СПРАЙТОВ ========================
var (
	gemSprites [GEM_TYPES]*ebiten.Image
	selSprite  *ebiten.Image
	starSprite *ebiten.Image
	coinSprite *ebiten.Image
	bgTile     *ebiten.Image
	groundTile *ebiten.Image
	hudBg      *ebiten.Image
)

func loadSprites() {
	gems := []string{"gem0.png", "gem1.png", "gem2.png", "gem3.png", "gem4.png", "gem5.png"}
	for i, name := range gems {
		img, _, err := ebitenutil.NewImageFromFile("assets/sprites/" + name)
		if err == nil {
			// Масштабируем до TILE
			gemSprites[i] = scaleImg(img, TILE, TILE)
		}
	}
	selSprite, _, _ = ebitenutil.NewImageFromFile("assets/sprites/selector.png")
	if selSprite != nil {
		selSprite = scaleImg(selSprite, TILE+4, TILE+4)
	}
	starSprite, _, _ = ebitenutil.NewImageFromFile("assets/sprites/star.png")
	coinSprite, _, _ = ebitenutil.NewImageFromFile("assets/sprites/coin.png")
	bgTile, _, _ = ebitenutil.NewImageFromFile("assets/sprites/bg_tile.png")
	groundTile, _, _ = ebitenutil.NewImageFromFile("assets/sprites/ground.png")
	hudBg = mkImg(WIN_W, HUD, color.RGBA{10, 10, 25, 255})
}

func scaleImg(src *ebiten.Image, w, h int) *ebiten.Image {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	dst := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(b.Dx()), float64(h)/float64(b.Dy()))
	dst.DrawImage(src, op)
	return dst
}

func mkImg(w, h int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	return img
}

// ======================== ЗВУКИ ========================
var (
	actx      *audio.Context
	sndMatch  *audio.Player
	sndSwap   *audio.Player
	sndBad    *audio.Player
	sndCombo  *audio.Player
	sndWin    *audio.Player
)

func initAudio() {
	actx = audio.NewContext(44100)
	sndMatch, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(tone(0.12, 800, 44100))))
	sndSwap, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(tone(0.06, 500, 44100))))
	sndBad, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(noise(0.1, 44100))))
	sndCombo, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(arp(0.3, []float64{600, 800, 1000}, 44100))))
	sndWin, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(arp(0.5, []float64{523, 659, 784, 1047}, 44100))))
}

func play(p *audio.Player) {
	if p == nil {
		return
	}
	p.Rewind()
	p.Play()
}

func playSoundP(p *audio.Player) {
	play(p)
}

func mkWAV(s []float64) []byte {
	buf := new(bytes.Buffer)
	n := uint32(len(s) * 2)
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+n))
	buf.WriteString("WAVEfmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint32(44100))
	binary.Write(buf, binary.LittleEndian, uint32(88200))
	binary.Write(buf, binary.LittleEndian, uint16(2))
	binary.Write(buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, n)
	for _, v := range s {
		if v > 1 {
			v = 1
		}
		if v < -1 {
			v = -1
		}
		binary.Write(buf, binary.LittleEndian, int16(v*32767))
	}
	return buf.Bytes()
}

func tone(dur, freq float64, sr int) []float64 {
	n := int(float64(sr) * dur)
	o := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		o[i] = math.Sin(6.2832*freq*t) * math.Exp(-t*8) * 0.4
	}
	return o
}

func noise(dur float64, sr int) []float64 {
	n := int(float64(sr) * dur)
	o := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		o[i] = (rand.Float64()*2 - 1) * math.Exp(-t*10) * 0.4
	}
	return o
}

func arp(dur float64, fs []float64, sr int) []float64 {
	n := int(float64(sr) * dur)
	o := make([]float64, n)
	st := dur / float64(len(fs))
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		idx := int(t / st)
		if idx >= len(fs) {
			idx = len(fs) - 1
		}
		o[i] = math.Sin(6.2832*fs[idx]*t) * math.Exp(-t*5) * 0.3
	}
	return o
}

// ======================== ЧАСТИЦЫ ========================
type Particle struct {
	x, y     float64
	vx, vy   float64
	life     int
	maxLife  int
	clr      color.Color
	sz       int
	rotation float64
	rotSpeed float64
}

func spawnParts(x, y float64, clr color.Color, n int) []Particle {
	ps := make([]Particle, 0, n)
	for i := 0; i < n; i++ {
		a := rand.Float64() * 6.2832
		sp := 2 + rand.Float64()*4
		ps = append(ps, Particle{
			x: x, y: y,
			vx: math.Cos(a) * sp,
			vy: math.Sin(a)*sp - 3,
			life:    25 + rand.Intn(20),
			maxLife: 45,
			clr:     clr,
			sz:      3 + rand.Intn(6),
			rotation: rand.Float64() * 6.2832,
			rotSpeed: (rand.Float64() - 0.5) * 0.3,
		})
	}
	return ps
}

// ======================== АНИМАЦИЯ ВЫДЕЛЕНИЯ ========================
type SelectAnim struct {
	r, c int    // позиция гема
	t    int    // кадр
}

func (s *SelectAnim) Update() {
	s.t++
}

func (s *SelectAnim) Draw(screen *ebiten.Image, x, y int) {
	pulse := 0.5 + 0.5*math.Sin(float64(s.t)/8)

	// Кольцо свечения
	ringR := TILE/2 + 4
	ringW := 3 + int(pulse*3)
	for i := 0; i < 32; i++ {
		angle := float64(i) * math.Pi * 2 / 32
		px := float64(x+TILE/2) + math.Cos(angle)*float64(ringR)
		py := float64(y+TILE/2) + math.Sin(angle)*float64(ringR)
		a := 0.4 + pulse*0.4
		rectAlphaC(screen, int(px)-ringW/2, int(py)-ringW/2, ringW, ringW,
			color.RGBA{100, 255, 200, 255}, a)
	}

	// Искорки по кругу
	for i := 0; i < 4; i++ {
		angle := float64(i)*math.Pi/2 + float64(s.t)/20
		px := float64(x+TILE/2) + math.Cos(angle)*float64(ringR+6)
		py := float64(y+TILE/2) + math.Sin(angle)*float64(ringR+6)
		sparkSz := 3 + int(pulse*3)
		rectAlphaC(screen, int(px)-sparkSz/2, int(py)-sparkSz/2, sparkSz, sparkSz,
			color.RGBA{255, 255, 255, 255}, 0.5+pulse*0.5)
	}
}

// ======================== АНИМАЦИЯ УДАЛЕНИЯ ========================
type RemoveAnim struct {
	r, c  int
	x, y  float64
	t     int
	total int // 25 кадров
	typeId int
}

func (a *RemoveAnim) Update() bool {
	a.t++
	// Частицы на каждом кадре
	if a.t%2 == 0 && a.t < 15 {
		clr := gemColor(a.typeId)
		x := a.x + float64(TILE/2)
		y := a.y + float64(TILE/2)
		ps := spawnParts(x, y, clr, 3)
		for i := range ps {
			ps[i].vx *= 1.5
			ps[i].vy *= 1.5
		}
	}
	return a.t >= a.total
}

func (a *RemoveAnim) Draw(screen *ebiten.Image) {
	progress := float64(a.t) / float64(a.total)

	// Фаза 1: увеличение (0-0.3)
	// Фаза 2: исчезновение (0.3-1.0)
	var scale, alpha float64
	if progress < 0.3 {
		t := progress / 0.3
		scale = 1.0 + t*0.5
		alpha = 1.0
	} else {
		t := (progress - 0.3) / 0.7
		scale = 1.5 - t*1.5
		alpha = 1.0 - t
	}

	if scale <= 0 || alpha <= 0 {
		return
	}

	spr := gemSprites[a.typeId-1]
	if spr != nil {
		b := spr.Bounds()
		ox := a.x + float64(TILE/2) - float64(b.Dx())*scale/2
		oy := a.y + float64(TILE/2) - float64(b.Dy())*scale/2
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(ox, oy)
		op.ColorM.Scale(1, 1, 1, alpha)
		screen.DrawImage(spr, op)
	}

	// Вспышка при начале удаления
	if a.t < 5 {
		flashA := (1.0 - float64(a.t)/5.0) * 0.6
		rectAlphaC(screen, int(a.x)-4, int(a.y)-4, TILE+8, TILE+8,
			color.RGBA{255, 255, 255, 255}, flashA)
	}
}

func gemColor(typeId int) color.Color {
	clrs := []color.Color{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 80, 80, 255},   // red
		color.RGBA{80, 255, 80, 255},   // green
		color.RGBA{80, 80, 255, 255},   // blue
		color.RGBA{255, 255, 80, 255},  // yellow
		color.RGBA{255, 80, 255, 255},  // purple
		color.RGBA{80, 255, 255, 255},  // cyan
	}
	if typeId >= 1 && typeId <= 6 {
		return clrs[typeId]
	}
	return color.RGBA{255, 255, 255, 255}
}

// ======================== АНИМАЦИЯ ========================
type SlideAnim struct {
	r, c  int    // целевая позиция
	sx, sy float64 // текущая позиция в пикселях
	tx, ty float64 // целевая позиция
	t    int
	spr  *ebiten.Image
	typeId int
}

func (a *SlideAnim) Update() bool {
	a.t++
	f := float64(a.t) / 10.0
	if f > 1 {
		f = 1
	}
	a.sx = a.sx + (a.tx-a.sx)*f*0.3
	a.sy = a.sy + (a.ty-a.sy)*f*0.3
	return a.t >= 10
}

// ======================== ИГРА ========================
var fCount int64

type Game struct {
	board   [ROWS][COLS]int // 0 = пусто, 1-6 = тип гема
	score   int
	combo   int
	moves   int
	state   State

	selR, selC int
	hovR, hovC int

	particles []Particle
	anims     []SlideAnim
	selectAnim *SelectAnim
	removeAnims []RemoveAnim

	busy bool
	flash int

	msg string

	// Menu
	menuSprites map[string]*ebiten.Image
	buttons     []*MenuButton
	menuAnim    int
	highScore   int
	enterPrev   bool
}

func NewGame() *Game {
	g := &Game{
		state:       S_MENU,
		selR:        -1, selC: -1,
		hovR: -1, hovC: -1,
		menuSprites: make(map[string]*ebiten.Image),
	}
	g.loadMenuSprites()
	initAudio()
	loadSprites()
	return g
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
		rectCachedP(s, b.x, b.y, b.w, b.h, bg)
		bw := text.BoundString(basicfont.Face7x13, b.label)
		text.Draw(s, b.label, basicfont.Face7x13, b.x+b.w/2-bw.Dx()/2, b.y+b.h/2+5, color.RGBA{255, 255, 255, 255})
	}
}

func (g *Game) loadMenuSprites() {
	tryLoadP := func(name, file string) {
		img, _, err := ebitenutil.NewImageFromFile("assets/sprites/menu/" + file)
		if err == nil {
			g.menuSprites[name] = img
		}
	}

	tryLoadP("play", "play button.png")
	tryLoadP("options", "Options Button.png")
	tryLoadP("exit", "Exit Button.png")
	tryLoadP("back", "Back Button.png")
	tryLoadP("stars", "stars.png")
	tryLoadP("stars_bg", "stars back.png")

	if len(g.menuSprites) > 0 {
		fmt.Printf("✓ Loaded %d menu sprites\n", len(g.menuSprites))
	}
}

func (g *Game) initMenuButtons() {
	g.buttons = nil

	if spr, ok := g.menuSprites["play"]; ok {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 90, y: WIN_H/2 - 60, w: 180, h: 44,
			label: "PLAY", spr: spr,
		})
	} else {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 60, y: WIN_H/2 - 60, w: 120, h: 40,
			label: "▶ PLAY",
		})
	}
	g.buttons = append(g.buttons, &MenuButton{
		x: WIN_W/2 - 90, y: WIN_H/2, w: 180, h: 44,
		label: "OPTIONS", spr: g.menuSprites["options"],
	})
	g.buttons = append(g.buttons, &MenuButton{
		x: WIN_W/2 - 90, y: WIN_H/2 + 60, w: 180, h: 44,
		label: "EXIT", spr: g.menuSprites["exit"],
	})
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

func (g *Game) initOptionsButtons() {
	g.buttons = nil
	if spr, ok := g.menuSprites["back"]; ok {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 80, y: WIN_H/2 + 80, w: 160, h: 44,
			label: "▶ BACK", spr: spr,
		})
	} else {
		g.buttons = append(g.buttons, &MenuButton{
			x: WIN_W/2 - 60, y: WIN_H/2 + 80, w: 120, h: 40,
			label: "▶ BACK",
		})
	}
}

func rectCachedP(s *ebiten.Image, x, y, w, h int, c color.Color) {
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	key := fmt.Sprintf("%d_%d_%d_%d_%d_%d", rgba.R, rgba.G, rgba.B, rgba.A, w, h)
	if img, ok := cachedRectsP[key]; ok {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		s.DrawImage(img, op)
		return
	}
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	cachedRectsP[key] = img
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

var cachedRectsP = make(map[string]*ebiten.Image)

func (g *Game) start() {
	g.board = [ROWS][COLS]int{}
	g.score = 0
	g.combo = 0
	g.moves = 0
	g.state = S_PLAY
	g.selR, g.selC = -1, -1
	g.particles = nil
	g.anims = nil
	g.busy = false
	g.msg = "Moves: 0 | Combo: x0"
	g.initMenuButtons()

	// Заполняем без начальных совпадений
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			for {
				g.board[r][c] = 1 + rand.Intn(GEM_TYPES)
				if !wouldMatch(g.board, r, c) {
					break
				}
			}
		}
	}
}

func wouldMatch(b [ROWS][COLS]int, r, c int) bool {
	v := b[r][c]
	// Горизонталь
	if c >= 2 && b[r][c-1] == v && b[r][c-2] == v {
		return true
	}
	// Вертикаль
	if r >= 2 && b[r-1][c] == v && b[r-2][c] == v {
		return true
	}
	return false
}

func (g *Game) Update() error {
	fCount++

	// ===== MENU =====
	if g.state == S_MENU {
		g.menuAnim++
		mx, my := ebiten.CursorPosition()
		for _, btn := range g.buttons {
			btn.hover = btn.contains(mx, my)
			if btn.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if btn.label == "PLAY" {
					g.start()
					playSoundP(sndSwap)
				} else if btn.label == "OPTIONS" {
					g.state = S_OPTIONS
					g.initOptionsButtons()
					playSoundP(sndSwap)
				} else if btn.label == "EXIT" {
					os.Exit(0)
				}
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.start()
			playSoundP(sndSwap)
		}
		return nil
	}

	// ===== OPTIONS =====
	if g.state == S_OPTIONS {
		esc := ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyBackspace)
		if esc && !g.enterPrev {
			g.state = S_MENU
			g.initMenuButtons()
			playSoundP(sndSwap)
		}
		mx, my := ebiten.CursorPosition()
		for _, btn := range g.buttons {
			btn.hover = btn.contains(mx, my)
			if btn.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if btn.label == "▶ BACK" {
					g.state = S_MENU
					g.initMenuButtons()
					playSoundP(sndSwap)
				}
			}
		}
		g.enterPrev = esc
		return nil
	}

	// ===== PAUSE =====
	if g.state == S_PAUSE {
		esc := ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyP)
		if esc {
			g.state = S_PLAY
			playSoundP(sndSwap)
		}
		mx, my := ebiten.CursorPosition()
		for _, btn := range g.buttons {
			btn.hover = btn.contains(mx, my)
			if btn.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if btn.label == "▶ RESUME" {
					g.state = S_PLAY
					playSoundP(sndSwap)
				} else if btn.label == "▶ RESTART" {
					g.start()
					playSoundP(sndSwap)
				}
			}
		}
		return nil
	}

	// ===== WIN =====
	if g.state == S_WIN {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = S_MENU
			g.initMenuButtons()
		}
		return nil
	}

	// Обновляем частицы
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.15
		p.rotation += p.rotSpeed
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Обновляем анимации
	if g.busy {
		done := true
		for i := len(g.anims) - 1; i >= 0; i-- {
			if g.anims[i].Update() {
				g.anims = append(g.anims[:i], g.anims[i+1:]...)
			} else {
				done = false
			}
		}
		if done {
			// Анимация завершена — проверяем совпадения
			matches := g.findMatches()
			if len(matches) > 0 {
				g.removeMatches(matches)
			} else {
				g.busy = false
				g.combo = 0
			}
		}
		return nil
	}

	// Обновляем selectAnim
	if g.selectAnim != nil {
		g.selectAnim.Update()
	}

	// Обновляем removeAnims
	for i := len(g.removeAnims) - 1; i >= 0; i-- {
		if g.removeAnims[i].Update() {
			g.removeAnims = append(g.removeAnims[:i], g.removeAnims[i+1:]...)
		}
	}

	// Flash
	if g.flash > 0 {
		g.flash--
	}

	// Hover
	mx, my := ebiten.CursorPosition()
	cr, cc := px2rc(mx, my)
	if cr >= 0 && cc >= 0 {
		g.hovR, g.hovC = cr, cc
	} else {
		g.hovR, g.hovC = -1, -1
	}

	// Клик
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if cr >= 0 && cc >= 0 {
			g.click(cr, cc)
		}
	}

	// Рестарт
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.start()
	}

	// Пауза
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.state = S_PAUSE
		g.initPauseButtons()
		playSoundP(sndSwap)
		return nil
	}

	return nil
}

func (g *Game) click(r, c int) {
	if g.busy {
		return
	}

	// Если ничего не выбрано
	if g.selR < 0 {
		g.selR, g.selC = r, c
		g.selectAnim = &SelectAnim{r: r, c: c, t: 0}
		return
	}

	// Кликнули на ту же клетку
	if g.selR == r && g.selC == c {
		g.selR, g.selC = -1, -1
		g.selectAnim = nil
		return
	}

	// Проверяем соседство
	dr := r - g.selR
	dc := c - g.selC
	if (abs(dr) == 1 && dc == 0) || (abs(dc) == 1 && dr == 0) {
		// Пробуем свап
		g.trySwap(g.selR, g.selC, r, c)
		g.selR, g.selC = -1, -1
		g.selectAnim = nil
	} else {
		// Выбрали другую
		g.selR, g.selC = r, c
		g.selectAnim = &SelectAnim{r: r, c: c, t: 0}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *Game) trySwap(r1, c1, r2, c2 int) {
	// Свап
	g.board[r1][c1], g.board[r2][c2] = g.board[r2][c2], g.board[r1][c1]

	// Проверяем совпадения
	matches := g.findMatches()
	if len(matches) > 0 {
		g.moves++
		play(sndSwap)
		g.busy = true
		g.removeMatches(matches)
	} else {
		// Откат
		g.board[r1][c1], g.board[r2][c2] = g.board[r2][c2], g.board[r1][c1]
		play(sndBad)
		g.msg = "Нет совпадений!"
	}
}

func (g *Game) findMatches() [][2]int {
	matched := make(map[string]bool)

	// Горизонталь
	for r := 0; r < ROWS; r++ {
		for c := 0; c <= COLS-MATCH_MIN; c++ {
			v := g.board[r][c]
			if v == 0 {
				continue
			}
			count := 1
			for c2 := c + 1; c2 < COLS && g.board[r][c2] == v; c2++ {
				count++
			}
			if count >= MATCH_MIN {
				for i := 0; i < count; i++ {
					key := fmt.Sprintf("%d,%d", r, c+i)
					matched[key] = true
				}
			}
		}
	}

	// Вертикаль
	for c := 0; c < COLS; c++ {
		for r := 0; r <= ROWS-MATCH_MIN; r++ {
			v := g.board[r][c]
			if v == 0 {
				continue
			}
			count := 1
			for r2 := r + 1; r2 < ROWS && g.board[r2][c] == v; r2++ {
				count++
			}
			if count >= MATCH_MIN {
				for i := 0; i < count; i++ {
					key := fmt.Sprintf("%d,%d", r+i, c)
					matched[key] = true
				}
			}
		}
	}

	result := make([][2]int, 0, len(matched))
	for key := range matched {
		var r, c int
		fmt.Sscanf(key, "%d,%d", &r, &c)
		result = append(result, [2]int{r, c})
	}
	return result
}

func (g *Game) removeMatches(matches [][2]int) {
	g.combo++
	bonus := len(matches) * 10 * g.combo
	g.score += bonus

	if g.combo > 1 {
		play(sndCombo)
		g.msg = fmt.Sprintf("COMBO x%d! +%d", g.combo, bonus)
	} else {
		play(sndMatch)
		g.msg = fmt.Sprintf("Ходов: %d | Combo: x%d | + %d", g.moves, g.combo, bonus)
	}

	// Удаляем с анимацией
	for _, m := range matches {
		typeId := g.board[m[0]][m[1]]
		px := float64(m[1]*TILE + BOARD_OFFX)
		py := float64(m[0]*TILE + BOARD_OFFY)

		// Анимация удаления
		g.removeAnims = append(g.removeAnims, RemoveAnim{
			r: m[0], c: m[1],
			x: px, y: py,
			t: 0, total: 25,
			typeId: typeId,
		})

		g.board[m[0]][m[1]] = 0

		// Частицы
		cx := px + float64(TILE/2)
		cy := py + float64(TILE/2)
		clr := gemColor(typeId)
		g.particles = append(g.particles, spawnParts(cx, cy, clr, 8)...)
	}

	// Вспышка экрана при комбо
	if g.combo >= 2 {
		g.flash = 8
	}

	// Падают вниз
	g.applyGravity()

	// Заполняем пустоты
	g.fillEmpty()

	// Проверка на ещё совпадения (каскад)
	go func() {
		time.Sleep(200 * time.Millisecond)
		matches2 := g.findMatches()
		if len(matches2) > 0 {
			g.busy = true
			g.removeMatches(matches2)
		} else {
			g.busy = false
			g.combo = 0
		}
	}()
}

func (g *Game) applyGravity() {
	for c := 0; c < COLS; c++ {
		writeR := ROWS - 1
		for r := ROWS - 1; r >= 0; r-- {
			if g.board[r][c] != 0 {
				g.board[writeR][c] = g.board[r][c]
				if writeR != r {
					g.board[r][c] = 0
				}
				writeR--
			}
		}
		for r := writeR; r >= 0; r-- {
			g.board[r][c] = 0
		}
	}
}

func (g *Game) fillEmpty() {
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			if g.board[r][c] == 0 {
				g.board[r][c] = 1 + rand.Intn(GEM_TYPES)
			}
		}
	}
}

func px2rc(px, py int) (int, int) {
	px -= BOARD_OFFX
	py -= BOARD_OFFY
	if px < 0 || py < 0 {
		return -1, -1
	}
	c := px / TILE
	r := py / TILE
	if r >= ROWS || c >= COLS {
		return -1, -1
	}
	return r, c
}

func (g *Game) Draw(s *ebiten.Image) {
	s.Fill(color.RGBA{15, 15, 25, 255})

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

	// Фон доски
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			x := c*TILE + BOARD_OFFX
			y := r*TILE + BOARD_OFFY

			// Тёмная клетка
			rect(s, x, y, TILE, TILE, color.RGBA{40, 40, 60, 255})
		}
	}

	// Рамка доски
	rect(s, BOARD_OFFX-2, BOARD_OFFY-2, COLS*TILE+4, 2, color.RGBA{100, 100, 150, 255})
	rect(s, BOARD_OFFX-2, BOARD_OFFY+ROWS*TILE, COLS*TILE+4, 2, color.RGBA{100, 100, 150, 255})
	rect(s, BOARD_OFFX-2, BOARD_OFFY, 2, ROWS*TILE, color.RGBA{100, 100, 150, 255})
	rect(s, BOARD_OFFX+COLS*TILE, BOARD_OFFY, 2, ROWS*TILE, color.RGBA{100, 100, 150, 255})

	// Гемы
	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			v := g.board[r][c]
			if v == 0 {
				continue
			}

			x := c*TILE + BOARD_OFFX
			y := r*TILE + BOARD_OFFY

			// Пропускаем гемы которые сейчас удаляются
			skipped := false
			for _, ra := range g.removeAnims {
				if ra.r == r && ra.c == c {
					skipped = true
					break
				}
			}
			if skipped {
				continue
			}

			spr := gemSprites[v-1]
			if spr != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x), float64(y))
				s.DrawImage(spr, op)
			}

			// Hover
			if g.hovR == r && g.hovC == c && !g.busy {
				op := &ebiten.DrawImageOptions{}
				op.ColorM.Scale(1, 1, 1, 0.3)
				op.GeoM.Translate(float64(x), float64(y))
				s.DrawImage(mkImg(TILE, TILE, color.White), op)
			}
		}
	}

	// Анимация выделения
	if g.selectAnim != nil && g.selR >= 0 && g.selC >= 0 {
		x := g.selC*TILE + BOARD_OFFX
		y := g.selR*TILE + BOARD_OFFY
		g.selectAnim.Draw(s, x, y)
	}

	// Анимации удаления
	for _, ra := range g.removeAnims {
		ra.Draw(s)
	}

	// Частицы
	for _, p := range g.particles {
		a := float64(p.life) / float64(p.maxLife)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x, p.y)
		op.GeoM.Rotate(p.rotation)
		op.GeoM.Translate(-float64(p.sz)/2, -float64(p.sz)/2)
		op.ColorM.Scale(1, 1, 1, a)
		s.DrawImage(mkImg(p.sz, p.sz, p.clr), op)
	}

	// Вспышка экрана при комбо
	if g.flash > 0 {
		a := float64(g.flash) / 8.0 * 0.3
		op := &ebiten.DrawImageOptions{}
		op.ColorM.Scale(1, 1, 1, a)
		s.DrawImage(mkImg(WIN_W, WIN_H, color.White), op)
	}

	// HUD
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, ROWS*TILE+BOARD_OFFY+20)
	s.DrawImage(hudBg, op)

	txt := fmt.Sprintf("Score: %d  |  %s", g.score, g.msg)
	text.Draw(s, txt, basicfont.Face7x13, 10, ROWS*TILE+BOARD_OFFY+52, color.RGBA{255, 255, 255, 255})

	// Иконка
	if coinSprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(WIN_W-80), float64(ROWS*TILE+BOARD_OFFY+30))
		s.DrawImage(coinSprite, op)
	}

	// Win overlay
	if g.state == 2 {
		rect(s, 0, 0, WIN_W, WIN_H, color.RGBA{0, 0, 0, 180})
		msg := "🏆 ОТЛИЧНО!"
		clr := color.RGBA{100, 255, 100, 255}
		bw := text.BoundString(basicfont.Face7x13, msg)
		text.Draw(s, msg, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-20, clr)
		sub := "ENTER — играть снова"
		bw2 := text.BoundString(basicfont.Face7x13, sub)
		text.Draw(s, sub, basicfont.Face7x13, WIN_W/2-bw2.Dx()/2, WIN_H/2+25, color.White)
	}
}

func (g *Game) drawMenu(s *ebiten.Image) {
	s.Fill(color.RGBA{10, 10, 30, 255})

	// Фон со звёздами
	if sprBg := g.menuSprites["stars_bg"]; sprBg != nil {
		for i := 0; i < 4; i++ {
			for j := 0; j < 3; j++ {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(i*200), float64(j*200))
				s.DrawImage(sprBg, op)
			}
		}
	}

	// Анимированные звёзды
	if sprStars := g.menuSprites["stars"]; sprStars != nil {
		t := fCount / 60
		for i := 0; i < 8; i++ {
			x := (i*90 + int(t*15)) % (WIN_W + 40) - 20
			y := 40 + int(math.Sin(float64(fCount)/30+float64(i))*15)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(0.4, 0.4)
			op.GeoM.Translate(float64(x), float64(y))
			s.DrawImage(sprStars, op)
		}
	}

	// Заголовок
	title := "PUZZLE GO"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2+2, WIN_H/2-98, color.RGBA{0, 0, 0, 150})
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-100, color.RGBA{255, 220, 100, 255})

	// Подзаголовок
	sub := "Go365 Challenge — Day 96"
	bwSub := text.BoundString(basicfont.Face7x13, sub)
	text.Draw(s, sub, basicfont.Face7x13, WIN_W/2-bwSub.Dx()/2, WIN_H/2-75, color.RGBA{150, 150, 150, 255})

	// Кнопки
	for _, btn := range g.buttons {
		btn.Draw(s)
	}

	// Управление
	text.Draw(s, "Click gem, then neighbor", basicfont.Face7x13, WIN_W/2-90, WIN_H/2+80, color.RGBA{255, 255, 255, 255})
	text.Draw(s, "ESC / P — Pause", basicfont.Face7x13, WIN_W/2-70, WIN_H/2+100, color.RGBA{255, 255, 255, 255})
	text.Draw(s, "R — Restart", basicfont.Face7x13, WIN_W/2-50, WIN_H/2+120, color.RGBA{255, 255, 255, 255})

	// Особенности
	text.Draw(s, "Match 3+ in a row!", basicfont.Face7x13, WIN_W/2-70, WIN_H/2+160, color.RGBA{255, 255, 100, 255})

	// Анимированные гемы
	for i := 0; i < 6; i++ {
		x := WIN_W/2 - 100 + i*40
		y := WIN_H/2 + 190 + int(fCount/20+int64(i)*8)%12
		spr := gemSprites[i]
		if spr != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			s.DrawImage(spr, op)
		}
	}

	// High Score
	if g.highScore > 0 {
		hs := fmt.Sprintf("High Score: %d", g.highScore)
		bwHS := text.BoundString(basicfont.Face7x13, hs)
		text.Draw(s, hs, basicfont.Face7x13, WIN_W/2-bwHS.Dx()/2, WIN_H/2+240, color.RGBA{255, 255, 80, 255})
	}
}

func (g *Game) drawOptions(s *ebiten.Image) {
	s.Fill(color.RGBA{10, 10, 30, 255})

	if sprBg := g.menuSprites["stars_bg"]; sprBg != nil {
		for i := 0; i < 4; i++ {
			for j := 0; j < 3; j++ {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(i*200), float64(j*200))
				s.DrawImage(sprBg, op)
			}
		}
	}

	frameW, frameH := 300, 120
	frameX, frameY := WIN_W/2-frameW/2, WIN_H/2-frameH/2-30
	rectCachedP(s, frameX, frameY, frameW, frameH, color.RGBA{30, 30, 60, 220})

	title := "CONTROLS"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, frameY+20, color.RGBA{255, 220, 100, 255})

	text.Draw(s, "Click gem, then neighbor to swap", basicfont.Face7x13, WIN_W/2-120, frameY+50, color.RGBA{255, 255, 255, 255})
	text.Draw(s, "ESC / P — Pause", basicfont.Face7x13, WIN_W/2-70, frameY+72, color.RGBA{255, 255, 255, 255})
	text.Draw(s, "R — Restart", basicfont.Face7x13, WIN_W/2-50, frameY+94, color.RGBA{255, 255, 255, 255})

	for _, btn := range g.buttons {
		btn.Draw(s)
	}
}

func (g *Game) drawPause(s *ebiten.Image) {
	// Затемнение
	img := ebiten.NewImage(WIN_W, WIN_H)
	img.Fill(color.RGBA{0, 0, 0, 180})
	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(1, 1, 1, 0.7)
	s.DrawImage(img, op)

	// Заголовок
	title := "PAUSED"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-100, color.RGBA{255, 255, 80, 255})

	// Кнопки
	for _, btn := range g.buttons {
		btn.Draw(s)
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	return WIN_W, WIN_H
}

func rect(s *ebiten.Image, x, y, w, h int, c color.Color) {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	s.DrawImage(img, op)
}

// ======================== MAIN ========================
func main() {
	fmt.Println("═══════════════════════════════════")
	fmt.Println("  PUZZLE GO — Match-3 — Go365 Day 96")
	fmt.Println("  New: Menu | Pause | High Score")
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	ebiten.SetWindowSize(WIN_W, WIN_H)
	ebiten.SetWindowTitle("Puzzle Go — Go365 Day 96")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
