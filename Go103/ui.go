package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// Цвета UI
var (
	colorWhite       = color.RGBA{255, 255, 255, 255}
	colorBlack       = color.RGBA{0, 0, 0, 255}
	colorYellow      = color.RGBA{255, 204, 0, 255}
	colorLightYellow = color.RGBA{255, 230, 100, 200}
	colorGreen       = color.RGBA{52, 199, 89, 255}
	colorRed         = color.RGBA{255, 69, 58, 255}
	colorBlue        = color.RGBA{0, 122, 255, 255}
	colorOverlay     = color.RGBA{0, 0, 0, 180} // Полупрозрачный чёрный
)

// UIRenderer отвечает за отрисовку всех UI элементов
type UIRenderer struct {
	score      int
	highScore  int
	timer      int
	gameOver   bool
	paused     bool
}

// NewUIRenderer создаёт новый UI рендерер
func NewUIRenderer() *UIRenderer {
	return &UIRenderer{
		score:     0,
		highScore: loadHighScore(),
		timer:     60,
	}
}

// Draw отрисовывает все UI элементы
func (ui *UIRenderer) Draw(screen *ebiten.Image, boardOffsetX, boardOffsetY, cellSize float64,
	hintPos1, hintPos2 [2]int, hintActive bool, animSystem *AnimationSystem) {

	// Рисуем фон
	ui.drawBackground(screen)

	// Рисуем игровое поле (рамка)
	ui.drawBoardFrame(screen, boardOffsetX, boardOffsetY, cellSize)

	// Рисуем фишки
	ui.drawTiles(screen, boardOffsetX, boardOffsetY, cellSize, animSystem)

	// Рисуем подсветку выбранной фишки
	// (это будет делать Game, т.к. там есть состояние selectedTile)

	// Рисуем подсказку
	if hintActive {
		ui.drawHint(screen, boardOffsetX, boardOffsetY, cellSize, hintPos1, hintPos2)
	}

	// Рисуем счёт и таймер
	ui.drawHUD(screen)

	// Рисуем кнопку "Новая игра"
	ui.drawNewGameButton(screen)

	// Рисуем экран Game Over
	if ui.gameOver {
		ui.drawGameOver(screen)
	}

	// Рисуем экран паузы
	if ui.paused && !ui.gameOver {
		ui.drawPaused(screen)
	}
}

// drawBackground рисует фон
func (ui *UIRenderer) drawBackground(screen *ebiten.Image) {
	if BackgroundSprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(800.0/float64(BackgroundSprite.Bounds().Dx()), 800.0/float64(BackgroundSprite.Bounds().Dy()))
		screen.DrawImage(BackgroundSprite, op)
	} else {
		// Fallback - просто тёмный фон
		vector.DrawFilledRect(screen, 0, 0, 800, 800, color.RGBA{30, 30, 60, 255}, false)
	}
}

// drawBoardFrame рисует рамку игрового поля
func (ui *UIRenderer) drawBoardFrame(screen *ebiten.Image, offsetX, offsetY, cellSize float64) {
	width := float64(BoardSize) * cellSize
	height := float64(BoardSize) * cellSize

	// Фон поля
	vector.DrawFilledRect(screen,
		float32(offsetX), float32(offsetY),
		float32(width), float32(height),
		color.RGBA{20, 20, 40, 255}, false)

	// Рамка
	vector.StrokeRect(screen,
		float32(offsetX), float32(offsetY),
		float32(width), float32(height),
		4, color.RGBA{100, 100, 150, 255}, false)
}

// drawTiles рисует все фишки на поле
func (ui *UIRenderer) drawTiles(screen *ebiten.Image, offsetX, offsetY, cellSize float64,
	animSystem *AnimationSystem) {

	// Получаем доступ к полю через глобальную переменную (Game передаст данные)
	// Эта функция будет вызываться из Game.Draw с актуальным состоянием
}

// DrawTilesWithBoard рисует фишки с данными доски
func (ui *UIRenderer) DrawTilesWithBoard(screen *ebiten.Image, offsetX, offsetY, cellSize float64,
	board *Board, selectedTile *[2]int, animSystem *AnimationSystem) {

	grid := board.Grid()

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			tileType := grid[row][col]
			if tileType == -1 {
				continue // Пустая ячейка
			}

			// Позиция на экране
			x := offsetX + float64(col)*cellSize
			y := offsetY + float64(row)*cellSize

			// Применяем анимации
			alpha := animSystem.GetMatchAlpha(row, col)
			scale := animSystem.GetMatchScale(row, col)
			dropOffset := animSystem.GetDropOffset(row, col)
			shakeX, shakeY := animSystem.GetShakeOffset(row, col)

			// Если фишка удаляется - пропускаем при alpha=0
			if alpha <= 0 {
				continue
			}

			// Смещение из-за падения
			y += dropOffset

			// Смещение из-за дрожания
			x += shakeX
			y += shakeY

			// Рисуем фишку
			ui.drawTile(screen, tileType, x, y, cellSize, scale, alpha)

			// Подсветка выбранной фишки
			if selectedTile != nil && (*selectedTile)[0] == row && (*selectedTile)[1] == col {
				ui.drawSelection(screen, x, y, cellSize)
			}
		}
	}
}

// drawTile рисует одну фишку
func (ui *UIRenderer) drawTile(screen *ebiten.Image, tileType int, x, y, cellSize, scale, alpha float64) {
	// Центр фишки
	centerX := x + cellSize/2
	centerY := y + cellSize/2

	op := &ebiten.DrawImageOptions{}

	// Масштабирование относительно центра
	op.GeoM.Translate(-cellSize/2, -cellSize/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(centerX, centerY)

	// Прозрачность
	op.ColorScale.ScaleAlpha(float32(alpha))

	if TileSprites[tileType] != nil {
		// Масштабируем спрайт под размер ячейки
		spriteSize := float64(TileSprites[tileType].Bounds().Dx())
		cellScale := cellSize / spriteSize
		op.GeoM.Reset()
		op.GeoM.Translate(-spriteSize/2, -spriteSize/2)
		op.GeoM.Scale(cellScale*scale, cellScale*scale)
		op.GeoM.Translate(centerX, centerY)
		op.ColorScale.ScaleAlpha(float32(alpha))

		screen.DrawImage(TileSprites[tileType], op)
	} else {
		// Fallback - рисуем цветной круг
		ui.drawFallbackTile(screen, tileType, centerX, centerY, cellSize*scale/2, alpha)
	}
}

// drawFallbackTile рисует фишку программно (круг с цветом)
func (ui *UIRenderer) drawFallbackTile(screen *ebiten.Image, tileType int, cx, cy, radius, alpha float64) {
	colors := []color.RGBA{
		{255, 69, 58, uint8(alpha * 255)},  // Красный
		{0, 122, 255, uint8(alpha * 255)},  // Синий
		{52, 199, 89, uint8(alpha * 255)},  // Зелёный
		{255, 204, 0, uint8(alpha * 255)},  // Жёлтый
		{175, 82, 222, uint8(alpha * 255)}, // Фиолетовый
		{255, 149, 0, uint8(alpha * 255)},  // Оранжевый
	}

	c := colors[tileType%len(colors)]

	// Рисуем круг (векторный)
	vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(radius), c, false)

	// Обводка
	vector.StrokeCircle(screen, float32(cx), float32(cy), float32(radius), 2,
		color.RGBA{255, 255, 255, uint8(alpha * 100)}, false)
}

// drawSelection рисует подсветку выбранной фишки
func (ui *UIRenderer) drawSelection(screen *ebiten.Image, x, y, cellSize float64) {
	// Жёлтая обводка
	vector.StrokeRect(screen,
		float32(x)+2, float32(y)+2,
		float32(cellSize)-4, float32(cellSize)-4,
		3, colorYellow, false)
}

// drawHint рисует подсказку (пульсирующая зелёная обводка)
func (ui *UIRenderer) drawHint(screen *ebiten.Image, offsetX, offsetY, cellSize float64,
	pos1, pos2 [2]int) {

	pulse := (animSystemGlobal.GetHintPulse() + 1) / 2 // 0-1

	c := color.RGBA{
		R: uint8(52),
		G: uint8(199),
		B: uint8(89),
		A: uint8(pulse * 200),
	}

	// Первая позиция
	x1 := offsetX + float64(pos1[1])*cellSize
	y1 := offsetY + float64(pos1[0])*cellSize
	vector.StrokeRect(screen,
		float32(x1)+2, float32(y1)+2,
		float32(cellSize)-4, float32(cellSize)-4,
		3, c, false)

	// Вторая позиция
	x2 := offsetX + float64(pos2[1])*cellSize
	y2 := offsetY + float64(pos2[0])*cellSize
	vector.StrokeRect(screen,
		float32(x2)+2, float32(y2)+2,
		float32(cellSize)-4, float32(cellSize)-4,
		3, c, false)
}

// drawHUD рисует счёт и таймер
func (ui *UIRenderer) drawHUD(screen *ebiten.Image) {
	// Счёт (левый верхний угол)
	scoreText := fmt.Sprintf("Score: %d", ui.score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 20, 30, colorWhite)

	// Рекорд
	if ui.highScore > 0 {
		highText := fmt.Sprintf("Best: %d", ui.highScore)
		text.Draw(screen, highText, basicfont.Face7x13, 20, 50, colorYellow)
	}

	// Таймер (правый верхний угол)
	minutes := ui.timer / 60
	seconds := ui.timer % 60
	timerText := fmt.Sprintf("Time: %02d:%02d", minutes, seconds)
	text.Draw(screen, timerText, basicfont.Face7x13, 680, 30, colorWhite)
}

// drawNewGameButton рисует кнопку "Новая игра"
func (ui *UIRenderer) drawNewGameButton(screen *ebiten.Image) {
	btnX, btnY, btnW, btnH := 300.0, 10.0, 200.0, 40.0

	// Фон кнопки
	vector.DrawFilledRect(screen,
		float32(btnX), float32(btnY),
		float32(btnW), float32(btnH),
		colorBlue, false)

	// Рамка
	vector.StrokeRect(screen,
		float32(btnX), float32(btnY),
		float32(btnW), float32(btnH),
		2, colorWhite, false)

	// Текст
	text.Draw(screen, "New Game (R)", basicfont.Face7x13, 340, 35, colorWhite)
}

// drawGameOver рисует экран Game Over
func (ui *UIRenderer) drawGameOver(screen *ebiten.Image) {
	// Затемнение
	vector.DrawFilledRect(screen, 0, 0, 800, 800, colorOverlay, false)

	// Game Over текст
	text.Draw(screen, "GAME OVER!", basicfont.Face7x13, 340, 350, colorRed)

	// Финальный счёт
	scoreText := fmt.Sprintf("Final Score: %d", ui.score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 340, 400, colorWhite)

	// Рекорд
	if ui.score >= ui.highScore && ui.score > 0 {
		text.Draw(screen, "NEW HIGH SCORE!", basicfont.Face7x13, 320, 430, colorYellow)
	}

	// Кнопка Play Again
	btnX, btnY, btnW, btnH := 300.0, 470.0, 200.0, 40.0
	vector.DrawFilledRect(screen,
		float32(btnX), float32(btnY),
		float32(btnW), float32(btnH),
		colorGreen, false)
	text.Draw(screen, "Play Again (R)", basicfont.Face7x13, 335, 495, colorWhite)
}

// drawPaused рисует экран паузы
func (ui *UIRenderer) drawPaused(screen *ebiten.Image) {
	// Затемнение
	vector.DrawFilledRect(screen, 0, 0, 800, 800, colorOverlay, false)

	// Paused текст
	text.Draw(screen, "PAUSED", basicfont.Face7x13, 350, 400, colorYellow)
	text.Draw(screen, "Press P to continue", basicfont.Face7x13, 310, 430, colorWhite)
}

// UpdateScore обновляет счёт
func (ui *UIRenderer) UpdateScore(score int) {
	ui.score = score
	if score > ui.highScore {
		ui.highScore = score
		saveHighScore(score)
	}
}

// UpdateTimer обновляет таймер
func (ui *UIRenderer) UpdateTimer(timer int) {
	ui.timer = timer
}

// SetGameOver активирует экран Game Over
func (ui *UIRenderer) SetGameOver() {
	ui.gameOver = true
}

// IsGameOver возвращает true если игра окончена
func (ui *UIRenderer) IsGameOver() bool {
	return ui.gameOver
}

// TogglePause переключает паузу
func (ui *UIRenderer) TogglePause() {
	if !ui.gameOver {
		ui.paused = !ui.paused
	}
}

// IsPaused возвращает true если пауза
func (ui *UIRenderer) IsPaused() bool {
	return ui.paused
}

// Reset сбрасывает UI для новой игры
func (ui *UIRenderer) Reset() {
	ui.score = 0
	ui.timer = 60
	ui.gameOver = false
	ui.paused = false
}

// GetNewGameButtonBounds возвращает границы кнопки "Новая игра"
func GetNewGameButtonBounds() (float64, float64, float64, float64) {
	return 300.0, 10.0, 200.0, 40.0
}

// GetGameOverButtonBounds возвращает границы кнопки "Play Again"
func GetGameOverButtonBounds() (float64, float64, float64, float64) {
	return 300.0, 470.0, 200.0, 40.0
}

// loadHighScore загружает рекорд из файла
func loadHighScore() int {
	// TODO: Реализовать загрузку из JSON файла
	// Пока возвращаем 0
	return 0
}

// saveHighScore сохраняет рекорд в файл
func saveHighScore(score int) {
	// TODO: Реализовать сохранение в JSON файл
	log.Printf("Новый рекорд: %d", score)
}

// Глобальная ссылка на AnimationSystem для drawHint
var animSystemGlobal *AnimationSystem

// SetGlobalAnimationSystem устанавливает глобальную ссылку
func SetGlobalAnimationSystem(as *AnimationSystem) {
	animSystemGlobal = as
}
