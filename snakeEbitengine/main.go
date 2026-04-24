package main

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"io"
	"log"
	"math"
	mathrand "math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
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

type Bomb struct {
	X, Y  int
	Timer float64
}

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

// Типы фруктов
const (
	FRUIT_APPLE = iota
	FRUIT_STRAWBERRY
)

type Game struct {
	rng            *mathrand.Rand
	state          GameState
	snake          []Vec
	dir            Vec
	nextDir        Vec
	ticker         float64
	speed          float64
	fruitX, fruitY int
	fruitType      int
	bombs          []Bomb
	score          int
	health         int
	particles      []Particle
	audioCtx       *audio.Context
	sndEat         *audio.Player
	sndBoom        *audio.Player
	sndHeal        *audio.Player
	sndPause       *audio.Player
	shake          float64
	menuPulse      float64
	pauseCooldown  float64

	menuSelected int
	menuButtons  []string

	fontFace font.Face
}

func NewGame() *Game {
	g := &Game{
		rng:          mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		state:        STATE_MENU,
		speed:        9,
		health:       maxHealth,
		menuPulse:    0,
		menuSelected: 0,
		menuButtons:  []string{"Начать игру", "Продолжить", "Новая игра", "Выйти из игры"},
	}
	g.reset()
	g.audioCtx = audio.NewContext(44100)
	g.sndEat = newSound(g.audioCtx, sndEat())
	g.sndBoom = newSound(g.audioCtx, sndBoom())
	g.sndHeal = newSound(g.audioCtx, sndHeal())
	g.sndPause = newSound(g.audioCtx, sndPause())

	if err := g.loadFont(); err != nil {
		log.Printf("Шрифт не загружен: %v. Русский текст не отобразится.", err)
	}
	return g
}

func (g *Game) loadFont() error {
	data, err := os.ReadFile("font.ttf")
	if err != nil {
		return err
	}
	tt, err := opentype.Parse(data)
	if err != nil {
		return err
	}
	const dpi = 72
	g.fontFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	return err
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
	g.placeFruit()
	g.bombs = nil
	g.score = 0
	g.ticker = 0
	g.shake = 0
	g.particles = nil
}

func (g *Game) placeFruit() {
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
			g.fruitX, g.fruitY = x, y
			// Случайный выбор фрукта: 0 - яблоко, 1 - клубника
			g.fruitType = g.rng.Intn(2)
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

	// ESC открывает меню
	if ebiten.IsKeyPressed(ebiten.KeyEscape) && g.pauseCooldown <= 0 {
		if g.state == STATE_PLAYING || g.state == STATE_PAUSED || g.state == STATE_GAMEOVER {
			g.state = STATE_MENU
			if g.state == STATE_PAUSED {
				g.menuSelected = 1
			} else {
				g.menuSelected = 0
			}
			g.sndPause.Rewind()
			g.sndPause.Play()
		}
		g.pauseCooldown = 0.3
	}

	// Клавиша P – простая пауза
	if ebiten.IsKeyPressed(ebiten.KeyP) && g.pauseCooldown <= 0 && (g.state == STATE_PLAYING || g.state == STATE_PAUSED) {
		if g.state == STATE_PLAYING {
			g.state = STATE_PAUSED
		} else {
			g.state = STATE_PLAYING
		}
		g.sndPause.Rewind()
		g.sndPause.Play()
		g.pauseCooldown = 0.3
	}

	// Меню
	if g.state == STATE_MENU {
		if ebiten.IsKeyPressed(ebiten.KeyUp) && g.pauseCooldown <= 0 {
			g.menuSelected--
			if g.menuSelected < 0 {
				g.menuSelected = len(g.menuButtons) - 1
			}
			g.pauseCooldown = 0.15
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) && g.pauseCooldown <= 0 {
			g.menuSelected++
			if g.menuSelected >= len(g.menuButtons) {
				g.menuSelected = 0
			}
			g.pauseCooldown = 0.15
		}
		if (ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace)) && g.pauseCooldown <= 0 {
			switch g.menuSelected {
			case 0: // Начать игру
				g.reset()
				g.state = STATE_PLAYING
			case 1: // Продолжить
				g.reset()
				g.state = STATE_PLAYING
			case 2: // Новая игра
				g.reset()
				g.state = STATE_PLAYING
			case 3: // Выйти из игры
				return ebiten.Termination
			}
			g.pauseCooldown = 0.3
		}
		return nil
	}

	// Пауза
	if g.state == STATE_PAUSED {
		return nil
	}

	// GAME OVER – любая клавиша в меню
	if g.state == STATE_GAMEOVER {
		if inputPressed() && g.pauseCooldown <= 0 {
			g.state = STATE_MENU
			g.menuSelected = 0
			g.pauseCooldown = 0.2
		}
		return nil
	}

	// Управление змейкой
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

	// Обновление таймеров бомб
	for i := 0; i < len(g.bombs); i++ {
		g.bombs[i].Timer -= dt
		if g.bombs[i].Timer <= 0 {
			g.bombExplode(i)
			i--
		}
	}

	// Частицы
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
	return nil
}

// ---------- Игровая логика ----------
func (g *Game) step() {
	head := g.snake[0]
	newHead := Vec{head.X + g.dir.X, head.Y + g.dir.Y}

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

	// Съедание фрукта
	if newHead.X == g.fruitX && newHead.Y == g.fruitY {
		if g.fruitType == FRUIT_APPLE {
			g.score++
			g.health = minInt(maxHealth, g.health+25)
		} else { // клубника
			g.score += 2
			g.health = minInt(maxHealth, g.health+40)
		}
		g.placeFruit()
		g.spawnBombRandom()
		g.sndHeal.Rewind()
		g.sndHeal.Play()
		g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 25, color.RGBA{50, 255, 80, 255}, true)
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}

	// Проверка столкновения с бомбами
	for i := 0; i < len(g.bombs); i++ {
		if g.bombs[i].X == newHead.X && g.bombs[i].Y == newHead.Y {
			g.health -= 35
			g.triggerExplosion(newHead, g.health <= 0)
			g.bombs = append(g.bombs[:i], g.bombs[i+1:]...)
			return
		}
	}
	g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 2, color.RGBA{0, 180, 220, 140}, false)
}

func (g *Game) triggerExplosion(v Vec, fatal bool) {
	if fatal {
		g.health = 0
	}
	g.shake = 18
	g.sndBoom.Rewind()
	g.sndBoom.Play()
	g.addParticles(float64(v.X*tileSize+tileSize/2), float64(v.Y*tileSize+tileSize/2), 80, color.RGBA{255, 120, 30, 255}, true)
	g.addParticles(float64(v.X*tileSize+tileSize/2), float64(v.Y*tileSize+tileSize/2), 40, color.RGBA{255, 255, 200, 200}, true)
	if fatal {
		g.state = STATE_GAMEOVER
	}
}

func (g *Game) bombExplode(idx int) {
	b := g.bombs[idx]
	g.bombs = append(g.bombs[:idx], g.bombs[idx+1:]...)

	// Проверяем урон змейке (если голова близко)
	head := g.snake[0]
	dx := math.Abs(float64(head.X - b.X))
	dy := math.Abs(float64(head.Y - b.Y))
	distance := dx + dy

	g.shake = 12
	g.sndBoom.Rewind()
	g.sndBoom.Play()

	cx := float64(b.X*tileSize + tileSize/2)
	cy := float64(b.Y*tileSize + tileSize/2)

	if distance <= 1.5 {
		g.health -= 25
		g.addParticles(cx, cy, 120, color.RGBA{255, 60, 30, 255}, true)
		g.addParticles(cx, cy, 60, color.RGBA{255, 200, 50, 200}, true)
		if g.health <= 0 {
			g.state = STATE_GAMEOVER
		}
	} else {
		g.addParticles(cx, cy, 80, color.RGBA{255, 100, 30, 255}, true)
	}
	// Дополнительные искры и дым
	g.addParticles(cx, cy, 30, color.RGBA{255, 200, 100, 200}, false)
}

func (g *Game) spawnBombRandom() {
	if g.rng.Float64() < 0.4 {
		for i := 0; i < 2000; i++ {
			x := g.rng.Intn(gridW)
			y := g.rng.Intn(gridH)
			ok := true
			if x == g.fruitX && y == g.fruitY {
				ok = false
			}
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
				g.bombs = append(g.bombs, Bomb{X: x, Y: y, Timer: 5.0})
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

// ---------- Отрисовка ----------
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 12, 20, 255})

	ox, oy := 0.0, 0.0
	if g.shake > 0.5 {
		ox = (mathrand.Float64()*2 - 1) * g.shake
		oy = (mathrand.Float64()*2 - 1) * g.shake
	}

	// Сетка
	for x := 0; x < gridW; x++ {
		for y := 0; y < gridH; y++ {
			c := color.RGBA{15, 15, 25, 255}
			if (x+y)%2 != 0 {
				c = color.RGBA{18, 18, 30, 255}
			}
			ebitenutil.DrawRect(screen, float64(x*tileSize)+ox, float64(y*tileSize)+oy, tileSize-1, tileSize-1, c)
		}
	}

	// Бомбы – увеличенные, пульсирующие
	for _, b := range g.bombs {
		cx := float64(b.X*tileSize+tileSize/2) + ox
		cy := float64(b.Y*tileSize+tileSize/2) + oy
		baseRadius := float64(tileSize) / 2 * 1.5
		t := b.Timer
		freq := 3.0 + 9.0*(1.0-math.Min(1.0, t/5.0))
		pulse := 1.0 + 0.15*math.Sin(g.menuPulse*20*freq)
		radius := baseRadius * pulse
		r := uint8(20)
		if t < 2.0 {
			r = uint8(80 + int(175*(1.0-t/2.0)))
		}
		ebitenutil.DrawCircle(screen, cx, cy, radius, color.RGBA{r, 20, 25, 255})
		ebitenutil.DrawCircle(screen, cx-2, cy-2, radius-2, color.RGBA{0, 0, 0, 100})
		ebitenutil.DrawCircle(screen, cx-radius*0.3, cy-radius*0.35, radius*0.25, color.RGBA{255, 255, 255, 180})
		ebitenutil.DrawCircle(screen, cx+radius*0.2, cy+radius*0.2, radius*0.2, color.RGBA{255, 80, 80, 120})
	}

	// Змейка (без изменений)
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
			eyex := float64(tileSize)/4 - 2
			eyey := float64(tileSize)/4 - 2
			ebitenutil.DrawRect(screen, x+eyex, y+eyey, 4, 4, color.White)
			ebitenutil.DrawRect(screen, x+float64(tileSize)-eyex-6, y+eyey, 4, 4, color.White)
			ebitenutil.DrawRect(screen, x+eyex+1, y+eyey+1, 2, 2, color.Black)
			ebitenutil.DrawRect(screen, x+float64(tileSize)-eyex-5, y+eyey+1, 2, 2, color.Black)

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

	// Фрукт (яблоко или клубника) – увеличен в 1.5 раза
	{
		cx := float64(g.fruitX*tileSize+tileSize/2) + ox
		cy := float64(g.fruitY*tileSize+tileSize/2) + oy
		baseRadius := (float64(tileSize)/2 - 2) * 1.5
		radius := baseRadius

		// Тень
		ebitenutil.DrawCircle(screen, cx-2, cy-2, radius-2, color.RGBA{0, 0, 0, 80})

		if g.fruitType == FRUIT_APPLE {
			// Яблоко
			ebitenutil.DrawCircle(screen, cx, cy, radius, color.RGBA{230, 40, 50, 255})
			ebitenutil.DrawCircle(screen, cx-3, cy-3, radius-4, color.RGBA{255, 100, 100, 150})
			ebitenutil.DrawCircle(screen, cx-radius*0.3, cy-radius*0.35, radius*0.2, color.RGBA{255, 255, 255, 220})
			// Хвостик
			ebitenutil.DrawRect(screen, cx+radius*0.5, cy-radius*0.8, 6, 3, color.RGBA{70, 180, 50, 255})
			ebitenutil.DrawRect(screen, cx+radius*0.7, cy-radius*0.9, 8, 2, color.RGBA{50, 150, 30, 255})
		} else {
			// Клубника
			ebitenutil.DrawCircle(screen, cx, cy, radius, color.RGBA{220, 30, 40, 255})
			ebitenutil.DrawCircle(screen, cx-2, cy-2, radius-3, color.RGBA{245, 80, 90, 200})
			// Семечки (жёлтые точки)
			for _, angle := range []float64{0, 1.2, 2.5, 3.8, 5.0} {
				sx := cx + math.Cos(angle)*radius*0.6
				sy := cy + math.Sin(angle)*radius*0.6
				ebitenutil.DrawRect(screen, sx-1, sy-1, 2, 2, color.RGBA{255, 220, 80, 255})
			}
			// Зелёные листочки
			ebitenutil.DrawRect(screen, cx-radius*0.2, cy-radius*0.7, 8, 4, color.RGBA{40, 160, 30, 255})
			ebitenutil.DrawRect(screen, cx+radius*0.2, cy-radius*0.75, 8, 4, color.RGBA{50, 170, 40, 255})
			ebitenutil.DrawCircle(screen, cx, cy-radius*0.65, radius*0.15, color.RGBA{30, 140, 20, 255})
		}
	}

	// Частицы
	for _, p := range g.particles {
		c := p.Color
		if p.Glow {
			ebitenutil.DrawRect(screen, p.X-p.Size*1.5+ox, p.Y-p.Size*1.5+oy, p.Size*3, p.Size*3, color.RGBA{c.R, c.G, c.B, uint8(float64(c.A) * 0.4 * p.Life)})
		}
		ebitenutil.DrawRect(screen, p.X-p.Size+ox, p.Y-p.Size+oy, p.Size*2, p.Size*2, c)
	}

	// Текст
	drawText := func(str string, x, y int, clr color.Color) {
		if g.fontFace != nil {
			text.Draw(screen, str, g.fontFace, x, y, clr)
		} else {
			ebitenutil.DebugPrintAt(screen, str, x, y)
		}
	}

	drawText("Счёт: "+strconv.Itoa(g.score), 10, 25, color.White)
	barX := float64(screenW - 20)
	barW := 150.0
	barH := 14.0
	healthPct := float64(g.health) / float64(maxHealth)
	ebitenutil.DrawRect(screen, barX-barW, 10, barW, barH, color.RGBA{30, 30, 40, 200})
	ebitenutil.DrawRect(screen, barX-barW, 10, barW*healthPct, barH, color.RGBA{50, 255, 80, 255})
	drawText("ЗДОРОВЬЕ", int(barX-barW+40), 25, color.White)
	drawText("ESC - меню", screenW-100, screenH-20, color.White)
	drawText("P - пауза", screenW-100, screenH-40, color.White)

	// Меню, пауза, game over
	switch g.state {
	case STATE_MENU:
		// Полностью непрозрачный фон
		ebitenutil.DrawRect(screen, 0, 0, screenW, screenH, color.RGBA{0, 0, 0, 255})
		drawText("S N A K E   R E V I V E D", screenW/2-180, 150, color.RGBA{255, 200, 100, 255})
		startY := 280
		step := 50
		for i, btn := range g.menuButtons {
			y := startY + i*step
			if i == g.menuSelected {
				ebitenutil.DrawRect(screen, screenW/2-150, float64(y)-15, 300, 35, color.RGBA{100, 100, 150, 255})
				drawText("→ "+btn, screenW/2-len(btn)*3, y, color.RGBA{255, 255, 0, 255})
			} else {
				drawText("  "+btn, screenW/2-len(btn)*3, y, color.White)
			}
		}
		drawText("Стрелки вверх/вниз, Enter - выбор", screenW/2-220, screenH-70, color.RGBA{200, 200, 200, 255})
	case STATE_PAUSED:
		ebitenutil.DrawRect(screen, 0, 0, screenW, screenH, color.RGBA{0, 0, 0, 200})
		drawText("ПАУЗА", screenW/2-40, screenH/2, color.RGBA{255, 255, 150, 255})
		drawText("Нажмите P для продолжения", screenW/2-150, screenH/2+40, color.White)
	case STATE_GAMEOVER:
		ebitenutil.DrawRect(screen, 100, 80, screenW-200, screenH-180, color.RGBA{40, 0, 0, 255})
		drawText("ИГРА ОКОНЧЕНА", screenW/2-80, screenH/2-40, color.RGBA{255, 100, 100, 255})
		drawText("Счёт: "+strconv.Itoa(g.score), screenW/2-60, screenH/2, color.White)
		drawText("Нажмите любую клавишу для меню", screenW/2-150, screenH/2+40, color.White)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}

// ---------- Вспомогательные функции ----------
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

// ---------- Аудио (без изменений) ----------
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
		att, dec, sus, rel := 0.005, 0.02, 0.6, dur*0.3
		env := 1.0
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
	ebiten.SetWindowTitle("Змейка: Возрождение")
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
