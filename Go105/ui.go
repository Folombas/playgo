package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// drawMenu рисует главное меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	title := "GEOMETRIC MATCH"
	subtitle := "Press ENTER or SPACE to start"

	// Заголовок с пульсацией
	scale := 1.0 + 0.05*math.Sin(float64(ebiten.ActualFPS()*60))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(screenWidth/2-200), 150)

	// Рисуем текст
	drawText(screen, title, screenWidth/2-150, 200, 2.0, color.RGBA{255, 255, 255, 255})
	drawText(screen, subtitle, screenWidth/2-150, 300, 1.0, color.RGBA{200, 200, 200, 255})

	if g.highScore > 0 {
		drawText(screen, fmt.Sprintf("High Score: %d", g.highScore), screenWidth/2-100, 400, 1.0, color.RGBA{255, 215, 0, 255})
	}
}

// drawGame рисует игровой процесс
func (g *Game) drawGame(screen *ebiten.Image) {
	// Игровое поле (фон)
	g.drawBoard(screen)

	// Сетка
	g.drawGrid(screen)

	// Тень падения
	if g.current != nil && g.board != nil {
		g.drawDropShadow(screen)
	}

	// Зафиксированные блоки
	g.drawPlacedBlocks(screen)

	// Текущая фигура
	if g.current != nil {
		g.drawTetromino(screen, g.current, 1.0)
	}

	// UI панель
	g.drawUI(screen)
}

// drawBoard рисует фон поля
func (g *Game) drawBoard(screen *ebiten.Image) {
	boardW := boardCols * cellSize
	boardH := boardRows * cellSize

	// Тёмный фон поля
	vector.DrawFilledRect(screen,
		float32(boardOffsetX),
		float32(boardOffsetY),
		float32(boardW),
		float32(boardH),
		color.RGBA{30, 30, 50, 255},
		false,
	)

	// Рамка
	vector.StrokeRect(screen,
		float32(boardOffsetX),
		float32(boardOffsetY),
		float32(boardW),
		float32(boardH),
		2,
		color.RGBA{100, 100, 150, 255},
		false,
	)
}

// drawGrid рисует сетку
func (g *Game) drawGrid(screen *ebiten.Image) {
	for x := 0; x <= boardCols; x++ {
		xPos := boardOffsetX + x*cellSize
		vector.StrokeLine(screen,
			float32(xPos), float32(boardOffsetY),
			float32(xPos), float32(boardOffsetY+boardRows*cellSize),
			1,
			color.RGBA{50, 50, 80, 128},
			false,
		)
	}
	for y := 0; y <= boardRows; y++ {
		yPos := boardOffsetY + y*cellSize
		vector.StrokeLine(screen,
			float32(boardOffsetX), float32(yPos),
			float32(boardOffsetX+boardCols*cellSize), float32(yPos),
			1,
			color.RGBA{50, 50, 80, 128},
			false,
		)
	}
}

// drawDropShadow рисует тень места падения
func (g *Game) drawDropShadow(screen *ebiten.Image) {
	dropY := g.board.GetDropY(g.current)
	shadow := &Tetromino{
		Type:     g.current.Type,
		Rotation: g.current.Rotation,
		X:        g.current.X,
		Y:        dropY,
	}
	g.drawTetromino(screen, shadow, 0.3)
}

// drawPlacedBlocks рисует зафиксированные блоки
func (g *Game) drawPlacedBlocks(screen *ebiten.Image) {
	for y := 0; y < boardRows; y++ {
		for x := 0; x < boardCols; x++ {
			if g.board.Grid[y][x] != nil {
				g.drawCell(screen, x, y, g.board.Grid[y][x].(color.RGBA), 1.0)
			}
		}
	}
}

// drawTetromino рисует фигуру
func (g *Game) drawTetromino(screen *ebiten.Image, t *Tetromino, alpha float64) {
	c := Colors[t.Type]
	c.A = uint8(float64(c.A) * alpha)

	for y, row := range t.Shape() {
		for x, cell := range row {
			if cell == 0 {
				continue
			}
			boardX := t.X + x
			boardY := t.Y + y
			if boardY >= 0 {
				g.drawCell(screen, boardX, boardY, c, alpha)
			}
		}
	}
}

// drawCell рисует один блок с закруглёнными углами
func (g *Game) drawCell(screen *ebiten.Image, x, y int, c color.RGBA, alpha float64) {
	px := boardOffsetX + x*cellSize
	py := boardOffsetY + y*cellSize

	// Основной блок
	vector.DrawFilledRect(screen,
		float32(px+1), float32(py+1),
		float32(cellSize-2), float32(cellSize-2),
		c,
		false,
	)

	// Блик (градиент сверху)
	highlight := color.RGBA{
		R: uint8(min(255, int(c.R)+50)),
		G: uint8(min(255, int(c.G)+50)),
		B: uint8(min(255, int(c.B)+50)),
		A: uint8(float64(c.A) * 0.3),
	}
	vector.DrawFilledRect(screen,
		float32(px+2), float32(py+2),
		float32(cellSize-4), float32(cellSize/3),
		highlight,
		false,
	)
}

// drawUI рисует интерфейс: счёт, уровень, следующая фигура
func (g *Game) drawUI(screen *ebiten.Image) {
	uiX := boardOffsetX + boardCols*cellSize + 30

	drawText(screen, "SCORE", uiX, boardOffsetY+20, 1.0, color.RGBA{200, 200, 200, 255})
	drawText(screen, fmt.Sprintf("%d", g.score), uiX, boardOffsetY+40, 1.5, color.RGBA{255, 255, 255, 255})

	drawText(screen, "LEVEL", uiX, boardOffsetY+100, 1.0, color.RGBA{200, 200, 200, 255})
	drawText(screen, fmt.Sprintf("%d", g.level), uiX, boardOffsetY+120, 1.5, color.RGBA{255, 255, 255, 255})

	drawText(screen, "LINES", uiX, boardOffsetY+180, 1.0, color.RGBA{200, 200, 200, 255})
	drawText(screen, fmt.Sprintf("%d", g.lines), uiX, boardOffsetY+200, 1.5, color.RGBA{255, 255, 255, 255})

	// Следующая фигура
	drawText(screen, "NEXT", uiX, boardOffsetY+280, 1.0, color.RGBA{200, 200, 200, 255})
	if g.next != nil {
		g.drawNextPiece(screen, uiX+20, boardOffsetY+310)
	}
}

// drawNextPiece рисует превью следующей фигуры
func (g *Game) drawNextPiece(screen *ebiten.Image, x, y int) {
	c := Colors[g.next.Type]
	for rowIdx, row := range g.next.Shape() {
		for colIdx, cell := range row {
			if cell == 0 {
				continue
			}
			vector.DrawFilledRect(screen,
				float32(x+colIdx*20), float32(y+rowIdx*20),
				18, 18,
				c,
				false,
			)
		}
	}
}

// drawPaused рисует экран паузы
func (g *Game) drawPaused(screen *ebiten.Image) {
	// Затемнение
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150}, false)
	drawText(screen, "PAUSED", screenWidth/2-80, screenHeight/2-30, 2.0, color.RGBA{255, 255, 255, 255})
	drawText(screen, "Press P to resume", screenWidth/2-100, screenHeight/2+30, 1.0, color.RGBA{200, 200, 200, 255})
}

// drawGameOver рисует экран конца игры
func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Затемнение с анимацией
	alpha := uint8(min(200, int(g.gameOverAnim*3)))
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, alpha}, false)

	drawText(screen, "GAME OVER", screenWidth/2-120, screenHeight/2-50, 2.0, color.RGBA{255, 50, 50, 255})
	drawText(screen, fmt.Sprintf("Score: %d", g.score), screenWidth/2-80, screenHeight/2+10, 1.5, color.RGBA{255, 255, 255, 255})

	if g.score >= g.highScore {
		drawText(screen, "NEW HIGH SCORE!", screenWidth/2-120, screenHeight/2+50, 1.0, color.RGBA{255, 215, 0, 255})
	}

	if g.gameOverAnim > 60 {
		drawText(screen, "Press ENTER to restart", screenWidth/2-120, screenHeight/2+100, 1.0, color.RGBA{200, 200, 200, 255})
	}
}

// drawText рисует текст
func drawText(screen *ebiten.Image, str string, x, y int, scale float64, c color.Color) {
	// Для простоты используем text/v1 API
	// Масштабирование пока не поддерживается (будет добавлено позже)
	text.Draw(screen, str, basicfont.Face7x13, x, y, c)
}
