package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"

	"match3_v2/internal/logic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 800
	gridSize     = 8
	tileSize     = 70
	gridOffset   = 130
	scoreY       = 60
	timerY       = 30
)

// DragState состояние перетаскивания
type DragState int

const (
	DragNone DragState = iota
	DragHovering      // Наведение на гем
	DragPicking       // Поднимаем гем
	Dragging          // Перетаскиваем
	DragSnapping      // Прилипаем к новой позиции
	DragReturning     // Возвращаем на место
	DragShaking       // Трясём при ошибке
)

// Game основной игровой объект
type Game struct {
	board       *logic.Board
	score       int
	combo       int
	maxCombo    int
	timeLeft    int
	gameOver    bool
	hintTime    time.Time
	hintPair    [2]image.Point
	showHint    bool
	particles   []Particle
	trails      []Trail      // Следы при перетаскивании
	gemImages   map[int]*ebiten.Image
	selectorImg *ebiten.Image
	rng         *rand.Rand

	// Drag & Drop
	dragState      DragState
	dragGem        image.Point  // Какой гем тащим
	dragStartPos   image.Point  // Начальная позиция
	dragCurrentX   float64      // Текущая X мыши
	dragCurrentY   float64      // Текущая Y мыши
	dragTargetPos  image.Point  // Целевая позиция
	dragAnimTime   float64      // Время анимации
	dragScale      float64      // Масштаб перетаскиваемого гема
	dragRotation   float64      // Вращение при перетаскивании
	dragGlowPulse  float64      // Пульсация свечения
	hoverGem       image.Point  // Гем под курсором
	hoverTime      float64      // Время наведения
}

// Particle частица для эффектов
type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    float64
	Color   color.Color
	Size    float64
	MaxLife float64
}

// Trail след при перетаскивании
type Trail struct {
	X, Y  float64
	Life  float64
	Color color.Color
	Size  float64
}

// NewGame создаёт новую игру
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		board:       logic.NewBoard(gridSize, rng),
		score:       0,
		combo:       0,
		maxCombo:    0,
		timeLeft:    120,
		gameOver:    false,
		hintTime:    time.Now(),
		showHint:    false,
		particles:   make([]Particle, 0),
		trails:      make([]Trail, 0),
		gemImages:   make(map[int]*ebiten.Image),
		rng:         rng,
		dragState:   DragNone,
		dragGem:     image.Point{-1, -1},
		hoverGem:    image.Point{-1, -1},
		dragScale:   1.0,
	}

	g.loadImages()

	// Таймер
	go func() {
		for !g.gameOver {
			time.Sleep(1 * time.Second)
			g.timeLeft--
			if g.timeLeft <= 0 {
				g.gameOver = true
			}
		}
	}()

	return g
}

// loadImages загружает спрайты
func (g *Game) loadImages() {
	gems := []string{
		"assets/gem_red.png",
		"assets/gem_blue.png",
		"assets/gem_green.png",
		"assets/gem_yellow.png",
		"assets/gem_purple.png",
	}

	for i, path := range gems {
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("Warning: could not load %s: %v\n", path, err)
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			continue
		}
		eimg := ebiten.NewImageFromImage(img)
		g.gemImages[i] = eimg
	}

	f, err := os.Open("assets/selector.png")
	if err == nil {
		img, _, err := image.Decode(f)
		f.Close()
		if err == nil {
			g.selectorImg = ebiten.NewImageFromImage(img)
		}
	}
}

func (g *Game) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.resetGame()
		}
		return nil
	}

	// Обработка drag & drop
	g.handleDragDrop()

	// Подсказки
	if time.Since(g.hintTime) > 5*time.Second && g.dragState == DragNone {
		g.showHint = true
		g.hintPair = g.board.FindHint()
	}

	// Обновление частиц и следов
	g.updateParticles()
	g.updateTrails()

	// Рестарт
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.resetGame()
	}

	return nil
}

func (g *Game) resetGame() {
	g.board = logic.NewBoard(gridSize, g.rng)
	g.score = 0
	g.combo = 0
	g.timeLeft = 120
	g.gameOver = false
	g.particles = nil
	g.trails = nil
	g.dragState = DragNone
	g.dragGem = image.Point{-1, -1}
	g.hoverGem = image.Point{-1, -1}
}

func (g *Game) handleDragDrop() {
	mx, my := ebiten.CursorPosition()

	// Определяем ячейку под курсором
	col := (mx - gridOffset) / tileSize
	row := (my - gridOffset) / tileSize
	hoverValid := row >= 0 && row < gridSize && col >= 0 && col < gridSize
	hoverPos := image.Point{row, col}

	switch g.dragState {
	case DragNone:
		// Начало перетаскивания
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && hoverValid {
			g.dragState = DragPicking
			g.dragGem = hoverPos
			g.dragStartPos = hoverPos
			g.dragCurrentX = float64(mx)
			g.dragCurrentY = float64(my)
			g.dragAnimTime = 0
			g.dragScale = 1.0
			g.dragRotation = 0
			g.showHint = false
			g.hintTime = time.Now()

			// Эффект поднятия - всплеск частиц
			g.spawnLiftEffect(hoverPos)
		} else if hoverValid {
			// Наведение
			g.hoverGem = hoverPos
			g.hoverTime += 1.0 / 60.0
		} else {
			g.hoverGem = image.Point{-1, -1}
			g.hoverTime = 0
		}

	case DragPicking:
		// Поднимаем гем (короткая анимация)
		g.dragAnimTime += 1.0 / 60.0
		g.dragScale = 1.0 + math.Sin(g.dragAnimTime*math.Pi)*0.3
		g.dragRotation = math.Sin(g.dragAnimTime*math.Pi*2) * 0.1

		if g.dragAnimTime > 0.15 {
			g.dragState = Dragging
			g.dragAnimTime = 0
			g.dragScale = 1.2
		}

	case Dragging:
		// Перетаскиваем
		g.dragCurrentX = float64(mx)
		g.dragCurrentY = float64(my)
		g.dragRotation = math.Sin(float64(time.Now().UnixMilli())*0.005) * 0.15
		g.dragGlowPulse = math.Sin(float64(time.Now().UnixMilli())*0.008)*0.3 + 0.7

		// Добавляем следы
		if len(g.trails) == 0 || time.Now().UnixMilli()%3 == 0 {
			gemType := g.board.Get(g.dragGem.X, g.dragGem.Y)
			g.trails = append(g.trails, Trail{
				X:     g.dragCurrentX,
				Y:     g.dragCurrentY,
				Life:  0.5,
				Color: g.getColorForGem(gemType),
				Size:  8,
			})
		}

		// Отпускание кнопки
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if hoverValid && g.isValidSwap(g.dragGem, hoverPos) {
				// Валидный обмен
				g.dragTargetPos = hoverPos
				g.dragState = DragSnapping
				g.dragAnimTime = 0

				// Выполнить обмен
				g.board.Swap(g.dragGem.X, g.dragGem.Y, hoverPos.X, hoverPos.Y)

				// Проверить матчи
				matches := g.board.FindMatches()
				if len(matches) > 0 {
					g.processMatches(matches)
					g.combo++
					if g.combo > g.maxCombo {
						g.maxCombo = g.combo
					}
					g.spawnSuccessEffect(hoverPos)
				} else {
					// Вернуть обратно
					g.board.Swap(g.dragGem.X, g.dragGem.Y, hoverPos.X, hoverPos.Y)
					g.dragState = DragShaking
					g.dragAnimTime = 0
					g.spawnFailEffect()
				}
			} else {
				// Вернуть на место
				g.dragTargetPos = g.dragStartPos
				g.dragState = DragReturning
				g.dragAnimTime = 0
			}
		}

	case DragSnapping:
		// Прилипаем к новой позиции
		g.dragAnimTime += 1.0 / 60.0
		t := math.Min(g.dragAnimTime/0.2, 1.0)
		g.dragScale = 1.2 - 0.2*t // Уменьшаем до нормального

		if t >= 1.0 {
			g.dragState = DragNone
			g.dragGem = image.Point{-1, -1}
		}

	case DragReturning:
		// Возвращаем на место
		g.dragAnimTime += 1.0 / 60.0
		t := math.Min(g.dragAnimTime/0.3, 1.0)
		g.dragScale = 1.2 - 0.2*t

		if t >= 1.0 {
			g.dragState = DragNone
			g.dragGem = image.Point{-1, -1}
		}

	case DragShaking:
		// Трясём при ошибке
		g.dragAnimTime += 1.0 / 60.0
		t := math.Min(g.dragAnimTime/0.4, 1.0)
		g.dragRotation = math.Sin(t*math.Pi*8) * 0.2 * (1 - t)
		g.dragScale = 1.0

		if t >= 1.0 {
			g.dragState = DragNone
			g.dragGem = image.Point{-1, -1}
			g.dragRotation = 0
		}
	}
}

func (g *Game) isValidSwap(from, to image.Point) bool {
	// Проверить что это соседние ячейки
	dr := abs(from.X - to.X)
	dc := abs(from.Y - to.Y)
	return dr+dc == 1
}

func (g *Game) processMatches(matches []image.Point) {
	score := len(matches) * 10

	if len(matches) == 4 {
		score = 50
	} else if len(matches) >= 5 {
		score = 100
	}

	if g.combo > 1 {
		score = score * (1 + g.combo/2)
	}

	g.score += score

	// Частицы
	for _, m := range matches {
		x := float64(gridOffset + m.Y*tileSize + tileSize/2)
		y := float64(gridOffset + m.X*tileSize + tileSize/2)
		for i := 0; i < 12; i++ {
			angle := float64(i) * math.Pi * 2 / 12
			speed := 3 + g.rng.Float64()*4
			g.particles = append(g.particles, Particle{
				X:       x,
				Y:       y,
				VX:      math.Cos(angle) * speed,
				VY:      math.Sin(angle) * speed - 3,
				Life:    1.0,
				MaxLife: 1.0,
				Size:    4 + g.rng.Float64()*6,
				Color:   g.getColorForGem(g.board.Get(m.X, m.Y)),
			})
		}
	}

	g.board.RemoveMatches(matches)
	g.board.FillEmpty()
}

func (g *Game) spawnLiftEffect(pos image.Point) {
	x := float64(gridOffset + pos.Y*tileSize + tileSize/2)
	y := float64(gridOffset + pos.X*tileSize + tileSize/2)
	gemType := g.board.Get(pos.X, pos.Y)
	c := g.getColorForGem(gemType)

	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi * 2 / 8
		g.particles = append(g.particles, Particle{
			X:       x,
			Y:       y,
			VX:      math.Cos(angle) * 2,
			VY:      math.Sin(angle) * 2,
			Life:    0.5,
			MaxLife: 0.5,
			Size:    3 + g.rng.Float64()*3,
			Color:   c,
		})
	}
}

func (g *Game) spawnSuccessEffect(pos image.Point) {
	x := float64(gridOffset + pos.Y*tileSize + tileSize/2)
	y := float64(gridOffset + pos.X*tileSize + tileSize/2)

	// Золотой всплеск
	for i := 0; i < 20; i++ {
		angle := float64(i) * math.Pi * 2 / 20
		speed := 4 + g.rng.Float64()*5
		g.particles = append(g.particles, Particle{
			X:       x,
			Y:       y,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed - 2,
			Life:    1.2,
			MaxLife: 1.2,
			Size:    5 + g.rng.Float64()*8,
			Color:   color.RGBA{255, 215, 0, 255},
		})
	}
}

func (g *Game) spawnFailEffect() {
	// Красные частицы
	for i := 0; i < 10; i++ {
		angle := g.rng.Float64() * math.Pi * 2
		speed := 2 + g.rng.Float64()*3
		g.particles = append(g.particles, Particle{
			X:       g.dragCurrentX,
			Y:       g.dragCurrentY,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed,
			Life:    0.6,
			MaxLife: 0.6,
			Size:    3 + g.rng.Float64()*4,
			Color:   color.RGBA{255, 50, 50, 255},
		})
	}
}

func (g *Game) getColorForGem(gem int) color.Color {
	colors := []color.Color{
		color.RGBA{255, 80, 80, 255},
		color.RGBA{80, 80, 255, 255},
		color.RGBA{80, 255, 80, 255},
		color.RGBA{255, 255, 80, 255},
		color.RGBA{180, 80, 255, 255},
	}
	if gem >= 0 && gem < len(colors) {
		return colors[gem]
	}
	return color.White
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.25
		p.Life -= 1.0 / 60.0

		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) updateTrails() {
	for i := len(g.trails) - 1; i >= 0; i-- {
		g.trails[i].Life -= 1.0 / 60.0
		if g.trails[i].Life <= 0 {
			g.trails = append(g.trails[:i], g.trails[i+1:]...)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Фон
	screen.Fill(color.RGBA{20, 10, 40, 255})

	// Заголовок
	g.drawHeader(screen)

	// Игровое поле
	g.drawBoard(screen)

	// Следы перетаскивания
	g.drawTrails(screen)

	// Частицы
	g.drawParticles(screen)

	// Перетаскиваемый гем (рисуем поверх всего)
	if g.dragState == DragPicking || g.dragState == Dragging ||
		g.dragState == DragSnapping || g.dragState == DragReturning ||
		g.dragState == DragShaking {
		g.drawDraggedGem(screen)
	}

	// Game Over
	if g.gameOver {
		g.drawGameOver(screen)
	}
}

func (g *Game) drawHeader(screen *ebiten.Image) {
	drawRoundedRect(screen, 20, 10, 760, 100, 15, color.RGBA{40, 30, 70, 200})

	g.drawCenteredText(screen, "SCORE", screenWidth/2, scoreY-15, 16, color.RGBA{150, 150, 200, 255})
	g.drawCenteredText(screen, fmt.Sprintf("%d", g.score), screenWidth/2, scoreY+15, 32, color.RGBA{255, 215, 0, 255})

	mins := g.timeLeft / 60
	secs := g.timeLeft % 60
	timeColor := color.RGBA{100, 255, 100, 255}
	if g.timeLeft < 30 {
		timeColor = color.RGBA{255, 100, 100, 255}
	}
	g.drawCenteredText(screen, "TIME", 120, timerY, 14, color.RGBA{150, 150, 200, 255})
	g.drawCenteredText(screen, fmt.Sprintf("%d:%02d", mins, secs), 120, timerY+20, 24, timeColor)

	if g.combo > 1 {
		g.drawCenteredText(screen, fmt.Sprintf("COMBO x%d!", g.combo), 680, timerY+10, 20, color.RGBA{255, 165, 0, 255})
	}

	g.drawCenteredText(screen, "R - Restart | Drag gems to swap", 680, timerY+35, 12, color.RGBA{150, 150, 200, 255})
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	// Фон доски
	drawRoundedRect(screen, float64(gridOffset-10), float64(gridOffset-10),
		float64(gridSize*tileSize+20), float64(gridSize*tileSize+20), 10,
		color.RGBA{30, 20, 60, 255})

	// Ячейки
	for r := 0; r < gridSize; r++ {
		for c := 0; c < gridSize; c++ {
			x := gridOffset + c*tileSize
			y := gridOffset + r*tileSize

			// Фон ячейки
			cellColor := color.RGBA{50, 40, 90, 255}
			if (r+c)%2 == 0 {
				cellColor = color.RGBA{60, 50, 100, 255}
			}

			// Подсветка при наведении
			if g.hoverGem.X == r && g.hoverGem.Y == c && g.dragState == DragNone {
				pulse := math.Sin(float64(time.Now().UnixMilli())*0.006)*0.15 + 0.15
				r2 := uint8(float64(cellColor.R) + 50*pulse)
				g2 := uint8(float64(cellColor.G) + 50*pulse)
				b2 := uint8(float64(cellColor.B) + 80*pulse)
				cellColor = color.RGBA{r2, g2, b2, 255}
			}

			// Подсветка целевой ячейки при перетаскивании
			if g.dragState == Dragging && g.isValidSwap(g.dragGem, image.Point{r, c}) {
				pulse := math.Sin(float64(time.Now().UnixMilli())*0.01)*0.3 + 0.5
				cellColor = color.RGBA{
					uint8(100 + 100*pulse),
					uint8(50 + 100*pulse),
					uint8(150 + 100*pulse),
					255,
				}
			}

			drawRoundedRect(screen, float64(x+2), float64(y+2), tileSize-4, tileSize-4, 8, cellColor)

			// Гем (не рисуем перетаскиваемый)
			gem := g.board.Get(r, c)
			if gem >= 0 && !(g.dragGem.X == r && g.dragGem.Y == c &&
				(g.dragState == DragPicking || g.dragState == Dragging ||
					g.dragState == DragSnapping || g.dragState == DragShaking)) {
				g.drawGem(screen, x, y, gem, 1.0, 0, 1.0)
			}

			// Подсказка
			if g.showHint {
				if (r == g.hintPair[0].X && c == g.hintPair[0].Y) ||
					(r == g.hintPair[1].X && c == g.hintPair[1].Y) {
					pulse := math.Sin(float64(time.Now().UnixMilli())*0.008)*0.3 + 0.7
					alpha := uint8(180 * pulse)
					op := &ebiten.DrawImageOptions{}
					op.ColorScale.ScaleAlpha(float32(alpha) / 255.0)
					op.GeoM.Translate(float64(x), float64(y))
					if g.selectorImg != nil {
						screen.DrawImage(g.selectorImg, op)
					}
				}
			}
		}
	}
}

func (g *Game) drawGem(screen *ebiten.Image, x, y int, gem int, scale, rotation, alpha float64) {
	if img, ok := g.gemImages[gem]; ok {
		op := &ebiten.DrawImageOptions{}
		baseScale := float64(tileSize-10) / float64(img.Bounds().Dx()) * scale
		op.GeoM.Scale(baseScale, baseScale)

		// Вращение вокруг центра
		if rotation != 0 {
			cx := float64(img.Bounds().Dx()) / 2
			cy := float64(img.Bounds().Dy()) / 2
			op.GeoM.Translate(-cx, -cy)
			op.GeoM.Rotate(rotation)
			op.GeoM.Translate(cx, cy)
		}

		op.ColorScale.ScaleAlpha(float32(alpha))
		op.GeoM.Translate(float64(x+5), float64(y+5))
		screen.DrawImage(img, op)
	} else {
		colors := []color.RGBA{
			{255, 80, 80, uint8(255 * alpha)},
			{80, 80, 255, uint8(255 * alpha)},
			{80, 255, 80, uint8(255 * alpha)},
			{255, 255, 80, uint8(255 * alpha)},
			{180, 80, 255, uint8(255 * alpha)},
		}
		c := colors[0]
		if gem < len(colors) {
			c = colors[gem]
		}
		drawRoundedRect(screen, float64(x+5), float64(y+5), tileSize-10, tileSize-10, 10, c)
	}
}

func (g *Game) drawDraggedGem(screen *ebiten.Image) {
	gemType := g.board.Get(g.dragGem.X, g.dragGem.Y)
	if gemType < 0 {
		return
	}

	// Позиция: прилипание или мышь
	var drawX, drawY float64
	if g.dragState == DragSnapping || g.dragState == DragReturning {
		drawX = float64(gridOffset + g.dragTargetPos.Y*tileSize + tileSize/2)
		drawY = float64(gridOffset + g.dragTargetPos.X*tileSize + tileSize/2)
	} else if g.dragState == DragShaking {
		shake := math.Sin(g.dragAnimTime*math.Pi*8) * 8 * (1 - g.dragAnimTime/0.4)
		drawX = g.dragCurrentX + shake
		drawY = g.dragCurrentY
	} else {
		drawX = g.dragCurrentX
		drawY = g.dragCurrentY
	}

	// Свечение
	if g.dragState == Dragging {
		glowSize := 40 + g.dragGlowPulse*10
		vector.DrawFilledCircle(screen, float32(drawX), float32(drawY), float32(glowSize),
			color.RGBA{255, 255, 255, 30}, false)
	}

	// Сам гем
	g.drawGem(screen, int(drawX-tileSize/2+5), int(drawY-tileSize/2+5), gemType,
		g.dragScale, g.dragRotation, 0.9)
}

func (g *Game) drawTrails(screen *ebiten.Image) {
	for _, t := range g.trails {
		alpha := uint8(255 * (t.Life / 0.5))
		if c, ok := t.Color.(color.RGBA); ok {
			vector.DrawFilledCircle(screen, float32(t.X), float32(t.Y), float32(t.Size*t.Life/0.5),
				color.RGBA{c.R, c.G, c.B, alpha}, false)
		}
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		alpha := uint8(255 * (p.Life / p.MaxLife))
		if c, ok := p.Color.(color.RGBA); ok {
			vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), float32(p.Size),
				color.RGBA{c.R, c.G, c.B, alpha}, false)
		}
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 180})
	drawRoundedRect(screen, 150, 250, 500, 300, 20, color.RGBA{40, 30, 70, 255})

	g.drawCenteredText(screen, "TIME'S UP!", screenWidth/2, 300, 48, color.RGBA{255, 100, 100, 255})
	g.drawCenteredText(screen, fmt.Sprintf("Final Score: %d", g.score), screenWidth/2, 380, 32, color.RGBA{255, 215, 0, 255})
	g.drawCenteredText(screen, fmt.Sprintf("Max Combo: x%d", g.maxCombo), screenWidth/2, 430, 24, color.RGBA{255, 165, 0, 255})
	g.drawCenteredText(screen, "Press R or ENTER to restart", screenWidth/2, 500, 20, color.RGBA{200, 200, 255, 255})
}

func (g *Game) drawCenteredText(screen *ebiten.Image, str string, x, y int, size int, c color.Color) {
	ebitenutil.DebugPrintAt(screen, str, x-len(str)*size/4, y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawRoundedRect(screen *ebiten.Image, x, y, w, h float64, cr float64, c color.Color) {
	vector.DrawFilledRect(screen, float32(x+cr), float32(y), float32(w-cr*2), float32(h), c, false)
	vector.DrawFilledRect(screen, float32(x), float32(y+cr), float32(w), float32(h-cr*2), c, false)
	vector.DrawFilledCircle(screen, float32(x+cr), float32(y+cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+w-cr), float32(y+cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+cr), float32(y+h-cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+w-cr), float32(y+h-cr), float32(cr), c, false)
}
