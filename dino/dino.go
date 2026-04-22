// dino.go
// Dino-style endless runner in terminal using termbox-go and faiface/beep.
// Russian UI: стартовое меню с выбором дня/ночи, визуальные элементы, синтезированные звуки.
// Controls: Space — прыжок, P — пауза, Q — выход. Монетки собираются при контакте.
package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/nsf/termbox-go"
)

const (
	fps           = 30
	groundYOffset = 3
	gravity       = 1
	jumpVel       = -9
	maxObstacles  = 4
	coinSpawnProb = 0.02
	sampleRate    = 44100
)

type GameState int

const (
	Menu GameState = iota
	Playing
	Paused
	Quit
)

type TimeOfDay int

const (
	Day TimeOfDay = iota
	Night
)

type Dino struct {
	x, y        int
	vy          int
	onGround    bool
	runFrame    int
	invincible  int
	coinCount   int
}

type Obstacle struct {
	x int
	t int // type: 0 small cactus, 1 tall cactus
}

type Cloud struct {
	x     float64
	y     int
	speed float64
}

type Coin struct {
	x         int
	y         int
	collected bool
}

var (
	state     = Menu
	dayNight  = Day
	width     int
	height    int
	groundY   int
	dino      Dino
	obstacles []Obstacle
	clouds    []Cloud
	coins     []Coin
	score     int
	speed     = 1
	tickDelay = time.Second / fps
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := termbox.Init(); err != nil {
		log.Fatalf("termbox init: %v", err)
	}
	defer termbox.Close()

	// audio init (non-fatal)
	sr := beep.SampleRate(sampleRate)
	if err := speaker.Init(sr, sr.N(time.Second/10)); err != nil {
		log.Println("audio init failed:", err)
	}

	resetWorld()
	drawMenu()

	// event handling
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	ticker := time.NewTicker(tickDelay)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-sigc:
			break loop
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				handleKey(ev)
			}
			if ev.Type == termbox.EventResize {
				width, height = termbox.Size()
				groundY = height - groundYOffset
			}
		case <-ticker.C:
			switch state {
			case Menu:
				drawMenu()
			case Playing:
				update()
				draw()
			case Paused:
				drawPaused()
			case Quit:
				break loop
			}
		}
	}

	termbox.Close()
	fmt.Println("До встречи — спасибо за игру!")
}

func resetWorld() {
	width, height = termbox.Size()
	groundY = height - groundYOffset
	dino = Dino{x: 6, y: groundY - 3, vy: 0, onGround: true, runFrame: 0, coinCount: 0}
	obstacles = []Obstacle{}
	coins = []Coin{}
	clouds = []Cloud{}
	// init clouds
	for i := 0; i < 6; i++ {
		clouds = append(clouds, Cloud{
			x:     float64(rand.Intn(width)),
			y:     rand.Intn(groundY/2) + 1,
			speed: 0.1 + rand.Float64()*0.4,
		})
	}
	score = 0
	speed = 1
}

func handleKey(ev termbox.Event) {
	switch state {
	case Menu:
		if ev.Ch == '1' {
			dayNight = Day
			state = Playing
			resetWorld()
		}
		if ev.Ch == '2' {
			dayNight = Night
			state = Playing
			resetWorld()
		}
		if ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC {
			state = Quit
		}
	case Playing:
		if ev.Ch == ' ' || ev.Key == termbox.KeySpace {
			if dino.onGround {
				dino.vy = jumpVel
				dino.onGround = false
				playTone(880, 120*time.Millisecond, 0.6)
			}
		}
		if ev.Ch == 'p' || ev.Ch == 'P' {
			state = Paused
		}
		if ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC {
			state = Quit
		}
	case Paused:
		if ev.Ch == 'p' || ev.Ch == 'P' {
			state = Playing
		}
		if ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC {
			state = Quit
		}
	}
}

func update() {
	// update clouds
	for i := range clouds {
		clouds[i].x += clouds[i].speed
		if clouds[i].x > float64(width)+10 {
			clouds[i].x = -10
			clouds[i].y = rand.Intn(groundY/2) + 1
			clouds[i].speed = 0.1 + rand.Float64()*0.5
		}
	}

	// spawn obstacles
	if len(obstacles) < maxObstacles && rand.Float64() < 0.02+float64(score)/2000.0 {
		t := 0
		if rand.Float64() < 0.4 {
			t = 1
		}
		obstacles = append(obstacles, Obstacle{x: width - 2, t: t})
	}

	// spawn coins randomly in mid-air
	if rand.Float64() < coinSpawnProb {
		coins = append(coins, Coin{x: width - 2, y: groundY - 6, collected: false})
	}

	// move obstacles & coins
	for i := range obstacles {
		obstacles[i].x -= speed
	}
	// remove off-screen obstacles
	filtered := obstacles[:0]
	for _, o := range obstacles {
		if o.x > -5 {
			filtered = append(filtered, o)
		}
	}
	obstacles = filtered

	for i := range coins {
		if !coins[i].collected {
			coins[i].x -= speed
		}
	}
	cf := coins[:0]
	for _, c := range coins {
		if c.x > -2 && !c.collected {
			cf = append(cf, c)
		}
	}
	coins = cf

	// update dino physics
	if !dino.onGround {
		dino.vy += gravity
		dino.y += dino.vy
		if dino.y >= groundY-3 {
			dino.y = groundY - 3
			dino.vy = 0
			dino.onGround = true
		}
	} else {
		// running animation frame
		dino.runFrame = (dino.runFrame + 1) % 6
	}
	// check collisions with obstacles
	for _, o := range obstacles {
		if collideDinoObstacle(o) {
			// collision -> end game (simple)
			playTone(150, 300*time.Millisecond, 0.9)
			state = Menu
			// show score briefly then return to menu
			drawGameOver()
			time.Sleep(800 * time.Millisecond)
			return
		}
	}
	// check coins
	for i := range coins {
		if !coins[i].collected && collideDinoCoin(coins[i]) {
			coins[i].collected = true
			dino.coinCount++
			playTone(1200, 80*time.Millisecond, 0.7)
		}
	}

	// increase difficulty slowly
	if score%200 == 0 && score > 0 {
		speed = 1 + score/400
	}
	score++
}

func collideDinoObstacle(o Obstacle) bool {
	// approximate bounding boxes
	dx := o.x - dino.x
	if dx < -3 || dx > 6 {
		return false
	}
	// obstacle height
	oh := 2
	if o.t == 1 {
		oh = 4
	}
	oy := groundY - oh
	// dino top y
	dt := dino.y
	if dt < oy {
		return false
	}
	// horizontal overlap
	if dino.x+3 >= o.x && dino.x <= o.x+2 {
		return true
	}
	return false
}

func collideDinoCoin(c Coin) bool {
	if c.x >= dino.x && c.x <= dino.x+4 {
		if c.y >= dino.y && c.y <= dino.y+3 {
			return true
		}
	}
	return false
}

func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	// sky background
	drawSky()
	// clouds
	if dayNight == Day {
		for _, cl := range clouds {
			drawCloud(int(cl.x), cl.y)
		}
	}
	// ground line
	for x := 0; x < width; x++ {
		termbox.SetCell(x, groundY+1, '─', termbox.ColorYellow, termbox.ColorDefault)
		if x%2 == 0 {
			termbox.SetCell(x, groundY+2, '_', termbox.ColorGreen, termbox.ColorDefault)
		}
	}
	// obstacles (cacti)
	for _, o := range obstacles {
		drawCactus(o.x, groundY, o.t)
	}
	// coins
	for _, c := range coins {
		if !c.collected {
			drawCoin(c.x, c.y)
		}
	}
	// dino
	drawDino(dino.x, dino.y, dino.runFrame, !dino.onGround)
	// HUD
	hud := fmt.Sprintf("Монетки: %d  Скорость: %d  Очки: %d", dino.coinCount, speed, score)
	for i, ch := range hud {
		termbox.SetCell(2+i, 1, ch, termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.Flush()
}

func drawMenu() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	width, height = termbox.Size()
	putCentered(2, "┏━━━━━━━━━━━━━━ DINO RUN ━━━━━━━━━━━━━━┓")
	putCentered(5, "Добро пожаловать! Выберите время суток:")
	putCentered(7, "1 — День (солнце и облака)")
	putCentered(8, "2 — Ночь (луна и звёзды)")
	putCentered(10, "Прыжок — <Space>   Пауза — P   Выход — Q")
	putCentered(12, "Нажмите 1 или 2 чтобы начать...")
	putCentered(height-3, "Автор: Ваш Терминал — Удачи!")
	// draw example sun/moon left-top
	if dayNight == Day {
		drawSun(8, 4)
		for _, cl := range clouds {
			drawCloud(int(cl.x)%width, cl.y)
		}
	} else {
		drawMoon(8, 4)
		drawStars()
	}
	termbox.Flush()
}

func drawPaused() {
	putCentered(height/2, "Пауза — нажмите 'P' чтобы продолжить")
	termbox.Flush()
}

func drawGameOver() {
	putCentered(height/2, "Упс! Вы врезались! Возврат в меню...")
	termbox.Flush()
}

func drawSky() {
	bg := termbox.ColorBlack
	fg := termbox.ColorBlue
	if dayNight == Day {
		bg = termbox.ColorCyan
		fg = termbox.ColorWhite
	} else {
		bg = termbox.ColorBlack
		fg = termbox.ColorWhite
	}
	for y := 0; y <= groundY; y++ {
		for x := 0; x < width; x++ {
			termbox.SetCell(x, y, ' ', fg, bg)
		}
	}
	// sun/moon
	if dayNight == Day {
		drawSun(6, 3)
	} else {
		drawMoon(6, 3)
		drawStars()
	}
}

func drawSun(cx, cy int) {
	// core
	termbox.SetCell(cx, cy, '☼', termbox.ColorYellow|termbox.AttrBold, termbox.ColorCyan)
	// rays
	rays := []struct{ x, y int }{
		{-2, 0}, {2, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1},
	}
	for _, r := range rays {
		if cx+r.x >= 0 && cx+r.x < width && cy+r.y >= 0 && cy+r.y < height {
			termbox.SetCell(cx+r.x, cy+r.y, '*', termbox.ColorYellow, termbox.ColorCyan)
		}
	}
}

func drawMoon(cx, cy int) {
	if cx >= 0 && cx < width && cy >= 0 && cy < height {
		termbox.SetCell(cx, cy, '☾', termbox.ColorWhite, termbox.ColorBlack)
	}
}

func drawStars() {
	for i := 0; i < 25; i++ {
		x := rand.Intn(width-1)
		y := rand.Intn(groundY/2) + 1
		termbox.SetCell(x, y, '✦', termbox.ColorWhite, termbox.ColorBlack)
	}
}

func drawCloud(x, y int) {
	// a simple cloud glyph cluster
	chars := []struct{ dx int; ch rune }{
		{-2, '░'}, {-1, '▒'}, {0, '▒'}, {1, '░'}, {2, ' '},
	}
	for _, c := range chars {
		xx := x + c.dx
		if xx >= 0 && xx < width && y >= 0 && y < groundY {
			termbox.SetCell(xx, y, c.ch, termbox.ColorWhite, termbox.ColorCyan)
		}
	}
}

func drawCactus(x, groundY, t int) {
	// simple ASCII cactus with vertical segments
	h := 2
	if t == 1 {
		h = 4
	}
	for i := 0; i < h; i++ {
		yy := groundY - i
		if x >= 0 && x < width && yy >= 0 && yy < height {
			termbox.SetCell(x, yy, '♣', termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault)
		}
		// small arms
		if i == 1 && x+1 < width && yy >= 0 && yy < height {
			termbox.SetCell(x+1, yy, '╭', termbox.ColorGreen, termbox.ColorDefault)
		}
	}
}

func drawCoin(x, y int) {
	if x >= 0 && x < width && y >= 0 && y < groundY {
		termbox.SetCell(x, y, '¤', termbox.ColorYellow, termbox.ColorDefault)
	}
}

func drawDino(x, y, frame int, jumping bool) {
	// multi-line dino ~ 4x4 block, stylized
	// simple body/leg variants
	var leg string
	if jumping {
		leg = " /\\ "
	} else {
		if frame%2 == 0 {
			leg = "/  \\"
		} else {
			leg = "\\  /"
		}
	}
	sprite := []string{
		"  __",
		" /  \\",
		" |()|",
		leg,
	}
	// draw sprite
	for ry, line := range sprite {
		for rx, ch := range line {
			xx := x + rx
			yy := y + ry
			if xx >= 0 && xx < width && yy >= 0 && yy < height {
				termbox.SetCell(xx, yy, ch, termbox.ColorWhite|termbox.AttrBold, termbox.ColorDefault)
			}
		}
	}
	// small eye as accent
	if x+2 >= 0 && x+2 < width && y+1 >= 0 && y+1 < height {
		termbox.SetCell(x+2, y+1, '•', termbox.ColorMagenta, termbox.ColorDefault)
	}
}

func putCentered(y int, s string) {
	width, _ = termbox.Size()
	x := (width - len(s)) / 2
	for i, ch := range s {
		termbox.SetCell(x+i, y, ch, termbox.ColorWhite, termbox.ColorDefault)
	}
}

// AUDIO: simple sine generator streamer
type Sine struct {
	freq   float64
	amp    float64
	pos    int
	length int
}

func NewSine(freq float64, dur time.Duration, amp float64) *Sine {
	length := int(float64(sampleRate) * dur.Seconds())
	return &Sine{freq: freq, amp: amp, pos: 0, length: length}
}

func (s *Sine) Stream(samples [][2]float64) (n int, ok bool) {
	for i := range samples {
		if s.pos >= s.length {
			return i, false
		}
		t := float64(s.pos) / float64(sampleRate)
		v := s.amp * math.Sin(2*math.Pi*s.freq*t) * envelope(s.pos, s.length)
		samples[i][0] = v
		samples[i][1] = v
		s.pos++
		n = i + 1
	}
	return n, true
}

func (s *Sine) Err() error { return nil }

func envelope(pos, total int) float64 {
	// simple linear attack/decay
	a := float64(pos) / float64(total)
	if a < 0.1 {
		return a / 0.1
	}
	if a > 0.8 {
		return (1 - a) / 0.2
	}
	return 1
}

func playTone(freq float64, dur time.Duration, amp float64) {
	stream := NewSine(freq, dur, amp)
	done := make(chan struct{})
	speaker.Play(beep.Seq(stream, beep.Callback(func() { close(done) })))
	<-done
}
