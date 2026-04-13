package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game - основная структура игры, реализует ebiten.Game
type Game struct {
	board         *Board
	ui            *UIRenderer
	animSystem    *AnimationSystem
	inputProc     *InputProcessor
	boardOffsetX  float64
	boardOffsetY  float64
	cellSize      float64

	// Состояние игры
	selectedTile *[2]int // Выбранная фишка (row, col)
	score        int
	timer        int
	gameOver     bool
	paused       bool

	// Таймеры
	lastInputTime   time.Time // Время последнего ввода игрока
	timerLastUpdate time.Time // Для отсчёта секунд
	frameCount      int       // Счётчик кадров (для Update)
}

// NewGame создаёт новую игру
func NewGame() *Game {
	g := &Game{
		board:      NewBoard(),
		ui:         NewUIRenderer(),
		animSystem: NewAnimationSystem(),
		score:      0,
		timer:      60,
		gameOver:   false,
		paused:     false,
	}

	// Вычисляем размеры
	g.calculateDimensions()

	// Создаём обработчик ввода
	g.inputProc = NewInputProcessor(g.boardOffsetX, g.boardOffsetY, g.cellSize)

	// Устанавливаем глобальную ссылку на анимации
	SetGlobalAnimationSystem(g.animSystem)

	// Инициализируем таймер
	g.timerLastUpdate = time.Now()
	g.lastInputTime = time.Now()

	log.Println("Игра инициализирована")
	return g
}

// calculateDimensions вычисляет размеры и позиции для адаптивности
func (g *Game) calculateDimensions() {
	w, h := ebiten.WindowSize()

	// Поле 8x8 должно помещаться с отступами для UI
	availableWidth := float64(w) - 40  // 20px отступы с боков
	availableHeight := float64(h) - 100 // 100px сверху для UI

	// Размер ячейки - минимум из доступных
	if availableWidth/BoardSize < availableHeight/BoardSize {
		g.cellSize = availableWidth / BoardSize
	} else {
		g.cellSize = availableHeight / BoardSize
	}

	// Центрируем поле
	boardWidth := g.cellSize * BoardSize
	boardHeight := g.cellSize * BoardSize
	g.boardOffsetX = (float64(w) - boardWidth) / 2
	g.boardOffsetY = (float64(h) - boardHeight) / 2 + 30 // Смещение вниз для UI
}

// Update обновляет логику игры (вызывается 60 раз в секунду)
func (g *Game) Update() error {
	// Обработка ввода
	if !g.paused && !g.gameOver && !g.animSystem.IsAnimating() {
		g.handleInput()
	}

	// Обработка клавиатуры (всегда)
	keyInput := ProcessKeyboardInput()
	if keyInput.NewGame && (g.gameOver || !g.animSystem.IsAnimating()) {
		g.startNewGame()
	}
	if keyInput.Pause {
		g.ui.TogglePause()
		g.paused = g.ui.IsPaused()
	}

	// Обновление таймера (раз в секунду)
	if !g.paused && !g.gameOver {
		elapsed := time.Since(g.timerLastUpdate)
		if elapsed >= time.Second {
			g.timer--
			g.ui.UpdateTimer(g.timer)
			g.timerLastUpdate = time.Now()

			if g.timer <= 0 {
				g.endGame()
			}
		}

		// Авто-подсказка через 5 секунд бездействия
		if time.Since(g.lastInputTime) > 5*time.Second && !g.animSystem.IsAnimating() {
			g.showHint()
		}
	}

	// Обновляем ввод
	g.inputProc.Update(g.boardOffsetX, g.boardOffsetY, g.cellSize)

	g.frameCount++
	return nil
}

// handleInput обрабатывает ввод мыши/тача
func (g *Game) handleInput() {
	// Мышь
	mouseAction := g.inputProc.ProcessMouseInput()
	if mouseAction.Type != ActionNone {
		g.processAction(mouseAction)
		return
	}

	// Тач
	touchAction := g.inputProc.ProcessTouchInput()
	if touchAction.Type != ActionNone {
		g.processAction(touchAction)
		return
	}
}

// processAction обрабатывает действие ввода
func (g *Game) processAction(action InputAction) {
	g.lastInputTime = time.Now()

	// Проверяем клик по кнопке "Новая игра"
	btnX, btnY, btnW, btnH := GetNewGameButtonBounds()
	if action.Type == ActionSelect && IsInsideButton(
		float64(action.Col1)*g.cellSize+g.boardOffsetX,
		float64(action.Row1)*g.cellSize+g.boardOffsetY,
		btnX, btnY, btnW, btnH) {
		g.startNewGame()
		return
	}

	// Проверяем клик по полю
	if action.Type == ActionSelect {
		row, col := action.Row1, action.Col1

		if g.selectedTile == nil {
			// Первый выбор
			g.selectedTile = &[2]int{row, col}
		} else {
			// Вторая фишка - пробуем обмен
			r1, c1 := (*g.selectedTile)[0], (*g.selectedTile)[1]
			r2, c2 := row, col

			// Проверяем что это соседние фишки
			if g.board.isAdjacent(r1, c1, r2, c2) {
				// Пробуем обмен
				success := g.board.Swap(r1, c1, r2, c2)
				if success {
					// Успешный обмен!
					g.board.PerformSwap(r1, c1, r2, c2)
					g.animSystem.AddSwap(r1, c1, r2, c2)
					PlaySwap()

					// После анимации обмена - разрешаем матчи
					g.resolveMatchesAndDrop()
				} else {
					// Невалидный обмен
					g.animSystem.AddShake(r1, c1, r2, c2)
					PlayError()
				}
			}

			// Сбрасываем выбор
			g.selectedTile = nil
		}
	}
}

// resolveMatchesAndDrop находит комбинации, удаляет их и опускает фишки
func (g *Game) resolveMatchesAndDrop() {
	// Ограничиваем глубину каскада
	maxCascade := 10

	for cascade := 0; cascade < maxCascade; cascade++ {
		// Находим все комбинации
		matches := g.board.findAllMatches()
		if len(matches) == 0 {
			break // Нет комбинаций - выходим
		}

		// Начисляем очки
		scoreDelta := g.calculateMatchScore(matches)
		g.score += scoreDelta
		g.ui.UpdateScore(g.score)

		// Преобразуем карту в срез позиций
		positions := make([][2]int, 0, len(matches))
		for pos := range matches {
			positions = append(positions, pos)
		}

		// Анимация удаления
		g.animSystem.AddMatch(positions)
		PlayMatch()

		// Ждём завершения анимации (простая симуляция)
		// В реальном UPDATE это будет асинхронно
		time.Sleep(200 * time.Millisecond)

		// Удаляем фишки
		for _, pos := range positions {
			g.board.SetTile(pos[0], pos[1], -1)
		}

		// Падение новых фишек
		newTiles := g.board.dropDown()
		if len(newTiles) > 0 {
			// Преобразуем карту в срез
			newTilePositions := make([][2]int, 0, len(newTiles))
			for pos := range newTiles {
				newTilePositions = append(newTilePositions, pos)
			}
			g.animSystem.AddDrop(newTilePositions)
			time.Sleep(250 * time.Millisecond)
		}
	}

	// Проверяем есть ли возможные ходы
	if !g.hasPossibleMoves() {
		// Перемешиваем поле
		log.Println("Нет возможных ходов - перемешиваем")
		g.board.fillNoMatches()
	}
}

// calculateMatchScore вычисляет очки за комбинации
func (g *Game) calculateMatchScore(matches map[[2]int]bool) int {
	// Группируем по связанным позициям (простой подсчёт)
	totalTiles := len(matches)

	if totalTiles == 0 {
		return 0
	}

	score := totalTiles * 10 // Базовые очки

	// Бонусы
	if totalTiles >= 5 {
		score += 100 // Бонус за 5+
	} else if totalTiles >= 4 {
		score += 50 // Бонус за 4
	}

	return score
}

// hasPossibleMoves проверяет есть ли возможные ходы
func (g *Game) hasPossibleMoves() bool {
	_, _, ok := g.board.FindHint()
	return ok
}

// showHint показывает подсказку
func (g *Game) showHint() {
	pos1, pos2, ok := g.board.FindHint()
	if ok {
		g.animSystem.AddHint(pos1[0], pos1[1], pos2[0], pos2[1])
	}
}

// startNewGame начинает новую игру
func (g *Game) startNewGame() {
	g.board = NewBoard()
	g.score = 0
	g.timer = 60
	g.gameOver = false
	g.paused = false
	g.selectedTile = nil
	g.animSystem.Clear()
	g.ui.Reset()
	g.ui.UpdateScore(0)
	g.ui.UpdateTimer(60)
	g.lastInputTime = time.Now()
	g.timerLastUpdate = time.Now()
	log.Println("Новая игра начата")
}

// endGame завершает игру
func (g *Game) endGame() {
	g.gameOver = true
	g.ui.SetGameOver()
	PlayGameOver()
	log.Printf("Игра окончена! Счёт: %d", g.score)
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// UI отрисовка
	g.ui.DrawTilesWithBoard(screen, g.boardOffsetX, g.boardOffsetY, g.cellSize,
		g.board, g.selectedTile, g.animSystem)

	// Рисуем остальной UI поверх
	hintPos1, hintPos2, hintActive := g.animSystem.GetHintPositions()
	g.ui.Draw(screen, g.boardOffsetX, g.boardOffsetY, g.cellSize,
		hintPos1, hintPos2, hintActive, g.animSystem)

	// Отладочная информация (FPS)
	// fps := ebiten.ActualFPS()
	// debugText := fmt.Sprintf("FPS: %.1f", fps)
	// (можно добавить через text.Draw если нужно)
}

// Layout возвращает логический размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
