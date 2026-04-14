package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game представляет основное состояние игры
type Game struct {
	Board         *Board
	Score         int
	HighScore     int
	TimeLeft      int
	AnimManager   *AnimationManager
	InputHandler  *UI
	GameOver      bool
	Paused        bool
	Animating     bool // Флаг блокировки ввода во время анимаций
	SelectedTile  *Position
	HintP1        *Position
	HintP2        *Position
	LastMoveTime  time.Time
	HintTimer     time.Duration
	WindowWidth   int
	WindowHeight  int
	BoardOffsetX  float64
	BoardOffsetY  float64
	timerTicker   *time.Ticker
	lastTick      time.Time
}

// NewGame создаёт новую игру
func NewGame(windowWidth, windowHeight int) *Game {
	// Рассчитываем смещение доски для центрирования
	boardOffsetX := float64(windowWidth-BoardPixel) / 2
	boardOffsetY := float64(windowHeight-BoardPixel)/2 + 30 // Небольшой сдвиг вниз для UI

	g := &Game{
		Board:        NewBoard(),
		Score:        0,
		HighScore:    0, // Загружается из файла
		TimeLeft:     60,
		AnimManager:  NewAnimationManager(),
		InputHandler: &UI{},
		WindowWidth:  windowWidth,
		WindowHeight: windowHeight,
		BoardOffsetX: boardOffsetX,
		BoardOffsetY: boardOffsetY,
		LastMoveTime: time.Now(),
		HintTimer:    0,
		timerTicker:  time.NewTicker(time.Second),
		lastTick:     time.Now(),
	}

	// Загружаем рекорд (заглушка)
	g.loadHighScore()

	return g
}

// Update обновляет состояние игры (вызывается 60 раз в секунду)
func (g *Game) Update() error {
	// Если игра окончена или на паузе - не обновляем
	if g.GameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.Reset()
		}
		return nil
	}

	if g.Paused {
		if ebiten.IsKeyPressed(ebiten.KeyP) {
			g.Paused = false
		}
		return nil
	}

	// Обновляем таймер (каждую секунду)
	g.updateTimer()

	// Проверяем ввод
	g.handleInput()

	// Обновляем анимации
	g.Animating = g.AnimManager.Update(g.Board)

	// Если нет активных анимаций, проверяем каскад
	if !g.Animating {
		g.resolveCascade()
	}

	// Подсказки
	g.updateHint()

	return nil
}

// updateTimer обновляет таймер
func (g *Game) updateTimer() {
	now := time.Now()
	if now.Sub(g.lastTick) >= time.Second {
		g.TimeLeft--
		g.lastTick = now

		if g.TimeLeft <= 0 {
			g.TimeLeft = 0
			g.GameOver = true
			g.saveHighScore()
		}
	}
}

// handleInput обрабатывает ввод
func (g *Game) handleInput() {
	// Клавиша сброса
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.Reset()
		return
	}

	// Клавиша паузы
	if ebiten.IsKeyPressed(ebiten.KeyP) {
		g.Paused = !g.Paused
		return
	}

	// Если анимация - блокируем ввод
	if g.Animating {
		return
	}

	// Клик мыши
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		g.handleClick(float64(mx), float64(my))
	}
}

// handleClick обрабатывает клик
func (g *Game) handleClick(screenX, screenY float64) {
	// Проверяем кнопку "Новая игра"
	buttonW := 150.0
	buttonH := 40.0
	buttonX := float64(g.WindowWidth)/2 - buttonW/2
	buttonY := 10.0

	if screenX >= buttonX && screenX <= buttonX+buttonW && screenY >= buttonY && screenY <= buttonY+buttonH {
		g.Reset()
		return
	}

	// Преобразуем в координаты сетки
	x := screenX - g.BoardOffsetX
	y := screenY - g.BoardOffsetY

	if x < 0 || x >= BoardPixel || y < 0 || y >= BoardPixel {
		return
	}

	col := int(x / CellSize)
	row := int(y / CellSize)

	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return
	}

	clickedPos := &Position{Row: row, Col: col}

	// Если ничего не выбрано - выбираем
	if g.SelectedTile == nil {
		g.SelectedTile = clickedPos
		g.Board.Grid[row][col].Selected = true
		g.LastMoveTime = time.Now()
		g.HintP1 = nil
		g.HintP2 = nil
		return
	}

	// Если кликнули на ту же фишку - снимаем выделение
	if g.SelectedTile.Row == row && g.SelectedTile.Col == col {
		g.Board.Grid[row][col].Selected = false
		g.SelectedTile = nil
		return
	}

	// Проверяем соседство
	if !g.SelectedTile.IsAdjacent(clickedPos) {
		// Снимаем выделение и выбираем новую
		g.Board.Grid[g.SelectedTile.Row][g.SelectedTile.Col].Selected = false
		g.SelectedTile = clickedPos
		g.Board.Grid[row][col].Selected = true
		g.LastMoveTime = time.Now()
		return
	}

	// Пробуем обмен
	g.trySwap(g.SelectedTile, clickedPos)
}

// trySwap пытается поменять фишки местами
func (g *Game) trySwap(p1, p2 *Position) {
	// Снимаем выделение
	if g.Board.Grid[p1.Row][p1.Col] != nil {
		g.Board.Grid[p1.Row][p1.Col].Selected = false
	}

	// Делаем обмен
	g.Board.Swap(p1, p2)

	// Проверяем совпадения
	matches := g.Board.FindMatches()

	if len(matches) > 0 {
		// Успешный обмен!
		g.AnimManager.AddSwap(p1, p2)
		g.processMatches(matches)
		g.SelectedTile = nil
		g.LastMoveTime = time.Now()
		g.HintP1 = nil
		g.HintP2 = nil
	} else {
		// Невалидный обмен - возвращаем
		g.Board.Swap(p1, p2)
		g.AnimManager.AddShake(p1, p2)
		g.SelectedTile = nil
	}
}

// processMatches обрабатывает найденные совпадения
func (g *Game) processMatches(matches map[string]*Position) {
	// Начисляем очки
	score := g.Board.RemoveTiles(matches)
	g.Score += score

	// Создаём позиции для анимации удаления
	positions := make([]*Position, 0, len(matches))
	for _, pos := range matches {
		positions = append(positions, pos)
	}

	g.AnimManager.AddRemove(positions)
	g.Animating = true
}

// resolveCascade разрешает каскадные совпадения
func (g *Game) resolveCascade() {
	// После удаления - сдвигаем вниз
	g.Board.DropTiles()

	// Проверяем новые совпадения
	matches := g.Board.FindMatches()

	if len(matches) > 0 {
		// Каскад!
		g.processMatches(matches)
	} else {
		// Проверяем, есть ли валидные ходы
		if !g.Board.HasValidMoves() {
			// Перегенерируем поле
			g.Board = NewBoard()
		}
	}
}

// updateHint обновляет подсказку
func (g *Game) updateHint() {
	// Если игрок не ходил 5 секунд - показываем подсказку
	elapsed := time.Since(g.LastMoveTime)
	if elapsed >= 5*time.Second && g.HintP1 == nil {
		p1, p2 := g.Board.FindHint()
		g.HintP1 = p1
		g.HintP2 = p2
		g.HintTimer = 0
	}

	// Подсказка исчезает через 2 секунды
	if g.HintP1 != nil {
		g.HintTimer += time.Second / 60 // Примерно
		if g.HintTimer >= 2*time.Second {
			g.HintP1 = nil
			g.HintP2 = nil
		}
	}
}

// Reset сбрасывает игру
func (g *Game) Reset() {
	g.Board = NewBoard()
	g.Score = 0
	g.TimeLeft = 60
	g.GameOver = false
	g.Paused = false
	g.SelectedTile = nil
	g.AnimManager = NewAnimationManager()
	g.LastMoveTime = time.Now()
	g.HintP1 = nil
	g.HintP2 = nil
	g.lastTick = time.Now()
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка экрана
	screen.Fill(color.RGBA{R: 30, G: 30, B: 40, A: 255})

	// Отрисовка доски
	ui := NewUI()
	ui.DrawBoard(screen, g.Board, g.BoardOffsetX, g.BoardOffsetY)

	// Подсказка
	ui.DrawHint(screen, g.HintP1, g.HintP2, g.BoardOffsetX, g.BoardOffsetY)

	// UI
	ui.DrawScore(screen, g.Score, g.HighScore)
	ui.DrawTimer(screen, g.TimeLeft)

	// Кнопка "Новая игра"
	buttonW := 150.0
	buttonH := 40.0
	buttonX := float64(g.WindowWidth)/2 - buttonW/2
	buttonY := 10.0
	ui.DrawNewGameButton(screen, buttonX, buttonY, buttonW, buttonH)

	// Экран паузы
	if g.Paused {
		ui.DrawPauseScreen(screen, g.WindowWidth, g.WindowHeight)
	}

	// Экран Game Over
	if g.GameOver {
		ui.DrawGameOver(screen, g.Score, g.WindowWidth, g.WindowHeight)
	}

	// Отладочная информация
	g.drawDebug(screen)
}

// drawDebug отрисовывает отладочную информацию
func (g *Game) drawDebug(screen *ebiten.Image) {
	debugText := fmt.Sprintf("FPS: %.1f", ebiten.ActualFPS)
	// Можно добавить больше отладочной информации
	_ = debugText
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// loadHighScore загружает рекорд (заглушка)
func (g *Game) loadHighScore() {
	// В реальной версии загружаем из файла
	g.HighScore = 0
}

// saveHighScore сохраняет рекорд (заглушка)
func (g *Game) saveHighScore() {
	if g.Score > g.HighScore {
		g.HighScore = g.Score
		// Сохраняем в файл
	}
}
