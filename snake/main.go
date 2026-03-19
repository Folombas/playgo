package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	tileSize     = 20
	screenWidth  = 800
	screenHeight = 600
	gridSizeX    = screenWidth / tileSize
	gridSizeY    = screenHeight / tileSize
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type GameState int

const (
	Menu GameState = iota
	SelectDifficulty
	Playing
	Paused
	GameOver
)

type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	default:
		return "Unknown"
	}
}

func (d Difficulty) EnemyCount() int {
	switch d {
	case Easy:
		return 2
	case Medium:
		return 3
	case Hard:
		return 5
	default:
		return 3
	}
}

// Particle represents a visual effect particle
type Particle struct {
	X, Y     float32
	VX, VY   float32
	Life     int
	MaxLife  int
	Color    color.RGBA
	Size     float32
	Gravity  float32
}

// ScreenShake handles screen shake effect
type ScreenShake struct {
	Intensity float32
	Duration  int
	Timer     int
	Angle     float64
}

func (ss *ScreenShake) Update() {
	if ss.Timer > 0 {
		ss.Timer--
		ss.Angle += math.Pi / 8
		if ss.Timer <= 0 {
			ss.Intensity = 0
		}
	}
}

func (ss *ScreenShake) IsActive() bool {
	return ss.Timer > 0
}

func (ss *ScreenShake) GetOffset() (float32, float32) {
	if !ss.IsActive() {
		return 0, 0
	}
	offset := ss.Intensity * float32(ss.Timer) / float32(ss.Duration)
	dx := offset * float32(math.Sin(ss.Angle))
	dy := offset * float32(math.Cos(ss.Angle))
	return dx, dy
}

func (ss *ScreenShake) Trigger(intensity float32, duration int) {
	ss.Intensity = intensity
	ss.Duration = duration
	ss.Timer = duration
	ss.Angle = rand.Float64() * math.Pi * 2
}

type Game struct {
	snake       []Point
	direction   Direction
	food        Point
	score       int
	gameOver    bool
	moveTimer   int
	moveDelay   int // тиков между движениями (60 тиков = 1 сек)
	enemies     []Enemy
	enemyTimer  int
	enemyDelay  int // скорость врагов
	bombs       []Bomb
	bombTimer   int
	bombDelay   int // время спавна бомб
	chest       *TreasureChest
	key         *Key
	coins       []Coin
	coinTimer   int
	coinDelay   int // время спавна монет
	arrows      []Arrow
	hasKey      bool
	arrowCount  int
	shootTimer  int
	state       GameState
	difficulty  Difficulty
	
	// Visual effects
	particles   []Particle
	screenShake ScreenShake
	foodTimer   int // для анимации появления еды
	backgroundGradient *ebiten.Image
}

type Point struct {
	X int
	Y int
}

type Enemy struct {
	pos       Point
	direction Direction
	animFrame int
}

type Bomb struct {
	pos      Point
	timer    int
	maxTime  int // время до взрыва
}

type TreasureChest struct {
	pos    Point
	open   bool
	arrows int // количество стрел в сундуке
}

type Key struct {
	pos Point
}

type Coin struct {
	pos      Point
	value    int // множитель очков (2 = x2 XP)
	collected bool
	pulsePhase float64 // для анимации пульсации
}

type Arrow struct {
	pos       Point
	direction Direction
	active    bool
	speed     int
}

func createGradientBackground() *ebiten.Image {
	gradient := ebiten.NewImage(screenWidth, screenHeight)
	
	// Создаём градиент от тёмно-синего к чёрному
	for y := 0; y < screenHeight; y++ {
		ratio := float32(y) / float32(screenHeight)
		// Тёмно-синий (10, 10, 30) к чёрному (0, 0, 0)
		r := uint8(float32(10) * (1 - ratio))
		g := uint8(float32(10) * (1 - ratio))
		b := uint8(float32(30) * (1 - ratio) + float32(0)*ratio)
		
		for x := 0; x < screenWidth; x++ {
			gradient.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	
	return gradient
}

func NewGame() *Game {
	g := &Game{
		snake:     []Point{{gridSizeX / 2, gridSizeY / 2}, {gridSizeX/2 - 1, gridSizeY / 2}, {gridSizeX/2 - 2, gridSizeY / 2}},
		direction: Right,
		score:     0,
		gameOver:  false,
		moveDelay: 8,
		enemyDelay: 12,
		bombDelay: 180,
		coinDelay: 300,
		enemies:   []Enemy{},
		bombs:     []Bomb{},
		coins:     []Coin{},
		arrows:    []Arrow{},
		hasKey:    false,
		arrowCount: 0,
		state:     Menu,
		particles: []Particle{},
		backgroundGradient: createGradientBackground(),
	}
	return g
}

func (g *Game) startGame() {
	g.placeFood()
	g.spawnChest()
	g.spawnKey()
	enemyCount := g.difficulty.EnemyCount()
	for i := 0; i < enemyCount; i++ {
		g.spawnEnemy()
	}
	g.state = Playing
	g.foodTimer = 20 // Анимация появления еды
}

func (g *Game) placeFood() {
	rand.Seed(time.Now().UnixNano())
	for {
		g.food = Point{
			X: rand.Intn(gridSizeX),
			Y: rand.Intn(gridSizeY),
		}
		onSnake := false
		for _, segment := range g.snake {
			if segment.X == g.food.X && segment.Y == g.food.Y {
				onSnake = true
				break
			}
		}
		if !onSnake {
			break
		}
	}
	g.foodTimer = 20 // Сброс таймера для анимации
}

func (g *Game) spawnEnemy() {
	rand.Seed(time.Now().UnixNano())
	for {
		pos := Point{
			X: rand.Intn(gridSizeX),
			Y: rand.Intn(gridSizeY),
		}
		tooClose := false
		for _, segment := range g.snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 20 && dy < 15 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			dir := Direction(rand.Intn(4))
			g.enemies = append(g.enemies, Enemy{pos: pos, direction: dir, animFrame: 0})
			break
		}
	}
}

func (g *Game) spawnBomb() {
	rand.Seed(time.Now().UnixNano())
	for {
		pos := Point{
			X: rand.Intn(gridSizeX),
			Y: rand.Intn(gridSizeY),
		}
		tooClose := false
		for _, segment := range g.snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 15 && dy < 10 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.bombs = append(g.bombs, Bomb{pos: pos, timer: 0, maxTime: 180})
			break
		}
	}
}

func (g *Game) spawnChest() {
	rand.Seed(time.Now().UnixNano())
	for {
		pos := Point{
			X: rand.Intn(gridSizeX),
			Y: rand.Intn(gridSizeY),
		}
		tooClose := false
		for _, segment := range g.snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 20 && dy < 15 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.chest = &TreasureChest{pos: pos, open: false, arrows: 5}
			break
		}
	}
}

func (g *Game) spawnKey() {
	rand.Seed(time.Now().UnixNano())
	for {
		pos := Point{
			X: rand.Intn(gridSizeX),
			Y: rand.Intn(gridSizeY),
		}
		tooClose := false
		for _, segment := range g.snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 20 && dy < 15 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.key = &Key{pos: pos}
			break
		}
	}
}

func (g *Game) spawnCoin() {
	rand.Seed(time.Now().UnixNano())
	for {
		pos := Point{
			X: rand.Intn(gridSizeX),
			Y: rand.Intn(gridSizeY),
		}
		tooClose := false
		for _, segment := range g.snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 15 && dy < 10 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.coins = append(g.coins, Coin{pos: pos, value: 2, collected: false, pulsePhase: 0})
			break
		}
	}
}

// spawnParticles создаёт частицы в указанной позиции
func (g *Game) spawnParticles(x, y float32, count int, baseColor color.RGBA, spread float32) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := rand.Float64() * float64(spread)
		particle := Particle{
			X: x,
			Y: y,
			VX: float32(math.Cos(angle) * speed),
			VY: float32(math.Sin(angle) * speed),
			Life: 20 + rand.Intn(10),
			MaxLife: 30,
			Color: baseColor,
			Size: 2 + rand.Float32()*3,
			Gravity: 0.1,
		}
		g.particles = append(g.particles, particle)
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += p.Gravity
		p.Life--
		
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		alpha := uint8(float32(p.Color.A) * float32(p.Life) / float32(p.MaxLife))
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		vector.DrawFilledCircle(screen, p.X, p.Y, p.Size, c, false)
	}
}

func (g *Game) Update() error {
	// Обновление тряски экрана
	g.screenShake.Update()
	
	// Обновление пульсации монет
	for i := range g.coins {
		g.coins[i].pulsePhase += 0.15
	}
	
	switch g.state {
	case Menu:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = SelectDifficulty
		}
		return nil

	case SelectDifficulty:
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			if g.difficulty == Easy {
				g.difficulty = Hard
			} else {
				g.difficulty--
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			if g.difficulty == Hard {
				g.difficulty = Easy
			} else {
				g.difficulty++
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.startGame()
		}
		return nil

	case Paused:
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = Playing
		}
		return nil

	case GameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			*g = *NewGame()
		}
		return nil

	case Playing:
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = Paused
			return nil
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && g.direction != Down {
		g.direction = Up
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && g.direction != Up {
		g.direction = Down
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && g.direction != Right {
		g.direction = Left
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && g.direction != Left {
		g.direction = Right
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && g.arrowCount > 0 {
		g.shootArrow()
	}

	g.moveTimer++
	if g.moveTimer < g.moveDelay {
		g.updateParticles()
		return nil
	}
	g.moveTimer = 0

	head := g.snake[0]
	newHead := head

	switch g.direction {
	case Up:
		newHead.Y--
	case Down:
		newHead.Y++
	case Left:
		newHead.X--
	case Right:
		newHead.X++
	}

	if newHead.X < 0 || newHead.X >= gridSizeX || newHead.Y < 0 || newHead.Y >= gridSizeY {
		g.gameOver = true
		g.state = GameOver
		g.screenShake.Trigger(5, 20) // Тряска при ударе о стену
		g.spawnParticles(float32(newHead.X*tileSize+tileSize/2), float32(newHead.Y*tileSize+tileSize/2), 20, color.RGBA{255, 100, 100, 255}, 3)
		return nil
	}

	for _, segment := range g.snake {
		if segment.X == newHead.X && segment.Y == newHead.Y {
			g.gameOver = true
			g.state = GameOver
			g.screenShake.Trigger(5, 20)
			g.spawnParticles(float32(newHead.X*tileSize+tileSize/2), float32(newHead.Y*tileSize+tileSize/2), 20, color.RGBA{255, 100, 100, 255}, 3)
			return nil
		}
	}

	g.snake = append([]Point{newHead}, g.snake...)

	if newHead.X == g.food.X && newHead.Y == g.food.Y {
		g.score++
		g.placeFood()
		if g.score%2 == 0 {
			g.spawnEnemy()
		}
		// Частицы при поедании еды
		g.spawnParticles(float32(g.food.X*tileSize+tileSize/2), float32(g.food.Y*tileSize+tileSize/2), 10, color.RGBA{255, 100, 0, 255}, 2)
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}

	g.enemyTimer++
	if g.enemyTimer >= g.enemyDelay {
		g.enemyTimer = 0
		g.updateEnemies()
	}

	g.bombTimer++
	if g.bombTimer >= g.bombDelay {
		g.bombTimer = 0
		g.spawnBomb()
	}

	g.coinTimer++
	if g.coinTimer >= g.coinDelay {
		g.coinTimer = 0
		g.spawnCoin()
	}

	g.updateBombs()

	if g.key != nil && newHead.X == g.key.pos.X && newHead.Y == g.key.pos.Y {
		g.hasKey = true
		g.key = nil
		g.spawnParticles(float32(g.key.pos.X*tileSize+tileSize/2), float32(g.key.pos.Y*tileSize+tileSize/2), 15, color.RGBA{255, 215, 0, 255}, float32(2.5))
	}

	for i := range g.coins {
		if !g.coins[i].collected && newHead.X == g.coins[i].pos.X && newHead.Y == g.coins[i].pos.Y {
			g.coins[i].collected = true
			g.score += g.coins[i].value
			g.spawnParticles(float32(g.coins[i].pos.X*tileSize+tileSize/2), float32(g.coins[i].pos.Y*tileSize+tileSize/2), 15, color.RGBA{255, 215, 0, 255}, float32(2.5))
		}
	}
	for i := len(g.coins) - 1; i >= 0; i-- {
		if g.coins[i].collected {
			g.coins = append(g.coins[:i], g.coins[i+1:]...)
		}
	}

	if g.chest != nil && !g.chest.open && newHead.X == g.chest.pos.X && newHead.Y == g.chest.pos.Y {
		if g.hasKey {
			g.chest.open = true
			g.arrowCount += g.chest.arrows
			g.hasKey = false
			g.spawnParticles(float32(g.chest.pos.X*tileSize+tileSize/2), float32(g.chest.pos.Y*tileSize+tileSize/2), 20, color.RGBA{255, 215, 0, 255}, 3)
		}
	}

	g.updateArrows()
	g.updateParticles()

	return nil
}

func (g *Game) updateEnemies() {
	for i := range g.enemies {
		enemy := &g.enemies[i]
		enemy.animFrame++

		newPos := enemy.pos
		switch enemy.direction {
		case Up:
			newPos.Y--
		case Down:
			newPos.Y++
		case Left:
			newPos.X--
		case Right:
			newPos.X++
		}

		if newPos.X < 0 || newPos.X >= gridSizeX || newPos.Y < 0 || newPos.Y >= gridSizeY {
			enemy.direction = Direction(rand.Intn(4))
			continue
		}

		enemy.pos = newPos

		if rand.Intn(10) < 2 {
			enemy.direction = Direction(rand.Intn(4))
		}

		for _, segment := range g.snake {
			if segment.X == enemy.pos.X && segment.Y == enemy.pos.Y {
				g.gameOver = true
				g.state = GameOver
				g.screenShake.Trigger(8, 25) // Сильная тряска при столкновении с врагом
				g.spawnParticles(float32(enemy.pos.X*tileSize+tileSize/2), float32(enemy.pos.Y*tileSize+tileSize/2), 30, color.RGBA{128, 0, 128, 255}, 4)
				return
			}
		}
	}
}

func (g *Game) updateBombs() {
	for i := len(g.bombs) - 1; i >= 0; i-- {
		bomb := &g.bombs[i]
		bomb.timer++

		for _, segment := range g.snake {
			if segment.X == bomb.pos.X && segment.Y == bomb.pos.Y {
				g.gameOver = true
				g.state = GameOver
				g.screenShake.Trigger(5, 20)
				return
			}
		}

		if bomb.timer >= bomb.maxTime {
			// Взрыв бомбы
			g.screenShake.Trigger(10, 30) // Очень сильная тряска при взрыве
			g.spawnParticles(float32(bomb.pos.X*tileSize+tileSize/2), float32(bomb.pos.Y*tileSize+tileSize/2), 40, color.RGBA{255, 100, 0, 255}, 5)
			
			for _, segment := range g.snake {
				dx := segment.X - bomb.pos.X
				if dx < 0 {
					dx = -dx
				}
				dy := segment.Y - bomb.pos.Y
				if dy < 0 {
					dy = -dy
				}
				if dx < 3 && dy < 3 {
					g.gameOver = true
					g.state = GameOver
					return
				}
			}
			g.bombs = append(g.bombs[:i], g.bombs[i+1:]...)
		}
	}
}

func (g *Game) shootArrow() {
	head := g.snake[0]
	arrow := Arrow{
		pos:       head,
		direction: g.direction,
		active:    true,
		speed:     0,
	}
	g.arrows = append(g.arrows, arrow)
	g.arrowCount--
}

func (g *Game) updateArrows() {
	for i := len(g.arrows) - 1; i >= 0; i-- {
		arrow := &g.arrows[i]
		if !arrow.active {
			g.arrows = append(g.arrows[:i], g.arrows[i+1:]...)
			continue
		}

		arrow.speed++
		if arrow.speed < 3 {
			continue
		}
		arrow.speed = 0

		switch arrow.direction {
		case Up:
			arrow.pos.Y--
		case Down:
			arrow.pos.Y++
		case Left:
			arrow.pos.X--
		case Right:
			arrow.pos.X++
		}

		if arrow.pos.X < 0 || arrow.pos.X >= gridSizeX || arrow.pos.Y < 0 || arrow.pos.Y >= gridSizeY {
			arrow.active = false
			continue
		}

		for j := len(g.enemies) - 1; j >= 0; j-- {
			enemy := &g.enemies[j]
			if arrow.pos.X == enemy.pos.X && arrow.pos.Y == enemy.pos.Y {
				g.enemies = append(g.enemies[:j], g.enemies[j+1:]...)
				arrow.active = false
				g.score += 1
				g.spawnParticles(float32(enemy.pos.X*tileSize+tileSize/2), float32(enemy.pos.Y*tileSize+tileSize/2), 25, color.RGBA{128, 0, 128, 255}, 4)
				break
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Применяем тряску экрана
	dx, dy := g.screenShake.GetOffset()

	// Рисуем градиентный фон
	screen.DrawImage(g.backgroundGradient, nil)

	// Создаём временную поверхность для игры со смещением
	gameScreen := ebiten.NewImage(screenWidth, screenHeight)

	switch g.state {
	case Menu:
		g.drawMenu(gameScreen)
		g.drawParticles(gameScreen)
		screen.DrawImage(gameScreen, nil)
		return
	case SelectDifficulty:
		g.drawDifficultySelection(gameScreen)
		g.drawParticles(gameScreen)
		screen.DrawImage(gameScreen, nil)
		return
	case Paused:
		g.drawGame(gameScreen)
		g.drawParticles(gameScreen)
		g.drawPauseOverlay(gameScreen)
		screen.DrawImage(gameScreen, nil)
		return
	case GameOver:
		g.drawGame(gameScreen)
		g.drawParticles(gameScreen)
		g.drawGameOverOverlay(gameScreen)
		screen.DrawImage(gameScreen, nil)
		return
	case Playing:
		g.drawGame(gameScreen)
		g.drawParticles(gameScreen)
	}

	// Рисуем игровую поверхность со смещением тряски
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dx), float64(dy))
	screen.DrawImage(gameScreen, op)
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	title := "SNAKE GAME"
	titleX := float32(screenWidth/2 - len(title)*10)
	titleY := float32(screenHeight/2 - 100)
	ebitenutil.DebugPrintAt(screen, title, int(titleX), int(titleY))

	subtitle := "Go365 Go79 - Ebitengine"
	subX := float32(screenWidth/2 - len(subtitle)*5)
	ebitenutil.DebugPrintAt(screen, subtitle, int(subX), int(titleY+40))

	startText := "Press ENTER or SPACE to Start"
	startX := float32(screenWidth/2 - len(startText)*6)
	ebitenutil.DebugPrintAt(screen, startText, int(startX), int(titleY+100))

	controls := []string{
		"Controls:",
		"Arrow Keys - Move",
		"SPACE - Shoot Arrow",
		"P - Pause",
		"",
		"Find the golden key and open the treasure chest!",
		"Collect arrows and shoot the bugs!",
	}
	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, screenWidth/2-150, int(titleY)+160+i*20)
	}
}

func (g *Game) drawDifficultySelection(screen *ebiten.Image) {
	title := "SNAKE GAME"
	titleX := float32(screenWidth/2 - len(title)*10)
	titleY := float32(screenHeight/2 - 150)
	ebitenutil.DebugPrintAt(screen, title, int(titleX), int(titleY))

	subtitle := "Go365 Go79 - Ebitengine"
	subX := float32(screenWidth/2 - len(subtitle)*5)
	ebitenutil.DebugPrintAt(screen, subtitle, int(subX), int(titleY+40))

	selectText := "Select Difficulty"
	selectX := float32(screenWidth/2 - len(selectText)*8)
	ebitenutil.DebugPrintAt(screen, selectText, int(selectX), int(titleY+100))

	difficulties := []struct {
		name        string
		enemyCount  int
	}{
		{"Easy", 2},
		{"Medium", 3},
		{"Hard", 5},
	}

	for i, diff := range difficulties {
		y := int(titleY) + 160 + i*40
		marker := "  "
		prefix := "  "

		if Difficulty(i) == g.difficulty {
			marker = ">> "
			prefix = "<<"
			highlight := fmt.Sprintf("%s%s - %d bugs %s", marker, diff.name, diff.enemyCount, prefix)
			ebitenutil.DebugPrintAt(screen, highlight, screenWidth/2-100, y)
		} else {
			text := fmt.Sprintf("%s%s - %d bugs %s", prefix, diff.name, diff.enemyCount, prefix)
			ebitenutil.DebugPrintAt(screen, text, screenWidth/2-100, y)
		}
	}

	controls := []string{
		"",
		"UP/DOWN - Change difficulty",
		"ENTER/SPACE - Start game",
	}
	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, screenWidth/2-120, int(titleY)+300+i*20)
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	vector.StrokeRect(screen, 0, 0, screenWidth, screenHeight, 2, color.RGBA{100, 100, 100, 255}, false)

	for i, segment := range g.snake {
		green := color.RGBA{0, 255, 0, 255}
		if i == 0 {
			green = color.RGBA{100, 255, 100, 255}
		}
		vector.DrawFilledRect(
			screen,
			float32(segment.X*tileSize),
			float32(segment.Y*tileSize),
			tileSize,
			tileSize,
			green,
			false,
		)
		if i == 0 {
			g.drawSnakeEyes(screen, segment, g.direction)
			g.drawSnakeTongue(screen, segment, g.direction)
		}
	}

	g.drawFood(screen)

	for _, enemy := range g.enemies {
		g.drawEnemy(screen, enemy)
	}

	for _, bomb := range g.bombs {
		g.drawBomb(screen, bomb)
	}

	if g.chest != nil {
		g.drawChest(screen, *g.chest)
	}

	if g.key != nil {
		g.drawKey(screen, *g.key)
	}

	for _, coin := range g.coins {
		g.drawCoin(screen, coin)
	}

	for _, arrow := range g.arrows {
		g.drawArrow(screen, arrow)
	}

	ebitenutil.DebugPrintAt(screen, "Score: "+string(rune('0'+g.score)), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Arrows: %d", g.arrowCount), 10, 25)
	if g.hasKey {
		ebitenutil.DebugPrintAt(screen, "KEY!", 10, 40)
	}
	if len(g.coins) > 0 {
		ebitenutil.DebugPrintAt(screen, "x2 XP COINS!", 10, 55)
	}
}

func (g *Game) drawPauseOverlay(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 128})

	pausedText := "PAUSED"
	pausedX := screenWidth/2 - len(pausedText)*8
	ebitenutil.DebugPrintAt(screen, pausedText, pausedX, screenHeight/2-50)

	continueText := "Press P to Continue"
	contX := screenWidth/2 - len(continueText)*6
	ebitenutil.DebugPrintAt(screen, continueText, contX, screenHeight/2)
}

func (g *Game) drawGameOverOverlay(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 0, 0, 180})

	gameOverText := "GAME OVER"
	gameOverX := screenWidth/2 - len(gameOverText)*8
	ebitenutil.DebugPrintAt(screen, gameOverText, gameOverX, screenHeight/2-80)

	scoreText := fmt.Sprintf("Final Score: %d", g.score)
	scoreX := screenWidth/2 - len(scoreText)*6
	ebitenutil.DebugPrintAt(screen, scoreText, scoreX, screenHeight/2-20)

	enemiesText := fmt.Sprintf("Enemies: %d", len(g.enemies))
	enemiesX := screenWidth/2 - len(enemiesText)*6
	ebitenutil.DebugPrintAt(screen, enemiesText, enemiesX, screenHeight/2+10)

	restartText := "Press ENTER to Restart"
	restartX := screenWidth/2 - len(restartText)*7
	ebitenutil.DebugPrintAt(screen, restartText, restartX, screenHeight/2+60)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) drawEnemy(screen *ebiten.Image, enemy Enemy) {
	x := float32(enemy.pos.X * tileSize)
	y := float32(enemy.pos.Y * tileSize)
	size := float32(tileSize) * 1.5

	vector.DrawFilledCircle(screen, x+size/2, y+size/2, size/2-2, color.RGBA{128, 0, 128, 255}, false)

	headX := x + size/2
	headY := y + size/2
	vector.DrawFilledCircle(screen, headX, headY, size/3, color.RGBA{100, 0, 100, 255}, false)

	legOffset := float32((enemy.animFrame % 20) / 10.0 * 3)
	if enemy.animFrame%40 < 20 {
		legOffset = -legOffset
	}

	for i := 0; i < 3; i++ {
		legY := y + size/4 + float32(i)*size/4
		vector.StrokeLine(screen, x+size/3, legY, x-size/4, legY+legOffset+float32(i)*2, 2, color.RGBA{100, 0, 100, 255}, false)
	}

	for i := 0; i < 3; i++ {
		legY := y + size/4 + float32(i)*size/4
		vector.StrokeLine(screen, x+2*size/3, legY, x+5*size/4, legY-legOffset+float32(i)*2, 2, color.RGBA{100, 0, 100, 255}, false)
	}

	antennaAngle := float32((enemy.animFrame % 30) / 30.0 * 1.0)
	if enemy.animFrame%60 < 30 {
		antennaAngle = -antennaAngle
	}

	vector.StrokeLine(screen, headX-size/6, headY-size/3, headX-size/2-antennaAngle*size, headY-size/2-antennaAngle*size, 1, color.RGBA{150, 50, 50, 255}, false)
	vector.StrokeLine(screen, headX+size/6, headY-size/3, headX+size/2+antennaAngle*size, headY-size/2-antennaAngle*size, 1, color.RGBA{150, 50, 50, 255}, false)

	mouthX := headX
	mouthY := headY + size/8
	vector.DrawFilledCircle(screen, mouthX, mouthY, size/8, color.RGBA{50, 0, 0, 255}, false)

	eyeSize := size / 5
	leftEyeX := headX - size/6
	leftEyeY := headY - size/8
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize+2, color.RGBA{255, 50, 0, 100}, false)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize, color.RGBA{255, 100, 0, 180}, false)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize-2, color.RGBA{255, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize/3, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, leftEyeX-eyeSize/4, leftEyeY-eyeSize/4, eyeSize/5, color.RGBA{255, 255, 255, 255}, false)

	rightEyeX := headX + size/6
	rightEyeY := headY - size/8
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize+2, color.RGBA{255, 50, 0, 100}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize, color.RGBA{255, 100, 0, 180}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize-2, color.RGBA{255, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize/3, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX+eyeSize/4, rightEyeY-eyeSize/4, eyeSize/5, color.RGBA{255, 255, 255, 255}, false)
}

func (g *Game) drawSnakeEyes(screen *ebiten.Image, head Point, direction Direction) {
	x := float32(head.X * tileSize)
	y := float32(head.Y * tileSize)
	size := float32(tileSize)
	eyeSize := size / 6
	pupilSize := eyeSize / 2

	var leftEyeX, leftEyeY, rightEyeX, rightEyeY float32

	switch direction {
	case Up:
		leftEyeX = x + size/3
		leftEyeY = y + size/3
		rightEyeX = x + 2*size/3
		rightEyeY = y + size/3
	case Down:
		leftEyeX = x + size/3
		leftEyeY = y + 2*size/3
		rightEyeX = x + 2*size/3
		rightEyeY = y + 2*size/3
	case Left:
		leftEyeX = x + size/3
		leftEyeY = y + size/3
		rightEyeX = x + size/3
		rightEyeY = y + 2*size/3
	case Right:
		leftEyeX = x + 2*size/3
		leftEyeY = y + size/3
		rightEyeX = x + 2*size/3
		rightEyeY = y + 2*size/3
	}

	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, pupilSize, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, pupilSize, color.RGBA{0, 0, 0, 255}, false)
}

func (g *Game) drawSnakeTongue(screen *ebiten.Image, head Point, direction Direction) {
	x := float32(head.X * tileSize)
	y := float32(head.Y * tileSize)
	size := float32(tileSize)

	tongueColor := color.RGBA{255, 50, 50, 255}
	tongueLength := size / 2
	tongueWidth := size / 12

	var startX, startY, endX, endY float32

	switch direction {
	case Up:
		startX = x + size/2
		startY = y + size/4
		endX = x + size/2
		endY = y - tongueLength
	case Down:
		startX = x + size/2
		startY = y + 3*size/4
		endX = x + size/2
		endY = y + size + tongueLength
	case Left:
		startX = x + size/4
		startY = y + size/2
		endX = x - tongueLength
		endY = y + size/2
	case Right:
		startX = x + 3*size/4
		startY = y + size/2
		endX = x + size + tongueLength
		endY = y + size/2
	}

	vector.StrokeLine(screen, startX, startY, endX, endY, tongueWidth, tongueColor, false)

	forkLength := size / 6
	var leftForkX, leftForkY, rightForkX, rightForkY float32

	switch direction {
	case Up:
		leftForkX = endX - forkLength/2
		leftForkY = endY + forkLength/2
		rightForkX = endX + forkLength/2
		rightForkY = endY + forkLength/2
	case Down:
		leftForkX = endX - forkLength/2
		leftForkY = endY - forkLength/2
		rightForkX = endX + forkLength/2
		rightForkY = endY - forkLength/2
	case Left:
		leftForkX = endX + forkLength/2
		leftForkY = endY - forkLength/2
		rightForkX = endX + forkLength/2
		rightForkY = endY + forkLength/2
	case Right:
		leftForkX = endX - forkLength/2
		leftForkY = endY - forkLength/2
		rightForkX = endX - forkLength/2
		rightForkY = endY + forkLength/2
	}

	vector.StrokeLine(screen, endX, endY, leftForkX, leftForkY, tongueWidth, tongueColor, false)
	vector.StrokeLine(screen, endX, endY, rightForkX, rightForkY, tongueWidth, tongueColor, false)
}

func (g *Game) drawFood(screen *ebiten.Image) {
	x := float32(g.food.X * tileSize)
	y := float32(g.food.Y * tileSize)
	size := float32(tileSize)

	centerX := x + size/2
	centerY := y + size/2 + 2
	radius := size/2 - 3

	// Анимация появления (пульсация в начале)
	pulseScale := 1.0
	if g.foodTimer > 0 {
		pulseScale = 1.0 + 0.3*math.Sin(float64(g.foodTimer)*math.Pi/10)
		g.foodTimer--
	}

	// Main red body
	vector.DrawFilledCircle(screen, centerX, centerY, radius*float32(pulseScale), color.RGBA{255, 0, 0, 255}, false)

	highlightX := centerX - radius/3
	highlightY := centerY - radius/3
	vector.DrawFilledCircle(screen, highlightX, highlightY, radius/3*float32(pulseScale), color.RGBA{255, 100, 100, 255}, false)

	vector.DrawFilledCircle(screen, centerX, centerY-radius+2, 2, color.RGBA{200, 0, 0, 255}, false)

	stemX := centerX
	stemY := centerY - radius
	vector.StrokeLine(screen, stemX, stemY, stemX, stemY-4, 2, color.RGBA{139, 69, 19, 255}, false)

	leafColor := color.RGBA{34, 139, 34, 255}
	leafBaseX := centerX + 1
	leafBaseY := stemY - 2
	
	// Leaf tip and bottom positions
	tipX := centerX + 5
	tipY := leafBaseY - 3
	botX := centerX + 2
	botY := leafBaseY + 2

	vector.DrawFilledRect(screen, leafBaseX, leafBaseY-1, 5, 2, leafColor, false)
	vector.DrawFilledCircle(screen, tipX-1, tipY, 2, leafColor, false)
	vector.DrawFilledCircle(screen, botX, botY, 2, leafColor, false)

	vector.StrokeLine(screen, leafBaseX+1, leafBaseY, tipX-1, tipY, 1, color.RGBA{100, 200, 100, 255}, false)
}

func (g *Game) drawCoin(screen *ebiten.Image, coin Coin) {
	x := float32(coin.pos.X * tileSize)
	y := float32(coin.pos.Y * tileSize)
	size := float32(tileSize)

	coinRadius := size/2 - 4
	centerX := x + size/2
	centerY := y + size/2

	// Пульсация монеты
	pulseScale := 1.0 + 0.1*math.Sin(coin.pulsePhase)
	
	vector.DrawFilledCircle(screen, centerX, centerY, coinRadius*float32(pulseScale), color.RGBA{255, 215, 0, 255}, false)

	innerRadius := coinRadius - 3
	vector.DrawFilledCircle(screen, centerX, centerY, innerRadius*float32(pulseScale), color.RGBA{255, 235, 100, 255}, false)

	dotRadius := innerRadius / 2
	vector.DrawFilledCircle(screen, centerX, centerY, dotRadius*float32(pulseScale), color.RGBA{255, 200, 0, 255}, false)

	sparkleOffset := coinRadius - 1
	vector.DrawFilledCircle(screen, centerX, centerY-sparkleOffset*float32(pulseScale), 1, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, centerX, centerY+sparkleOffset*float32(pulseScale), 1, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, centerX-sparkleOffset*float32(pulseScale), centerY, 1, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, centerX+sparkleOffset*float32(pulseScale), centerY, 1, color.RGBA{255, 255, 255, 255}, false)

	// Анимированное свечение
	glowPhase := (time.Now().UnixMilli() / 200) % 2
	glowIntensity := uint8(100 + glowPhase*50)
	vector.DrawFilledCircle(screen, centerX, centerY, coinRadius*float32(pulseScale)+2, color.RGBA{255, 215, 0, glowIntensity}, false)
}

func (g *Game) drawBomb(screen *ebiten.Image, bomb Bomb) {
	x := float32(bomb.pos.X * tileSize)
	y := float32(bomb.pos.Y * tileSize)
	size := float32(tileSize)

	vector.DrawFilledCircle(screen, x+size/2, y+size/2, size/2-2, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, x+size/3, y+size/3, size/6, color.RGBA{50, 50, 50, 255}, false)

	fuseX := x + size/2
	fuseY := y + size/4
	vector.StrokeLine(screen, fuseX, fuseY, fuseX, fuseY-size/3, 2, color.RGBA{139, 69, 19, 255}, false)

	// Мигание искры перед взрывом
	blinkPhase := float32(bomb.timer%5) * 0.4
	sparkSize := size/6 + float32(blinkPhase)*size/8

	vector.DrawFilledCircle(screen, fuseX, fuseY-size/3, sparkSize, color.RGBA{255, 200, 0, 200}, false)
	vector.DrawFilledCircle(screen, fuseX, fuseY-size/3, sparkSize/2, color.RGBA{255, 255, 255, 255}, false)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		particleX := fuseX + float32(rand.Intn(8)-4)
		particleY := fuseY - size/3 + float32(rand.Intn(8)-4)
		vector.DrawFilledCircle(screen, particleX, particleY, 1, color.RGBA{255, 100, 0, 255}, false)
	}
}

func (g *Game) drawChest(screen *ebiten.Image, chest TreasureChest) {
	x := float32(chest.pos.X * tileSize)
	y := float32(chest.pos.Y * tileSize)
	size := float32(tileSize)

	chestColor := color.RGBA{139, 69, 19, 255}
	if chest.open {
		chestColor = color.RGBA{100, 50, 10, 255}
	}
	vector.DrawFilledRect(screen, x+2, y+4, size-4, size-6, chestColor, false)

	lidColor := color.RGBA{255, 215, 0, 255}
	if chest.open {
		vector.StrokeLine(screen, x+2, y+4, x+size-2, y+4, 2, lidColor, false)
	} else {
		vector.DrawFilledRect(screen, x+2, y+2, size-4, size/3, lidColor, false)
	}

	if !chest.open {
		vector.DrawFilledCircle(screen, x+size/2, y+size/2, 3, color.RGBA{255, 215, 0, 255}, false)
	}
}

func (g *Game) drawKey(screen *ebiten.Image, key Key) {
	x := float32(key.pos.X * tileSize)
	y := float32(key.pos.Y * tileSize)
	size := float32(tileSize)

	keyColor := color.RGBA{255, 215, 0, 255}

	headSize := size / 3
	vector.DrawFilledCircle(screen, x+size/2, y+size/3, headSize, keyColor, false)

	shaftWidth := size / 8
	shaftHeight := size / 2
	vector.DrawFilledRect(screen, x+size/2-shaftWidth/2, y+size/2, shaftWidth, shaftHeight, keyColor, false)

	toothSize := size / 6
	vector.DrawFilledRect(screen, x+size/2-shaftWidth/2, y+size/2+shaftHeight-toothSize, shaftWidth, toothSize, keyColor, false)
	vector.DrawFilledRect(screen, x+size/2, y+size/2+shaftHeight-toothSize, shaftWidth, toothSize/2, keyColor, false)
}

func (g *Game) drawArrow(screen *ebiten.Image, arrow Arrow) {
	x := float32(arrow.pos.X * tileSize)
	y := float32(arrow.pos.Y * tileSize)
	size := float32(tileSize)

	arrowColor := color.RGBA{192, 192, 192, 255}

	shaftLength := size / 2
	shaftWidth := float32(2)

	var startX, startY, endX, endY float32

	switch arrow.direction {
	case Up:
		startX = x + size/2
		startY = y + size/2
		endX = x + size/2
		endY = y + size/2 - shaftLength
	case Down:
		startX = x + size/2
		startY = y + size/2
		endX = x + size/2
		endY = y + size/2 + shaftLength
	case Left:
		startX = x + size/2
		startY = y + size/2
		endX = x + size/2 - shaftLength
		endY = y + size/2
	case Right:
		startX = x + size/2
		startY = y + size/2
		endX = x + size/2 + shaftLength
		endY = y + size/2
	}

	vector.StrokeLine(screen, startX, startY, endX, endY, shaftWidth, arrowColor, false)

	headSize := size / 6
	var headX1, headY1, headX2, headY2, headX3, headY3 float32

	switch arrow.direction {
	case Up:
		headX1 = endX
		headY1 = endY
		headX2 = endX - headSize/2
		headY2 = endY + headSize
		headX3 = endX + headSize/2
		headY3 = endY + headSize
	case Down:
		headX1 = endX
		headY1 = endY
		headX2 = endX - headSize/2
		headY2 = endY - headSize
		headX3 = endX + headSize/2
		headY3 = endY - headSize
	case Left:
		headX1 = endX
		headY1 = endY
		headX2 = endX + headSize
		headY2 = endY - headSize/2
		headX3 = endX + headSize
		headY3 = endY + headSize/2
	case Right:
		headX1 = endX
		headY1 = endY
		headX2 = endX - headSize
		headY2 = endY - headSize/2
		headX3 = endX - headSize
		headY3 = endY + headSize/2
	}

	vector.StrokeLine(screen, headX1, headY1, headX2, headY2, shaftWidth, arrowColor, false)
	vector.StrokeLine(screen, headX2, headY2, headX3, headY3, shaftWidth, arrowColor, false)
	vector.StrokeLine(screen, headX3, headY3, headX1, headY1, shaftWidth, arrowColor, false)
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Simple Snake - Go365 Go79")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
