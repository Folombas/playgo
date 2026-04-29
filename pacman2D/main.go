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
	tileSize   = 20
	gridW      = 19
	gridH      = 21
)

const (
	cellWall = iota
	cellEmpty
	cellPellet
	cellPowerPellet
)

var maze = [gridH][gridW]int{
	{1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1},
	{1,2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2,1},
	{1,2,1,1,1,2,1,1,2,1,1,2,1,1,2,1,1,2,1},
	{1,3,1,1,1,2,1,1,2,1,1,2,1,1,2,1,1,3,1},
	{1,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,1},
	{1,2,1,1,1,2,1,2,1,1,1,1,2,1,2,1,1,2,1},
	{1,2,2,2,2,2,1,2,2,2,2,2,2,1,2,2,2,2,1},
	{1,1,1,1,1,2,1,1,1,2,1,1,1,1,2,1,1,1,1},
	{1,1,1,1,1,2,1,2,2,2,2,2,2,1,2,1,1,1,1},
	{0,0,0,0,1,2,1,2,1,1,1,1,2,1,2,1,0,0,0},
	{1,1,1,1,1,2,1,2,2,2,2,2,2,1,2,1,1,1,1},
	{1,1,1,1,1,2,1,1,1,1,1,1,1,1,2,1,1,1,1},
	{1,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,1},
	{1,2,1,1,1,2,1,1,1,2,1,1,1,1,2,1,1,2,1},
	{1,2,2,2,2,2,1,2,2,2,2,2,2,1,2,2,2,2,1},
	{1,2,1,1,1,2,1,1,1,1,1,1,1,1,2,1,1,2,1},
	{1,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,1},
	{1,3,1,1,1,2,1,1,2,1,1,2,1,1,2,1,1,3,1},
	{1,2,1,1,1,2,1,1,2,1,1,2,1,1,2,1,1,2,1},
	{1,2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2,1},
	{1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1},
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
	pacmanPos      pos
	pacmanDir      direction
	nextDir        direction
	score          int
	powerMode      bool
	powerTimer     float64
	ghosts         []*Ghost
	gameOver       bool
	win            bool
	remainingPellets int
	moveTimer      float64
	moveDelay      float64
	ghostMoveTimers []float64
	frameCounter   int
	pacmanFrame    int
}

type Ghost struct {
	pos   pos
	dir   direction
	color color.Color
	name  string
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		pacmanPos:      pos{9, 15},
		pacmanDir:      dirRight,
		nextDir:        dirRight,
		score:          0,
		powerMode:      false,
		powerTimer:     0,
		gameOver:       false,
		win:            false,
		remainingPellets: countPellets(),
		moveTimer:      0,
		moveDelay:      0.12,
		ghostMoveTimers: make([]float64, 4),
		pacmanFrame:    0,
	}
	g.ghosts = []*Ghost{
		{pos: pos{9, 9}, dir: dirLeft, color: color.RGBA{255, 0, 0, 255}, name: "Blinky"},
		{pos: pos{10, 9}, dir: dirRight, color: color.RGBA{255, 184, 255, 255}, name: "Pinky"},
		{pos: pos{8, 10}, dir: dirUp, color: color.RGBA{0, 255, 255, 255}, name: "Inky"},
		{pos: pos{10, 10}, dir: dirDown, color: color.RGBA{255, 184, 82, 255}, name: "Clyde"},
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

	g.frameCounter++
	if g.frameCounter >= 6 {
		g.frameCounter = 0
		g.pacmanFrame = (g.pacmanFrame + 1) % 3
	}

	g.moveTimer += dt
	if g.moveTimer >= g.moveDelay {
		g.moveTimer = 0
		g.movePacman()
	}

	if g.powerMode {
		g.powerTimer -= dt
		if g.powerTimer <= 0 {
			g.powerMode = false
		}
	}

	for i, ghost := range g.ghosts {
		g.ghostMoveTimers[i] += dt
		if g.ghostMoveTimers[i] >= g.moveDelay {
			g.ghostMoveTimers[i] = 0
			g.moveGhost(ghost)
		}
	}

	for _, ghost := range g.ghosts {
		if ghost.pos == g.pacmanPos {
			if g.powerMode {
				g.score += 200
				switch ghost.name {
				case "Blinky":
					ghost.pos = pos{9, 9}
				case "Pinky":
					ghost.pos = pos{10, 9}
				case "Inky":
					ghost.pos = pos{8, 10}
				case "Clyde":
					ghost.pos = pos{10, 10}
				}
				ghost.dir = dirLeft
			} else {
				g.gameOver = true
			}
		}
	}
	return nil
}

func (g *Game) movePacman() {
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
	if !g.isWall(newX, newY) {
		g.pacmanDir = g.nextDir
	}
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
	if !g.isWall(newX, newY) {
		g.pacmanPos.x = newX
		g.pacmanPos.y = newY
	}
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

func (g *Game) moveGhost(ghost *Ghost) {
	dirs := []direction{dirUp, dirDown, dirLeft, dirRight}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
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
			ghost.dir = d
			ghost.pos.x = nx
			ghost.pos.y = ny
			break
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

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

	for _, ghost := range g.ghosts {
		gx := offsetX + float64(ghost.pos.x)*cellSize + cellSize/2
		gy := offsetY + float64(ghost.pos.y)*cellSize + cellSize/2
		radius := cellSize * 0.4
		ebitenutil.DrawCircle(screen, gx, gy, radius, ghost.color)
		eyeSize := radius * 0.3
		ebitenutil.DrawCircle(screen, gx- radius*0.3, gy- radius*0.2, eyeSize, color.White)
		ebitenutil.DrawCircle(screen, gx+ radius*0.3, gy- radius*0.2, eyeSize, color.White)
	}

	// Рисуем Pacman
	pacX := offsetX + float64(g.pacmanPos.x)*cellSize + cellSize/2
	pacY := offsetY + float64(g.pacmanPos.y)*cellSize + cellSize/2
	radius := cellSize * 0.4
	ebitenutil.DrawCircle(screen, pacX, pacY, radius, color.RGBA{255, 255, 0, 255})

	// Рисуем рот чёрными линиями от центра к краям
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
		dirAngle = -math.Pi / 2
	}
	start := dirAngle + mouthAngle
	end := dirAngle + 2*math.Pi - mouthAngle
	segments := 20
	for i := 0; i <= segments; i++ {
		theta := start + float64(i)/float64(segments)*(end-start)
		cx := pacX + radius*math.Cos(theta)
		cy := pacY + radius*math.Sin(theta)
		ebitenutil.DrawLine(screen, pacX, pacY, cx, cy, color.RGBA{0, 0, 0, 255})
	}
	// Дополнительно можно нарисовать чёрный треугольник, но линии дают достаточный эффект.

	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(g.score), 10, 10)

	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER! Press R to restart", windowWidth/2-150, windowHeight/2)
	}
	if g.win {
		ebitenutil.DebugPrintAt(screen, "YOU WIN! Press R to play again", windowWidth/2-130, windowHeight/2)
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
	cellSize = min(cellSizeX, cellSizeY)
	offsetX = (float64(windowWidth) - float64(gridW)*cellSize) / 2
	offsetY = (float64(windowHeight) - float64(gridH)*cellSize) / 2
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Pacman Classic")
	ebiten.SetFullscreen(true)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}