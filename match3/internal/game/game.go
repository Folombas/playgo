package game

import (
	"fmt"
	"match3/internal/logic"
	"match3/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

// GameState определяет текущее состояние игры
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateGameOver
)

// Game - основная структура игры, реализует ebiten.Game
type Game struct {
	state     GameState
	board     *logic.Board
	uiManager *ui.Manager
	score     int
	moves     int
}

// NewGame создаёт новую игру
func NewGame() *Game {
	g := &Game{
		state:     StateMenu,
		uiManager: ui.NewManager(),
		score:     0,
		moves:     0,
	}
	return g
}

// Update обновляет логику игры каждый кадр
func (g *Game) Update() error {
	switch g.state {
	case StateMenu:
		// Обновление меню
		if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsGamepadButtonPressed(0, 0) {
			g.startGame()
		}
	case StatePlaying:
		// Обновление игровой логики
		g.board.Update()
		
		// Проверка на доступные ходы
		if !g.board.HasValidMoves() {
			g.state = StateGameOver
		}
	case StateGameOver:
		// Обновление экрана Game Over
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			g.startGame()
		}
	}
	return nil
}

// Draw отрисовывает текущий кадр
func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка экрана
	screen.Fill(ui.ColorBackground)

	switch g.state {
	case StateMenu:
		g.uiManager.DrawMenu(screen)
	case StatePlaying:
		g.drawGame(screen)
	case StateGameOver:
		g.uiManager.DrawGameOver(screen, g.score)
	}
}

// Layout возвращает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// startGame начинает новую игру
func (g *Game) startGame() {
	g.state = StatePlaying
	g.board = logic.NewBoard(8, 8)
	g.score = 0
	g.moves = 0
	fmt.Println("Новая игра началась!")
}

// drawGame отрисовывает игровое поле и UI
func (g *Game) drawGame(screen *ebiten.Image) {
	// Отрисовка доски
	g.board.Draw(screen)
	
	// Отрисовка UI (счёт, ходы)
	g.uiManager.DrawHUD(screen, g.score, g.moves)
}
