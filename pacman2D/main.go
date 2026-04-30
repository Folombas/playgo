package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	tileSize = 20
	gridW    = 19
	gridH    = 21
)

const (
	cellWall = iota
	cellEmpty
	cellPellet
	cellPowerPellet
)

var maze = [gridH][gridW]int{
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1},
	{1, 3, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 3, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 1, 2, 1, 2, 1, 1, 1, 1, 2, 1, 2, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 1},
	{1, 1, 1, 1, 1, 2, 1, 1, 1, 2, 1, 1, 1, 1, 2, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 2, 1, 2, 2, 2, 2, 2, 2, 1, 2, 1, 1, 1, 1},
	{0, 0, 0, 0, 1, 2, 1, 2, 1, 1, 1, 1, 2, 1, 2, 1, 0, 0, 0},
	{1, 1, 1, 1, 1, 2, 1, 2, 2, 2, 2, 2, 2, 1, 2, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 1, 2, 1, 1, 1, 2, 1, 1, 1, 1, 2, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 3, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 3, 1},
	{1, 2, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
}

var (
	windowWidth  int
	windowHeight int
	cellSize     float64
	offsetX      float64
	offsetY      float64
)

type direction int

const (
	dirUp direction = iota
	dirDown
	dirLeft
	dirRight
)

type pos struct{ x, y int }

type Game struct {
	pacmanPos         pos
	pacmanDir         direction
	nextDir           direction
	score             int
	powerMode         bool
	powerTimer        float64
	ghosts            []*Ghost
	lives             int
	gameOver          bool
	win               bool
	remainingPellets  int
	moveTimer         float64
	moveDelay         float64
	ghostMoveTimers   []float64
	frameCounter      int
	pacmanFrame       int
	powerFlashCounter int
	powerBlink        bool
}

type Ghost struct {
	pos      pos
	dir      direction
	color    color.Color
	name     string
	homePos  pos
	dead     bool
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		pacmanPos:         pos{9, 15},
		pacmanDir:         dirRight,
		nextDir:           dirRight,
		score:             0,
		powerMode:         false,
		powerTimer:        0,
		lives:             3,
		gameOver:          false,
		win:               false,
		remainingPellets:  countPellets(),
		moveTimer:         0,
		moveDelay:         0.12,
		ghostMoveTimers:   make([]float64, 4),
		pacmanFrame:       0,
		powerFlashCounter: 0,
		powerBlink:        false,
	}
	g.ghosts = []*Ghost{
		{pos: pos{9, 9}, dir: dirLeft, color: color.RGBA{255, 0, 0, 255}, name: "Blinky", homePos: pos{9, 9}},
		{pos: pos{10, 9}, dir: dirRight, color: color.RGBA{255, 184, 255, 255}, name: "Pinky", homePos: pos{10, 9}},
		{pos: pos{8, 10}, dir: dirUp, color: color.RGBA{0, 255, 255, 255}, name: "Inky", homePos: pos{8, 10}},
		{pos: pos{10, 10}, dir: dirDown, color: color.RGBA{255, 184, 82, 255}, name: "Clyde", homePos: pos{10, 10}},
	}
	return g
}

func countPellets() int {
	cnt := 0
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			if maze[y][x] == cellPellet || maze[y][x] == cellPowerPellet {
				cnt++
			}
		}
	}
	return cnt
}

func (g *Game) Update() error {
	dt := 1.0 / 60.0

	if g.gameOver || g.win {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			*g = *NewGame()
			resize()
		}
		return nil
	}
	if g.lives <= 0 {
		g.gameOver = true
	}

	// Ввод
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.nextDir = dirUp
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.nextDir = dirDown
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.nextDir = dirLeft
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.nextDir = dirRight
	}

	// Анимация рта Pacman
	g.frameCounter++
	if g.frameCounter >= 6 {
		g.frameCounter = 0
		g.pacmanFrame = (g.pacmanFrame + 1) % 3
	}

	// Движение Pacman
	g.moveTimer += dt
	if g.moveTimer >= g.moveDelay {
		g.moveTimer = 0
		g.movePacman()
	}

	// Режим страха
	if g.powerMode {
		g.powerTimer -= dt
		if g.powerTimer <= 0 {
			g.powerMode = false
		}
		// Мигание призраков в последние 3 секунды
		if g.powerTimer < 3.0 {
			g.powerFlashCounter++
			if g.powerFlashCounter >= 3 {
				g.powerFlashCounter = 0
				g.powerBlink = !g.powerBlink
			}
		}
	}

	// Движение призраков
	for i, ghost := range g.ghosts {
		g.ghostMoveTimers[i] += dt
		if g.ghostMoveTimers[i] >= g.moveDelay {
			g.ghostMoveTimers[i] = 0
			g.moveGhost(ghost)
		}
	}

	// Столкновение Pacman с призраками
	for _, ghost := range g.ghosts {
		if ghost.pos == g.pacmanPos && !ghost.dead {
			if g.powerMode {
				ghost.dead = true
				g.score += 200
				ghost.pos = ghost.homePos // возврат в дом
				ghost.dir = dirLeft
			} else {
				g.lives--
				if g.lives <= 0 {
					g.gameOver = true
				} else {
					g.respawn()
				}
			}
		}
	}

	return nil
}

func (g *Game) respawn() {
	// Сброс Pacman в начальную позицию
	g.pacmanPos = pos{9, 15}
	g.pacmanDir = dirRight
	g.nextDir = dirRight
	// Сброс призраков
	for _, ghost := range g.ghosts {
		ghost.dead = false
	}
	g.ghosts[0].pos = pos{9, 9}
	g.ghosts[1].pos = pos{10, 9}
	g.ghosts[2].pos = pos{8, 10}
	g.ghosts[3].pos = pos{10, 10}
	// Сброс таймеров
	g.moveTimer = 0
	for i := range g.ghostMoveTimers {
		g.ghostMoveTimers[i] = 0
	}
	// Короткая пауза даст время игроку
	time.Sleep(500 * time.Millisecond)
}

func (g *Game) movePacman() {
	// Проверка телепортации
	if g.pacmanPos.x == 0 && g.pacmanPos.y == 9 && g.nextDir == dirLeft {
		g.pacmanPos.x = gridW - 1
	} else if g.pacmanPos.x == gridW-1 && g.pacmanPos.y == 9 && g.nextDir == dirRight {
		g.pacmanPos.x = 0
	} else {
		// Попробовать направление nextDir
		newX, newY := g.pacmanPos.x, g.pacmanPos.y
		switch g.nextDir {
		case dirUp:
			newY--
		case dirDown:
			newY++
		case dirLeft:
			newX--
		case dirRight:
			newX++
		}
		if !g.isWall(newX, newY) && !g.isTeleport(newX, newY) {
			g.pacmanDir = g.nextDir
		}
		// Движение в текущем направлении
		newX, newY = g.pacmanPos.x, g.pacmanPos.y
		switch g.pacmanDir {
		case dirUp:
			newY--
		case dirDown:
			newY++
		case dirLeft:
			newX--
		case dirRight:
			newX++
		}
		if !g.isWall(newX, newY) && !g.isTeleport(newX, newY) {
			g.pacmanPos.x = newX
			g.pacmanPos.y = newY
		}
	}

	// Поедание еды
	if maze[g.pacmanPos.y][g.pacmanPos.x] == cellPellet {
		maze[g.pacmanPos.y][g.pacmanPos.x] = cellEmpty
		g.score += 10
		g.remainingPellets--
	} else if maze[g.pacmanPos.y][g.pacmanPos.x] == cellPowerPellet {
		maze[g.pacmanPos.y][g.pacmanPos.x] = cellEmpty
		g.score += 50
		g.remainingPellets--
		g.powerMode = true
		g.powerTimer = 10.0
		g.powerBlink = false
		g.powerFlashCounter = 0
	}
	if g.remainingPellets == 0 {
		g.win = true
	}
}

func (g *Game) isWall(x, y int) bool {
	if x < 0 || x >= gridW || y < 0 || y >= gridH {
		return true
	}
	return maze[y][x] == cellWall
}

func (g *Game) isTeleport(x, y int) bool {
	// тоннель только на клетках (0,9) и (18,9)
	return (x == -1 && y == 9) || (x == gridW && y == 9)
}

func (g *Game) moveGhost(ghost *Ghost) {
	// Простой случайный поиск пути (можно улучшить AI)
	dirs := []direction{dirUp, dirDown, dirLeft, dirRight}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	// Если призрак в режиме страха, то он разбегается от Pacman (просто идёт в противоположную сторону)
	// Упростим: в режиме страха выбираем направление, которое максимально удаляет от Pacman
	bestDir := ghost.dir
	minDist := math.Inf(1)
	for _, d := range dirs {
		nx, ny := ghost.pos.x, ghost.pos.y
		switch d {
		case dirUp:
			ny--
		case dirDown:
			ny++
		case dirLeft:
			nx--
		case dirRight:
			nx++
		}
		if !g.isWall(nx, ny) {
			dx := float64(nx - g.pacmanPos.x)
			dy := float64(ny - g.pacmanPos.y)
			dist := dx*dx + dy*dy
			if g.powerMode {
				// В режиме страха убегаем – чем больше расстояние, тем лучше
				if dist > minDist {
					minDist = dist
					bestDir = d
				}
			} else {
				if dist < minDist {
					minDist = dist
					bestDir = d
				}
			}
		}
	}
	if bestDir != ghost.dir || g.isWall(ghost.pos.x, ghost.pos.y) {
		ghost.dir = bestDir
	}
	// Движение
	switch ghost.dir {
	case dirUp:
		ghost.pos.y--
	case dirDown:
		ghost.pos.y++
	case dirLeft:
		ghost.pos.x--
	case dirRight:
		ghost.pos.x++
	}
	// Телепорт для привидений
	if ghost.pos.x < 0 && ghost.pos.y == 9 {
		ghost.pos.x = gridW - 1
	} else if ghost.pos.x >= gridW && ghost.pos.y == 9 {
		ghost.pos.x = 0
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Стены и еда
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			px := offsetX + float64(x)*cellSize
			py := offsetY + float64(y)*cellSize
			switch maze[y][x] {
			case cellWall:
				ebitenutil.DrawRect(screen, px, py, cellSize, cellSize, color.RGBA{33, 33, 255, 255})
			case cellPellet:
				rad := cellSize * 0.1
				ebitenutil.DrawCircle(screen, px+cellSize/2, py+cellSize/2, rad, color.RGBA{255, 255, 0, 255})
			case cellPowerPellet:
				rad := cellSize * 0.3
				ebitenutil.DrawCircle(screen, px+cellSize/2, py+cellSize/2, rad, color.RGBA{255, 255, 0, 255})
			}
		}
	}

	// Привидения
	for _, ghost := range g.ghosts {
		if ghost.dead {
			continue
		}
		gx := offsetX + float64(ghost.pos.x)*cellSize + cellSize/2
		gy := offsetY + float64(ghost.pos.y)*cellSize + cellSize/2
		radius := cellSize * 0.4
		var curColor color.Color = ghost.color
		if g.powerMode {
			if g.powerTimer < 3.0 && g.powerBlink {
				curColor = color.RGBA{255, 255, 255, 255}
			} else {
				curColor = color.RGBA{100, 100, 255, 255}
			}
		}
		ebitenutil.DrawCircle(screen, gx, gy, radius, curColor)
		// Глаза
		eyeSize := radius * 0.3
		ebitenutil.DrawCircle(screen, gx-radius*0.3, gy-radius*0.2, eyeSize, color.White)
		ebitenutil.DrawCircle(screen, gx+radius*0.3, gy-radius*0.2, eyeSize, color.White)
		// Зрачки
		pupil := eyeSize * 0.5
		ebitenutil.DrawCircle(screen, gx-radius*0.3, gy-radius*0.2, pupil, color.Black)
		ebitenutil.DrawCircle(screen, gx+radius*0.3, gy-radius*0.2, pupil, color.Black)
	}

	// Pacman
	pacX := offsetX + float64(g.pacmanPos.x)*cellSize + cellSize/2
	pacY := offsetY + float64(g.pacmanPos.y)*cellSize + cellSize/2
	radius := cellSize * 0.4
	ebitenutil.DrawCircle(screen, pacX, pacY, radius, color.RGBA{255, 255, 0, 255})

	// Рисуем рот (треугольник)
	mouthAngle := 0.0
	switch g.pacmanFrame {
	case 0:
		mouthAngle = 0.3
	case 1:
		mouthAngle = 0.5
	default:
		mouthAngle = 0.7
	}
	dirAngle := 0.0
	switch g.pacmanDir {
	case dirRight:
		dirAngle = 0
	case dirDown:
		dirAngle = math.Pi / 2
	case dirLeft:
		dirAngle = math.Pi
	case dirUp:
		dirAngle = 3 * math.Pi / 2
	}
	startAng := dirAngle - mouthAngle
	endAng := dirAngle + mouthAngle
	// Рисуем чёрный сектор (треугольник на угле)
	numPoints := 20
	for i := 0; i <= numPoints; i++ {
		theta := startAng + float64(i)/float64(numPoints)*(endAng-startAng)
		cx := pacX + radius*math.Cos(theta)
		cy := pacY + radius*math.Sin(theta)
		ebitenutil.DrawLine(screen, pacX, pacY, cx, cy, color.RGBA{0, 0, 0, 255})
	}
	// Также дорисовываем треугольник для заполнения
	// Для простоты используем заливку сектора, но в ebiten сложно; линии дают эффект.

	// Счёт, жизни
	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(g.score), 10, 10)
	livesStr := "Lives: "
	for i := 0; i < g.lives; i++ {
		livesStr += "❤️ "
	}
	ebitenutil.DebugPrintAt(screen, livesStr, 10, 30)

	winX := windowWidth/2 - 150
	winY := windowHeight / 2
	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER! Press R to restart", winX, winY)
	}
	if g.win {
		ebitenutil.DebugPrintAt(screen, "YOU WIN! Press R to play again", winX, winY)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	windowWidth, windowHeight = outsideWidth, outsideHeight
	resize()
	return windowWidth, windowHeight
}

func resize() {
	cellSizeX := float64(windowWidth) / float64(gridW+2)
	cellSizeY := float64(windowHeight) / float64(gridH+2)
	cellSize = math.Min(cellSizeX, cellSizeY)
	offsetX = (float64(windowWidth) - float64(gridW)*cellSize) / 2
	offsetY = (float64(windowHeight) - float64(gridH)*cellSize) / 2
}

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Pacman Classic - Fresh Look")
	ebiten.SetFullscreen(true)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}