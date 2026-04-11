package game

import (
	"fmt"
	"match3/internal/logic"
	"match3/internal/ui"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// GameState определяет текущее состояние игры
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateSettings
	StateGameOver
	StateLevelComplete
)

// Game - основная структура игры, реализует ebiten.Game
type Game struct {
	state           GameState
	board           *logic.Board
	uiManager       *ui.Manager
	score           int
	moves           int
	selectedTile    *logic.Tile
	isSwapping      bool
	swapFrom        *logic.Tile
	swapTo          *logic.Tile
	mouseX          int
	mouseY          int
	effectSystem    *logic.EffectSystem
	scorePopups     *logic.ScorePopupSystem
	comboCounter    int
	lastMatchTime   float64
	spriteManager   *SpriteManager
	soundManager    SoundPlayer
	saveManager     SaveStorage
	levelManager    *logic.LevelManager
	levelStartTime  time.Time
	hintSystem      *logic.HintSystem
}

// GameConfig содержит зависимости для создания игры
// Это позволяет легко инжектить моки для тестирования
type GameConfig struct {
	SoundPlayer SoundPlayer
	SaveStorage SaveStorage
}

// NewGame создаёт новую игру с опциональной конфигурацией
// Если config nil, создаются реализации по умолчанию
func NewGame(config ...GameConfig) *Game {
	var cfg GameConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	g := &Game{
		state:         StateMenu,
		uiManager:     ui.NewManager(),
		score:         0,
		moves:         0,
		effectSystem:  logic.NewEffectSystem(),
		scorePopups:   logic.NewScorePopupSystem(),
		spriteManager: NewSpriteManager(),
	}

	// Dependency Injection для звуковой системы
	if cfg.SoundPlayer != nil {
		g.soundManager = cfg.SoundPlayer
	} else {
		g.soundManager = NewSoundManager()
	}

	// Dependency Injection для системы сохранений
	if cfg.SaveStorage != nil {
		g.saveManager = cfg.SaveStorage
	} else {
		g.saveManager = NewSaveManager()
	}

	g.levelManager = logic.NewLevelManager()
	g.hintSystem = logic.NewHintSystem()

	return g
}

// Update обновляет логику игры каждый кадр
func (g *Game) Update() error {
	// Отслеживание позиции мыши
	g.mouseX, g.mouseY = ebiten.CursorPosition()

	switch g.state {
	case StateMenu:
		// Обновление меню
		if ebiten.IsKeyPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.startGame()
		}
	case StatePlaying:
		g.handleInput()
		// Обновление игровой логики
		g.board.Update()

		// Обновление эффектов
		g.effectSystem.Update()
		g.scorePopups.Update()
		
		// Обновление системы подсказок
		g.hintSystem.Update(g.board)

		// Проверка на доступные ходы
		if !g.board.HasValidMoves() {
			g.state = StateGameOver
		}

		// Пауза
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = StatePaused
		}
	case StatePaused:
		// Обновление паузы
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = StatePlaying
		}
	case StateSettings:
		// Обновление настроек
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StatePaused
		}
	case StateGameOver:
		// Обновление экрана Game Over
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.startGame()
		}
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			g.state = StateMenu
		}
	case StateLevelComplete:
		// Обновление экрана завершения уровня
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.nextLevel()
		}
	}
	return nil
}

// Layout возвращает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// startGame начинает новую игру
func (g *Game) startGame() {
	g.state = StatePlaying
	level := g.levelManager.GetCurrentLevel()
	g.board = logic.NewBoard(level.BoardRows, level.BoardCols)
	
	// Передаём спрайты доске
	if g.spriteManager.IsLoaded() {
		sprites := make(map[int]*ebiten.Image)
		for i := 0; i < level.GemTypes; i++ {
			sprites[i] = g.spriteManager.GetGemSprite(i)
		}
		g.board.SetGemSprites(sprites)
	}
	
	g.score = 0
	g.moves = 0
	g.comboCounter = 0
	g.levelStartTime = time.Now()
	g.levelManager.Reset()
	
	fmt.Printf("Новая игра! Уровень %d\n", level.Number)
}

// drawGame отрисовывает игровое поле и UI
func (g *Game) drawGame(screen *ebiten.Image) {
	// Отрисовка доски
	g.board.Draw(screen)
	
	// Отрисовка подсказок
	g.hintSystem.Draw(screen, g.board.OffsetX, g.board.OffsetY, g.board.TileSize)

	// Отрисовка UI (счёт, ходы)
	g.uiManager.DrawHUD(screen, g.score, g.moves)
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
	case StatePaused:
		g.drawGame(screen)
		g.uiManager.DrawPaused(screen)
	case StateSettings:
		g.uiManager.DrawSettings(screen, g.soundManager)
	case StateGameOver:
		g.uiManager.DrawGameOver(screen, g.score)
	case StateLevelComplete:
		g.uiManager.DrawLevelComplete(screen, g.levelManager.GetCurrentLevelNumber())
	}
}

// handleInput обрабатывает ввод игрока
func (g *Game) handleInput() {
	// Выход в меню
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StateMenu
		return
	}

	// Обработка клика мыши
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		tile := g.board.GetTileAt(g.mouseX, g.mouseY)
		if tile != nil {
			g.handleTileClick(tile)
		}
	}

	// Обработка клавиш для обмена (Shift + стрелки)
	if ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight) {
		if g.selectedTile != nil {
			targetRow, targetCol := g.selectedTile.Row, g.selectedTile.Col
			
			if inpututil.IsKeyJustPressed(ebiten.KeyUp) && targetRow > 0 {
				targetRow--
			} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) && targetRow < g.board.Rows-1 {
				targetRow++
			} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && targetCol > 0 {
				targetCol--
			} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) && targetCol < g.board.Cols-1 {
				targetCol++
			} else {
				return
			}
			
			targetTile := g.board.Tiles[targetRow][targetCol]
			g.executeSwap(g.selectedTile, targetTile)
		}
	}
}

// handleTileClick обрабатывает клик по камню
func (g *Game) handleTileClick(tile *logic.Tile) {
	// Сбрасываем подсказку при любом взаимодействии
	g.hintSystem.HideHint()
	
	if g.selectedTile == nil {
		// Первый клик - выбор камня
		g.selectedTile = tile
		tile.Selected = true
	} else if g.selectedTile == tile {
		// Повторный клик - отмена выбора
		g.selectedTile.Selected = false
		g.selectedTile = nil
	} else {
		// Клик по другому камню - попытка обмена
		g.executeSwap(g.selectedTile, tile)
	}
}

// executeSwap выполняет обмен между двумя камнями
func (g *Game) executeSwap(from, to *logic.Tile) {
	// Сброс выделения
	if g.selectedTile != nil {
		g.selectedTile.Selected = false
	}
	g.selectedTile = nil

	// Проверка на соседство
	dr := abs(from.Row - to.Row)
	dc := abs(from.Col - to.Col)
	
	if dr+dc != 1 {
		// Не соседние - просто выбрать новый
		to.Selected = true
		g.selectedTile = to
		return
	}

	// Выполнение обмена
	success := g.board.SwapTiles(from, to)
	
	if success {
		g.moves++
		
		// Звук обмена
		g.soundManager.Play(SoundSwap)
		
		// Проверка на матчи
		matches := g.board.FindAllMatches()
		if len(matches) > 0 {
			// Есть матчи - удаляем и начисляем очки
			score := g.board.RemoveMatches()
			
			// Звук матча
			g.soundManager.Play(SoundMatch)
			
			// Комбо система
			g.comboCounter++
			
			// Бонус за комбо
			if g.comboCounter > 1 {
				score *= g.comboCounter
				g.effectSystem.SpawnComboEffect(from.Col*60+40, from.Row*60+150, g.comboCounter)
				g.soundManager.Play(SoundCombo)
			}
			
			// Эффекты для каждого матча
			for _, m := range matches {
				x := g.board.OffsetX + m.Col*g.board.TileSize
				y := g.board.OffsetY + m.Row*g.board.TileSize
				g.effectSystem.SpawnMatchEffect(x, y, logic.GemColors[m.Gem], 1)
				g.scorePopups.AddScorePopup(x, y, score/len(matches))
			}
			
			g.score += score
			fmt.Printf("Матч! +%d очков (комбо x%d, всего: %d)\n", score, g.comboCounter, g.score)
		} else {
			// Нет матчей - отменить обмен
			g.board.SwapTiles(from, to)
			g.comboCounter = 0 // Сброс комбо
			g.soundManager.Play(SoundInvalid)
			fmt.Println("Нет матча - обмен отменён")
		}
	}
}

// abs возвращает абсолютное значение
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// nextLevel переходит к следующему уровню
func (g *Game) nextLevel() {
	g.state = StatePlaying
	g.board = logic.NewBoard(8, 8)
	g.score = 0
	g.moves = 0
	g.comboCounter = 0
	fmt.Println("Следующий уровень!")
}

// toggleMute переключает режим mute
func (g *Game) toggleMute() {
	muted := g.soundManager.ToggleMute()
	fmt.Printf("Sound muted: %v\n", muted)
}
