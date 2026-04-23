// main.go
package main

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"io"
	"log"
	"math"
	mathrand "math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenW       = 640
	screenH       = 480
	tileSize      = 20
	gridW         = screenW / tileSize
	gridH         = screenH / tileSize
	initialLength = 5
)

type Vec struct{ X, Y int }

type GameState int

const (
	STATE_MENU GameState = iota
	STATE_PLAYING
	STATE_GAMEOVER
)

type Particle struct {
	X, Y     float64
	VX, VY   float64
	Life     float64
	Color    color.RGBA
	Size     float64
}

type Game struct {
	rng       *mathrand.Rand
	state     GameState
	snake     []Vec
	dir       Vec
	nextDir   Vec
	ticker    float64
	speed     float64
	apple     Vec
	bombs     []Vec
	score     int
	particles []Particle
	audioCtx  *audio.Context
	sndEat    *audio.Player
	sndBoom   *audio.Player
	sndTick   *audio.Player
	shake     float64
	menuPulse float64
}

func NewGame() *Game {
	g := &Game{
		rng:       mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		state:     STATE_MENU,
		speed:     8,
		menuPulse: 0,
	}
	g.reset()
	// audio
	g.audioCtx = audio.NewContext(44100)
	g.sndEat = newSound(g.audioCtx, beepEat())
	g.sndBoom = newSound(g.audioCtx, beepBoom())
	g.sndTick = newSound(g.audioCtx, beepTick())
	return g
}

func (g *Game) reset() {
	g.snake = nil
	cx, cy := gridW/2, gridH/2
	for i := 0; i < initialLength; i++ {
		g.snake = append(g.snake, Vec{cx - i, cy})
	}
	g.dir = Vec{1, 0}
	g.nextDir = g.dir
	g.placeApple()
	g.bombs = nil
	g.score = 0
	g.ticker = 0
	g.shake = 0
	g.particles = nil
}

func (g *Game) placeApple() {
	for {
		x := g.rng.Intn(gridW)
		y := g.rng.Intn(gridH)
		ok := true
		for _, s := range g.snake {
			if s.X == x && s.Y == y {
				ok = false
				break
			}
		}
		if ok {
			g.apple = Vec{x, y}
			return
		}
	}
}

func (g *Game) Update() error {
	g.menuPulse += 0.06

	if g.state == STATE_MENU {
		if inputPressed() {
			g.state = STATE_PLAYING
			g.reset()
		}
		return nil
	}

	if g.state == STATE_GAMEOVER {
		if inputPressed() {
			g.state = STATE_MENU
		}
		return nil
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) && g.dir.Y != 1 {
		g.nextDir = Vec{0, -1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) && g.dir.Y != -1 {
		g.nextDir = Vec{0, 1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && g.dir.X != 1 {
		g.nextDir = Vec{-1, 0}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) && g.dir.X != -1 {
		g.nextDir = Vec{1, 0}
	}

	g.ticker += g.speed * (1.0 / 60.0)
	if g.ticker >= 1 {
		g.ticker = 0
		g.dir = g.nextDir
		g.step()
	}

	for i := 0; i < len(g.particles); i++ {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.Life -= 0.04
		p.VY += 0.02
	}
	j := 0
	for _, p := range g.particles {
		if p.Life > 0 {
			g.particles[j] = p
			j++
		}
	}
	g.particles = g.particles[:j]

	g.shake *= 0.9
	return nil
}

func (g *Game) step() {
	head := g.snake[0]
	newHead := Vec{head.X + g.dir.X, head.Y + g.dir.Y}
	// bounds -> game over (no wrap)
	if newHead.X < 0 || newHead.X >= gridW || newHead.Y < 0 || newHead.Y >= gridH {
		g.explodeAt(head)
		g.playBoom()
		g.state = STATE_GAMEOVER
		return
	}
	for _, s := range g.snake {
		if s == newHead {
			g.explodeAt(newHead)
			g.playBoom()
			g.state = STATE_GAMEOVER
			return
		}
	}
	g.snake = append([]Vec{newHead}, g.snake...)
	ate := newHead == g.apple
	if ate {
		g.score++
		g.placeApple()
		g.spawnBombRandom()
		g.playEat()
		g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 20, color.RGBA{255, 50, 50, 255})
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}
	for _, b := range g.bombs {
		if b == newHead {
			g.explodeAt(newHead)
			g.playBoom()
			g.state = STATE_GAMEOVER
			return
		}
	}
	g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 3, color.RGBA{0, 200, 255, 160})
}

func (g *Game) spawnBombRandom() {
	if g.rng.Float64() < 0.35 {
		for i := 0; i < 1000; i++ {
			x := g.rng.Intn(gridW)
			y := g.rng.Intn(gridH)
			ok := true
			if x == g.apple.X && y == g.apple.Y {
				ok = false
			}
			for _, s := range g.snake {
				if s.X == x && s.Y == y {
					ok = false
					break
				}
			}
			if ok {
				g.bombs = append(g.bombs, Vec{x, y})
				return
			}
		}
	}
}

func (g *Game) explodeAt(v Vec) {
	g.addParticles(float64(v.X*tileSize+tileSize/2), float64(v.Y*tileSize+tileSize/2), 60, color.RGBA{255, 200, 60, 255})
	g.shake = 12
}

func (g *Game) addParticles(x, y float64, n int, c color.RGBA) {
	for i := 0; i < n; i++ {
		a := g.rng.Float64() * 2 * math.Pi
		s := g.rng.Float64()*3 + 1
		g.particles = append(g.particles, Particle{
			X:     x,
			Y:     y,
			VX:    math.Cos(a) * s,
			VY:    math.Sin(a) * s,
			Life:  g.rng.Float64()*1.2 + 0.3,
			Color: c,
			Size:  g.rng.Float64()*3 + 1,
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 20, 255})

	ox, oy := 0.0, 0.0
	if g.shake > 0.05 {
		ox = (mathrand.Float64()*2 - 1) * g.shake
		oy = (mathrand.Float64()*2 - 1) * g.shake
	}

	// grid background subtle
	for x := 0; x < gridW; x++ {
		for y := 0; y < gridH; y++ {
			r := (x+y)%2 == 0
			c := color.RGBA{14, 14, 24, 255}
			if r {
				c = color.RGBA{16, 16, 28, 255}
			}
			ebitenutil.DrawRect(screen, float64(x*tileSize)+ox, float64(y*tileSize)+oy, tileSize-1, tileSize-1, c)
		}
	}

	// draw helpers
	drawApple := func(x, y int) {
		cx := float64(x*tileSize + tileSize/2)
		cy := float64(y*tileSize + tileSize/2)
		ebitenutil.DrawRect(screen, cx-7+ox, cy-6+oy, 14, 12, color.RGBA{200, 20, 30, 255})
		ebitenutil.DrawRect(screen, cx-5+ox, cy-10+oy, 10, 6, color.RGBA{220, 30, 40, 255})
		ebitenutil.DrawRect(screen, cx+6+ox, cy-12+oy, 6, 3, color.RGBA{30, 160, 50, 255})
		ebitenutil.DrawRect(screen, cx-4+ox, cy-8+oy, 4, 2, color.RGBA{255, 120, 120, 160})
	}
	drawBomb := func(x, y int) {
		cx := float64(x*tileSize + tileSize/2)
		cy := float64(y*tileSize + tileSize/2)
		ebitenutil.DrawRect(screen, cx-8+ox, cy-8+oy, 16, 16, color.RGBA{10, 10, 10, 255})
		ebitenutil.DrawRect(screen, cx+8+ox, cy-11+oy, 8, 2, color.RGBA{120, 60, 10, 255})
		flame := math.Sin(float64(time.Now().UnixNano())/1e9*10+float64(x+y))*2 + 4
		ebitenutil.DrawRect(screen, cx+14+ox, cy-11+oy, flame, 3, color.RGBA{255, 140, 20, 255})
	}

	for _, b := range g.bombs {
		drawBomb(b.X, b.Y)
	}

	for i, s := range g.snake {
		x := float64(s.X*tileSize) + ox
		y := float64(s.Y*tileSize) + oy
		if i == 0 {
			ebitenutil.DrawRect(screen, x, y, tileSize-1, tileSize-1, color.RGBA{20, 200, 80, 255})
			eyex := float64(tileSize)/4 - 2
			eyey := float64(tileSize)/4 - 2
			ebitenutil.DrawRect(screen, x+eyex, y+eyey, 4, 4, color.White)
			ebitenutil.DrawRect(screen, x+float64(tileSize)-eyex-6, y+eyey, 4, 4, color.White)
			ebitenutil.DrawRect(screen, x+eyex+1, y+eyey+1, 2, 2, color.Black)
			ebitenutil.DrawRect(screen, x+float64(tileSize)-eyex-5, y+eyey+1, 2, 2, color.Black)
			// Draw tongue based on direction
			tongueLen := 6.0
			var tx, ty float64
			switch g.dir {
			case Vec{1, 0}: // right
				tx, ty = float64(tileSize)-6, float64(tileSize)/2-2
			case Vec{-1, 0}: // left
				tx, ty = 2, float64(tileSize)/2-2
			case Vec{0, 1}: // down
				tx, ty = float64(tileSize)/2-2, float64(tileSize)-6
			case Vec{0, -1}: // up
				tx, ty = float64(tileSize)/2-2, 2
			}
			ebitenutil.DrawRect(screen, x+tx+ox, y+ty+oy, 4, tongueLen, color.RGBA{255, 80, 120, 255})
		} else {
			shade := uint8(120 + (i*3)%120)
			c := color.RGBA{20, shade, 80, 255}
			ebitenutil.DrawRect(screen, x, y, tileSize-1, tileSize-1, c)
		}
	}

	drawApple(g.apple.X, g.apple.Y)

	for _, p := range g.particles {
		ebitenutil.DrawRect(screen, p.X-p.Size+ox, p.Y-p.Size+oy, p.Size*2, p.Size*2, p.Color)
	}

	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(g.score), 8, 8)

	switch g.state {
	case STATE_MENU:
		ebitenutil.DrawRect(screen, 80+ox, 60+oy, screenW-160, screenH-140, color.RGBA{0, 0, 0, 180})
		title := "S N A K E  —  R E V I V E D"
		ebitenutil.DebugPrintAt(screen, title, screenW/2-160, 100)
		ebitenutil.DebugPrintAt(screen, "Press any arrow key or Enter to start", screenW/2-140, 160+int(10*math.Sin(g.menuPulse)))
		ebitenutil.DebugPrintAt(screen, "Collect red apples, avoid black bombs!", screenW/2-140, 200)
	case STATE_GAMEOVER:
		ebitenutil.DrawRect(screen, 80+ox, 60+oy, screenW-160, screenH-140, color.RGBA{30, 0, 0, 220})
		ebitenutil.DebugPrintAt(screen, "GAME OVER", screenW/2-60, screenH/2-20)
		ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(g.score), screenW/2-40, screenH/2+10)
		ebitenutil.DebugPrintAt(screen, "Press any key to return to menu", screenW/2-120, screenH/2+40)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}

/* --- Utilities & procedural audio --- */

func inputPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyEnter) ||
		ebiten.IsKeyPressed(ebiten.KeySpace) ||
		ebiten.IsKeyPressed(ebiten.KeyUp) ||
		ebiten.IsKeyPressed(ebiten.KeyDown) ||
		ebiten.IsKeyPressed(ebiten.KeyLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyRight)
}

func newSound(ctx *audio.Context, data []byte) *audio.Player {
	d, err := wav.Decode(ctx, bytes.NewReader(data))
	if err != nil {
		log.Printf("failed to decode wav: %v", err)
		return nil
	}
	p, err := audio.NewPlayer(ctx, d)
	if err != nil {
		log.Printf("failed to create audio player: %v", err)
		return nil
	}
	return p
}

func playPlayer(p *audio.Player) {
	if p == nil {
		return
	}
	_ = p.Rewind()
	_ = p.Play()
}

func (g *Game) playEat()  { playPlayer(g.sndEat) }
func (g *Game) playBoom() { playPlayer(g.sndBoom) }
func (g *Game) playTick() { playPlayer(g.sndTick) }

/* Procedural WAV generator and helpers */

func writeLEUint16(w io.Writer, v uint16) {
	_ = binary.Write(w, binary.LittleEndian, v)
}
func writeLEUint32(w io.Writer, v uint32) {
	_ = binary.Write(w, binary.LittleEndian, v)
}

// synthRawPCM returns mono int16 samples.
func synthRawPCM(sampleRate int, durSec float64, freq float64, amp float64, wave string) []int16 {
	numSamples := int(float64(sampleRate) * durSec)
	out := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		var s float64
		switch wave {
		case "sine":
			s = math.Sin(2*math.Pi*freq*t)
		case "square":
			if math.Sin(2*math.Pi*freq*t) >= 0 {
				s = 1
			} else {
				s = -1
			}
		case "saw":
			s = 2*(t*freq-math.Floor(t*freq+0.5))
		case "noise":
			s = mathrand.NormFloat64() * 0.6
		default:
			s = math.Sin(2*math.Pi*freq*t)
		}
		env := 1.0
		att := math.Min(0.01, durSec*0.15)
		dec := math.Min(0.05, durSec*0.2)
		if t < att {
			env = t / att
		} else if t < att+dec {
			env = 1 - (t-att)/dec*0.5
		} else {
			rem := durSec - t
			if rem < 0 {
				env = 0
			} else {
				env = math.Max(0, rem/(durSec*0.6))
			}
		}
		val := s * amp * env
		if val > 1 {
			val = 1
		}
		if val < -1 {
			val = -1
		}
		out[i] = int16(val * 32767)
	}
	return out
}

func mixToWAV(sampleRate int, tracks [][]int16) []byte {
	maxLen := 0
	for _, t := range tracks {
		if len(t) > maxLen {
			maxLen = len(t)
		}
	}
	mix := make([]int32, maxLen)
	for _, t := range tracks {
		for i := 0; i < len(t); i++ {
			mix[i] += int32(t[i])
		}
	}
	var peak int32
	for i := 0; i < len(mix); i++ {
		if abs32(mix[i]) > peak {
			peak = abs32(mix[i])
		}
	}
	scale := 1.0
	if peak == 0 {
		scale = 1
	} else if peak > 32767 {
		scale = 32767.0 / float64(peak)
	}

	buf := &bytes.Buffer{}
	dataSize := maxLen * 2
	riffSize := 36 + dataSize
	buf.WriteString("RIFF")
	writeLEUint32(buf, uint32(riffSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	writeLEUint32(buf, 16)
	writeLEUint16(buf, 1)
	writeLEUint16(buf, 1)
	writeLEUint32(buf, uint32(sampleRate))
	writeLEUint32(buf, uint32(sampleRate*2))
	writeLEUint16(buf, 2)
	writeLEUint16(buf, 16)
	buf.WriteString("data")
	writeLEUint32(buf, uint32(dataSize))
	for i := 0; i < maxLen; i++ {
		v := int32(0)
		if i < len(mix) {
			v = mix[i]
		}
		vScaled := int16(float64(v) * scale)
		_ = binary.Write(buf, binary.LittleEndian, vScaled)
	}
	return buf.Bytes()
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// Specific sound constructors using synthRawPCM + mixToWAV

func beepEat() []byte {
	const sr = 44100
	t1 := synthRawPCM(sr, 0.12, 880, 0.6, "sine")
	t2 := synthRawPCM(sr, 0.12, 1320, 0.35, "sine")
	return mixToWAV(sr, [][]int16{t1, t2})
}

func beepBoom() []byte {
	const sr = 44100
	low := synthRawPCM(sr, 0.28, 110, 0.9, "sine")
	n := synthRawPCM(sr, 0.28, 0, 0.6, "noise")
	return mixToWAV(sr, [][]int16{low, n})
}

func beepTick() []byte {
	const sr = 44100
	sq := synthRawPCM(sr, 0.06, 1200, 0.6, "square")
	hi := synthRawPCM(sr, 0.06, 4000, 0.25, "sine")
	return mixToWAV(sr, [][]int16{sq, hi})
}

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Snake — Ebitengine")
	g := NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
