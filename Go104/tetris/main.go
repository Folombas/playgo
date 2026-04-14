// Tetris Game - Go365 Challenge (Go104)
// Классический тетрис с геометрическими фигурами
// Управление: стрелки/WSAD для движения, вверх/W для вращения, пробел для ускорения

package main

import (
	"bytes"
	"embed"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

//go:embed sprites/*.png
var spritesFS embed.FS

// Константы игры
const (
	BoardWidth   = 10
	BoardHeight  = 20
	CellSize     = 32
	ScreenWidth  = 480
	ScreenHeight = 720
)

// Цвета фигур (I, J, L, O, T, S, Z)
var PieceColors = []color.RGBA{
	{R: 0, G: 255, B: 255, A: 255},   // I - cyan
	{R: 0, G: 0, B: 255, A: 255},     // J - blue
	{R: 255, G: 165, B: 0, A: 255},   // L - orange
	{R: 255, G: 255, B: 0, A: 255},   // O - yellow
	{R: 128, G: 0, B: 128, A: 255},   // T - purple
	{R: 0, G: 255, B: 0, A: 255},     // S - green
	{R: 255, G: 0, B: 0, A: 255},     // Z - red
}

// Формы фигур (каждая - список позиций блоков)
var PieceShapes = [][][]int{
	// I
	{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
	// J
	{{0, 0}, {0, 1}, {1, 1}, {2, 1}},
	// L
	{{2, 0}, {0, 1}, {1, 1}, {2, 1}},
	// O
	{{0, 0}, {1, 0}, {0, 1}, {1, 1}},
	// T
	{{1, 0}, {0, 1}, {1, 1}, {2, 1}},
	// S
	{{1, 0}, {2, 0}, {0, 1}, {1, 1}},
	// Z
	{{0, 0}, {1, 0}, {1, 1}, {2, 1}},
}

// Block представляет один блок
type Block struct {
	X, Y int
}

// Piece представляет текущую фигуру
type Piece struct {
	Blocks []Block
	Color  color.RGBA
	X, Y   int // позиция на доске
	Type   int
}

// Game - основная структура игры
type Game struct {
	board         [][]int // 0 = пусто, 1-7 = цвет
	score         int
	level         int
	lines         int
	gameOver      bool
	paused        bool
	currentPiece  *Piece
	nextPieceType int
	dropTimer     float64
	dropInterval  float64
	blockImages   []*ebiten.Image
	boardOffsetX  float64
	boardOffsetY  float64
}

// NewGame создаёт новую игру
func NewGame() *Game {
	g := &Game{
		board:        make([][]int, BoardWidth),
		dropInterval: 1.0, // секунд между падениями
		dropTimer:    0,
		score:        0,
		level:        1,
		lines:        0,
		gameOver:     false,
		paused:       false,
	}

	// Инициализация доски
	for i := range g.board {
		g.board[i] = make([]int, BoardHeight)
	}

	// Загрузка спрайтов
	g.loadSprites()

	// Расчёт смещения доски для центрирования
	g.boardOffsetX = float64(ScreenWidth-BoardWidth*CellSize) / 2
	g.boardOffsetY = 60

	// Создание первой фигуры
	g.nextPieceType = rand.Intn(len(PieceShapes))
	g.spawnPiece()

	return g
}

// loadSprites загружает спрайты блоков
func (g *Game) loadSprites() {
	g.blockImages = make([]*ebiten.Image, 8)

	colorNames := []string{"gray", "cyan", "blue", "orange", "yellow", "purple", "green", "red"}
	for i, name := range colorNames {
		data, err := spritesFS.ReadFile("sprites/block_" + name + ".png")
		if err != nil {
			log.Printf("Warning: could not load sprite %s", name)
			g.blockImages[i] = g.createFallbackImage(PieceColors[i%len(PieceColors)])
			continue
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			log.Printf("Warning: could not decode sprite %s", name)
			g.blockImages[i] = g.createFallbackImage(PieceColors[i%len(PieceColors)])
			continue
		}
		g.blockImages[i] = ebiten.NewImageFromImage(img)
	}
}

// createFallbackImage создаёт цветной блок если спрайт не загрузился
func (g *Game) createFallbackImage(clr color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(CellSize, CellSize)
	vector.DrawFilledRect(img, 2, 2, float32(CellSize-4), float32(CellSize-4), clr, false)
	vector.StrokeRect(img, 2, 2, float32(CellSize-4), float32(CellSize-4), 2, color.Black, false)
	return img
}

// spawnPiece создаёт новую фигуру
func (g *Game) spawnPiece() {
	shape := PieceShapes[g.nextPieceType]

	blocks := make([]Block, len(shape))
	for i, pos := range shape {
		blocks[i] = Block{X: pos[0], Y: pos[1]}
	}

	g.currentPiece = &Piece{
		Blocks: blocks,
		Color:  PieceColors[g.nextPieceType],
		X:      BoardWidth/2 - 2,
		Y:      0,
		Type:   g.nextPieceType,
	}

	// Следующая фигура
	g.nextPieceType = rand.Intn(len(PieceShapes))

	// Проверка на game over
	if !g.isValidPosition(g.currentPiece.X, g.currentPiece.Y, g.currentPiece.Blocks) {
		g.gameOver = true
	}
}

// isValidPosition проверяет валидность позиции фигуры
func (g *Game) isValidPosition(px, py int, blocks []Block) bool {
	for _, b := range blocks {
		x := px + b.X
		y := py + b.Y

		if x < 0 || x >= BoardWidth || y >= BoardHeight {
			return false
		}
		if y >= 0 && g.board[x][y] != 0 {
			return false
		}
	}
	return true
}

// rotatePiece вращает фигуру
func (g *Game) rotatePiece() {
	if g.currentPiece == nil || g.gameOver || g.paused {
		return
	}

	// O не вращается
	if g.currentPiece.Type == 3 {
		return
	}

	// Находим центр вращения
	centerX, centerY := 1, 1

	// Вращаем на 90 градусов: (x,y) -> (-y,x) относительно центра
	newBlocks := make([]Block, len(g.currentPiece.Blocks))
	for i, b := range g.currentPiece.Blocks {
		dx := b.X - centerX
		dy := b.Y - centerY
		newBlocks[i] = Block{X: -dy + centerX, Y: dx + centerY}
	}

	// Проверяем валидность
	if g.isValidPosition(g.currentPiece.X, g.currentPiece.Y, newBlocks) {
		g.currentPiece.Blocks = newBlocks
	} else {
		// Wall kick - пробуем сдвинуть
		for _, offset := range []int{-1, 1, -2, 2} {
			if g.isValidPosition(g.currentPiece.X+offset, g.currentPiece.Y, newBlocks) {
				g.currentPiece.X += offset
				g.currentPiece.Blocks = newBlocks
				return
			}
		}
	}
}

// movePiece двигает фигуру
func (g *Game) movePiece(dx, dy int) {
	if g.currentPiece == nil || g.gameOver || g.paused {
		return
	}

	newX := g.currentPiece.X + dx
	newY := g.currentPiece.Y + dy

	if g.isValidPosition(newX, newY, g.currentPiece.Blocks) {
		g.currentPiece.X = newX
		g.currentPiece.Y = newY
	} else if dy > 0 {
		// Не можем двигаться вниз - фиксируем фигуру
		g.lockPiece()
	}
}

// lockPiece фиксирует фигуру на доске
func (g *Game) lockPiece() {
	if g.currentPiece == nil {
		return
	}

	for _, b := range g.currentPiece.Blocks {
		x := g.currentPiece.X + b.X
		y := g.currentPiece.Y + b.Y
		if y >= 0 && y < BoardHeight && x >= 0 && x < BoardWidth {
			g.board[x][y] = g.currentPiece.Type + 1
		}
	}

	// Проверяем заполненные линии
	g.clearLines()

	// Создаём новую фигуру
	g.spawnPiece()
	g.dropTimer = 0
}

// clearLines очищает заполненные линии
func (g *Game) clearLines() {
	clearedLines := 0

	for y := BoardHeight - 1; y >= 0; y-- {
		full := true
		for x := 0; x < BoardWidth; x++ {
			if g.board[x][y] == 0 {
				full = false
				break
			}
		}

		if full {
			clearedLines++
			// Сдвигаем всё вниз
			for yy := y; yy > 0; yy-- {
				for x := 0; x < BoardWidth; x++ {
					g.board[x][yy] = g.board[x][yy-1]
				}
			}
			// Очищаем верхнюю линию
			for x := 0; x < BoardWidth; x++ {
				g.board[x][0] = 0
			}
			y++ // Проверяем ту же строку снова
		}
	}

	// Начисляем очки
	if clearedLines > 0 {
		lineScores := []int{0, 100, 300, 500, 800}
		g.score += lineScores[clearedLines] * g.level
		g.lines += clearedLines
		g.level = g.lines/10 + 1
		g.dropInterval = max(0.1, 1.0-float64(g.level-1)*0.1)
	}
}

// hardDrop мгновенное падение
func (g *Game) hardDrop() {
	if g.currentPiece == nil || g.gameOver || g.paused {
		return
	}

	for g.isValidPosition(g.currentPiece.X, g.currentPiece.Y+1, g.currentPiece.Blocks) {
		g.currentPiece.Y++
		g.score += 2 // бонус за жёсткое падение
	}
	g.lockPiece()
}

// Update обновляет игру
func (g *Game) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyR) {
			*g = *NewGame()
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
		return nil
	}

	if g.paused {
		return nil
	}

	// Ввод
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.movePiece(-1, 0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.movePiece(1, 0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.movePiece(0, 1)
		g.score += 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.rotatePiece()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.hardDrop()
	}

	// Таймер падения
	g.dropTimer += 1.0 / 60.0 // Update вызывается 60 раз в секунду
	if g.dropTimer >= g.dropInterval {
		g.movePiece(0, 1)
		g.dropTimer = 0
	}

	return nil
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Фон
	screen.Fill(color.RGBA{R: 20, G: 20, B: 30, A: 255})

	// Доска
	g.drawBoard(screen)

	// Текущая фигура
	if g.currentPiece != nil && !g.gameOver {
		g.drawPiece(screen, g.currentPiece, g.currentPiece.X, g.currentPiece.Y)
		// Ghost piece (тень)
		g.drawGhostPiece(screen)
	}

	// UI
	g.drawUI(screen)

	// Game Over
	if g.gameOver {
		g.drawGameOver(screen)
	}

	// Pause
	if g.paused {
		g.drawPause(screen)
	}
}

// drawBoard отрисовывает доску
func (g *Game) drawBoard(screen *ebiten.Image) {
	// Фон доски
	vector.DrawFilledRect(
		screen,
		float32(g.boardOffsetX-2), float32(g.boardOffsetY-2),
		float32(BoardWidth*CellSize+4), float32(BoardHeight*CellSize+4),
		color.RGBA{R: 0, G: 0, B: 0, A: 255},
		false,
	)

	// Ячейки
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			if g.board[x][y] > 0 {
				imgIdx := g.board[x][y]
				if imgIdx < len(g.blockImages) && g.blockImages[imgIdx] != nil {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(
						g.boardOffsetX+float64(x*CellSize),
						g.boardOffsetY+float64(y*CellSize),
					)
					screen.DrawImage(g.blockImages[imgIdx], op)
				}
			}
		}
	}

	// Сетка
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			vector.StrokeRect(
				screen,
				float32(g.boardOffsetX+float64(x*CellSize)),
				float32(g.boardOffsetY+float64(y*CellSize)),
				float32(CellSize), float32(CellSize),
				1,
				color.RGBA{R: 40, G: 40, B: 60, A: 100},
				false,
			)
		}
	}
}

// drawPiece отрисовывает фигуру
func (g *Game) drawPiece(screen *ebiten.Image, piece *Piece, px, py int) {
	for _, b := range piece.Blocks {
		x := px + b.X
		y := py + b.Y

		if y >= 0 && x >= 0 && x < BoardWidth {
			imgIdx := piece.Type + 1
			if imgIdx < len(g.blockImages) && g.blockImages[imgIdx] != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(
					g.boardOffsetX+float64(x*CellSize),
					g.boardOffsetY+float64(y*CellSize),
				)
				screen.DrawImage(g.blockImages[imgIdx], op)
			}
		}
	}
}

// drawGhostPiece отрисовывает тень (где упадёт фигура)
func (g *Game) drawGhostPiece(screen *ebiten.Image) {
	if g.currentPiece == nil {
		return
	}

	ghostY := g.currentPiece.Y
	for g.isValidPosition(g.currentPiece.X, ghostY+1, g.currentPiece.Blocks) {
		ghostY++
	}

	// Рисуем полупрозрачную тень
	for _, b := range g.currentPiece.Blocks {
		x := g.currentPiece.X + b.X
		y := ghostY + b.Y

		if y >= 0 && x >= 0 && x < BoardWidth {
			vector.StrokeRect(
				screen,
				float32(g.boardOffsetX+float64(x*CellSize)+2),
				float32(g.boardOffsetY+float64(y*CellSize)+2),
				float32(CellSize-4), float32(CellSize-4),
				2,
				color.RGBA{R: 128, G: 128, B: 128, A: 100},
				false,
			)
		}
	}
}

// drawText простая отрисовка текста (векторная заглушка)
func drawText(screen *ebiten.Image, text string, x, y float64, clr color.RGBA) {
	// Рисуем каждую букву как прямоугольник (упрощённо)
	for i, ch := range text {
		if ch >= '0' && ch <= '9' {
			vector.DrawFilledRect(
				screen,
				float32(x+float64(i*12)), float32(y),
				10, 18,
				clr,
				false,
			)
		}
	}
}

// drawUI отрисовывает интерфейс
func (g *Game) drawUI(screen *ebiten.Image) {
	// Панель справа
	panelX := g.boardOffsetX + float64(BoardWidth*CellSize) + 20
	pX := float32(panelX)

	// Score panel
	vector.DrawFilledRect(screen, pX, 40, 140, 50, color.RGBA{R: 40, G: 40, B: 60, A: 200}, false)
	drawText(screen, "SCORE:", panelX+10, 50, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	// Level panel
	vector.DrawFilledRect(screen, pX, 120, 140, 50, color.RGBA{R: 40, G: 40, B: 60, A: 200}, false)
	drawText(screen, "LVL:", panelX+10, 130, color.RGBA{R: 255, G: 255, B: 0, A: 255})
	
	// Lines panel
	vector.DrawFilledRect(screen, pX, 180, 140, 50, color.RGBA{R: 40, G: 40, B: 60, A: 200}, false)
	drawText(screen, "LINES:", panelX+10, 190, color.RGBA{R: 0, G: 255, B: 255, A: 255})

	// Next Piece panel
	vector.DrawFilledRect(screen, pX, 260, 140, 120, color.RGBA{R: 40, G: 40, B: 60, A: 200}, false)
	drawText(screen, "NEXT:", panelX+10, 270, color.RGBA{R: 200, G: 200, B: 200, A: 255})

	// Preview next piece
	shape := PieceShapes[g.nextPieceType]
	for _, pos := range shape {
		imgIdx := g.nextPieceType + 1
		if imgIdx < len(g.blockImages) && g.blockImages[imgIdx] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(panelX+float64(pos[0]*CellSize), 300+float64(pos[1]*CellSize))
			screen.DrawImage(g.blockImages[imgIdx], op)
		}
	}

	// Controls panel
	vector.DrawFilledRect(screen, pX, 450, 140, 200, color.RGBA{R: 30, G: 30, B: 50, A: 150}, false)
}

// drawGameOver отрисовывает экран Game Over
func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Затемнение
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(ScreenWidth), float32(ScreenHeight),
		color.RGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	// Game Over text panel
	vector.DrawFilledRect(
		screen,
		float32(ScreenWidth)/2-120, float32(ScreenHeight)/2-80,
		240, 160,
		color.RGBA{R: 60, G: 20, B: 20, A: 255},
		false,
	)

	// Restart panel
	vector.DrawFilledRect(
		screen,
		float32(ScreenWidth)/2-150, float32(ScreenHeight)/2+60,
		300, 40,
		color.RGBA{R: 20, G: 60, B: 20, A: 255},
		false,
	)
}

// drawPause отрисовывает паузу
func (g *Game) drawPause(screen *ebiten.Image) {
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(ScreenWidth), float32(ScreenHeight),
		color.RGBA{R: 0, G: 0, B: 0, A: 128},
		false,
	)

	vector.DrawFilledRect(
		screen,
		float32(ScreenWidth)/2-80, float32(ScreenHeight)/2-30,
		160, 60,
		color.RGBA{R: 40, G: 40, B: 60, A: 255},
		false,
	)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// max возвращает максимальное значение
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Tetris - Go365 Challenge (Go104)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
