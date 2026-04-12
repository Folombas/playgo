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

// Game основной игровой объект
type Game struct {
	board       *logic.Board
	score       int
	combo       int
	maxCombo    int
	timeLeft    int
	gameOver    bool
	selected    image.Point
	hintTime    time.Time
	hintPair    [2]image.Point
	showHint    bool
	particles   []Particle
	gemImages   map[int]*ebiten.Image
	selectorImg *ebiten.Image
	rng         *rand.Rand
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

// NewGame создаёт новую игру
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		board:     logic.NewBoard(gridSize, rng),
		score:     0,
		combo:     0,
		maxCombo:  0,
		timeLeft:  120, // 2 минуты
		gameOver:  false,
		selected:  image.Point{-1, -1},
		hintTime:  time.Now(),
		showHint:  false,
		particles: make([]Particle, 0),
		gemImages: make(map[int]*ebiten.Image),
		rng:       rng,
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

	// Загрузить селектор
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
			// Рестарт
			g.board = logic.NewBoard(gridSize, g.rng)
			g.score = 0
			g.combo = 0
			g.timeLeft = 120
			g.gameOver = false
			g.particles = nil
		}
		return nil
	}

	// Обработка кликов
	g.handleInput()

	// Проверка подсказок
	if time.Since(g.hintTime) > 5*time.Second {
		g.showHint = true
		g.hintPair = g.board.FindHint()
	}

	// Обновление частиц
	g.updateParticles()

	// ESC для рестарта
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.board = logic.NewBoard(gridSize, g.rng)
		g.score = 0
		g.combo = 0
		g.timeLeft = 120
		g.particles = nil
		g.selected = image.Point{-1, -1}
	}

	return nil
}

func (g *Game) handleInput() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		col := (x - gridOffset) / tileSize
		row := (y - gridOffset) / tileSize

		if row < 0 || row >= gridSize || col < 0 || col >= gridSize {
			g.selected = image.Point{-1, -1}
			return
		}

		// Сбросить подсказку при действии
		g.showHint = false
		g.hintTime = time.Now()

		if g.selected.X == -1 {
			// Первый выбор
			g.selected = image.Point{row, col}
		} else {
			// Второй выбор - попытка обмена
			r1, c1 := g.selected.X, g.selected.Y
			r2, c2 := row, col

			// Проверить соседство
			if abs(r1-r2)+abs(c1-c2) == 1 {
				g.board.Swap(r1, c1, r2, c2)

				matches := g.board.FindMatches()
				if len(matches) > 0 {
					// Успешный обмен
					g.processMatches(matches)
					g.combo++
					if g.combo > g.maxCombo {
						g.maxCombo = g.combo
					}
				} else {
					// Неудачный обмен - вернуть обратно
					g.board.Swap(r1, c1, r2, c2)
					// Анимация тряски
					g.board.Shake(r1, c1)
					g.board.Shake(r2, c2)
				}
			} else {
				// Не соседние - выбрать новый
				g.selected = image.Point{row, col}
			}
		}
	}
}

func (g *Game) processMatches(matches []image.Point) {
	score := len(matches) * 10

	// Бонус за размер комбинации
	if len(matches) == 4 {
		score = 50 // Бонус за 4
	} else if len(matches) >= 5 {
		score = 100 // Бонус за 5+
	}

	// Комбо множитель
	if g.combo > 1 {
		score = score * (1 + g.combo/2)
	}

	g.score += score

	// Создать частицы
	for _, m := range matches {
		x := float64(gridOffset + m.Y*tileSize + tileSize/2)
		y := float64(gridOffset + m.X*tileSize + tileSize/2)
		for i := 0; i < 8; i++ {
			angle := float64(i) * math.Pi * 2 / 8
			g.particles = append(g.particles, Particle{
				X:       x,
				Y:       y,
				VX:      math.Cos(angle) * 4,
				VY:      math.Sin(angle) * 4 - 2,
				Life:    1.0,
				MaxLife: 1.0,
				Size:    5 + g.rng.Float64()*5,
				Color:   g.getColorForGem(g.board.Get(m.X, m.Y)),
			})
		}
	}

	// Удалить и заполнить
	g.board.RemoveMatches(matches)
	g.board.FillEmpty()
}

func (g *Game) getColorForGem(gem int) color.Color {
	colors := []color.Color{
		color.RGBA{255, 80, 80, 255},   // красный
		color.RGBA{80, 80, 255, 255},   // синий
		color.RGBA{80, 255, 80, 255},   // зелёный
		color.RGBA{255, 255, 80, 255},  // жёлтый
		color.RGBA{180, 80, 255, 255},  // фиолетовый
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
		p.VY += 0.3 // гравитация
		p.Life -= 1.0 / 60.0

		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
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

	// Частицы
	g.drawParticles(screen)

	// Game Over экран
	if g.gameOver {
		g.drawGameOver(screen)
	}
}

func (g *Game) drawHeader(screen *ebiten.Image) {
	// Фон заголовка
	drawRoundedRect(screen, 20, 10, 760, 100, 15, color.RGBA{40, 30, 70, 200})

	// Счёт
	g.drawCenteredText(screen, "SCORE", screenWidth/2, scoreY-15, 16, color.RGBA{150, 150, 200, 255})
	g.drawCenteredText(screen, fmt.Sprintf("%d", g.score), screenWidth/2, scoreY+15, 32, color.RGBA{255, 215, 0, 255})

	// Таймер
	mins := g.timeLeft / 60
	secs := g.timeLeft % 60
	timeColor := color.RGBA{100, 255, 100, 255}
	if g.timeLeft < 30 {
		timeColor = color.RGBA{255, 100, 100, 255}
	}
	g.drawCenteredText(screen, "TIME", 120, timerY, 14, color.RGBA{150, 150, 200, 255})
	g.drawCenteredText(screen, fmt.Sprintf("%d:%02d", mins, secs), 120, timerY+20, 24, timeColor)

	// Комбо
	if g.combo > 1 {
		g.drawCenteredText(screen, fmt.Sprintf("COMBO x%d!", g.combo), 680, timerY+10, 20, color.RGBA{255, 165, 0, 255})
	}

	// Подсказка
	g.drawCenteredText(screen, "R - Restart", 680, timerY+35, 12, color.RGBA{150, 150, 200, 255})
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
			drawRoundedRect(screen, float64(x+2), float64(y+2), tileSize-4, tileSize-4, 8, cellColor)

			// Гем
			gem := g.board.Get(r, c)
			if gem >= 0 {
				g.drawGem(screen, x, y, gem)
			}

			// Выделение
			if g.selected.X == r && g.selected.Y == c {
				g.drawSelector(screen, x, y)
			}

			// Подсветка подсказки
			if g.showHint {
				if (r == g.hintPair[0].X && c == g.hintPair[0].Y) ||
					(r == g.hintPair[1].X && c == g.hintPair[1].Y) {
					pulse := math.Sin(float64(time.Now().UnixMilli())*0.008)*0.3 + 0.7
					alpha := uint8(150 * pulse)
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

func (g *Game) drawGem(screen *ebiten.Image, x, y int, gem int) {
	if img, ok := g.gemImages[gem]; ok {
		op := &ebiten.DrawImageOptions{}
		// Масштабировать до tileSize
		scale := float64(tileSize-10) / float64(img.Bounds().Dx())
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(x+5), float64(y+5))
		screen.DrawImage(img, op)
	} else {
		// Если картинка не загрузилась - рисовать цветной квадрат
		colors := []color.Color{
			color.RGBA{255, 80, 80, 255},
			color.RGBA{80, 80, 255, 255},
			color.RGBA{80, 255, 80, 255},
			color.RGBA{255, 255, 80, 255},
			color.RGBA{180, 80, 255, 255},
		}
		c := color.RGBA{100, 100, 100, 255}
		if gem < len(colors) {
			c = colors[gem].(color.RGBA)
		}
		drawRoundedRect(screen, float64(x+5), float64(y+5), tileSize-10, tileSize-10, 10, c)
	}
}

func (g *Game) drawSelector(screen *ebiten.Image, x, y int) {
	if g.selectorImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(g.selectorImg, op)
	} else {
		// Нарисовать обводку
		vector.StrokeRect(screen, float32(x+2), float32(y+2), float32(tileSize-4), float32(tileSize-4), 3,
			color.RGBA{255, 215, 0, 255}, false)
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
	// Затемнение
	screen.Fill(color.RGBA{0, 0, 0, 180})

	// Панель
	drawRoundedRect(screen, 150, 250, 500, 300, 20, color.RGBA{40, 30, 70, 255})

	g.drawCenteredText(screen, "TIME'S UP!", screenWidth/2, 300, 48, color.RGBA{255, 100, 100, 255})
	g.drawCenteredText(screen, fmt.Sprintf("Final Score: %d", g.score), screenWidth/2, 380, 32, color.RGBA{255, 215, 0, 255})
	g.drawCenteredText(screen, fmt.Sprintf("Max Combo: x%d", g.maxCombo), screenWidth/2, 430, 24, color.RGBA{255, 165, 0, 255})
	g.drawCenteredText(screen, "Press R or ENTER to restart", screenWidth/2, 500, 20, color.RGBA{200, 200, 255, 255})
}

func (g *Game) drawCenteredText(screen *ebiten.Image, str string, x, y int, size int, c color.Color) {
	// Центрирование - используем простой подход
	ebitenutil.DebugPrintAt(screen, str, x-len(str)*size/4, y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawRoundedRect рисует скруглённый прямоугольник
func drawRoundedRect(screen *ebiten.Image, x, y, w, h float64, cr float64, c color.Color) {
	vector.DrawFilledRect(screen, float32(x+cr), float32(y), float32(w-cr*2), float32(h), c, false)
	vector.DrawFilledRect(screen, float32(x), float32(y+cr), float32(w), float32(h-cr*2), c, false)
	vector.DrawFilledCircle(screen, float32(x+cr), float32(y+cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+w-cr), float32(y+cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+cr), float32(y+h-cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+w-cr), float32(y+h-cr), float32(cr), c, false)
}
