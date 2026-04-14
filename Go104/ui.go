package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// UI управляет отрисовкой пользовательского интерфейса
type UI struct {
	font      text.Face
	fontSize  float64
	gameOver  bool
	paused    bool
}

// NewUI создаёт новый UI
func NewUI() *UI {
	// Загружаем шрифт (используем встроенный шрифт)
	fontFace := text.NewFace(24)

	return &UI{
		font:     fontFace,
		fontSize: 24,
	}
}

// DrawBoard отрисовывает игровое поле
func (ui *UI) DrawBoard(screen *ebiten.Image, board *Board, boardOffsetX, boardOffsetY float64) {
	// Фон доски
	vector.DrawFilledRect(
		screen,
		float32(boardOffsetX-5), float32(boardOffsetY-5),
		float32(BoardPixel+10), float32(BoardPixel+10),
		color.RGBA{R: 60, G: 60, B: 80, A: 255},
		false,
	)

	// Ячейки доски
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			x := boardOffsetX + float64(c*CellSize)
			y := boardOffsetY + float64(r*CellSize)

			// Чередующиеся цвета ячеек
			if (r+c)%2 == 0 {
				vector.DrawFilledRect(
					screen,
					float32(x), float32(y),
					float32(CellSize), float32(CellSize),
					color.RGBA{R: 40, G: 40, B: 60, A: 255},
					false,
				)
			} else {
				vector.DrawFilledRect(
					screen,
					float32(x), float32(y),
					float32(CellSize), float32(CellSize),
					color.RGBA{R: 50, G: 50, B: 70, A: 255},
					false,
				)
			}
		}
	}

	// Отрисовка фишек
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			tile := board.Grid[r][c]
			if tile == nil || tile.Removing {
				continue
			}

			ui.drawTile(screen, tile, boardOffsetX, boardOffsetY)
		}
	}
}

// drawTile отрисовывает одну фишку
func (ui *UI) drawTile(screen *ebiten.Image, tile *Tile, boardOffsetX, boardOffsetY float64) {
	// Позиция с учётом анимации
	x := boardOffsetX + tile.X
	y := boardOffsetY + tile.Y

	// Применяем масштаб и альфа
	scale := tile.Scale
	alpha := tile.Alpha

	if scale <= 0 || alpha <= 0 {
		return
	}

	// Подсветка выбранной фишки
	if tile.Selected {
		vector.DrawFilledRect(
			screen,
			float32(x-3), float32(y-3),
			float32(CellSize+6), float32(CellSize+6),
			color.RGBA{R: 255, G: 255, B: 0, A: uint8(200 * alpha)},
			false,
		)
	}

	// Спрайт фишки
	if tile.Color < len(TileImages) && TileImages[tile.Color] != nil {
		img := TileImages[tile.Color]

		// Масштабируем и рисуем
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x+4, y+4) // Небольшой отступ
		op.ColorScale.ScaleAlpha(float32(alpha))

		screen.DrawImage(img, op)
	} else {
		// Fallback: цветной прямоугольник
		colors := []color.RGBA{
			{R: 255, G: 50, B: 50, A: 255},  // красный
			{R: 50, G: 50, B: 255, A: 255},   // синий
			{R: 50, G: 255, B: 50, A: 255},   // зелёный
			{R: 255, G: 255, B: 50, A: 255},  // жёлтый
			{R: 180, G: 50, B: 255, A: 255},  // фиолетовый
			{R: 255, G: 165, B: 0, A: 255},   // оранжевый
		}

		c := colors[tile.Color%len(colors)]
		c.A = uint8(float64(c.A) * alpha)

		vector.DrawFilledRect(
			screen,
			float32(x+8*scale), float32(y+8*scale),
			float32((CellSize-16)*scale), float32((CellSize-16)*scale),
			c,
			false,
		)
	}
}

// DrawScore отрисовывает счёт
func (ui *UI) DrawScore(screen *ebiten.Image, score int, highScore int) {
	scoreText := fmt.Sprintf("Score: %d", score)
	highText := fmt.Sprintf("Best: %d", highScore)

	// Счёт
	text.Draw(screen, scoreText, ui.font, 20, 30, color.White)
	// Рекорд
	text.Draw(screen, highText, ui.font, 20, 60, color.RGBA{R: 255, G: 215, B: 0, A: 255})
}

// DrawTimer отрисовывает таймер
func (ui *UI) DrawTimer(screen *ebiten.Image, timeLeft int) {
	minutes := timeLeft / 60
	seconds := timeLeft % 60
	timerText := fmt.Sprintf("Time: %02d:%02d", minutes, seconds)

	// Красный, если мало времени
	timerColor := color.White
	if timeLeft <= 10 {
		timerColor = color.RGBA{R: 255, G: 50, B: 50, A: 255}
	}

	text.Draw(screen, timerText, ui.font, 680, 30, timerColor)
}

// DrawNewGameButton отрисовывает кнопку "Новая игра"
func (ui *UI) DrawNewGameButton(screen *ebiten.Image, x, y, w, h float64) {
	// Фон кнопки
	vector.DrawFilledRect(
		screen,
		float32(x), float32(y),
		float32(w), float32(h),
		color.RGBA{R: 70, G: 130, B: 180, A: 255},
		false,
	)

	// Обводка
	vector.StrokeRect(
		screen,
		float32(x), float32(y),
		float32(w), float32(h),
		2,
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
		false,
	)

	// Текст
	buttonText := "New Game"
	textWidth := float64(len(buttonText) * 12) // Приблизительно
	textX := x + (w-textWidth)/2
	textY := y + h/2 + 8

	text.Draw(screen, buttonText, ui.font, int(textX), int(textY), color.White)
}

// DrawGameOver отрисовывает экран окончания игры
func (ui *UI) DrawGameOver(screen *ebiten.Image, finalScore int, windowWidth, windowHeight int) {
	// Затемнённый фон
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(windowWidth), float32(windowHeight),
		color.RGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	// Текст "Game Over"
	gameOverText := "Game Over!"
	gowX := float64(windowWidth)/2 - 100
	gowY := float64(windowHeight)/2 - 80
	text.Draw(screen, gameOverText, ui.font, int(gowX), int(gowY), color.RGBA{R: 255, G: 50, B: 50, A: 255})

	// Финальный счёт
	scoreText := fmt.Sprintf("Final Score: %d", finalScore)
	scoreX := float64(windowWidth)/2 - 80
	scoreY := float64(windowHeight)/2 - 20
	text.Draw(screen, scoreText, ui.font, int(scoreX), int(scoreY), color.White)

	// Кнопка "Play Again"
	buttonW := 200.0
	buttonH := 50.0
	buttonX := float64(windowWidth)/2 - buttonW/2
	buttonY := float64(windowHeight)/2 + 40

	vector.DrawFilledRect(
		screen,
		float32(buttonX), float32(buttonY),
		float32(buttonW), float32(buttonH),
		color.RGBA{R: 70, G: 130, B: 180, A: 255},
		false,
	)

	vector.StrokeRect(
		screen,
		float32(buttonX), float32(buttonY),
		float32(buttonW), float32(buttonH),
		2,
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
		false,
	)

	playText := "Play Again"
	playX := buttonX + (buttonW - float64(len(playText)*12))/2
	playY := buttonY + buttonH/2 + 8
	text.Draw(screen, playText, ui.font, int(playX), int(playY), color.White)
}

// DrawPauseScreen отрисовывает экран паузы
func (ui *UI) DrawPauseScreen(screen *ebiten.Image, windowWidth, windowHeight int) {
	// Затемнённый фон
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(windowWidth), float32(windowHeight),
		color.RGBA{R: 0, G: 0, B: 0, A: 128},
		false,
	)

	// Текст "Paused"
	pauseText := "PAUSED"
	pX := float64(windowWidth)/2 - 60
	pY := float64(windowHeight)/2
	text.Draw(screen, pauseText, ui.font, int(pX), int(pY), color.White)
}

// DrawHint отрисовывает подсказку
func (ui *UI) DrawHint(screen *ebiten.Image, p1, p2 *Position, boardOffsetX, boardOffsetY float64) {
	if p1 == nil || p2 == nil {
		return
	}

	// Зелёная обводка для подсказки
	positions := []*Position{p1, p2}
	for _, pos := range positions {
		x := boardOffsetX + float64(pos.Col*CellSize)
		y := boardOffsetY + float64(pos.Row*CellSize)

		vector.StrokeRect(
			screen,
			float32(x-2), float32(y-2),
			float32(CellSize+4), float32(CellSize+4),
			3,
			color.RGBA{R: 0, G: 255, B: 0, A: 200},
			false,
		)
	}
}

func init() {
	// Инициализация шрифта
	log.Println("UI package initialized")
}
