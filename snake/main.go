package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 800
	screenHeight = 600
	gridSize     = 20
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Snake struct {
	body      []struct{ x, y int }
	direction Direction
	growing   bool
}

type Food struct {
	x, y  int
	color color.RGBA
	score int
}

type Particle struct {
	x, y     float64
	vx, vy   float64
	life     int
	color    color.RGBA
	size     float32
}

type Game struct {
	snake       Snake
	food        Food
	state       GameState
	score       int
	highScore   int
	level       int
	particles   []*Particle
	grid        [][]bool
	frameCount  int
	
	// Assets
	snakeHead  *ebiten.Image
	snakeBody  *ebiten.Image
	foodImages map[int]*ebiten.Image
	gameFont   font.Face
	audioCtx   *audio.Context
	sounds     map[int][]byte
}

func LoadFont(path string, size int) font.Face {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	ttFont, err := opentype.Parse(data)
	if err != nil {
		return nil
	}
	face, _ := opentype.NewFace(ttFont, &opentype.FaceOptions{
		Size: float64(size),
		DPI:  72,
	})
	return face
}

func InitAudio() *audio.Context {
	return audio.NewContext(44100)
}

func GenerateSound(frequency, duration float64) []byte {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]byte, numSamples*2)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		envelope := 1.0 - float64(i)/float64(numSamples)
		value := math.Sin(2*math.Pi*frequency*t) * envelope * 0.3
		sample := int16(value * 32767)
		samples[i*2] = byte(sample)
		samples[i*2+1] = byte(sample >> 8)
	}
	return samples
}

func LoadSounds() map[int][]byte {
	sounds := make(map[int][]byte)
	sounds[0] = GenerateSound(600, 0.1)  // Eat
	sounds[1] = GenerateSound(200, 0.3)  // Die
	sounds[2] = GenerateSound(800, 0.15) // Level up
	sounds[3] = GenerateSound(400, 0.1)  // Move
	return sounds
}

func PlaySound(g *Game, soundType int) {
	if g.audioCtx == nil {
		return
	}
	samples, ok := g.sounds[soundType]
	if !ok || len(samples) == 0 {
		return
	}
	player := g.audioCtx.NewPlayerFromBytes(samples)
	player.SetVolume(0.4)
	player.Play()
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	
	g := &Game{
		snake: Snake{
			body: []struct{ x, y int }{{10, 10}},
			direction: Right,
		},
		food: Food{
			x: 15,
			y: 15,
			color: color.RGBA{255, 50, 50, 255},
			score: 10,
		},
		state: StateMenu,
		highScore: 0,
		level: 1,
		particles: make([]*Particle, 0),
		grid: make([][]bool, screenWidth/gridSize),
		frameCount: 0,
	}
	
	// Initialize grid
	for x := range g.grid {
		g.grid[x] = make([]bool, screenHeight/gridSize)
	}
	
	// Load assets
	g.gameFont = LoadFont("assets/fonts/SuperFeel-JpZqa.ttf", 32)
	g.audioCtx = InitAudio()
	g.sounds = LoadSounds()
	
	// Load food images
	g.foodImages = make(map[int]*ebiten.Image)
	foodTypes := []string{"coinGold.png", "gemRed.png", "gemBlue.png", "gemGreen.png", "star.png"}
	for i, foodFile := range foodTypes {
		path := filepath.Join("assets", "ui", foodFile)
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err == nil {
			g.foodImages[i+1] = img
		}
	}
	
	return g
}

func (g *Game) Update() error {
	g.frameCount++
	
	switch g.state {
	case StateMenu:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = StatePlaying
			PlaySound(g, 2)
		}
		return nil
		
	case StatePaused:
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = StatePlaying
		}
		return nil
		
	case StateGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g = NewGame()
		}
		return nil
	}
	
	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.2
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
	
	// Move snake (every 10 frames based on level)
	speed := 10 - g.level
	if speed < 3 {
		speed = 3
	}
	
	if g.frameCount%speed != 0 {
		return nil
	}
	
	// Calculate new head position
	head := g.snake.body[0]
	newX, newY := head.x, head.y
	
	switch g.snake.direction {
	case Up:
		newY--
	case Down:
		newY++
	case Left:
		newX--
	case Right:
		newX++
	}
	
	// Check wall collision
	if newX < 0 || newX >= screenWidth/gridSize || newY < 0 || newY >= screenHeight/gridSize {
		g.state = StateGameOver
		PlaySound(g, 1)
		return nil
	}
	
	// Check self collision
	for _, segment := range g.snake.body {
		if segment.x == newX && segment.y == newY {
			g.state = StateGameOver
			PlaySound(g, 1)
			return nil
		}
	}
	
	// Add new head
	g.snake.body = append([]struct{ x, y int }{{newX, newY}}, g.snake.body...)
	
	if !g.snake.growing {
		g.snake.body = g.snake.body[:len(g.snake.body)-1]
	} else {
		g.snake.growing = false
	}
	
	// Check food collision
	if newX == g.food.x && newY == g.food.y {
		g.snake.growing = true
		g.score += g.food.score
		if g.score > g.highScore {
			g.highScore = g.score
		}
		
		// Level up every 50 points
		if g.score%50 == 0 {
			g.level++
			PlaySound(g, 2)
		} else {
			PlaySound(g, 0)
		}
		
		// Spawn particles
		g.spawnParticles(float64(newX*gridSize), float64(newY*gridSize), 10, g.food.color)
		
		// Spawn new food
		g.spawnFood()
	}
	
	return nil
}

func (g *Game) spawnFood() {
	for {
		g.food.x = rand.Intn(screenWidth / gridSize)
		g.food.y = rand.Intn(screenHeight / gridSize)
		
		// Check if food spawned on snake
		onSnake := false
		for _, segment := range g.snake.body {
			if segment.x == g.food.x && segment.y == g.food.y {
				onSnake = true
				break
			}
		}
		
		if !onSnake {
			// Random food type
			foodTypes := []struct {
				color color.RGBA
				score int
			}{
				{color.RGBA{255, 50, 50, 255}, 10},   // Apple
				{color.RGBA{50, 255, 50, 255}, 20},   // Berry
				{color.RGBA{255, 255, 50, 255}, 30},  // Star
				{color.RGBA{50, 50, 255, 255}, 40},   // Gem
			}
			food := foodTypes[rand.Intn(len(foodTypes))]
			g.food.color = food.color
			g.food.score = food.score
			break
		}
	}
}

func (g *Game) spawnParticles(x, y float64, count int, c color.RGBA) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			x: x + float64(gridSize)/2,
			y: y + float64(gridSize)/2,
			vx: float64(rand.Intn(10)-5) * 0.5,
			vy: float64(rand.Intn(10)-5) * 0.5,
			life: 30 + rand.Intn(20),
			color: c,
			size: float32(rand.Intn(4)+2),
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{20, 20, 40, 255})
	
	// Draw grid
	for x := 0; x < screenWidth/gridSize; x++ {
		for y := 0; y < screenHeight/gridSize; y++ {
			c := color.RGBA{30, 30, 60, 255}
			if (x+y)%2 == 0 {
				c = color.RGBA{40, 40, 80, 255}
			}
			vector.DrawFilledRect(screen, float32(x*gridSize), float32(y*gridSize), gridSize, gridSize, c, false)
		}
	}
	
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying, StatePaused:
		g.drawGame(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	}
	
	// Draw particles
	for _, p := range g.particles {
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), p.size, p.color, false)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Animated background
	for i := 0; i < 50; i++ {
		x := float32((i*37 + g.frameCount/2) % screenWidth)
		y := float32((i*23) % screenHeight)
		size := float32((i%3)+1) * float32(0.5+0.5*math.Sin(float64(g.frameCount+i)/20))
		vector.DrawFilledCircle(screen, x, y, size, color.RGBA{100, 100, 255, 100}, false)
	}
	
	if g.gameFont != nil {
		title := "🐍 SNAKE"
		text.Draw(screen, title, g.gameFont, screenWidth/2-80, 200, color.RGBA{50, 255, 50, 255})
		
		subtitle := "Classic Arcade Game"
		text.Draw(screen, subtitle, g.gameFont, screenWidth/2-100, 250, color.RGBA{200, 200, 200, 255})
		
		instructions := []string{
			"Arrow Keys - Move",
			"P - Pause",
			"Enter/Space - Start",
			fmt.Sprintf("High Score: %d", g.highScore),
		}
		
		for i, line := range instructions {
			text.Draw(screen, line, g.gameFont, screenWidth/2-100, 320+i*40, color.RGBA{255, 255, 255, 255})
		}
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Draw food
	foodX := float64(g.food.x * gridSize)
	foodY := float64(g.food.y * gridSize)
	
	if img, ok := g.foodImages[g.level%5+1]; ok && img != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(foodX, foodY)
		screen.DrawImage(img, op)
	} else {
		vector.DrawFilledCircle(screen, float32(foodX)+float32(gridSize)/2, float32(foodY)+float32(gridSize)/2, float32(gridSize)/2-2, g.food.color, false)
	}
	
	// Draw snake
	for i, segment := range g.snake.body {
		x := float64(segment.x * gridSize)
		y := float64(segment.y * gridSize)
		
		// Gradient color from head to tail
		green := uint8(255 - i*5)
		if green > 200 {
			green = 200
		}
		c := color.RGBA{50, green, 50, 255}
		
		if i == 0 {
			// Head
			c = color.RGBA{100, 255, 100, 255}
			vector.DrawFilledRect(screen, float32(x)+1, float32(y)+1, gridSize-2, gridSize-2, c, false)
			
			// Eyes
			eyeOffset := 0
			switch g.snake.direction {
			case Up:
				eyeOffset = -5
			case Down:
				eyeOffset = 5
			case Left:
				eyeOffset = -5
			case Right:
				eyeOffset = 5
			}
			vector.DrawFilledCircle(screen, float32(x)+5, float32(y)+8+float32(eyeOffset), 3, color.RGBA{0, 0, 0, 255}, false)
			vector.DrawFilledCircle(screen, float32(x)+15, float32(y)+8+float32(eyeOffset), 3, color.RGBA{0, 0, 0, 255}, false)
		} else {
			vector.DrawFilledRect(screen, float32(x)+2, float32(y)+2, gridSize-4, gridSize-4, c, false)
		}
	}
	
	// Draw UI
	if g.gameFont != nil {
		scoreText := fmt.Sprintf("Score: %d", g.score)
		text.Draw(screen, scoreText, g.gameFont, 20, 40, color.RGBA{255, 255, 255, 255})
		
		levelText := fmt.Sprintf("Level: %d", g.level)
		text.Draw(screen, levelText, g.gameFont, 20, 80, color.RGBA{255, 215, 0, 255})
		
		highScoreText := fmt.Sprintf("High: %d", g.highScore)
		text.Draw(screen, highScoreText, g.gameFont, screenWidth-200, 40, color.RGBA{255, 215, 0, 255})
	}
	
	if g.state == StatePaused {
		vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150}, false)
		if g.gameFont != nil {
			text.Draw(screen, "PAUSED", g.gameFont, screenWidth/2-80, screenHeight/2, color.RGBA{255, 255, 255, 255})
		}
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{50, 0, 0, 200}, false)
	
	if g.gameFont != nil {
		title := "GAME OVER"
		text.Draw(screen, title, g.gameFont, screenWidth/2-120, screenHeight/2-50, color.RGBA{255, 50, 50, 255})
		
		scoreText := fmt.Sprintf("Score: %d", g.score)
		text.Draw(screen, scoreText, g.gameFont, screenWidth/2-60, screenHeight/2+10, color.RGBA{255, 255, 255, 255})
		
		restartText := "Press ENTER to restart"
		text.Draw(screen, restartText, g.gameFont, screenWidth/2-130, screenHeight/2+60, color.RGBA{200, 200, 200, 255})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("🐍 Snake - Classic Arcade | Go365 Day 84")
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
