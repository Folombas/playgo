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
	screenWidth  = 640
	screenHeight = 720
	tileSize     = 16
	gridW        = 20
	gridH        = 20
)

// Типы клеток
const (
	cellWall = iota
	cellEmpty
	cellPellet
	cellPowerPellet
)

type direction int

const (
	dirUp direction = iota
	dirDown
	dirLeft
	dirRight
)

type pos struct{ x, y int }

var (
	// Простой лабиринт (1 - стена, 2 - точка, 3 - большая точка)
	maze = [gridH][gridW]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 1},
		{1, 2, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 1, 2, 1},
		{1, 2, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 1, 2, 1},
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
		{1, 2, 1, 1, 1, 2, 1, 2, 1, 1, 1, 1, 2, 1, 2, 1, 1, 1, 2, 1},
		{1, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 1},
		{1, 2, 1, 1, 1, 2, 1, 1, 1, 2, 1, 1, 1, 1, 2, 1, 1, 1, 2, 1},
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
		{1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1},
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
		{1, 2, 1, 1, 1, 2, 1, 1, 1, 2, 1, 1, 1, 1, 2, 1, 1, 1, 2, 1},
		{1, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 1},
		{1, 1, 1, 1, 1, 2, 1, 2, 1, 1, 1, 1, 2, 1, 2, 1, 1, 1, 1, 1},
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
		{1, 2, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 2, 1},
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
		{1, 2, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 2, 1},
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}
)

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
}

type Ghost struct {
	pos   pos
	dir   direction
	color color.Color
	name  string
}

func NewGame() *Game {
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
	if g.gameOver || g.win {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			*g = *NewGame()
		}
		return nil
	}

	// Управление
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

	g.movePacman()

	if g.powerMode {
		g.powerTimer -= 1.0 / 60.0
		if g.powerTimer <= 0 {
			g.powerMode = false
		}
	}

	for _, ghost := range g.ghosts {
		g.moveGhost(ghost)
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
	// Попытка повернуть
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
	if !g.isWall(newX, newY) {
		g.pacmanPos.x = newX
		g.pacmanPos.y = newY
	}

	// Съедание еды
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
			px := float64(x * tileSize)
			py := float64(y * tileSize)
			switch maze[y][x] {
			case cellWall:
				ebitenutil.DrawRect(screen, px, py, tileSize, tileSize, color.RGBA{33, 33, 255, 255})
			case cellPellet:
				ebitenutil.DrawCircle(screen, px+float64(tileSize/2), py+float64(tileSize/2), 2, color.RGBA{255, 255, 0, 255})
			case cellPowerPellet:
				ebitenutil.DrawCircle(screen, px+float64(tileSize/2), py+float64(tileSize/2), 6, color.RGBA{255, 255, 0, 255})
			}
		}
	}

	// Pacman
	pacX := float64(g.pacmanPos.x*tileSize + tileSize/2)
	pacY := float64(g.pacmanPos.y*tileSize + tileSize/2)
	ebitenutil.DrawCircle(screen, pacX, pacY, tileSize/2-1, color.RGBA{255, 255, 0, 255})
	// Рот (просто чёрный треугольник, упрощённо)
	// Для простоты оставим Pacman круглым – не критично.

	// Призраки
	for _, ghost := range g.ghosts {
		gx := float64(ghost.pos.x*tileSize + tileSize/2)
		gy := float64(ghost.pos.y*tileSize + tileSize/2)
		ebitenutil.DrawCircle(screen, gx, gy, tileSize/2-1, ghost.color)
		ebitenutil.DrawCircle(screen, gx-3, gy-2, 2, color.White)
		ebitenutil.DrawCircle(screen, gx+3, gy-2, 2, color.White)
	}

	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(g.score), 10, screenHeight-30)

	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER! Press R to restart", screenWidth/2-100, screenHeight/2)
	}
	if g.win {
		ebitenutil.DebugPrintAt(screen, "YOU WIN! Press R to play again", screenWidth/2-100, screenHeight/2)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Pacman Classic")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}