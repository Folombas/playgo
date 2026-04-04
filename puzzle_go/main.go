package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
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
	state   int // 0=menu, 1=play, 2=win

	selR, selC int // выбранная клетка
	hovR, hovC int

	particles []Particle
	anims     []SlideAnim

	busy bool // анимация идёт

	msg string
}

func NewGame() *Game {
	g := &Game{state: 0}
	g.selR, g.selC = -1, -1
	g.hovR, g.hovC = -1, -1
	initAudio()
	loadSprites()
	return g
}

func (g *Game) start() {
	g.board = [ROWS][COLS]int{}
	g.score = 0
	g.combo = 0
	g.moves = 0
	g.state = 1
	g.selR, g.selC = -1, -1
	g.particles = nil
	g.anims = nil
	g.busy = false
	g.msg = "Ходов: 0 | Combo: x0"

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

	if g.state == 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.start()
		}
		return nil
	}

	if g.state == 2 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.start()
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

	return nil
}

func (g *Game) click(r, c int) {
	if g.busy {
		return
	}

	// Если ничего не выбрано
	if g.selR < 0 {
		g.selR, g.selC = r, c
		return
	}

	// Кликнули на ту же клетку
	if g.selR == r && g.selC == c {
		g.selR, g.selC = -1, -1
		return
	}

	// Проверяем соседство
	dr := r - g.selR
	dc := c - g.selC
	if (abs(dr) == 1 && dc == 0) || (abs(dc) == 1 && dr == 0) {
		// Пробуем свап
		g.trySwap(g.selR, g.selC, r, c)
		g.selR, g.selC = -1, -1
	} else {
		// Выбрали другую
		g.selR, g.selC = r, c
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

	// Удаляем
	for _, m := range matches {
		g.board[m[0]][m[1]] = 0
		// Частицы
		px := float64(m[1]*TILE + BOARD_OFFX + TILE/2)
		py := float64(m[0]*TILE + BOARD_OFFY + TILE/2)
		clr := gemColor(g.board[m[0]][m[1]])
		g.particles = append(g.particles, spawnParts(px, py, clr, 8)...)
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

func gemColor(t int) color.Color {
	clrs := []color.Color{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{60, 120, 255, 255},   // blue
		color.RGBA{255, 80, 80, 255},    // red
		color.RGBA{80, 200, 80, 255},    // green
		color.RGBA{255, 255, 80, 255},   // yellow
		color.RGBA{180, 80, 255, 255},   // purple
		color.RGBA{180, 180, 180, 255},  // grey
	}
	if t >= 0 && t < len(clrs) {
		return clrs[t]
	}
	return color.White
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

	if g.state == 0 {
		g.drawMenu(s)
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

			// Выбор
			if g.selR == r && g.selC == c {
				if selSprite != nil {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(float64(x-2), float64(y-2))
					s.DrawImage(selSprite, op)
				}
			}
		}
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
	s.Fill(color.RGBA{15, 15, 30, 255})

	title := "PUZZLE GO"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WIN_W/2-bw.Dx()/2, WIN_H/2-100, color.White)

	t := fCount
	c := color.RGBA{100, 255, 100, 255}
	if (t/30)%2 == 0 {
		c = color.RGBA{255, 255, 100, 255}
	}
	text.Draw(s, "ENTER или SPACE — начать", basicfont.Face7x13, WIN_W/2-100, WIN_H/2-40, c)
	text.Draw(s, "Кликни на гем, затем на соседний", basicfont.Face7x13, WIN_W/2-120, WIN_H/2+5, color.White)
	text.Draw(s, "Собирай 3+ в ряд!", basicfont.Face7x13, WIN_W/2-70, WIN_H/2+30, color.White)
	text.Draw(s, "R — рестарт", basicfont.Face7x13, WIN_W/2-50, WIN_H/2+60, color.White)

	// Анимированные гемы
	for i := 0; i < 6; i++ {
		x := WIN_W/2 - 100 + i*40
		y := WIN_H/2 + 120 + int(t/20+int64(i)*8)%12
		spr := gemSprites[i]
		if spr != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			s.DrawImage(spr, op)
		}
	}

	text.Draw(s, "Go365 Challenge — Day 95", basicfont.Face7x13, WIN_W/2-95, WIN_H-30, color.RGBA{150, 150, 150, 255})
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
	rand.Seed(time.Now().UnixNano())
	fmt.Println("═══════════════════════════════════")
	fmt.Println("  PUZZLE GO — Match-3 — Go365 Day 95")
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	ebiten.SetWindowSize(WIN_W, WIN_H)
	ebiten.SetWindowTitle("Puzzle Go — Go365 Day 95")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
