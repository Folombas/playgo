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
	N    = 8
	SZ   = 72
	BDPX = N * SZ
	HUD  = 50
	WW   = BDPX
	WH   = BDPX + HUD
)

const (
	NONE  = iota
	WHITE
	BLACK
)

// ======================== КЭШИРОВАННЫЕ СПРАЙТЫ ========================
var (
	// Доска
	sprLight *ebiten.Image
	sprDark  *ebiten.Image

	// Фигуры (вырезаны из спрайтшитов)
	sprWP *ebiten.Image // белая шашка
	sprBP *ebiten.Image // чёрная шашка
	sprWK *ebiten.Image // белая дамка
	sprBK *ebiten.Image // чёрная дамка

	// Маркеры
	sprMv *ebiten.Image // ход
	sprCp *ebiten.Image // взятие
	sprSl *ebiten.Image // подсветка выбора
	sprSH *ebiten.Image // hover

	// HUD
	sprHBG *ebiten.Image

	// Загружено?
	loaded bool
)

func loadAllSprites() {
	if loaded {
		return
	}
	loaded = true

	// Загружаем ВСЕ спрайтшиты
	checkerImg, _, _ := ebitenutil.NewImageFromFile("assets/sprites/checker.png")
	checkersImg, _, _ := ebitenutil.NewImageFromFile("assets/sprites/checkers.png")
	chessImg, _, _ := ebitenutil.NewImageFromFile("assets/sprites/chess.png")
	_, _, _ = ebitenutil.NewImageFromFile("assets/sprites/wood.png")
	_, _, _ = ebitenutil.NewImageFromFile("assets/sprites/glass.png")
	_, _, _ = ebitenutil.NewImageFromFile("assets/sprites/marble.png")
	_, _, _ = ebitenutil.NewImageFromFile("assets/sprites/plastic.png")
	stonesImg, _, _ := ebitenutil.NewImageFromFile("assets/sprites/stones.png")
	marblesImg, _, _ := ebitenutil.NewImageFromFile("assets/sprites/marbles.png")

	// === КЛЕТКИ ДОСКИ — красивая шахматная доска ===
	// Светлая — кремовая с лёгкой текстурой
	sprLight = mkChessBoardTile(SZ, true)
	// Тёмная — тёмно-зелёная классическая
	sprDark = mkChessBoardTile(SZ, false)
	// Fallback
	if sprLight == nil {
		sprLight = mkImg(SZ, SZ, color.RGBA{240, 217, 181, 255})
	}
	if sprDark == nil {
		sprDark = mkImg(SZ, SZ, color.RGBA{130, 160, 140, 255})
	}

	// === ФИГУРЫ ===
	// chess.png 192x128 — 12 шахматных фигур (2 ряда x 6), каждая 32x32
	// Row 0: light pawn,rook,knight,bishop,queen,king
	// Row 1: dark pawn,rook,knight,bishop,queen,king
	if chessImg != nil {
		cb := chessImg.Bounds()
		pw := cb.Dx() / 6 // 32
		ph := cb.Dy() / 2 // 64

		// Белые = светлая пешка (0,0)
		sprWP = scaleTo(subImg(chessImg, 0, 0, pw, ph), SZ-8, SZ-8)
		// Чёрные = тёмная пешка (0,1)
		sprBP = scaleTo(subImg(chessImg, 0, ph, pw, ph), SZ-8, SZ-8)
		// Белая дамка = светлая королева (4,0)
		sprWK = scaleTo(subImg(chessImg, 4*pw, 0, pw, ph), SZ-8, SZ-8)
		// Чёрная дамка = тёмная королева (4,1)
		sprBK = scaleTo(subImg(chessImg, 4*pw, ph, pw, ph), SZ-8, SZ-8)
	}

	// checker.png 96x64 — 6 шашек (3x2)
	if checkerImg != nil && sprWP == nil {
		cb := checkerImg.Bounds()
		pw := cb.Dx() / 3
		ph := cb.Dy() / 2
		sprWP = scaleTo(subImg(checkerImg, 1*pw, 1*ph, pw, ph), SZ-10, SZ-10)
		sprBP = scaleTo(subImg(checkerImg, 0, 1*ph, pw, ph), SZ-10, SZ-10)
		sprWK = scaleTo(subImg(checkerImg, 1*pw, 0, pw, ph), SZ-10, SZ-10)
		sprBK = scaleTo(subImg(checkerImg, 0, 0, pw, ph), SZ-10, SZ-10)
	}

	// stones.png 96x64 — 6 камней (3x2): black,red,white x 2 rows
	if stonesImg != nil && sprWP == nil {
		sb := stonesImg.Bounds()
		pw := sb.Dx() / 3
		ph := sb.Dy() / 2
		sprWP = scaleTo(subImg(stonesImg, 2*pw, ph, pw, ph), SZ-10, SZ-10)
		sprBP = scaleTo(subImg(stonesImg, 0, ph, pw, ph), SZ-10, SZ-10)
		sprWK = scaleTo(subImg(stonesImg, 2*pw, 0, pw, ph), SZ-10, SZ-10)
		sprBK = scaleTo(subImg(stonesImg, 0, 0, pw, ph), SZ-10, SZ-10)
	}

	// marbles.png 96x64 — 6 шариков (3x2): red,black,white x 2
	if marblesImg != nil && sprWP == nil {
		mb := marblesImg.Bounds()
		pw := mb.Dx() / 3
		ph := mb.Dy() / 2
		sprWP = scaleTo(subImg(marblesImg, 2*pw, ph, pw, ph), SZ-10, SZ-10)
		sprBP = scaleTo(subImg(marblesImg, 1*pw, ph, pw, ph), SZ-10, SZ-10)
		sprWK = scaleTo(subImg(marblesImg, 2*pw, 0, pw, ph), SZ-10, SZ-10)
		sprBK = scaleTo(subImg(marblesImg, 1*pw, 0, pw, ph), SZ-10, SZ-10)
	}

	// checkers.png 128x32 — 4 токена (4x1 по 32px)
	if checkersImg != nil {
		// Токен 0: black+white crown, Токен 1: white+black crown
		// Используем как дополнительные дамки
		tk0 := scaleTo(subImg(checkersImg, 0, 0, 32, 32), SZ-10, SZ-10)
		tk1 := scaleTo(subImg(checkersImg, 32, 0, 32, 32), SZ-10, SZ-10)
		if sprWK == nil {
			sprWK = tk1
		}
		if sprBK == nil {
			sprBK = tk0
		}
		_ = tk0
	}

	// Fallback — генерация
	if sprWP == nil {
		sprWP = mkPc(color.RGBA{255, 248, 230, 255}, color.RGBA{210, 195, 170, 255}, false)
	}
	if sprBP == nil {
		sprBP = mkPc(color.RGBA{55, 40, 30, 255}, color.RGBA{85, 65, 50, 255}, false)
	}
	if sprWK == nil {
		sprWK = mkPc(color.RGBA{255, 248, 230, 255}, color.RGBA{210, 195, 170, 255}, true)
	}
	if sprBK == nil {
		sprBK = mkPc(color.RGBA{55, 40, 30, 255}, color.RGBA{85, 65, 50, 255}, true)
	}

	// Маркеры
	sprMv = mkImg(14, 14, color.RGBA{80, 220, 80, 200})
	sprCp = mkImg(SZ-10, SZ-10, color.RGBA{255, 80, 80, 120})
	sprSl = mkImg(SZ, SZ, color.RGBA{255, 255, 0, 80})
	sprSH = mkImg(SZ, SZ, color.RGBA{100, 255, 100, 50})
	sprHBG = mkImg(WW, HUD, color.RGBA{15, 15, 25, 255})

	fmt.Printf("✓ Sprites loaded:\n")
	fmt.Printf("  Board: light=%dx%d dark=%dx%d\n", sprLight.Bounds().Dx(), sprLight.Bounds().Dy(), sprDark.Bounds().Dx(), sprDark.Bounds().Dy())
	fmt.Printf("  White piece: %dx%d\n", sprWP.Bounds().Dx(), sprWP.Bounds().Dy())
	fmt.Printf("  Black piece: %dx%d\n", sprBP.Bounds().Dx(), sprBP.Bounds().Dy())
	fmt.Printf("  White king: %dx%d\n", sprWK.Bounds().Dx(), sprWK.Bounds().Dy())
	fmt.Printf("  Black king: %dx%d\n", sprBK.Bounds().Dx(), sprBK.Bounds().Dy())
}

func extractCenterTile(src *ebiten.Image, sz int) *ebiten.Image {
	b := src.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	dst := ebiten.NewImage(sz, sz)
	op := &ebiten.DrawImageOptions{}
	sx := float64(sz) / float64(srcW)
	sy := float64(sz) / float64(srcH)
	scale := sx
	if sy < scale {
		scale = sy
	}
	op.GeoM.Scale(scale, scale)
	offX := (float64(sz) - float64(srcW)*scale) / 2
	offY := (float64(sz) - float64(srcH)*scale) / 2
	op.GeoM.Translate(offX, offY)
	dst.DrawImage(src, op)
	return dst
}

// mkChessBoardTile создаёт красивую клетку шахматной доски
func mkChessBoardTile(sz int, light bool) *ebiten.Image {
	img := ebiten.NewImage(sz, sz)

	if light {
		// Светлая клетка — кремовая/слоновая кость
		img.Fill(color.RGBA{240, 225, 200, 255})
		// Лёгкая текстура дерева (линии)
		for i := 0; i < sz; i += 8 {
			for x := 0; x < sz; x++ {
				c := color.RGBA{220, 205, 180, 30}
				img.Set(x, i, c)
			}
		}
		// Рамка
		drawBorder(img, sz, color.RGBA{200, 180, 150, 255})
	} else {
		// Тёмная клетка — классическая зелёная
		img.Fill(color.RGBA{80, 120, 80, 255})
		// Текстура
		for i := 0; i < sz; i += 8 {
			for x := 0; x < sz; x++ {
				c := color.RGBA{60, 100, 60, 40}
				img.Set(x, i, c)
			}
		}
		// Рамка
		drawBorder(img, sz, color.RGBA{50, 90, 50, 255})
	}

	return img
}

func drawBorder(img *ebiten.Image, sz int, clr color.Color) {
	for i := 0; i < sz; i++ {
		img.Set(i, 0, clr)
		img.Set(i, sz-1, clr)
		img.Set(0, i, clr)
		img.Set(sz-1, i, clr)
	}
}

func subImg(src *ebiten.Image, x, y, w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-x), float64(-y))
	img.DrawImage(src, op)
	return img
}

func scaleTo(src *ebiten.Image, w, h int) *ebiten.Image {
	sb := src.Bounds()
	if sb.Dx() == w && sb.Dy() == h {
		return src
	}
	dst := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(sb.Dx()), float64(h)/float64(sb.Dy()))
	dst.DrawImage(src, op)
	return dst
}

func mkImg(w, h int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	return img
}

// Fallback: круглая шашка
func mkPc(body, edge color.Color, king bool) *ebiten.Image {
	d := SZ - 10
	r := d / 2
	img := ebiten.NewImage(d, d)
	for y := 0; y < d; y++ {
		for x := 0; x < d; x++ {
			dx, dy := x-r, y-r
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, body)
			}
		}
	}
	for y := 0; y < d; y++ {
		for x := 0; x < d; x++ {
			dx, dy := x-r, y-r
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > float64(r)-2 && dist <= float64(r) {
				img.Set(x, y, edge)
			}
		}
	}
	if king {
		kr := r / 2
		for y := r - kr; y <= r+kr; y++ {
			for x := r - kr; x <= r+kr; x++ {
				if (x-r)*(x-r)+(y-r)*(y-r) <= kr*kr {
					img.Set(x, y, color.RGBA{255, 215, 0, 255})
				}
			}
		}
	}
	return img
}

// ======================== ЛОГИКА ========================
type Board [N][N]int

func newBoard() Board {
	var b Board
	for r := 0; r < 3; r++ {
		for c := 0; c < N; c++ {
			if (r+c)%2 == 1 {
				b[r][c] = BLACK
			}
		}
	}
	for r := 5; r < 8; r++ {
		for c := 0; c < N; c++ {
			if (r+c)%2 == 1 {
				b[r][c] = WHITE
			}
		}
	}
	return b
}

func isKing(v int) bool  { return v == 3 || v == 4 }
func isWhite(v int) bool { return v == 1 || v == 3 }
func isBlack(v int) bool { return v == 2 || v == 4 }
func mySide(v int) int {
	if isWhite(v) {
		return WHITE
	}
	if isBlack(v) {
		return BLACK
	}
	return NONE
}

type Move struct {
	FR, FC, TR, TC int
	Caps           [][2]int
}

func dirs(v int) [][2]int {
	if isKing(v) {
		return [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	}
	if isWhite(v) {
		return [][2]int{{-1, -1}, {-1, 1}}
	}
	return [][2]int{{1, -1}, {1, 1}}
}

func canCap(b Board, r, c int) bool {
	v := b[r][c]
	if v == NONE {
		return false
	}
	for _, d := range dirs(v) {
		if isKing(v) {
			er, ec := -1, -1
			mr, mc := r+d[0], c+d[1]
			for mr >= 0 && mr < N && mc >= 0 && mc < N {
				if b[mr][mc] != NONE {
					if mySide(b[mr][mc]) != mySide(v) && er < 0 {
						er, ec = mr, mc
					} else {
						break
					}
				} else if er >= 0 {
					_ = ec
					return true
				}
				mr += d[0]
				mc += d[1]
			}
		} else {
			mr, mc := r+d[0]*2, c+d[1]*2
			if mr >= 0 && mr < N && mc >= 0 && mc < N {
				mid := b[r+d[0]][c+d[1]]
				if mid != NONE && mySide(mid) != mySide(v) && b[mr][mc] == NONE {
					return true
				}
			}
		}
	}
	return false
}

func anyCap(b Board, s int) bool {
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			if mySide(b[r][c]) == s && canCap(b, r, c) {
				return true
			}
		}
	}
	return false
}

func pieceMoves(b Board, r, c int, chain [][2]int) []Move {
	v := b[r][c]
	if v == NONE {
		return nil
	}
	var out []Move
	d := dirs(v)
	must := anyCap(b, mySide(v)) || len(chain) > 0

	if must {
		for _, dd := range d {
			if isKing(v) {
				er, ec := -1, -1
				mr, mc := r+dd[0], c+dd[1]
				for mr >= 0 && mr < N && mc >= 0 && mc < N {
					if b[mr][mc] != NONE {
						if mySide(b[mr][mc]) != mySide(v) && er < 0 {
							er, ec = mr, mc
						} else {
							break
						}
					} else if er >= 0 {
						taken := false
						for _, t := range chain {
							if t[0] == er && t[1] == ec {
								taken = true
								break
							}
						}
						if !taken {
							out = append(out, Move{FR: r, FC: c, TR: mr, TC: mc, Caps: append([][2]int{{er, ec}}, chain...)})
						}
					}
					mr += dd[0]
					mc += dd[1]
				}
			} else {
				mr, mc := r+dd[0]*2, c+dd[1]*2
				if mr >= 0 && mr < N && mc >= 0 && mc < N {
					mid := b[r+dd[0]][c+dd[1]]
					if mid != NONE && mySide(mid) != mySide(v) && b[mr][mc] == NONE {
						taken := false
						for _, t := range chain {
							if t[0] == r+dd[0] && t[1] == c+dd[1] {
								taken = true
								break
							}
						}
						if !taken {
							out = append(out, Move{FR: r, FC: c, TR: mr, TC: mc, Caps: append([][2]int{{r + dd[0], c + dd[1]}}, chain...)})
						}
					}
				}
			}
		}
	} else {
		for _, dd := range d {
			if isKing(v) {
				mr, mc := r+dd[0], c+dd[1]
				for mr >= 0 && mr < N && mc >= 0 && mc < N && b[mr][mc] == NONE {
					out = append(out, Move{FR: r, FC: c, TR: mr, TC: mc})
					mr += dd[0]
					mc += dd[1]
				}
			} else {
				mr, mc := r+dd[0], c+dd[1]
				if mr >= 0 && mr < N && mc >= 0 && mc < N && b[mr][mc] == NONE {
					out = append(out, Move{FR: r, FC: c, TR: mr, TC: mc})
				}
			}
		}
	}
	return out
}

func allMoves(b Board, s int) []Move {
	var out []Move
	mc := anyCap(b, s)
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			if mySide(b[r][c]) == s {
				if mc && !canCap(b, r, c) {
					continue
				}
				out = append(out, pieceMoves(b, r, c, nil)...)
			}
		}
	}
	return out
}

func apply(b Board, m Move) Board {
	nb := b
	v := nb[m.FR][m.FC]
	nb[m.FR][m.FC] = NONE
	for _, cap := range m.Caps {
		nb[cap[0]][cap[1]] = NONE
	}
	if isWhite(v) && m.TR == 0 {
		v = 3
	}
	if isBlack(v) && m.TR == N-1 {
		v = 4
	}
	nb[m.TR][m.TC] = v
	return nb
}

// ======================== AI ========================
func aiMove(b Board) *Move {
	moves := allMoves(b, BLACK)
	if len(moves) == 0 {
		return nil
	}
	best := -99999
	var bestM []Move
	for _, m := range moves {
		s := minimax(apply(b, m), 3, -99999, 99999, false)
		s += rand.Intn(3)
		if s > best {
			best = s
			bestM = []Move{m}
		} else if s == best {
			bestM = append(bestM, m)
		}
	}
	return &bestM[rand.Intn(len(bestM))]
}

func minimax(b Board, depth, alpha, beta int, max bool) int {
	if depth == 0 {
		return eval(b)
	}
	s := BLACK
	if !max {
		s = WHITE
	}
	moves := allMoves(b, s)
	if len(moves) == 0 {
		if max {
			return -5000
		}
		return 5000
	}
	if max {
		v := -99999
		for _, m := range moves {
			v2 := minimax(apply(b, m), depth-1, alpha, beta, false)
			if v2 > v {
				v = v2
			}
			if v > alpha {
				alpha = v
			}
			if alpha >= beta {
				break
			}
		}
		return v
	}
	v := 99999
	for _, m := range moves {
		v2 := minimax(apply(b, m), depth-1, alpha, beta, true)
		if v2 < v {
			v = v2
		}
		if v < beta {
			beta = v
		}
		if alpha >= beta {
			break
		}
	}
	return v
}

func eval(b Board) int {
	sc := 0
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			v := b[r][c]
			if v == NONE {
				continue
			}
			val := 10
			if isKing(v) {
				val = 30
			}
			if c >= 2 && c <= 5 && r >= 2 && r <= 5 {
				val += 2
			}
			if isBlack(v) {
				val += r
			} else {
				val += 7 - r
			}
			if isBlack(v) {
				sc += val
			} else {
				sc -= val
			}
		}
	}
	return sc
}

// ======================== ЗВУКИ ========================
var (
	actx  *audio.Context
	sMove *audio.Player
	sCap  *audio.Player
	sKing *audio.Player
	sWin  *audio.Player
	sLose *audio.Player
)

func initAudio() {
	actx = audio.NewContext(44100)
	sMove, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(tone(0.08, 600, 44100))))
	sCap, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(ns(0.12, 44100))))
	sKing, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(arp(0.25, []float64{523, 659, 784}, 44100))))
	sWin, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(arp(0.5, []float64{523, 659, 784, 1047}, 44100))))
	sLose, _ = audio.NewPlayer(actx, bytes.NewReader(mkWAV(sweep(0.4, 400, 100, 44100))))
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
		o[i] = math.Sin(6.2832*freq*t) * math.Exp(-t*10) * 0.4
	}
	return o
}

func ns(dur float64, sr int) []float64 {
	n := int(float64(sr) * dur)
	o := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		o[i] = (rand.Float64()*2 - 1) * math.Exp(-t*12) * 0.5
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

func sweep(dur, f1, f2 float64, sr int) []float64 {
	n := int(float64(sr) * dur)
	o := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		p := float64(i) / float64(n)
		f := f1 + (f2-f1)*p
		o[i] = math.Sin(6.2832*f*t) * math.Exp(-t*6) * 0.4
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
}

func spawnParts(x, y float64, clr color.Color, n int) []Particle {
	ps := make([]Particle, 0, n)
	for i := 0; i < n; i++ {
		a := rand.Float64() * 6.2832
		sp := 1.5 + rand.Float64()*3
		ps = append(ps, Particle{
			x: x, y: y,
			vx: math.Cos(a) * sp,
			vy: math.Sin(a)*sp - 2,
			life:    15 + rand.Intn(20),
			maxLife: 35,
			clr: clr,
			sz:   2 + rand.Intn(5),
		})
	}
	return ps
}

// ======================== ИГРА ========================
var fCount int64

type Game struct {
	board Board
	turn  int
	state int

	selR, selC int
	vMoves     []Move

	hovR, hovC int

	lastFR, lastFC, lastTR, lastTC int

	msg     string
	wCount  int
	bCount  int
	wKings  int
	bKings  int
	moveNum int

	particles []Particle

	aiResult chan *Move
	aiBusy   bool
}

func NewGame() *Game {
	g := &Game{state: 0, turn: WHITE, aiResult: make(chan *Move, 1)}
	initAudio()
	loadAllSprites()
	return g
}

func (g *Game) start() {
	g.board = newBoard()
	g.turn = WHITE
	g.state = 1
	g.selR, g.selC = -1, -1
	g.vMoves = nil
	g.lastFR, g.lastFC, g.lastTR, g.lastTC = -1, -1, -1, -1
	g.msg = "Ваш ход (белые)"
	g.moveNum = 1
	g.aiBusy = false
	g.particles = nil
	g.countPieces()
}

func (g *Game) countPieces() {
	g.wCount, g.bCount, g.wKings, g.bKings = 0, 0, 0, 0
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			v := g.board[r][c]
			if isWhite(v) {
				g.wCount++
				if isKing(v) {
					g.wKings++
				}
			}
			if isBlack(v) {
				g.bCount++
				if isKing(v) {
					g.bKings++
				}
			}
		}
	}
}

func (g *Game) Update() error {
	fCount++

	if g.state == 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.start()
		}
		return nil
	}

	if g.state == 2 || g.state == 3 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.start()
		}
		return nil
	}

	// Частицы
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.12
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// AI результат
	if g.aiBusy {
		select {
		case m := <-g.aiResult:
			g.aiBusy = false
			if m == nil {
				g.state = 2
				g.msg = "Вы победили!"
				play(sWin)
				for i := 0; i < 40; i++ {
					g.particles = append(g.particles, spawnParts(WW/2, WH/2, color.RGBA{100, 255, 100, 255}, 1)...)
				}
			} else {
				g.execMove(*m)
				if len(m.Caps) > 0 {
					p := g.board[m.TR][m.TC]
					if mySide(p) == BLACK && canCap(g.board, m.TR, m.TC) {
						go g.aiChain(m.TR, m.TC, m.Caps)
						return nil
					}
				}
				g.turn = WHITE
				g.moveNum++
				g.msg = "Ваш ход (белые)"
				g.countPieces()
				if len(allMoves(g.board, WHITE)) == 0 {
					g.state = 3
					g.msg = "AI победил!"
					play(sLose)
				}
			}
		default:
		}
		return nil
	}

	if g.turn == BLACK {
		g.aiBusy = true
		g.msg = "AI думает..."
		go func() {
			time.Sleep(400 * time.Millisecond)
			m := aiMove(g.board)
			g.aiResult <- m
		}()
		return nil
	}

	// ИГРОК
	if g.turn == WHITE {
		mx, my := ebiten.CursorPosition()
		cr, cc := px2rc(mx, my)
		if cr >= 0 && cc >= 0 {
			g.hovR, g.hovC = cr, cc
		} else {
			g.hovR, g.hovC = -1, -1
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if cr >= 0 && cc >= 0 {
				g.click(cr, cc)
			}
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.start()
	}

	return nil
}

func (g *Game) aiChain(r, c int, caps [][2]int) {
	time.Sleep(300 * time.Millisecond)
	moves := pieceMoves(g.board, r, c, caps)
	if len(moves) == 0 {
		g.turn = WHITE
		g.moveNum++
		g.msg = "Ваш ход (белые)"
		g.countPieces()
		if len(allMoves(g.board, WHITE)) == 0 {
			g.state = 3
			g.msg = "AI победил!"
			play(sLose)
		}
		return
	}
	m := &moves[rand.Intn(len(moves))]
	g.execMove(*m)
	if canCap(g.board, m.TR, m.TC) {
		g.aiChain(m.TR, m.TC, m.Caps)
	} else {
		g.turn = WHITE
		g.moveNum++
		g.msg = "Ваш ход (белые)"
		g.countPieces()
		if len(allMoves(g.board, WHITE)) == 0 {
			g.state = 3
			g.msg = "AI победил!"
			play(sLose)
		}
	}
}

func (g *Game) execMove(m Move) {
	oldV := g.board[m.FR][m.FC]
	g.board = apply(g.board, m)
	newV := g.board[m.TR][m.TC]

	g.lastFR, g.lastFC = m.FR, m.FC
	g.lastTR, g.lastTC = m.TR, m.TC

	cx := float64(m.TC*SZ + SZ/2)
	cy := float64(m.TR*SZ + SZ/2)
	if len(m.Caps) > 0 {
		play(sCap)
		for _, cap := range m.Caps {
			cx2 := float64(cap[1]*SZ + SZ/2)
			cy2 := float64(cap[0]*SZ + SZ/2)
			g.particles = append(g.particles, spawnParts(cx2, cy2, color.RGBA{255, 200, 100, 255}, 15)...)
		}
	} else {
		play(sMove)
	}
	if isKing(newV) && !isKing(oldV) {
		play(sKing)
		g.particles = append(g.particles, spawnParts(cx, cy, color.RGBA{255, 215, 0, 255}, 20)...)
	}
}

func (g *Game) click(r, c int) {
	v := g.board[r][c]

	if g.selR >= 0 {
		for _, m := range g.vMoves {
			if m.TR == r && m.TC == c {
				g.execMove(m)
				if len(m.Caps) > 0 && canCap(g.board, m.TR, m.TC) {
					g.selR, g.selC = m.TR, m.TC
					g.vMoves = pieceMoves(g.board, m.TR, m.TC, m.Caps)
					g.msg = "Продолжайте взятие!"
					g.countPieces()
					return
				}
				g.selR, g.selC = -1, -1
				g.vMoves = nil
				g.turn = BLACK
				g.moveNum++
				g.msg = "AI думает..."
				g.countPieces()
				if len(allMoves(g.board, BLACK)) == 0 {
					g.state = 2
					g.msg = "Вы победили!"
					play(sWin)
				}
				return
			}
		}
		if isWhite(v) {
			g.pick(r, c)
			return
		}
		g.selR, g.selC = -1, -1
		g.vMoves = nil
		g.msg = "Ваш ход (белые)"
		return
	}

	if isWhite(v) {
		g.pick(r, c)
	}
}

func (g *Game) pick(r, c int) {
	must := anyCap(g.board, WHITE)
	if must && !canCap(g.board, r, c) {
		g.msg = "Обязаны бить!"
		return
	}
	moves := pieceMoves(g.board, r, c, nil)
	if len(moves) == 0 {
		g.msg = "Нет ходов"
		return
	}
	g.selR, g.selC = r, c
	g.vMoves = moves
	g.msg = fmt.Sprintf("Ходов: %d", len(moves))
}

// ======================== РИСОВАНИЕ ========================
func (g *Game) Draw(s *ebiten.Image) {
	s.Fill(color.RGBA{18, 18, 28, 255})

	if g.state == 0 {
		g.drawMenu(s)
		return
	}

	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			x := c * SZ
			y := r * SZ

			// Клетка
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			if (r+c)%2 == 1 {
				s.DrawImage(sprDark, op)
			} else {
				s.DrawImage(sprLight, op)
			}

			// Hover
			if g.hovR == r && g.hovC == c && isWhite(g.board[r][c]) && g.turn == WHITE && g.selR < 0 {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x), float64(y))
				s.DrawImage(sprSH, op)
			}

			// Выбор
			if g.selR == r && g.selC == c {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x), float64(y))
				s.DrawImage(sprSl, op)
			}

			// Последний ход
			if g.lastTR >= 0 {
				if (g.lastFR == r && g.lastFC == c) || (g.lastTR == r && g.lastTC == c) {
					op := &ebiten.DrawImageOptions{}
					op.ColorM.Scale(1, 1, 0.5, 0.15)
					op.GeoM.Translate(float64(x), float64(y))
					s.DrawImage(sprSl, op)
				}
			}

			// Маркеры
			for _, m := range g.vMoves {
				if m.TR == r && m.TC == c {
					op := &ebiten.DrawImageOptions{}
					if len(m.Caps) > 0 {
						op.GeoM.Translate(float64(x+5), float64(y+5))
						s.DrawImage(sprCp, op)
					} else {
						op.GeoM.Translate(float64(x+SZ/2-7), float64(y+SZ/2-7))
						s.DrawImage(sprMv, op)
					}
				}
			}

			// Фигура
			v := g.board[r][c]
			if v != NONE {
				var spr *ebiten.Image
				switch v {
				case 1:
					spr = sprWP
				case 2:
					spr = sprBP
				case 3:
					spr = sprWK
				case 4:
					spr = sprBK
				}
				if spr != nil {
					op := &ebiten.DrawImageOptions{}
					// Центрируем
					sb := spr.Bounds()
					ox := (SZ - sb.Dx()) / 2
					oy := (SZ - sb.Dy()) / 2
					op.GeoM.Translate(float64(x+ox), float64(y+oy))
					s.DrawImage(spr, op)
				}
			}
		}
	}

	// Частицы
	for _, p := range g.particles {
		a := float64(p.life) / float64(p.maxLife)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x-float64(p.sz)/2, p.y-float64(p.sz)/2)
		op.ColorM.Scale(1, 1, 1, a)
		s.DrawImage(mkImg(p.sz, p.sz, p.clr), op)
	}

	// HUD
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, BDPX)
	s.DrawImage(sprHBG, op)

	txt := fmt.Sprintf("⚪Вы:%d(👑%d) | ⚫AI:%d(👑%d) | Ход#%d | %s",
		g.wCount, g.wKings, g.bCount, g.bKings, g.moveNum, g.msg)
	text.Draw(s, txt, basicfont.Face7x13, 6, BDPX+32, color.RGBA{255, 255, 255, 255})

	if g.state == 2 || g.state == 3 {
		op := &ebiten.DrawImageOptions{}
		op.ColorM.Scale(0, 0, 0, 0.7)
		s.DrawImage(sprLight, op)

		msg := "🏆 ВЫ ПОБЕДИЛИ!"
		clr := color.RGBA{100, 255, 100, 255}
		if g.state == 3 {
			msg = "💀 AI ПОБЕДИЛ!"
			clr = color.RGBA{255, 100, 100, 255}
		}
		bw := text.BoundString(basicfont.Face7x13, msg)
		text.Draw(s, msg, basicfont.Face7x13, WW/2-bw.Dx()/2, WH/2-20, clr)
		sub := "ENTER — играть снова"
		bw2 := text.BoundString(basicfont.Face7x13, sub)
		text.Draw(s, sub, basicfont.Face7x13, WW/2-bw2.Dx()/2, WH/2+25, color.RGBA{255, 255, 255, 255})
	}
}

func (g *Game) drawMenu(s *ebiten.Image) {
	s.Fill(color.RGBA{15, 15, 30, 255})

	title := "CHECKERS GO"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WW/2-bw.Dx()/2, WH/2-100, color.RGBA{255, 255, 255, 255})

	t := fCount
	c := color.RGBA{100, 255, 100, 255}
	if (t/30)%2 == 0 {
		c = color.RGBA{255, 255, 100, 255}
	}
	text.Draw(s, "ENTER или SPACE — начать", basicfont.Face7x13, WW/2-105, WH/2-40, c)
	text.Draw(s, "Клик мышкой — выбрать и ходить", basicfont.Face7x13, WW/2-120, WH/2+5, color.RGBA{255, 255, 255, 255})
	text.Draw(s, "Вы играете БЕЛЫМИ (внизу)", basicfont.Face7x13, WW/2-100, WH/2+30, color.RGBA{255, 255, 255, 255})
	text.Draw(s, "R — рестарт", basicfont.Face7x13, WW/2-50, WH/2+60, color.RGBA{255, 255, 255, 255})
	text.Draw(s, "Летающие дамки | Обязательное взятие", basicfont.Face7x13, WW/2-135, WH/2+100, color.RGBA{255, 255, 100, 255})

	// Анимированные фигуры — используем спрайты
	for i := 0; i < 6; i++ {
		x := WW/2 - 100 + i*40
		y := WH/2 + 140 + int(t/20+int64(i)*8)%12
		var spr *ebiten.Image
		if i < 3 {
			spr = sprWP
		} else {
			spr = sprBP
		}
		if spr != nil {
			op := &ebiten.DrawImageOptions{}
			sb := spr.Bounds()
			op.GeoM.Translate(float64(x-sb.Dx()/2), float64(y-sb.Dy()/2))
			s.DrawImage(spr, op)
		}
	}

	text.Draw(s, "Go365 Challenge — Day 95", basicfont.Face7x13, WW/2-95, WH-30, color.RGBA{150, 150, 150, 255})
}

func (g *Game) Layout(w, h int) (int, int) {
	return WW, WH
}

func px2rc(px, py int) (int, int) {
	if px < 0 || px >= BDPX || py < 0 || py >= BDPX {
		return -1, -1
	}
	return py / SZ, px / SZ
}

// ======================== MAIN ========================
func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  CHECKERS GO — Go365 Day 95")
	fmt.Println("  Real Sprites | Flying Kings | AI")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	ebiten.SetWindowSize(WW, WH)
	ebiten.SetWindowTitle("Checkers Go — Go365 Day 95")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
