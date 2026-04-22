// main.go
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
	fps         = 30
	paddleLen   = 12
	paddleSpeed = 2
	sampleRate  = 44100
)

type Paddle struct {
	x int
	y int
	w int
}

type Ball struct {
	x, y   int
	vx, vy int
}

type GameState int

const (
	Menu GameState = iota
	Playing
	Paused
	Quit
)

var (
	topPaddle    Paddle
	bottomPaddle Paddle
	ball         Ball
	scoreTop     int
	scoreBottom  int
	state        GameState = Menu
	width        int
	height       int
	tickDelay             = time.Second / fps
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	// init audio speaker
	sr := beep.SampleRate(sampleRate)
	if err := speaker.Init(sr, sr.N(time.Second/10)); err != nil {
		// if audio init fails, continue without crashing — game still playable
		log.Println("audio init failed:", err)
	}

	resetGame()
	drawMenu()

	// event loop
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
			}
		case <-ticker.C:
			switch state {
			case Menu:
				drawMenu()
			case Playing:
				update()
				draw()
			case Paused:
				draw()
			case Quit:
				break loop
			}
		}
	}
	termbox.Close()
	fmt.Println("Выход. Спасибо за игру!")
}

func resetGame() {
	width, height = termbox.Size()
	topPaddle = Paddle{x: (width - paddleLen) / 2, y: 1, w: paddleLen}
	bottomPaddle = Paddle{x: (width - paddleLen) / 2, y: height - 2, w: paddleLen}
	ball = Ball{x: width / 2, y: height / 2, vx: randChoice(-1, 1), vy: randChoice(-1, 1)}
	scoreTop = 0
	scoreBottom = 0
}

func randChoice(a, b int) int {
	if rand.Intn(2) == 0 {
		return a
	}
	return b
}

func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	// borders
	for x := 0; x < width; x++ {
		termbox.SetCell(x, 0, '═', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(x, height-1, '═', termbox.ColorWhite, termbox.ColorDefault)
	}
	// paddles (horizontal)
	for i := 0; i < topPaddle.w; i++ {
		termbox.SetCell(clampX(topPaddle.x+i), topPaddle.y, '█', termbox.ColorGreen, termbox.ColorDefault)
	}
	for i := 0; i < bottomPaddle.w; i++ {
		termbox.SetCell(clampX(bottomPaddle.x+i), bottomPaddle.y, '█', termbox.ColorGreen, termbox.ColorDefault)
	}
	// ball
	if ball.y > 0 && ball.y < height {
		termbox.SetCell(clampX(ball.x), clampY(ball.y), '●', termbox.ColorYellow, termbox.ColorDefault)
	}
	// scores
	scoreStr := fmt.Sprintf("Верх: %d    Низ: %d", scoreTop, scoreBottom)
	for i, ch := range scoreStr {
		termbox.SetCell(2+i, 0, ch, termbox.ColorCyan, termbox.ColorDefault)
	}
	if state == Paused {
		putCentered(height/2, "Пауза — нажмите 'p' чтобы продолжить")
	}
	termbox.Flush()
}

func clampX(x int) int {
	if x < 0 {
		return 0
	}
	if x >= width {
		return width - 1
	}
	return x
}

func clampY(y int) int {
	if y < 0 {
		return 0
	}
	if y >= height {
		return height - 1
	}
	return y
}

func putCentered(y int, s string) {
	x := (width - len(s)) / 2
	for i, ch := range s {
		termbox.SetCell(x+i, y, ch, termbox.ColorMagenta, termbox.ColorDefault)
	}
}

func update() {
	ball.x += ball.vx
	ball.y += ball.vy

	// walls
	if ball.x <= 1 || ball.x >= width-2 {
		ball.vx = -ball.vx
		playTone(440, 60*time.Millisecond, 0.5)
	}

	// top paddle collision
	if ball.y == topPaddle.y+1 {
		if ball.x >= topPaddle.x && ball.x <= topPaddle.x+topPaddle.w {
			ball.vy = -ball.vy
			offset := ball.x - (topPaddle.x + topPaddle.w/2)
			if offset != 0 {
				ball.vx += sign(offset)
			}
			playTone(880, 40*time.Millisecond, 0.6)
		}
	}

	// bottom paddle collision
	if ball.y == bottomPaddle.y-1 {
		if ball.x >= bottomPaddle.x && ball.x <= bottomPaddle.x+bottomPaddle.w {
			ball.vy = -ball.vy
			offset := ball.x - (bottomPaddle.x + bottomPaddle.w/2)
			if offset != 0 {
				ball.vx += sign(offset)
			}
			playTone(880, 40*time.Millisecond, 0.6)
		}
	}

	// scoring
	if ball.y <= 0 {
		scoreBottom++
		playTone(220, 150*time.Millisecond, 0.9)
		resetPositionsAfterScore()
	}
	if ball.y >= height-1 {
		scoreTop++
		playTone(220, 150*time.Millisecond, 0.9)
		resetPositionsAfterScore()
	}

	// clamp vx
	if ball.vx > 3 {
		ball.vx = 3
	}
	if ball.vx < -3 {
		ball.vx = -3
	}
}

func resetPositionsAfterScore() {
	width, height = termbox.Size()
	ball.x = width / 2
	ball.y = height / 2
	ball.vx = randChoice(-1, 1)
	ball.vy = randChoice(-1, 1)
	topPaddle.x = (width - topPaddle.w) / 2
	bottomPaddle.x = (width - bottomPaddle.w) / 2
}

func sign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

func handleKey(ev termbox.Event) {
	switch state {
	case Menu:
		if ev.Ch == '1' {
			state = Playing
			resetGame()
		}
		if ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC {
			state = Quit
		}
	case Playing:
		if ev.Key == termbox.KeyArrowLeft {
			bottomPaddle.x -= paddleSpeed
		}
		if ev.Key == termbox.KeyArrowRight {
			bottomPaddle.x += paddleSpeed
		}
		if ev.Ch == 'a' || ev.Ch == 'A' {
			topPaddle.x -= paddleSpeed
		}
		if ev.Ch == 'd' || ev.Ch == 'D' {
			topPaddle.x += paddleSpeed
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
	// clamp paddles
	if topPaddle.x < 1 {
		topPaddle.x = 1
	}
	if topPaddle.x+topPaddle.w > width-1 {
		topPaddle.x = width - 1 - topPaddle.w
	}
	if bottomPaddle.x < 1 {
		bottomPaddle.x = 1
	}
	if bottomPaddle.x+bottomPaddle.w > width-1 {
		bottomPaddle.x = width - 1 - bottomPaddle.w
	}
}

func drawMenu() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	putCentered(3, "┏━━━━━━━━━━━━━━ PONG ━━━━━━━━━━━━━━┓")
	putCentered(6, "Добро пожаловать в ретро-Pong!")
	putCentered(8, "Управление:")
	putCentered(9, "Верхняя ракетка: A (влево), D (вправо)")
	putCentered(10, "Нижняя ракетка: ← (влево), → (вправо)")
	putCentered(12, "1 - Начать игру")
	putCentered(13, "P - Пауза (во время игры)")
	putCentered(14, "Q - Выход")
	putCentered(16, "Нажмите '1' чтобы начать...")
	termbox.Flush()
}

// AUDIO: generate simple sine wave streamer (mono -> stereo)
type Sine struct {
	freq     float64
	duration time.Duration
	amp      float64
	sr       beep.SampleRate
	pos      int
	total    int
}

func NewSine(freq float64, dur time.Duration, amp float64) *Sine {
	total := int(float64(sampleRate) * dur.Seconds())
	return &Sine{freq: freq, duration: dur, amp: amp, sr: beep.SampleRate(sampleRate), pos: 0, total: total}
}

func (s *Sine) Stream(samples [][2]float64) (n int, ok bool) {
	for i := range samples {
		if s.pos >= s.total {
			return i, false
		}
		t := float64(s.pos) / float64(sampleRate)
		v := s.amp * math.Sin(2*math.Pi*s.freq*t)
		// stereo same on both channels
		samples[i][0] = v
		samples[i][1] = v
		s.pos++
		n = i + 1
	}
	return n, true
}

func (s *Sine) Err() error { return nil }

// playTone - non-blocking play of generated tone, but wait until done to keep timing safe
func playTone(freq float64, dur time.Duration, amp float64) {
	stream := NewSine(freq, dur, amp)
	done := make(chan struct{})
	speaker.Play(beep.Seq(stream, beep.Callback(func() { done <- struct{}{} })))
	// wait in a goroutine to avoid blocking game loop; but since sound is short it's fine to block briefly
	<-done
}
