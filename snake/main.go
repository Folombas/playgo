package main

import (
	"fmt"
	"image/color"
	"log"
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
}

type Arrow struct {
	pos       Point
	direction Direction
	active    bool
	speed     int
}

func NewGame() *Game {
	g := &Game{
		snake:     []Point{{gridSizeX / 2, gridSizeY / 2}, {gridSizeX/2 - 1, gridSizeY / 2}, {gridSizeX/2 - 2, gridSizeY / 2}},
		direction: Right,
		score:     0,
		gameOver:  false,
		moveDelay: 8,   // скорость движения змейки
		enemyDelay: 12,  // скорость врагов (медленнее змейки)
		bombDelay: 180,  // спавн бомбы каждые 180 тиков (~3 сек)
		coinDelay: 300,  // спавн монетки каждые 300 тиков (~5 сек)
		enemies:   []Enemy{},
		bombs:     []Bomb{},
		coins:     []Coin{},
		arrows:    []Arrow{},
		hasKey:    false,
		arrowCount: 0,
		state:     Menu,
	}
	return g
}

func (g *Game) startGame() {
	g.placeFood()
	g.spawnChest()
	g.spawnKey()
	// Spawn enemies based on difficulty
	enemyCount := g.difficulty.EnemyCount()
	for i := 0; i < enemyCount; i++ {
		g.spawnEnemy()
	}
	g.state = Playing
}

func (g *Game) placeFood() {
	rand.Seed(time.Now().UnixNano())
	for {
		g.food = Point{
			X: rand.Intn(gridSizeX),
			Y: rand.Intn(gridSizeY),
		}
		// Check if food is not on snake
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
}

func (g *Game) spawnEnemy() {
	rand.Seed(time.Now().UnixNano())
	for {
		pos := Point{
			X: rand.Intn(gridSizeX),
			Y: rand.Intn(gridSizeY),
		}
		// Don't spawn on snake or too close to player
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
		// Don't spawn on snake
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
			g.bombs = append(g.bombs, Bomb{pos: pos, timer: 0, maxTime: 180}) // 3 секунды до взрыва
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
		// Don't spawn too close to snake
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
		// Don't spawn too close to snake
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
		// Don't spawn too close to snake
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
			g.coins = append(g.coins, Coin{pos: pos, value: 2, collected: false})
			break
		}
	}
}

func (g *Game) Update() error {
	switch g.state {
	case Menu:
		// Go to difficulty selection with Enter or Space
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = SelectDifficulty
		}
		return nil

	case SelectDifficulty:
		// Navigate difficulty with Up/Down arrows
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
		// Confirm difficulty with Enter or Space
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.startGame()
		}
		return nil

	case Paused:
		// Unpause with P
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = Playing
		}
		return nil

	case GameOver:
		// Restart with Enter
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			*g = *NewGame()
		}
		return nil

	case Playing:
		// Pause with P
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = Paused
			return nil
		}
	}

	// Handle input
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && g.direction != Down {
		g.direction = Up
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && g.direction != Up {
		g.direction = Down
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && g.direction != Right {
		g.direction = Left
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && g.direction != Left {
		g.direction = Right
	}

	// Shoot arrow with Space key
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && g.arrowCount > 0 {
		g.shootArrow()
	}

	// Update move timer
	g.moveTimer++
	if g.moveTimer < g.moveDelay {
		return nil
	}
	g.moveTimer = 0

	// Move snake
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

	// Check wall collision
	if newHead.X < 0 || newHead.X >= gridSizeX || newHead.Y < 0 || newHead.Y >= gridSizeY {
		g.gameOver = true
		g.state = GameOver
		return nil
	}

	// Check self collision
	for _, segment := range g.snake {
		if segment.X == newHead.X && segment.Y == newHead.Y {
			g.gameOver = true
			g.state = GameOver
			return nil
		}
	}

	// Add new head
	g.snake = append([]Point{newHead}, g.snake...)

	// Check food collision
	if newHead.X == g.food.X && newHead.Y == g.food.Y {
		g.score++
		g.placeFood()
		// Spawn new enemy every 2 points
		if g.score%2 == 0 {
			g.spawnEnemy()
		}
	} else {
		// Remove tail
		g.snake = g.snake[:len(g.snake)-1]
	}

	// Update enemies
	g.enemyTimer++
	if g.enemyTimer >= g.enemyDelay {
		g.enemyTimer = 0
		g.updateEnemies()
	}

	// Spawn bombs periodically
	g.bombTimer++
	if g.bombTimer >= g.bombDelay {
		g.bombTimer = 0
		g.spawnBomb()
	}

	// Spawn coins periodically
	g.coinTimer++
	if g.coinTimer >= g.coinDelay {
		g.coinTimer = 0
		g.spawnCoin()
	}

	// Update bombs
	g.updateBombs()

	// Check key collision
	if g.key != nil && newHead.X == g.key.pos.X && newHead.Y == g.key.pos.Y {
		g.hasKey = true
		g.key = nil
	}

	// Check coin collision
	for i := range g.coins {
		if !g.coins[i].collected && newHead.X == g.coins[i].pos.X && newHead.Y == g.coins[i].pos.Y {
			g.coins[i].collected = true
			// x2 XP bonus for next food collection
			g.score += g.coins[i].value
		}
	}
	// Remove collected coins
	for i := len(g.coins) - 1; i >= 0; i-- {
		if g.coins[i].collected {
			g.coins = append(g.coins[:i], g.coins[i+1:]...)
		}
	}

	// Check chest collision
	if g.chest != nil && !g.chest.open && newHead.X == g.chest.pos.X && newHead.Y == g.chest.pos.Y {
		if g.hasKey {
			g.chest.open = true
			g.arrowCount += g.chest.arrows
			g.hasKey = false
		}
	}

	// Update arrows
	g.updateArrows()

	return nil
}

func (g *Game) updateEnemies() {
	for i := range g.enemies {
		enemy := &g.enemies[i]
		enemy.animFrame++

		// Move enemy
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

		// Check bounds - reverse direction if hitting wall
		if newPos.X < 0 || newPos.X >= gridSizeX || newPos.Y < 0 || newPos.Y >= gridSizeY {
			enemy.direction = Direction(rand.Intn(4))
			continue
		}

		enemy.pos = newPos

		// Random direction change
		if rand.Intn(10) < 2 {
			enemy.direction = Direction(rand.Intn(4))
		}

		// Check collision with snake
		for _, segment := range g.snake {
			if segment.X == enemy.pos.X && segment.Y == enemy.pos.Y {
				g.gameOver = true
				g.state = GameOver
				return
			}
		}
	}
}

func (g *Game) updateBombs() {
	for i := len(g.bombs) - 1; i >= 0; i-- {
		bomb := &g.bombs[i]
		bomb.timer++

		// Check collision with snake
		for _, segment := range g.snake {
			if segment.X == bomb.pos.X && segment.Y == bomb.pos.Y {
				g.gameOver = true
				g.state = GameOver
				return
			}
		}

		// Bomb explodes after maxTime
		if bomb.timer >= bomb.maxTime {
			// Check if snake is near explosion
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
			// Remove exploded bomb
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

		// Move arrow
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

		// Check bounds
		if arrow.pos.X < 0 || arrow.pos.X >= gridSizeX || arrow.pos.Y < 0 || arrow.pos.Y >= gridSizeY {
			arrow.active = false
			continue
		}

		// Check collision with enemies
		for j := len(g.enemies) - 1; j >= 0; j-- {
			enemy := &g.enemies[j]
			if arrow.pos.X == enemy.pos.X && arrow.pos.Y == enemy.pos.Y {
				g.enemies = append(g.enemies[:j], g.enemies[j+1:]...)
				arrow.active = false
				g.score += 1 // Bonus for killing enemy
				break
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw based on game state
	switch g.state {
	case Menu:
		g.drawMenu(screen)
		return
	case SelectDifficulty:
		g.drawDifficultySelection(screen)
		return
	case Paused:
		// Draw game paused
		g.drawGame(screen)
		g.drawPauseOverlay(screen)
		return
	case GameOver:
		// Draw game over
		g.drawGame(screen)
		g.drawGameOverOverlay(screen)
		return
	case Playing:
		g.drawGame(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Title
	title := "SNAKE GAME"
	titleX := float32(screenWidth/2 - len(title)*10)
	titleY := float32(screenHeight/2 - 100)
	ebitenutil.DebugPrintAt(screen, title, int(titleX), int(titleY))

	// Subtitle
	subtitle := "Go365 Go76 - Ebitengine"
	subX := float32(screenWidth/2 - len(subtitle)*5)
	ebitenutil.DebugPrintAt(screen, subtitle, int(subX), int(titleY+40))

	// Start button prompt
	startText := "Press ENTER or SPACE to Start"
	startX := float32(screenWidth/2 - len(startText)*6)
	ebitenutil.DebugPrintAt(screen, startText, int(startX), int(titleY+100))

	// Controls info
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
	// Title
	title := "SNAKE GAME"
	titleX := float32(screenWidth/2 - len(title)*10)
	titleY := float32(screenHeight/2 - 150)
	ebitenutil.DebugPrintAt(screen, title, int(titleX), int(titleY))

	// Subtitle
	subtitle := "Go365 Go76 - Ebitengine"
	subX := float32(screenWidth/2 - len(subtitle)*5)
	ebitenutil.DebugPrintAt(screen, subtitle, int(subX), int(titleY+40))

	// Select difficulty prompt
	selectText := "Select Difficulty"
	selectX := float32(screenWidth/2 - len(selectText)*8)
	ebitenutil.DebugPrintAt(screen, selectText, int(selectX), int(titleY+100))

	// Difficulty options with highlight
	difficulties := []struct {
		name        string
		enemyCount  int
		color       string
	}{
		{"Easy", 2, "Green"},
		{"Medium", 3, "Yellow"},
		{"Hard", 5, "Red"},
	}

	for i, diff := range difficulties {
		y := int(titleY) + 160 + i*40
		marker := "  "
		prefix := "  "
		
		if Difficulty(i) == g.difficulty {
			marker = ">> "
			prefix = "<< "
			// Highlight current selection
			highlight := fmt.Sprintf("%s%s - %d bugs %s", marker, diff.name, diff.enemyCount, prefix)
			ebitenutil.DebugPrintAt(screen, highlight, screenWidth/2-100, y)
		} else {
			text := fmt.Sprintf("%s%s - %d bugs %s", prefix, diff.name, diff.enemyCount, prefix)
			ebitenutil.DebugPrintAt(screen, text, screenWidth/2-100, y)
		}
	}

	// Controls info
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
	// Draw border around play area
	vector.StrokeRect(screen, 0, 0, screenWidth, screenHeight, 2, color.RGBA{100, 100, 100, 255}, false)

	// Draw snake
	for i, segment := range g.snake {
		green := color.RGBA{0, 255, 0, 255}
		if i == 0 {
			// Head is brighter
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
		// Draw eyes on head
		if i == 0 {
			g.drawSnakeEyes(screen, segment, g.direction)
			g.drawSnakeTongue(screen, segment, g.direction)
		}
	}

	// Draw food
	g.drawFood(screen)

	// Draw enemies (bugs with legs and antennae)
	for _, enemy := range g.enemies {
		g.drawEnemy(screen, enemy)
	}

	// Draw bombs
	for _, bomb := range g.bombs {
		g.drawBomb(screen, bomb)
	}

	// Draw treasure chest
	if g.chest != nil {
		g.drawChest(screen, *g.chest)
	}

	// Draw key
	if g.key != nil {
		g.drawKey(screen, *g.key)
	}

	// Draw coins
	for _, coin := range g.coins {
		g.drawCoin(screen, coin)
	}

	// Draw arrows
	for _, arrow := range g.arrows {
		g.drawArrow(screen, arrow)
	}

	// Draw score
	ebitenutil.DebugPrintAt(screen, "Score: "+string(rune('0'+g.score)), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Arrows: %d", g.arrowCount), 10, 25)
	if g.hasKey {
		ebitenutil.DebugPrintAt(screen, "KEY!", 10, 40)
	}
	// Show coin bonus indicator
	if len(g.coins) > 0 {
		ebitenutil.DebugPrintAt(screen, "x2 XP COINS!", 10, 55)
	}
}

func (g *Game) drawPauseOverlay(screen *ebiten.Image) {
	// Semi-transparent overlay
	screen.Fill(color.RGBA{0, 0, 0, 128})

	// PAUSED text
	pausedText := "PAUSED"
	pausedX := screenWidth/2 - len(pausedText)*8
	ebitenutil.DebugPrintAt(screen, pausedText, pausedX, screenHeight/2-50)

	// Continue prompt
	continueText := "Press P to Continue"
	contX := screenWidth/2 - len(continueText)*6
	ebitenutil.DebugPrintAt(screen, continueText, contX, screenHeight/2)
}

func (g *Game) drawGameOverOverlay(screen *ebiten.Image) {
	// Semi-transparent overlay
	screen.Fill(color.RGBA{50, 0, 0, 180})

	// GAME OVER text
	gameOverText := "GAME OVER"
	gameOverX := screenWidth/2 - len(gameOverText)*8
	ebitenutil.DebugPrintAt(screen, gameOverText, gameOverX, screenHeight/2-80)

	// Final score
	scoreText := fmt.Sprintf("Final Score: %d", g.score)
	scoreX := screenWidth/2 - len(scoreText)*6
	ebitenutil.DebugPrintAt(screen, scoreText, scoreX, screenHeight/2-20)

	// Enemies killed
	enemiesText := fmt.Sprintf("Enemies: %d", len(g.enemies))
	enemiesX := screenWidth/2 - len(enemiesText)*6
	ebitenutil.DebugPrintAt(screen, enemiesText, enemiesX, screenHeight/2+10)

	// Restart prompt
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
	size := float32(tileSize) * 1.5 // Increase bug size by 1.5x

	// Body (dark purple oval)
	vector.DrawFilledCircle(screen, x+size/2, y+size/2, size/2-2, color.RGBA{128, 0, 128, 255}, false)

	// Head
	headX := x + size/2
	headY := y + size/2
	vector.DrawFilledCircle(screen, headX, headY, size/3, color.RGBA{100, 0, 100, 255}, false)

	// Animated legs (6 legs - 3 on each side)
	legOffset := float32((enemy.animFrame % 20) / 10.0 * 3)
	if enemy.animFrame%40 < 20 {
		legOffset = -legOffset
	}

	// Left legs
	for i := 0; i < 3; i++ {
		legY := y + size/4 + float32(i)*size/4
		vector.StrokeLine(screen, x+size/3, legY, x-size/4, legY+legOffset+float32(i)*2, 2, color.RGBA{100, 0, 100, 255}, false)
	}

	// Right legs
	for i := 0; i < 3; i++ {
		legY := y + size/4 + float32(i)*size/4
		vector.StrokeLine(screen, x+2*size/3, legY, x+5*size/4, legY-legOffset+float32(i)*2, 2, color.RGBA{100, 0, 100, 255}, false)
	}

	// Antennae (animated)
	antennaAngle := float32((enemy.animFrame % 30) / 30.0 * 1.0)
	if enemy.animFrame%60 < 30 {
		antennaAngle = -antennaAngle
	}

	// Left antenna
	vector.StrokeLine(screen, headX-size/6, headY-size/3, headX-size/2-antennaAngle*size, headY-size/2-antennaAngle*size, 1, color.RGBA{150, 50, 50, 255}, false)
	// Right antenna
	vector.StrokeLine(screen, headX+size/6, headY-size/3, headX+size/2+antennaAngle*size, headY-size/2-antennaAngle*size, 1, color.RGBA{150, 50, 50, 255}, false)

	// Menacing mouth with two front teeth
	mouthX := headX
	mouthY := headY + size/8

	// Dark mouth opening (circle)
	vector.DrawFilledCircle(screen, mouthX, mouthY, size/8, color.RGBA{50, 0, 0, 255}, false)

	// Big scary glowing eyes (red with glow effect)
	eyeSize := size / 5
	// Left eye glow (multiple layers for glow effect)
	leftEyeX := headX - size/6
	leftEyeY := headY - size/8
	// Outer glow (red)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize+2, color.RGBA{255, 50, 0, 100}, false)
	// Middle glow
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize, color.RGBA{255, 100, 0, 180}, false)
	// Inner bright eye
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize-2, color.RGBA{255, 0, 0, 255}, false)
	// Pupil (black)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize/3, color.RGBA{0, 0, 0, 255}, false)
	// White highlight for scary look
	vector.DrawFilledCircle(screen, leftEyeX-eyeSize/4, leftEyeY-eyeSize/4, eyeSize/5, color.RGBA{255, 255, 255, 255}, false)

	// Right eye glow (multiple layers for glow effect)
	rightEyeX := headX + size/6
	rightEyeY := headY - size/8
	// Outer glow (red)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize+2, color.RGBA{255, 50, 0, 100}, false)
	// Middle glow
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize, color.RGBA{255, 100, 0, 180}, false)
	// Inner bright eye
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize-2, color.RGBA{255, 0, 0, 255}, false)
	// Pupil (black)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize/3, color.RGBA{0, 0, 0, 255}, false)
	// White highlight for scary look
	vector.DrawFilledCircle(screen, rightEyeX+eyeSize/4, rightEyeY-eyeSize/4, eyeSize/5, color.RGBA{255, 255, 255, 255}, false)

	// Eyes (yellow dots) - removed, now have big red glowing eyes
}

func (g *Game) drawSnakeEyes(screen *ebiten.Image, head Point, direction Direction) {
	x := float32(head.X * tileSize)
	y := float32(head.Y * tileSize)
	size := float32(tileSize)
	eyeSize := size / 6
	pupilSize := eyeSize / 2

	// Eye positions based on direction
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

	// Draw whites of eyes
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize, color.RGBA{255, 255, 255, 255}, false)

	// Draw pupils (black)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, pupilSize, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, pupilSize, color.RGBA{0, 0, 0, 255}, false)
}

func (g *Game) drawSnakeTongue(screen *ebiten.Image, head Point, direction Direction) {
	x := float32(head.X * tileSize)
	y := float32(head.Y * tileSize)
	size := float32(tileSize)

	// Tongue color (red/pink)
	tongueColor := color.RGBA{255, 50, 50, 255}

	// Tongue dimensions
	tongueLength := size / 2
	tongueWidth := size / 12

	// Calculate tongue start and end based on direction
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

	// Draw tongue (thin line)
	vector.StrokeLine(screen, startX, startY, endX, endY, tongueWidth, tongueColor, false)

	// Forked tongue (two prongs)
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

	// Draw fork prongs
	vector.StrokeLine(screen, endX, endY, leftForkX, leftForkY, tongueWidth, tongueColor, false)
	vector.StrokeLine(screen, endX, endY, rightForkX, rightForkY, tongueWidth, tongueColor, false)
}

func (g *Game) drawFood(screen *ebiten.Image) {
	x := float32(g.food.X * tileSize)
	y := float32(g.food.Y * tileSize)
	size := float32(tileSize)

	// Apple body (red circle)
	centerX := x + size/2
	centerY := y + size/2 + 2
	radius := size/2 - 3

	// Main red body
	vector.DrawFilledCircle(screen, centerX, centerY, radius, color.RGBA{255, 0, 0, 255}, false)

	// Shine/highlight on apple (lighter red)
	highlightX := centerX - radius/3
	highlightY := centerY - radius/3
	vector.DrawFilledCircle(screen, highlightX, highlightY, radius/3, color.RGBA{255, 100, 100, 255}, false)

	// Small indentation at top
	vector.DrawFilledCircle(screen, centerX, centerY-radius+2, 2, color.RGBA{200, 0, 0, 255}, false)

	// Brown stem
	stemX := centerX
	stemY := centerY - radius
	vector.StrokeLine(screen, stemX, stemY, stemX, stemY-4, 2, color.RGBA{139, 69, 19, 255}, false)

	// Green leaf
	leafColor := color.RGBA{34, 139, 34, 255} // Forest green
	leafBaseX := centerX + 1
	leafBaseY := stemY - 2

	// Draw leaf as a filled ellipse (using polygon approximation)
	leafTipX := leafBaseX + 5
	leafTipY := leafBaseY - 3
	leafBottomX := leafBaseX + 2
	leafBottomY := leafBaseY + 2

	// Main leaf shape (triangle-like)
	vector.DrawFilledRect(screen, leafBaseX, leafBaseY-1, 5, 2, leafColor, false)
	vector.DrawFilledCircle(screen, leafTipX-1, leafTipY, 2, leafColor, false)
	vector.DrawFilledCircle(screen, leafBottomX, leafBottomY, 2, leafColor, false)

	// Leaf vein (lighter green line)
	vector.StrokeLine(screen, leafBaseX+1, leafBaseY, leafTipX-1, leafTipY, 1, color.RGBA{100, 200, 100, 255}, false)
}

func (g *Game) drawCoin(screen *ebiten.Image, coin Coin) {
	x := float32(coin.pos.X * tileSize)
	y := float32(coin.pos.Y * tileSize)
	size := float32(tileSize)

	// Coin dimensions
	coinRadius := size/2 - 4
	centerX := x + size/2
	centerY := y + size/2

	// Outer gold ring
	vector.DrawFilledCircle(screen, centerX, centerY, coinRadius, color.RGBA{255, 215, 0, 255}, false)

	// Inner lighter gold (shine effect)
	innerRadius := coinRadius - 3
	vector.DrawFilledCircle(screen, centerX, centerY, innerRadius, color.RGBA{255, 235, 100, 255}, false)

	// Dollar sign or "X2" indicator (golden center dot)
	dotRadius := innerRadius / 2
	vector.DrawFilledCircle(screen, centerX, centerY, dotRadius, color.RGBA{255, 200, 0, 255}, false)

	// Sparkle effect (small white dots around)
	sparkleOffset := coinRadius - 1
	// Top sparkle
	vector.DrawFilledCircle(screen, centerX, centerY-sparkleOffset, 1, color.RGBA{255, 255, 255, 255}, false)
	// Bottom sparkle
	vector.DrawFilledCircle(screen, centerX, centerY+sparkleOffset, 1, color.RGBA{255, 255, 255, 255}, false)
	// Left sparkle
	vector.DrawFilledCircle(screen, centerX-sparkleOffset, centerY, 1, color.RGBA{255, 255, 255, 255}, false)
	// Right sparkle
	vector.DrawFilledCircle(screen, centerX+sparkleOffset, centerY, 1, color.RGBA{255, 255, 255, 255}, false)

	// Animated glow (pulsing effect)
	glowPhase := (time.Now().UnixMilli() / 200) % 2
	glowIntensity := uint8(100 + glowPhase*50)
	vector.DrawFilledCircle(screen, centerX, centerY, coinRadius+2, color.RGBA{255, 215, 0, glowIntensity}, false)
}

func (g *Game) drawBomb(screen *ebiten.Image, bomb Bomb) {
	x := float32(bomb.pos.X * tileSize)
	y := float32(bomb.pos.Y * tileSize)
	size := float32(tileSize)

	// Bomb body (black circle)
	vector.DrawFilledCircle(screen, x+size/2, y+size/2, size/2-2, color.RGBA{0, 0, 0, 255}, false)

	// Shine on bomb
	vector.DrawFilledCircle(screen, x+size/3, y+size/3, size/6, color.RGBA{50, 50, 50, 255}, false)

	// Fuse (brown stick)
	fuseX := x + size/2
	fuseY := y + size/4
	vector.StrokeLine(screen, fuseX, fuseY, fuseX, fuseY-size/3, 2, color.RGBA{139, 69, 19, 255}, false)

	// Spark at end of fuse (animated - blinking)
	sparkPhase := (bomb.timer % 10) / 5.0
	sparkSize := size/6 + float32(sparkPhase)*size/8

	// Yellow/orange spark glow
	vector.DrawFilledCircle(screen, fuseX, fuseY-size/3, sparkSize, color.RGBA{255, 200, 0, 200}, false)

	// White hot center
	vector.DrawFilledCircle(screen, fuseX, fuseY-size/3, sparkSize/2, color.RGBA{255, 255, 255, 255}, false)

	// Spark particles (random sparks around)
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

	// Chest body (brown rectangle)
	chestColor := color.RGBA{139, 69, 19, 255}
	if chest.open {
		chestColor = color.RGBA{100, 50, 10, 255} // Darker when open
	}
	vector.DrawFilledRect(screen, x+2, y+4, size-4, size-6, chestColor, false)

	// Chest lid (gold trim)
	lidColor := color.RGBA{255, 215, 0, 255}
	if chest.open {
		// Open lid - draw it tilted up
		vector.StrokeLine(screen, x+2, y+4, x+size-2, y+4, 2, lidColor, false)
	} else {
		// Closed lid - draw rounded top
		vector.DrawFilledRect(screen, x+2, y+2, size-4, size/3, lidColor, false)
	}

	// Lock (gold circle in center)
	if !chest.open {
		vector.DrawFilledCircle(screen, x+size/2, y+size/2, 3, color.RGBA{255, 215, 0, 255}, false)
	}
}

func (g *Game) drawKey(screen *ebiten.Image, key Key) {
	x := float32(key.pos.X * tileSize)
	y := float32(key.pos.Y * tileSize)
	size := float32(tileSize)

	// Key color (gold)
	keyColor := color.RGBA{255, 215, 0, 255}

	// Key head (circle)
	headSize := size / 3
	vector.DrawFilledCircle(screen, x+size/2, y+size/3, headSize, keyColor, false)

	// Key shaft (rectangle)
	shaftWidth := size / 8
	shaftHeight := size / 2
	vector.DrawFilledRect(screen, x+size/2-shaftWidth/2, y+size/2, shaftWidth, shaftHeight, keyColor, false)

	// Key teeth (two notches at bottom)
	toothSize := size / 6
	vector.DrawFilledRect(screen, x+size/2-shaftWidth/2, y+size/2+shaftHeight-toothSize, shaftWidth, toothSize, keyColor, false)
	vector.DrawFilledRect(screen, x+size/2, y+size/2+shaftHeight-toothSize, shaftWidth, toothSize/2, keyColor, false)
}

func (g *Game) drawArrow(screen *ebiten.Image, arrow Arrow) {
	x := float32(arrow.pos.X * tileSize)
	y := float32(arrow.pos.Y * tileSize)
	size := float32(tileSize)

	// Arrow color (silver/gray)
	arrowColor := color.RGBA{192, 192, 192, 255}

	// Arrow shaft (line in direction of travel)
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

	// Arrow head (triangle at front)
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

	// Draw arrow head as triangle outline
	vector.StrokeLine(screen, headX1, headY1, headX2, headY2, shaftWidth, arrowColor, false)
	vector.StrokeLine(screen, headX2, headY2, headX3, headY3, shaftWidth, arrowColor, false)
	vector.StrokeLine(screen, headX3, headY3, headX1, headY1, shaftWidth, arrowColor, false)
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Simple Snake - Go365 Go75")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
