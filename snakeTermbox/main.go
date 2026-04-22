// main.go
package main

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/nsf/termbox-go"
)

type Point struct{ X, Y int }

const (
	width  = 40
	height = 20
)

var (
	snake        []Point
	dir          = Point{1, 0}
	apple        Point
	bombs        map[Point]bool
	alive        = true
	score        = 0
	playCmdsEat  = [][]string{{"afplay", "eat.wav"}, {"play", "eat.wav"}, {"aplay", "eat.wav"}}
	playCmdsHurt = [][]string{{"afplay", "hit.wav"}, {"play", "hit.wav"}, {"aplay", "hit.wav"}}
	sampleRate   = 44100
	bitsPerSample = 16
)

func writeWavMono16(filename string, samples []int16) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	numSamples := uint32(len(samples))
	subchunk1Size := uint32(16)
	audioFormat := uint16(1)
	numChannels := uint16(1)
	byteRate := uint32(sampleRate * int(numChannels) * bitsPerSample / 8)
	blockAlign := uint16(numChannels * uint16(bitsPerSample/8))
	subchunk2Size := numSamples * uint32(bitsPerSample/8)
	chunkSize := 4 + (8 + subchunk1Size) + (8 + subchunk2Size)

	binary.Write(f, binary.LittleEndian, []byte("RIFF"))
	binary.Write(f, binary.LittleEndian, chunkSize)
	binary.Write(f, binary.LittleEndian, []byte("WAVE"))

	binary.Write(f, binary.LittleEndian, []byte("fmt "))
	binary.Write(f, binary.LittleEndian, subchunk1Size)
	binary.Write(f, binary.LittleEndian, audioFormat)
	binary.Write(f, binary.LittleEndian, numChannels)
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))
	binary.Write(f, binary.LittleEndian, byteRate)
	binary.Write(f, binary.LittleEndian, blockAlign)
	binary.Write(f, binary.LittleEndian, uint16(bitsPerSample))

	binary.Write(f, binary.LittleEndian, []byte("data"))
	binary.Write(f, binary.LittleEndian, subchunk2Size)
	for _, s := range samples {
		binary.Write(f, binary.LittleEndian, s)
	}
	return nil
}

func synthSine(freq float64, durMs int, vol float64, decay bool) []int16 {
	ns := sampleRate * durMs / 1000
	out := make([]int16, ns)
	for i := 0; i < ns; i++ {
		t := float64(i) / float64(sampleRate)
		env := 1.0
		if decay {
			env = math.Exp(-3.0 * t)
		}
		s := math.Sin(2*math.Pi*freq*t) * env * vol
		if s > 1 {
			s = 1
		}
		if s < -1 {
			s = -1
		}
		out[i] = int16(s * 32767)
	}
	return out
}

func synthBuzzy(freq float64, durMs int, vol float64) []int16 {
	ns := sampleRate * durMs / 1000
	out := make([]int16, ns)
	for i := 0; i < ns; i++ {
		t := float64(i) / float64(sampleRate)
		env := math.Exp(-3.0 * t)
		s := (0.7*math.Sin(2*math.Pi*freq*t) +
			0.2*math.Sin(2*math.Pi*2*freq*t) +
			0.1*math.Sin(2*math.Pi*3*freq*t)) * env * vol
		if s > 1 {
			s = 1
		}
		if s < -1 {
			s = -1
		}
		out[i] = int16(s * 32767)
	}
	return out
}

func concat(samples ...[]int16) []int16 {
	var all []int16
	for _, s := range samples {
		all = append(all, s...)
	}
	return all
}

func ensureSounds() {
	if _, err := os.Stat("eat.wav"); os.IsNotExist(err) {
		a := synthSine(880, 90, 0.8, true)
		b := synthSine(1320, 80, 0.7, true)
		_ = writeWavMono16("eat.wav", concat(a, b))
	}
	if _, err := os.Stat("hit.wav"); os.IsNotExist(err) {
		b1 := synthBuzzy(120, 260, 0.9)
		b2 := synthBuzzy(80, 120, 0.6)
		_ = writeWavMono16("hit.wav", concat(b1, b2))
	}
}

func playSound(cands [][]string) {
	for _, c := range cands {
		cmd := exec.Command(c[0], c[1])
		if err := cmd.Start(); err == nil {
			go cmd.Wait()
			return
		}
	}
}

func placeApple() {
	for {
		p := Point{rand.Intn(width), rand.Intn(height)}
		ok := true
		for _, s := range snake {
			if s == p {
				ok = false
				break
			}
		}
		if ok && !bombs[p] {
			apple = p
			return
		}
	}
}

func initGame() {
	snake = []Point{{width / 2, height / 2}, {width/2 - 1, height / 2}, {width/2 - 2, height / 2}}
	dir = Point{1, 0}
	bombs = make(map[Point]bool)
	for i := 0; i < 5; i++ {
		bombs[Point{rand.Intn(width), rand.Intn(height)}] = true
	}
	placeApple()
	alive = true
	score = 0
}

func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for x := 0; x < width+2; x++ {
		termbox.SetCell(x, 0, '#', termbox.ColorBlue, termbox.ColorDefault)
		termbox.SetCell(x, height+1, '#', termbox.ColorBlue, termbox.ColorDefault)
	}
	for y := 1; y < height+1; y++ {
		termbox.SetCell(0, y, '#', termbox.ColorBlue, termbox.ColorDefault)
		termbox.SetCell(width+1, y, '#', termbox.ColorBlue, termbox.ColorDefault)
	}
	termbox.SetCell(apple.X+1, apple.Y+1, 'O', termbox.ColorRed, termbox.ColorDefault)
	for b := range bombs {
		termbox.SetCell(b.X+1, b.Y+1, 'X', termbox.ColorMagenta, termbox.ColorDefault)
	}
	for i, s := range snake {
		ch := 'o'
		if i == 0 {
			ch = 'Q'
		}
		termbox.SetCell(s.X+1, s.Y+1, ch, termbox.ColorGreen, termbox.ColorDefault)
	}
	str := "Счёт: " + itoa(score) + "  (Esc - выход)"
	for i, r := range str {
		termbox.SetCell(i, height+3, r, termbox.ColorYellow, termbox.ColorDefault)
	}
	termbox.Flush()
}

func itoa(i int) string {
	b, _ := json.Marshal(i)
	return string(b)
}

func step() {
	head := Point{snake[0].X + dir.X, snake[0].Y + dir.Y}
	if head.X < 0 || head.X >= width || head.Y < 0 || head.Y >= height {
		playSound(playCmdsHurt)
		alive = false
		return
	}
	for _, s := range snake {
		if s == head {
			playSound(playCmdsHurt)
			alive = false
			return
		}
	}
	if bombs[head] {
		playSound(playCmdsHurt)
		alive = false
		return
	}
	snake = append([]Point{head}, snake...)
	if head == apple {
		score++
		playSound(playCmdsEat)
		placeApple()
		if rand.Intn(4) == 0 {
			bombs[Point{rand.Intn(width), rand.Intn(height)}] = true
		}
	} else {
		snake = snake[:len(snake)-1]
	}
}

func mainMenu() bool {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	msg := []string{
		"=== Змейка ===",
		"",
		"1) Начать игру",
		"2) Инструкции",
		"3) Выход",
		"",
		"Выберите опцию (1-3):",
	}
	for y, line := range msg {
		for x, r := range line {
			termbox.SetCell(x+2, y+2, r, termbox.ColorWhite, termbox.ColorDefault)
		}
	}
	termbox.Flush()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return false
			default:
				switch ev.Ch {
				case '1':
					return true
				case '2':
					showInstructions()
					return mainMenu()
				case '3':
					return false
				}
			}
		case termbox.EventError:
			log.Fatal(ev.Err)
		}
	}
}

func showInstructions() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	lines := []string{
		"Инструкции:",
		"- Управление стрелками (влево/вправо/вверх/вниз).",
		"- Ешьте яблоки (O) — за каждое +1 очко.",
		"- Избегайте стен, бомб (X) и собственного тела.",
		"- Esc для выхода.",
		"",
		"Нажмите любую клавишу, чтобы вернуться.",
	}
	for y, line := range lines {
		for x, r := range line {
			termbox.SetCell(x+2, y+2, r, termbox.ColorCyan, termbox.ColorDefault)
		}
	}
	termbox.Flush()
	termbox.PollEvent()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	ensureSounds()

	for {
		start := mainMenu()
		if !start {
			break
		}
		initGame()
		ticker := time.NewTicker(120 * time.Millisecond)
		eventCh := make(chan termbox.Event)
		go func() {
			for {
				eventCh <- termbox.PollEvent()
			}
		}()
		for alive {
			draw()
			select {
			case ev := <-eventCh:
				if ev.Type == termbox.EventKey {
					switch ev.Key {
					case termbox.KeyArrowLeft:
						if dir.X != 1 {
							dir = Point{-1, 0}
						}
					case termbox.KeyArrowRight:
						if dir.X != -1 {
							dir = Point{1, 0}
						}
					case termbox.KeyArrowUp:
						if dir.Y != 1 {
							dir = Point{0, -1}
						}
					case termbox.KeyArrowDown:
						if dir.Y != -1 {
							dir = Point{0, 1}
						}
					case termbox.KeyEsc:
						alive = false
					}
					switch ev.Ch {
					case 'a':
						if dir.X != 1 {
							dir = Point{-1, 0}
						}
					case 'd':
						if dir.X != -1 {
							dir = Point{1, 0}
						}
					case 'w':
						if dir.Y != 1 {
							dir = Point{0, -1}
						}
					case 's':
						if dir.Y != -1 {
							dir = Point{0, 1}
						}
					}
				}
			case <-ticker.C:
				step()
			}
		}
		ticker.Stop()
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		msg := "Игра окончена! Счёт: " + itoa(score) + "  Нажмите любую клавишу..."
		for i, r := range msg {
			termbox.SetCell(i+2, height/2, r, termbox.ColorRed, termbox.ColorDefault)
		}
		termbox.Flush()
		termbox.PollEvent()
	}
}
