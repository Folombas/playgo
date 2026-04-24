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
	screenW       = 1280
	screenH       = 800
	tileSize      = 32
	gridW         = screenW / tileSize
	gridH         = screenH / tileSize
	initialLength = 5
	maxHealth     = 100
)

type Vec struct{ X, Y int }

type GameState int

const (
	STATE_MENU GameState = iota
	STATE_PLAYING
	STATE_PAUSED
	STATE_GAMEOVER
)

type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
	Color  color.RGBA
	Size   float64
	Glow   bool
}

type Game struct {
	rng           *mathrand.Rand
	state         GameState
	snake         []Vec
	dir           Vec
	nextDir       Vec
	ticker        float64
	speed         float64
	apple         Vec
	bombs         []Vec
	score         int
	health        int
	particles     []Particle
	audioCtx      *audio.Context
	sndEat        *audio.Player
	sndBoom       *audio.Player
	sndHeal       *audio.Player
	sndPause      *audio.Player
	shake         float64
	flash         float64
	menuPulse     float64
	pauseCooldown float64
}

func NewGame() *Game {
	g := &Game{
		rng:       mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		state:     STATE_MENU,
		speed:     9,
		health:    maxHealth,
		menuPulse: 0,
	}
	g.reset()
	g.audioCtx = audio.NewContext(44100)
	g.sndEat = newSound(g.audioCtx, sndEat())
	g.sndBoom = newSound(g.audioCtx, sndBoom())
	g.sndHeal = newSound(g.audioCtx, sndHeal())
	g.sndPause = newSound(g.audioCtx, sndPause())
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
	g.health = maxHealth
	g.placeApple()
	g.bombs = nil
	g.score = 0
	g.ticker = 0
	g.shake = 0
	g.flash = 0
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
		for _, b := range g.bombs {
			if b.X == x && b.Y == y {
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
	dt := 1.0 / 60.0
	g.menuPulse += 0.05
	if g.pauseCooldown > 0 {
		g.pauseCooldown -= dt
	}

	// Pause toggle
	if (ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyP)) && g.pauseCooldown <= 0 {
		if g.state == STATE_PLAYING {
			g.state = STATE_PAUSED
			g.sndPause.Rewind()
			g.sndPause.Play()
		} else if g.state == STATE_PAUSED {
			g.state = STATE_PLAYING
		} else if g.state == STATE_MENU || g.state == STATE_GAMEOVER {
			g.state = STATE_PLAYING
			g.reset()
		}
		g.pauseCooldown = 0.3
	}

	if g.state == STATE_PAUSED {
		return nil
	}

	if g.state == STATE_MENU {
		if inputPressed() && g.pauseCooldown <= 0 {
			g.state = STATE_PLAYING
			g.reset()
			g.pauseCooldown = 0.2
		}
		return nil
	}

	if g.state == STATE_GAMEOVER {
		if inputPressed() && g.pauseCooldown <= 0 {
			g.state = STATE_MENU
			g.pauseCooldown = 0.2
		}
		return nil
	}

	// Direction
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

	g.ticker += g.speed * dt
	if g.ticker >= 1 {
		g.ticker = 0
		g.dir = g.nextDir
		g.step()
	}

	// Particles
	for i := 0; i < len(g.particles); i++ {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.05
		p.Life -= 0.02
		p.Size *= 0.98
	}
	j := 0
	for _, p := range g.particles {
		if p.Life > 0 {
			g.particles[j] = p
			j++
		}
	}
	g.particles = g.particles[:j]

	g.shake *= 0.88
	g.flash *= 0.85
	return nil
}

func (g *Game) step() {
	head := g.snake[0]
	newHead := Vec{head.X + g.dir.X, head.Y + g.dir.Y}

	// Wall & self collision -> instant game over
	if newHead.X < 0 || newHead.X >= gridW || newHead.Y < 0 || newHead.Y >= gridH {
		g.triggerExplosion(head, true)
		return
	}
	for _, s := range g.snake {
		if s == newHead {
			g.triggerExplosion(newHead, true)
			return
		}
	}

	g.snake = append([]Vec{newHead}, g.snake...)

	// Apple
	if newHead == g.apple {
		g.score++
		g.health = minInt(maxHealth, g.health+25)
		g.placeApple()
		g.spawnBombRandom()
		g.sndHeal.Rewind()
		g.sndHeal.Play()
		g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 25, color.RGBA{50, 255, 80, 255}, true)
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}

	// Bombs
	for _, b := range g.bombs {
		if b == newHead {
			g.health -= 35
			g.triggerExplosion(b, g.health <= 0)
			return
		}
	}

	// Trail particles
	g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 2, color.RGBA{0, 180, 220, 140}, false)
}

func (g *Game) triggerExplosion(v Vec, fatal bool) {
	if fatal {
		g.health = 0
	}
	g.flash = 1.0
	g.shake = 18
	g.sndBoom.Rewind()
	g.sndBoom.Play()
	g.addParticles(float64(v.X*tileSize+tileSize/2), float64(v.Y*tileSize+tileSize/2), 80, color.RGBA{255, 120, 30, 255}, true)
	g.addParticles(float64(v.X*tileSize+tileSize/2), float64(v.Y*tileSize+tileSize/2), 40, color.RGBA{255, 255, 200, 200}, true)
	if fatal {
		g.state = STATE_GAMEOVER
	}
}

func (g *Game) spawnBombRandom() {
	if g.rng.Float64() < 0.4 {
		for i := 0; i < 2000; i++ {
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

func (g *Game) addParticles(x, y float64, n int, c color.RGBA, glow bool) {
	for i := 0; i < n; i++ {
		a := g.rng.Float64() * 2 * math.Pi
		s := g.rng.Float64()*4 + 1.5
		g.particles = append(g.particles, Particle{
			X:     x,
			Y:     y,
			VX:    math.Cos(a) * s,
			VY:    math.Sin(a) * s,
			Life:  g.rng.Float64()*1.5 + 0.4,
			Color: c,
			Size:  g.rng.Float64()*4 + 2,
			Glow:  glow,
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 12, 20, 255})

	ox, oy := 0.0, 0.0
	if g.shake > 0.5 {
		ox = (mathrand.Float64()*2 - 1) * g.shake
		oy = (mathrand.Float64()*2 - 1) * g.shake
	}

	// Grid
	for x := 0; x < gridW; x++ {
		for y := 0; y < gridH; y++ {
			c := color.RGBA{15, 15, 25, 255}
			if (x+y)%2 != 0 {
				c = color.RGBA{18, 18, 30, 255}
			}
			ebitenutil.DrawRect(screen, float64(x*tileSize)+ox, float64(y*tileSize)+oy, tileSize-1, tileSize-1, c)
		}
	}

	// --- BOMBS (round black with highlights) ---
	pulse := math.Sin(g.menuPulse*3)*0.15 + 1.0
	for _, b := range g.bombs {
		cx := float64(b.X*tileSize+tileSize/2) + ox
		cy := float64(b.Y*tileSize+tileSize/2) + oy
		radius := float64(tileSize) / 2 * pulse * 0.85

		// Main black circle
		ebitenutil.DrawCircle(screen, cx, cy, radius, color.RGBA{20, 20, 25, 255})
		// Shadow
		ebitenutil.DrawCircle(screen, cx-2, cy-2, radius-2, color.RGBA{0, 0, 0, 100})
		// Specular highlight
		ebitenutil.DrawCircle(screen, cx-radius*0.3, cy-radius*0.35, radius*0.25, color.RGBA{255, 255, 255, 180})
		// Red glow
		ebitenutil.DrawCircle(screen, cx+radius*0.2, cy+radius*0.2, radius*0.2, color.RGBA{255, 80, 80, 120})
		// Fuse
		ebitenutil.DrawRect(screen, cx+radius*0.7, cy-radius*1.1, 4, 6, color.RGBA{60, 50, 40, 255})
		flicker := mathrand.Float64()*3 + 2
		ebitenutil.DrawRect(screen, cx+radius*0.7, cy-radius*1.4, flicker, flicker, color.RGBA{255, 120, 20, 255})
	}

	// --- SNAKE ---
	for i, s := range g.snake {
		x := float64(s.X*tileSize) + ox
		y := float64(s.Y*tileSize) + oy
		base := color.RGBA{20, 220, 90, 255}
		if i > 0 {
			shade := uint8(100 + (i*4)%100)
			base = color.RGBA{15, shade, 70, 255}
		}
		if i == 0 {
			ebitenutil.DrawRect(screen, x-3, y-3, tileSize+6, tileSize+6, color.RGBA{0, 200, 80, 40})
		}
		ebitenutil.DrawRect(screen, x, y, tileSize-1, tileSize-1, base)

		if i == 0 {
			// Eyes
			eyex := float64(tileSize)/4 - 2
			eyey := float64(tileSize)/4 - 2
			ebitenutil.DrawRect(screen, x+eyex, y+eyey, 4, 4, color.White)
			ebitenutil.DrawRect(screen, x+float64(tileSize)-eyex-6, y+eyey, 4, 4, color.White)
			ebitenutil.DrawRect(screen, x+eyex+1, y+eyey+1, 2, 2, color.Black)
			ebitenutil.DrawRect(screen, x+float64(tileSize)-eyex-5, y+eyey+1, 2, 2, color.Black)
			// Tongue
			var tx, ty, w, h float64
			switch g.dir {
			case Vec{1, 0}:
				tx, ty, w, h = tileSize-4, tileSize/2-2, 6, 4
			case Vec{-1, 0}:
				tx, ty, w, h = -2, tileSize/2-2, 6, 4
			case Vec{0, 1}:
				tx, ty, w, h = tileSize/2-2, tileSize-4, 4, 6
			case Vec{0, -1}:
				tx, ty, w, h = tileSize/2-2, -2, 4, 6
			}
			ebitenutil.DrawRect(screen, x+tx+ox, y+ty+oy, w, h, color.RGBA{255, 70, 100, 255})
		}
	}

	// --- APPLE (round) ---
	{
		cx := float64(g.apple.X*tileSize+tileSize/2) + ox
		cy := float64(g.apple.Y*tileSize+tileSize/2) + oy
		radius := float64(tileSize)/2 - 2

		// Shadow
		ebitenutil.DrawCircle(screen, cx-2, cy-2, radius-1, color.RGBA{0, 0, 0, 80})
		// Main red circle
		ebitenutil.DrawCircle(screen, cx, cy, radius, color.RGBA{230, 40, 50, 255})
		// Bright inner part
		ebitenutil.DrawCircle(screen, cx-3, cy-3, radius-4, color.RGBA{255, 100, 100, 150})
		// Highlight
		ebitenutil.DrawCircle(screen, cx-radius*0.3, cy-radius*0.35, radius*0.2, color.RGBA{255, 255, 255, 220})
		// Stem
		ebitenutil.DrawRect(screen, cx+radius*0.5, cy-radius*0.8, 6, 3, color.RGBA{70, 180, 50, 255})
		ebitenutil.DrawRect(screen, cx+radius*0.7, cy-radius*0.9, 8, 2, color.RGBA{50, 150, 30, 255})
	}

	// --- PARTICLES ---
	for _, p := range g.particles {
		c := p.Color
		if p.Glow {
			ebitenutil.DrawRect(screen, p.X-p.Size*1.5+ox, p.Y-p.Size*1.5+oy, p.Size*3, p.Size*3, color.RGBA{c.R, c.G, c.B, uint8(float64(c.A) * 0.4 * p.Life)})
		}
		ebitenutil.DrawRect(screen, p.X-p.Size+ox, p.Y-p.Size+oy, p.Size*2, p.Size*2, c)
	}

	// --- UI ---
	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(g.score), 10, 10)
	// Health bar
	barX := float64(screenW - 20)
	barW := 150.0
	barH := 14.0
	healthPct := float64(g.health) / float64(maxHealth)
	ebitenutil.DrawRect(screen, barX-barW, 10, barW, barH, color.RGBA{30, 30, 40, 200})
	ebitenutil.DrawRect(screen, barX-barW, 10, barW*healthPct, barH, color.RGBA{50, 255, 80, 255})
	ebitenutil.DebugPrintAt(screen, "HEALTH", int(barX-barW+40), 12)

	// Overlays
	switch g.state {
	case STATE_MENU:
		ebitenutil.DrawRect(screen, 100, 80, screenW-200, screenH-180, color.RGBA{0, 0, 0, 160})
		ebitenutil.DebugPrintAt(screen, "S N A K E  —  R E V I V E D", screenW/2-180, 120)
		ebitenutil.DebugPrintAt(screen, "Press ESC/P to start", screenW/2-120, 180+int(10*math.Sin(g.menuPulse)))
		ebitenutil.DebugPrintAt(screen, "Collect red apples, dodge black bombs!", screenW/2-160, 220)
	case STATE_PAUSED:
		ebitenutil.DrawRect(screen, 0, 0, screenW, screenH, color.RGBA{0, 0, 0, 120})
		ebitenutil.DebugPrintAt(screen, "PAUSED", screenW/2-40, screenH/2-20)
	case STATE_GAMEOVER:
		ebitenutil.DrawRect(screen, 100, 80, screenW-200, screenH-180, color.RGBA{40, 0, 0, 180})
		ebitenutil.DebugPrintAt(screen, "GAME OVER", screenW/2-60, screenH/2-40)
		ebitenutil.DebugPrintAt(screen, "Final Score: "+strconv.Itoa(g.score), screenW/2-80, screenH/2)
		ebitenutil.DebugPrintAt(screen, "Press ESC/P to continue", screenW/2-130, screenH/2+40)
	}

	// Screen flash
	if g.flash > 0.01 {
		alpha := uint8(g.flash * 150)
		ebitenutil.DrawRect(screen, 0, 0, screenW, screenH, color.RGBA{255, 255, 255, alpha})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}

/* --- Utils & Input --- */
func inputPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyEnter) ||
		ebiten.IsKeyPressed(ebiten.KeySpace) ||
		ebiten.IsKeyPressed(ebiten.KeyUp) ||
		ebiten.IsKeyPressed(ebiten.KeyDown) ||
		ebiten.IsKeyPressed(ebiten.KeyLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyRight)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

/* --- Audio --- */
func newSound(ctx *audio.Context, data []byte) *audio.Player {
	d, err := wav.Decode(ctx, bytes.NewReader(data))
	if err != nil {
		log.Printf("wav decode err: %v", err)
		return nil
	}
	p, err := audio.NewPlayer(ctx, d)
	if err != nil {
		log.Printf("audio player err: %v", err)
		return nil
	}
	return p
}

func play(p *audio.Player) {
	if p != nil {
		p.Rewind()
		p.Play()
	}
}

/* Procedural Audio Synthesis */
func synthWave(sr int, dur, freq, amp float64, wave string, freqSweep float64) []int16 {
	n := int(float64(sr) * dur)
	out := make([]int16, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		f := freq + freqSweep*t
		var s float64
		switch wave {
		case "sine":
			s = math.Sin(2 * math.Pi * f * t)
		case "square":
			if math.Sin(2*math.Pi*f*t) >= 0 {
				s = 1
			} else {
				s = -1
			}
		case "noise":
			s = mathrand.NormFloat64()
		default:
			s = math.Sin(2 * math.Pi * f * t)
		}
		// ADSR envelope
		env := 1.0
		att, dec, sus, rel := 0.005, 0.02, 0.6, dur*0.3
		if t < att {
			env = t / att
		} else if t < att+dec {
			env = 1 - (t-att)/dec*(1-sus)
		} else if t > dur-rel {
			env = sus * (dur - t) / rel
		} else {
			env = sus
		}
		val := s * amp * env
		if val > 1 {
			val = 1
		} else if val < -1 {
			val = -1
		}
		out[i] = int16(val * 32767)
	}
	return out
}

func mixToWAV(sr int, tracks [][]int16) []byte {
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
	for _, v := range mix {
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
	}
	scale := 1.0
	if peak > 32767 {
		scale = 32767.0 / float64(peak)
	}

	buf := &bytes.Buffer{}
	dataSize := maxLen * 2
	buf.WriteString("RIFF")
	writeLEUint32(buf, uint32(36+dataSize))
	buf.WriteString("WAVEfmt ")
	writeLEUint32(buf, 16)
	writeLEUint16(buf, 1)
	writeLEUint16(buf, 1)
	writeLEUint32(buf, uint32(sr))
	writeLEUint32(buf, uint32(sr*2))
	writeLEUint16(buf, 2)
	writeLEUint16(buf, 16)
	buf.WriteString("data")
	writeLEUint32(buf, uint32(dataSize))
	for i := 0; i < maxLen; i++ {
		v := int16(float64(mix[i]) * scale)
		_ = binary.Write(buf, binary.LittleEndian, v)
	}
	return buf.Bytes()
}

func writeLEUint16(w io.Writer, v uint16) { _ = binary.Write(w, binary.LittleEndian, v) }
func writeLEUint32(w io.Writer, v uint32) { _ = binary.Write(w, binary.LittleEndian, v) }

func sndEat() []byte {
	sr := 44100
	t1 := synthWave(sr, 0.1, 600, 0.5, "sine", 400)
	t2 := synthWave(sr, 0.1, 1200, 0.3, "sine", 200)
	return mixToWAV(sr, [][]int16{t1, t2})
}

func sndBoom() []byte {
	sr := 44100
	low := synthWave(sr, 0.3, 80, 0.9, "sine", -30)
	n := synthWave(sr, 0.3, 0, 0.7, "noise", 0)
	mid := synthWave(sr, 0.15, 300, 0.4, "square", -200)
	return mixToWAV(sr, [][]int16{low, n, mid})
}

func sndHeal() []byte {
	sr := 44100
	t1 := synthWave(sr, 0.15, 400, 0.4, "sine", 300)
	t2 := synthWave(sr, 0.2, 800, 0.3, "sine", 100)
	return mixToWAV(sr, [][]int16{t1, t2})
}

func sndPause() []byte {
	sr := 44100
	t := synthWave(sr, 0.08, 220, 0.4, "square", -50)
	return mixToWAV(sr, [][]int16{t})
}

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Snake Revived")
	ebiten.SetFullscreen(true)
	g := NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
