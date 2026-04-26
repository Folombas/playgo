package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
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

const (
	FRUIT_APPLE = iota
	FRUIT_STRAWBERRY
	FRUIT_ORANGE
	FRUIT_BANANA
	FRUIT_PINEAPPLE
)

type IceBlock struct {
	X, Y int
}

type Viking struct {
	X, Y   int
	Frame  int
	Timer  float64
	Active bool
}

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
	ice            IceBlock
	iceActive      bool
	frozenTimer    float64
	score          int
	health         int
	particles      []Particle
	audioCtx       *audio.Context
	sndEat         *audio.Player
	sndBoom        *audio.Player
	sndHeal        *audio.Player
	sndPause       *audio.Player
	sndMenuMove    *audio.Player
	sndMenuSelect  *audio.Player
	sndGhost       *audio.Player
	shake          float64
	menuPulse      float64
	pauseCooldown  float64
	menuSelected   int
	menuButtons    []string
	buttonFlash    int
	fontFace       font.Face

	appleImg      *ebiten.Image
	strawberryImg *ebiten.Image
	orangeImg     *ebiten.Image
	bananaImg     *ebiten.Image
	pineappleImg  *ebiten.Image

	ghostFrames     []*ebiten.Image
	ghostFrameIdx   int
	ghostAnimTimer  float64
	ghostActive     bool
	ghostX, ghostY  int
	ghostMoveTimer  float64
	ghostModeTimer  float64

	roachFrames    []*ebiten.Image
	roachFrameIdx  int
	roachActive    bool
	roachX, roachY int
	roachMoveTimer float64

	vikingFrames    []*ebiten.Image
	vikingList      []Viking
	vikingSpawnTimer float64
}

// Исправленная функция загрузки PNG
func loadPNG(path string) (*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// Исправленная загрузка спрайт-листа с явными размерами
func loadSpriteSheet(path string, frameW, frameH, cols, rows int) ([]*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	sheetW := bounds.Dx()
	sheetH := bounds.Dy()
	var frames []*ebiten.Image
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := col * frameW
			y := row * frameH
			if x+frameW > sheetW || y+frameH > sheetH {
				continue
			}
			subImg := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(x, y, x+frameW, y+frameH))
			frames = append(frames, ebiten.NewImageFromImage(subImg))
		}
	}
	return frames, nil
}

// Исправленная загрузка спрайт-листа с автоматическим определением размеров кадра
func loadSpriteSheetAuto(path string, cols, rows int) ([]*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	sheetW := bounds.Dx()
	sheetH := bounds.Dy()
	frameW := sheetW / cols
	frameH := sheetH / rows
	var frames []*ebiten.Image
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := col * frameW
			y := row * frameH
			subImg := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(x, y, x+frameW, y+frameH))
			frames = append(frames, ebiten.NewImageFromImage(subImg))
		}
	}
	return frames, nil
}

func NewGame() *Game {
	g := &Game{
		rng:             mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		state:           STATE_MENU,
		speed:           9,
		health:          maxHealth,
		menuPulse:       0,
		menuSelected:    0,
		menuButtons:     []string{"Начать игру", "Продолжить", "Новая игра", "Выйти из игры"},
		iceActive:       false,
		frozenTimer:     0,
		ghostActive:     false,
		ghostModeTimer:  0,
		ghostFrameIdx:   0,
		ghostAnimTimer:  0,
		roachActive:     false,
		vikingSpawnTimer: 0,
	}
	g.reset()
	g.audioCtx = audio.NewContext(44100)
	g.sndEat = newSound(g.audioCtx, sndEat())
	g.sndBoom = newSound(g.audioCtx, sndBoom())
	g.sndHeal = newSound(g.audioCtx, sndHeal())
	g.sndPause = newSound(g.audioCtx, sndPause())
	g.sndMenuMove = newSound(g.audioCtx, sndMenuMove())
	g.sndMenuSelect = newSound(g.audioCtx, sndMenuSelect())
	g.sndGhost = newSound(g.audioCtx, sndGhost())

	if err := g.loadFont(); err != nil {
		log.Printf("Шрифт не загружен: %v", err)
	}

	var err error
	g.appleImg, err = loadPNG("apple.png")
	if err != nil {
		log.Printf("apple.png не загружен: %v", err)
	}
	g.strawberryImg, err = loadPNG("strawberry.png")
	if err != nil {
		log.Printf("strawberry.png не загружен: %v", err)
	}
	g.orangeImg, err = loadPNG("orange.png")
	if err != nil {
		log.Printf("orange.png не загружен: %v", err)
	}
	g.bananaImg, err = loadPNG("banana.png")
	if err != nil {
		log.Printf("banana.png не загружен: %v", err)
	}
	g.pineappleImg, err = loadPNG("pineapple.png")
	if err != nil {
		log.Printf("pineapple.png не загружен: %v", err)
	}

	g.ghostFrames = make([]*ebiten.Image, 11)
	for i := 0; i <= 10; i++ {
		filename := fmt.Sprintf("skeleton-animation_%02d.png", i)
		img, err := loadPNG(filename)
		if err != nil {
			log.Printf("Не удалось загрузить %s: %v", filename, err)
			img = ebiten.NewImage(tileSize, tileSize)
			img.Fill(color.White)
		}
		g.ghostFrames[i] = img
	}

	roachFrames, err := loadSpriteSheet("roach.png", 32, 32, 4, 5)
	if err != nil {
		log.Printf("roach.png не загружен: %v", err)
	} else {
		g.roachFrames = roachFrames
		g.roachActive = true
		g.roachX = g.rng.Intn(gridW)
		g.roachY = g.rng.Intn(gridH)
		g.roachMoveTimer = 0.5
	}

	vikingFrames, err := loadSpriteSheetAuto("2204_w053_n004_9_medicharacters_p1_9.jpg", 5, 2)
	if err != nil {
		log.Printf("Не удалось загрузить викингов: %v", err)
	} else {
		g.vikingFrames = vikingFrames
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
	g.iceActive = false
	g.frozenTimer = 0
	g.score = 0
	g.ticker = 0
	g.shake = 0
	g.particles = nil
	g.ghostActive = false
	g.ghostModeTimer = 0
	g.ghostFrameIdx = 0
	g.ghostAnimTimer = 0
	if g.roachFrames != nil {
		g.roachActive = true
		g.roachX = g.rng.Intn(gridW)
		g.roachY = g.rng.Intn(gridH)
		g.roachMoveTimer = 0.5
		g.roachFrameIdx = 0
	}
	g.vikingList = nil
	g.vikingSpawnTimer = 3.0
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
		if g.iceActive && g.ice.X == x && g.ice.Y == y {
			ok = false
		}
		if ok {
			g.fruitX, g.fruitY = x, y
			g.fruitType = g.rng.Intn(5)
			return
		}
	}
}

func (g *Game) spawnIce() {
	if !g.iceActive && g.rng.Float64() < 0.3 {
		for i := 0; i < 500; i++ {
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
				g.ice = IceBlock{X: x, Y: y}
				g.iceActive = true
				return
			}
		}
	}
}

func (g *Game) spawnGhost() {
	if !g.ghostActive && g.rng.Float64() < 0.2 {
		for i := 0; i < 200; i++ {
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
			if g.iceActive && g.ice.X == x && g.ice.Y == y {
				ok = false
			}
			if ok {
				g.ghostActive = true
				g.ghostX, g.ghostY = x, y
				g.ghostMoveTimer = 0.5
				g.ghostFrameIdx = 0
				g.ghostAnimTimer = 0
				return
			}
		}
	}
}

func (g *Game) spawnViking() {
	if len(g.vikingList) >= 3 {
		return
	}
	for i := 0; i < 200; i++ {
		x := g.rng.Intn(gridW)
		y := g.rng.Intn(gridH)
		if !g.isCellOccupied(x, y) {
			g.vikingList = append(g.vikingList, Viking{
				X:      x,
				Y:      y,
				Frame:  0,
				Timer:  0,
				Active: true,
			})
			return
		}
	}
}

func (g *Game) isCellOccupied(x, y int) bool {
	for _, s := range g.snake {
		if s.X == x && s.Y == y {
			return true
		}
	}
	for _, b := range g.bombs {
		if b.X == x && b.Y == y {
			return true
		}
	}
	if g.iceActive && g.ice.X == x && g.ice.Y == y {
		return true
	}
	if g.ghostActive && g.ghostX == x && g.ghostY == y {
		return true
	}
	for _, v := range g.vikingList {
		if v.X == x && v.Y == y {
			return true
		}
	}
	return false
}

func (g *Game) Update() error {
	dt := 1.0 / 60.0
	g.menuPulse += 0.05
	if g.pauseCooldown > 0 {
		g.pauseCooldown -= dt
	}
	if g.buttonFlash > 0 {
		g.buttonFlash--
	}
	if g.frozenTimer > 0 {
		g.frozenTimer -= dt
		if g.frozenTimer < 0 {
			g.frozenTimer = 0
		}
	}
	if g.ghostModeTimer > 0 {
		g.ghostModeTimer -= dt
		if g.ghostModeTimer < 0 {
			g.ghostModeTimer = 0
		}
	}

	if g.ghostActive {
		g.ghostAnimTimer += dt
		if g.ghostAnimTimer >= 0.1 {
			g.ghostAnimTimer = 0
			g.ghostFrameIdx = (g.ghostFrameIdx + 1) % len(g.ghostFrames)
		}
	}

	if g.ghostActive && g.state == STATE_PLAYING {
		g.ghostMoveTimer -= dt
		if g.ghostMoveTimer <= 0 {
			dx, dy := 0, 0
			switch g.rng.Intn(4) {
			case 0:
				dx = 1
			case 1:
				dx = -1
			case 2:
				dy = 1
			case 3:
				dy = -1
			}
			nx, ny := g.ghostX+dx, g.ghostY+dy
			if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH {
				g.ghostX, g.ghostY = nx, ny
			}
			g.ghostMoveTimer = 0.5
		}
	}

	if g.roachActive && g.state == STATE_PLAYING {
		g.roachMoveTimer -= dt
		if g.roachMoveTimer <= 0 {
			dx, dy := 0, 0
			switch g.rng.Intn(4) {
			case 0:
				dx = 1
			case 1:
				dx = -1
			case 2:
				dy = 1
			case 3:
				dy = -1
			}
			nx, ny := g.roachX+dx, g.roachY+dy
			if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH && !g.isCellOccupied(nx, ny) {
				g.roachX, g.roachY = nx, ny
				if len(g.roachFrames) > 0 {
					g.roachFrameIdx = (g.roachFrameIdx + 1) % len(g.roachFrames)
				}
			}
			g.roachMoveTimer = 0.8
		}
	}

	if g.state == STATE_PLAYING {
		g.vikingSpawnTimer -= dt
		if g.vikingSpawnTimer <= 0 && len(g.vikingList) < 3 {
			g.spawnViking()
			g.vikingSpawnTimer = 10.0
		}
		for i := 0; i < len(g.vikingList); i++ {
			v := &g.vikingList[i]
			v.Timer += dt
			if v.Timer >= 0.1 {
				v.Timer = 0
				v.Frame = (v.Frame + 1) % len(g.vikingFrames)
			}
			if g.rng.Float64() < 0.02 {
				dx, dy := 0, 0
				switch g.rng.Intn(4) {
				case 0:
					dx = 1
				case 1:
					dx = -1
				case 2:
					dy = 1
				case 3:
					dy = -1
				}
				nx, ny := v.X+dx, v.Y+dy
				if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH && !g.isCellOccupied(nx, ny) {
					v.X, v.Y = nx, ny
				}
			}
			if v.Active && g.snake[0].X == v.X && g.snake[0].Y == v.Y {
				g.health -= 20
				g.addParticles(float64(v.X*tileSize+tileSize/2), float64(v.Y*tileSize+tileSize/2), 40, color.RGBA{200, 50, 50, 255}, true)
				g.sndHeal.Rewind()
				g.sndHeal.Play()
				g.vikingList = append(g.vikingList[:i], g.vikingList[i+1:]...)
				i--
				if g.health <= 0 {
					g.state = STATE_GAMEOVER
				}
			}
		}
	}

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

	if g.state == STATE_MENU {
		prev := g.menuSelected
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
		if prev != g.menuSelected {
			g.sndMenuMove.Rewind()
			g.sndMenuMove.Play()
		}
		if (ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace)) && g.pauseCooldown <= 0 {
			g.sndMenuSelect.Rewind()
			g.sndMenuSelect.Play()
			g.buttonFlash = 5
			switch g.menuSelected {
			case 0, 1, 2:
				g.reset()
				g.state = STATE_PLAYING
			case 3:
				return ebiten.Termination
			}
			g.pauseCooldown = 0.3
		}
		return nil
	}

	if g.state == STATE_PAUSED {
		return nil
	}
	if g.state == STATE_GAMEOVER {
		if inputPressed() && g.pauseCooldown <= 0 {
			g.state = STATE_MENU
			g.menuSelected = 0
			g.pauseCooldown = 0.2
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

	g.ticker += g.speed * dt
	if g.ticker >= 1 {
		g.ticker = 0
		g.dir = g.nextDir
		g.step()
	}

	for i := 0; i < len(g.bombs); i++ {
		g.bombs[i].Timer -= dt
		if g.bombs[i].Timer <= 0 {
			g.bombExplode(i)
			i--
		}
	}

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

func (g *Game) step() {
	head := g.snake[0]
	newHead := Vec{head.X + g.dir.X, head.Y + g.dir.Y}

	if newHead.X < 0 || newHead.X >= gridW || newHead.Y < 0 || newHead.Y >= gridH {
		g.triggerExplosion(head, true)
		return
	}
	if !g.ghostModeActive() {
		for _, s := range g.snake {
			if s == newHead {
				g.triggerExplosion(newHead, true)
				return
			}
		}
	}

	g.snake = append([]Vec{newHead}, g.snake...)

	ateFruit := false
	if newHead.X == g.fruitX && newHead.Y == g.fruitY {
		ateFruit = true
		switch g.fruitType {
		case FRUIT_APPLE:
			g.score += 1
			g.health = minInt(maxHealth, g.health+25)
		case FRUIT_STRAWBERRY:
			g.score += 2
			g.health = minInt(maxHealth, g.health+40)
		case FRUIT_ORANGE:
			g.score += 3
			g.health = minInt(maxHealth, g.health+35)
		case FRUIT_BANANA:
			g.score += 1
			g.health = minInt(maxHealth, g.health+20)
		case FRUIT_PINEAPPLE:
			g.score += 4
			g.health = minInt(maxHealth, g.health+45)
		}
		g.placeFruit()
		g.spawnBombRandom()
		g.spawnIce()
		g.spawnGhost()
		g.sndHeal.Rewind()
		g.sndHeal.Play()
		g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 25, color.RGBA{50, 255, 80, 255}, true)
	}

	if g.frozenTimer > 0 && ateFruit {
		g.snake = g.snake[:len(g.snake)-1]
	} else if !ateFruit {
		g.snake = g.snake[:len(g.snake)-1]
	}

	for i := 0; i < len(g.bombs); i++ {
		if g.bombs[i].X == newHead.X && g.bombs[i].Y == newHead.Y {
			g.health -= 35
			g.triggerExplosion(newHead, g.health <= 0)
			g.bombs = append(g.bombs[:i], g.bombs[i+1:]...)
			return
		}
	}

	if g.iceActive && newHead.X == g.ice.X && newHead.Y == g.ice.Y {
		g.frozenTimer = 5.0
		g.iceActive = false
		g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 50, color.RGBA{100, 200, 255, 255}, true)
		g.sndHeal.Rewind()
		g.sndHeal.Play()
	}

	if g.ghostActive && newHead.X == g.ghostX && newHead.Y == g.ghostY {
		g.ghostModeTimer = 5.0
		g.ghostActive = false
		g.sndGhost.Rewind()
		g.sndGhost.Play()
		g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 60, color.RGBA{200, 200, 255, 200}, true)
	}

	g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 2, color.RGBA{0, 180, 220, 140}, false)
	if g.frozenTimer > 0 {
		g.addParticles(float64(newHead.X*tileSize+tileSize/2), float64(newHead.Y*tileSize+tileSize/2), 1, color.RGBA{150, 220, 255, 180}, true)
	}
}

func (g *Game) ghostModeActive() bool {
	return g.ghostModeTimer > 0
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
	head := g.snake[0]
	dx := math.Abs(float64(head.X - b.X))
	dy := math.Abs(float64(head.Y - b.Y))
	dist := dx + dy
	g.shake = 12
	g.sndBoom.Rewind()
	g.sndBoom.Play()
	cx := float64(b.X*tileSize + tileSize/2)
	cy := float64(b.Y*tileSize + tileSize/2)
	if dist <= 1.5 {
		g.health -= 25
		g.addParticles(cx, cy, 120, color.RGBA{255, 60, 30, 255}, true)
		g.addParticles(cx, cy, 60, color.RGBA{255, 200, 50, 200}, true)
		if g.health <= 0 {
			g.state = STATE_GAMEOVER
		}
	} else {
		g.addParticles(cx, cy, 80, color.RGBA{255, 100, 30, 255}, true)
	}
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
			if g.iceActive && g.ice.X == x && g.ice.Y == y {
				ok = false
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

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 12, 20, 255})

	ox, oy := 0.0, 0.0
	if g.shake > 0.5 {
		ox = (mathrand.Float64()*2 - 1) * g.shake
		oy = (mathrand.Float64()*2 - 1) * g.shake
	}

	for x := 0; x < gridW; x++ {
		for y := 0; y < gridH; y++ {
			c := color.RGBA{15, 15, 25, 255}
			if (x+y)%2 != 0 {
				c = color.RGBA{18, 18, 30, 255}
			}
			ebitenutil.DrawRect(screen, float64(x*tileSize)+ox, float64(y*tileSize)+oy, tileSize-1, tileSize-1, c)
		}
	}

	for _, b := range g.bombs {
		cx := float64(b.X*tileSize+tileSize/2) + ox
		cy := float64(b.Y*tileSize+tileSize/2) + oy
		baseRadius := float64(tileSize)/2 * 1.5
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

		fuseLen := 20.0 * (t / 5.0)
		fuseStartX := cx + radius*0.7
		fuseStartY := cy - radius*1.1
		fuseEndX := fuseStartX + fuseLen*0.7
		fuseEndY := fuseStartY - fuseLen*0.5
		ebitenutil.DrawLine(screen, fuseStartX, fuseStartY, fuseEndX, fuseEndY, color.RGBA{80, 70, 50, 255})
		fireSize := 3.0 + 2*math.Sin(g.menuPulse*50)
		ebitenutil.DrawCircle(screen, fuseEndX, fuseEndY, fireSize, color.RGBA{255, 100, 20, 255})
		ebitenutil.DrawCircle(screen, fuseEndX, fuseEndY, fireSize*0.6, color.RGBA{255, 255, 100, 255})
	}

	for i, s := range g.snake {
		x := float64(s.X*tileSize) + ox
		y := float64(s.Y*tileSize) + oy
		var base color.RGBA
		if g.frozenTimer > 0 {
			base = color.RGBA{80, 180, 255, 255}
		} else if g.ghostModeActive() {
			base = color.RGBA{20, 220, 90, 180}
		} else {
			base = color.RGBA{20, 220, 90, 255}
		}
		if i > 0 {
			shade := uint8(100 + (i*4)%100)
			if g.frozenTimer > 0 {
				base = color.RGBA{80, 180, 255, shade}
			} else if g.ghostModeActive() {
				base = color.RGBA{20, 220, 90, uint8(180)}
			} else {
				base = color.RGBA{15, shade, 70, 255}
			}
		}
		if i == 0 {
			ebitenutil.DrawRect(screen, x-3, y-3, tileSize+6, tileSize+6, color.RGBA{0, 200, 80, 40})
			if g.frozenTimer > 0 {
				ebitenutil.DrawRect(screen, x-3, y-3, tileSize+6, tileSize+6, color.RGBA{100, 200, 255, 80})
			}
			if g.ghostModeActive() {
				ebitenutil.DrawRect(screen, x-3, y-3, tileSize+6, tileSize+6, color.RGBA{255, 255, 255, 80})
			}
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

	{
		cx := float64(g.fruitX*tileSize+tileSize/2) + ox
		cy := float64(g.fruitY*tileSize+tileSize/2) + oy
		var img *ebiten.Image
		switch g.fruitType {
		case FRUIT_APPLE:
			img = g.appleImg
		case FRUIT_STRAWBERRY:
			img = g.strawberryImg
		case FRUIT_ORANGE:
			img = g.orangeImg
		case FRUIT_BANANA:
			img = g.bananaImg
		case FRUIT_PINEAPPLE:
			img = g.pineappleImg
		}
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			w, h := img.Bounds().Dx(), img.Bounds().Dy()
			scale := 1.5
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
			op.GeoM.Translate(ox, oy)
			screen.DrawImage(img, op)
		} else {
			ebitenutil.DrawRect(screen, cx-float64(tileSize)/2+ox, cy-float64(tileSize)/2+oy, float64(tileSize), float64(tileSize), color.RGBA{200, 100, 50, 255})
		}
	}

	if g.iceActive {
		cx := float64(g.ice.X*tileSize+tileSize/2) + ox
		cy := float64(g.ice.Y*tileSize+tileSize/2) + oy
		radius := float64(tileSize)/2 * 1.2
		ebitenutil.DrawCircle(screen, cx, cy, radius, color.RGBA{150, 220, 255, 255})
		ebitenutil.DrawCircle(screen, cx-2, cy-2, radius-3, color.RGBA{100, 200, 240, 200})
		ebitenutil.DrawCircle(screen, cx+radius*0.3, cy-radius*0.3, radius*0.25, color.RGBA{255, 255, 255, 200})
	}

	if g.ghostActive && len(g.ghostFrames) > 0 {
		cx := float64(g.ghostX*tileSize+tileSize/2) + ox
		cy := float64(g.ghostY*tileSize+tileSize/2) + oy
		frame := g.ghostFrames[g.ghostFrameIdx]
		if frame != nil {
			op := &ebiten.DrawImageOptions{}
			w, h := frame.Bounds().Dx(), frame.Bounds().Dy()
			scale := float64(tileSize) / float64(w)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
			op.GeoM.Translate(ox, oy)
			screen.DrawImage(frame, op)
		}
	}

	if g.roachActive && len(g.roachFrames) > 0 {
		cx := float64(g.roachX*tileSize+tileSize/2) + ox
		cy := float64(g.roachY*tileSize+tileSize/2) + oy
		frame := g.roachFrames[g.roachFrameIdx]
		if frame != nil {
			op := &ebiten.DrawImageOptions{}
			w, h := frame.Bounds().Dx(), frame.Bounds().Dy()
			scale := float64(tileSize) / float64(w)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
			op.GeoM.Translate(ox, oy)
			screen.DrawImage(frame, op)
		}
	}

	if len(g.vikingFrames) > 0 {
		for _, v := range g.vikingList {
			cx := float64(v.X*tileSize+tileSize/2) + ox
			cy := float64(v.Y*tileSize+tileSize/2) + oy
			frame := g.vikingFrames[v.Frame%len(g.vikingFrames)]
			op := &ebiten.DrawImageOptions{}
			w, h := frame.Bounds().Dx(), frame.Bounds().Dy()
			scale := float64(tileSize) / float64(w)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
			op.GeoM.Translate(ox, oy)
			screen.DrawImage(frame, op)
		}
	}

	for _, p := range g.particles {
		c := p.Color
		if p.Glow {
			ebitenutil.DrawRect(screen, p.X-p.Size*1.5+ox, p.Y-p.Size*1.5+oy, p.Size*3, p.Size*3, color.RGBA{c.R, c.G, c.B, uint8(float64(c.A) * 0.4 * p.Life)})
		}
		ebitenutil.DrawRect(screen, p.X-p.Size+ox, p.Y-p.Size+oy, p.Size*2, p.Size*2, c)
	}

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

	if g.frozenTimer > 0 {
		drawText("ЗАМОРОЗКА", screenW/2-60, screenH-30, color.RGBA{100, 200, 255, 255})
	}
	if g.ghostModeActive() {
		drawText("ПРИЗРАЧНЫЙ РЕЖИМ", screenW/2-100, screenH-60, color.RGBA{200, 200, 255, 255})
	}

	switch g.state {
	case STATE_MENU:
		ebitenutil.DrawRect(screen, 0, 0, screenW, screenH, color.RGBA{0, 0, 0, 255})
		drawText("S N A K E   R E V I V E D", screenW/2-180, 150, color.RGBA{255, 200, 100, 255})
		startY := 280
		step := 50
		for i, btn := range g.menuButtons {
			y := startY + i*step
			if i == g.menuSelected {
				bg := color.RGBA{100, 100, 150, 255}
				if g.buttonFlash > 0 {
					bg = color.RGBA{200, 200, 255, 255}
				}
				ebitenutil.DrawRect(screen, screenW/2-150, float64(y)-15, 300, 35, bg)
				drawText(btn, screenW/2-len(btn)*3, y, color.RGBA{255, 255, 0, 255})
			} else {
				drawText(btn, screenW/2-len(btn)*3, y, color.White)
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

// ------------------------------------------------
// Аудио (синтез)
// ------------------------------------------------
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
func sndMenuMove() []byte {
	sr := 44100
	t := synthWave(sr, 0.05, 800, 0.3, "sine", 100)
	return mixToWAV(sr, [][]int16{t})
}
func sndMenuSelect() []byte {
	sr := 44100
	t1 := synthWave(sr, 0.1, 400, 0.5, "sine", 200)
	t2 := synthWave(sr, 0.1, 700, 0.5, "sine", -100)
	return mixToWAV(sr, [][]int16{t1, t2})
}
func sndGhost() []byte {
	sr := 44100
	t1 := synthWave(sr, 0.2, 500, 0.4, "sine", -200)
	t2 := synthWave(sr, 0.2, 800, 0.3, "sine", 100)
	return mixToWAV(sr, [][]int16{t1, t2})
}

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Змейка: Возрождение")
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}