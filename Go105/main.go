package main

import (
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// Размеры окна
	screenWidth  = 800
	screenHeight = 600

	// Размеры игрового поля
	boardCols    = 10
	boardRows    = 20
	cellSize     = 25
	boardOffsetX = 250 // Смещение поля вправо
	boardOffsetY = 50

	// Цвета (RGBA)
	colorBgR = 20
	colorBgG = 20
	colorBgB = 40
)

// Game представляет основную игру
type Game struct {
	state    GameState
	board    *Board
	current  *Tetromino
	next     *Tetromino
	score    int
	level    int
	lines    int
	highScore int
	dropTimer float64
	gameOverAnim float64
	stars    *BackgroundStars
	lineAnims []*LineAnim
}

// GameState — состояние игры
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

func NewGame() *Game {
	g := &Game{
		state: StateMenu,
		stars: NewBackgroundStars(),
	}
	g.loadHighScore()
	return g
}

func (g *Game) startNewGame() {
	g.board = NewBoard()
	g.score = 0
	g.level = 1
	g.lines = 0
	g.dropTimer = 0
	g.gameOverAnim = 0
	g.current = NewRandomTetromino()
	g.next = NewRandomTetromino()
	g.state = StatePlaying
}

func (g *Game) Update() error {
	// Обновить звёзды
	g.stars.Update()

	switch g.state {
	case StateMenu:
		if inputStart() {
			g.startNewGame()
		}
	case StatePlaying:
		g.updatePlaying()
	case StatePaused:
		if inputPause() {
			g.state = StatePlaying
		}
	case StateGameOver:
		g.gameOverAnim++
		if g.gameOverAnim > 60 && inputRestart() {
			g.startNewGame()
		}
	}
	return nil
}

func (g *Game) updatePlaying() {
	// Пауза
	if inputPause() {
		g.state = StatePaused
		return
	}

	// Ввод движения
	inputMovement(g)

	// Таймер падения
	g.dropTimer++
	dropInterval := g.getDropInterval()

	if g.dropTimer >= dropInterval {
		g.dropTimer = 0
		g.dropCurrent()
	}
}

func (g *Game) getDropInterval() float64 {
	// Ускорение с уровнем: от 60 до 5 фреймов
	base := 60.0
	reduction := float64(g.level-1) * 5.0
	interval := base - reduction
	if interval < 5 {
		interval = 5
	}
	return interval
}

func (g *Game) dropCurrent() {
	if g.current == nil || g.board == nil {
		return
	}

	// Попробовать сдвинуть вниз
	g.current.Y++
	if g.board.Collides(g.current) {
		g.current.Y--
		// Зафиксировать фигуру
		g.board.Place(g.current)
		// Проверить линии
		cleared := g.board.ClearLines()
		if cleared > 0 {
			g.lines += cleared
			g.score += g.calculateScore(cleared)
			g.level = g.lines/10 + 1
			PlaySound(SoundLineClear)
		} else {
			PlaySound(SoundPlace)
		}
		// Следующая фигура
		g.current = g.next
		g.next = NewRandomTetromino()
		// Проверить Game Over
		if g.board.Collides(g.current) {
			g.state = StateGameOver
			g.gameOverAnim = 0
			g.saveHighScore()
			PlaySound(SoundGameOver)
		}
	}
}

func (g *Game) calculateScore(lines int) int {
	// Классическая система очков
	points := []int{0, 100, 300, 500, 800}
	if lines > 4 {
		lines = 4
	}
	return points[lines] * g.level
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Фон (звёзды)
	screen.Fill(colorBg())
	g.stars.Draw(screen)

	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying, StatePaused:
		g.drawGame(screen)
		if g.state == StatePaused {
			g.drawPaused(screen)
		}
	case StateGameOver:
		g.drawGame(screen)
		g.drawGameOver(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func colorBg() color.Color {
	return color.RGBA{colorBgR, colorBgG, colorBgB, 255}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Инициализация аудио и звуков
	initAudio()
	initSounds()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Geometric Match Game — Go105")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
