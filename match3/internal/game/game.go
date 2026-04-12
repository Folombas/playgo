package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"match3/internal/logic"
	"match3/internal/ui"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 450
	screenHeight = 800
)

type GameState int

const (
	StateMap GameState = iota
	StatePlaying
	StatePause
	StateWin
	StateLose
)

type Game struct {
	state        GameState
	board        *logic.Board
	score        int
	moves        int
	combo        int
	maxCombo     int
	level        int
	targetScore  int
	energy       int
	coins        int
	stars        map[int]int
	particles    []Particle
	screenShake  float64
	rng          *rand.Rand
	lastAction   time.Time

	// Mouse input
	selectedTile *logic.Tile
	hintTimer    time.Time
	hintActive   bool
	hintTile1    *logic.Tile
	hintTile2    *logic.Tile
}

type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    float64
	MaxLife float64
	Size    float64
	Color   color.RGBA
}

func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	g := &Game{
		state:       StateMap,
		level:       1,
		energy:      5,
		coins:       1000,
		stars:       make(map[int]int),
		rng:         rng,
		lastAction:  time.Now(),
	}
	
	g.loadProgress()
	return g
}

func (g *Game) Update() error {
	switch g.state {
	case StateMap:
		g.updateMap()
	case StatePlaying:
		g.updatePlaying()
	case StatePause:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StatePlaying
		}
	case StateWin:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.nextLevel()
		}
	case StateLose:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.retryLevel()
		}
	}
	
	g.updateParticles()
	return nil
}

func (g *Game) updateMap() {
	if g.inputClicked() {
		if g.energy > 0 {
			g.startLevel(g.level)
		}
	}
	
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.level = min(g.level+1, 50)
		time.Sleep(150 * time.Millisecond)
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.level = max(1, g.level-1)
		time.Sleep(150 * time.Millisecond)
	}
}

func (g *Game) updatePlaying() {
	if g.board != nil {
		// Обработка кликов мыши
		g.handleMouseInput()

		// Проверка на бездействие для подсказок
		if time.Since(g.lastAction) > 5*time.Second && !g.hintActive {
			g.showHint()
		}

		g.board.Update()

		matches := g.board.ProcessMatches()
		if len(matches) > 0 {
			g.combo++
			multiplier := 1.0 + float64(g.combo-1)*0.5
			score := int(float64(len(matches)*10) * multiplier)
			g.score += score

			g.spawnParticles(matches)

			if g.combo > g.maxCombo {
				g.maxCombo = g.combo
			}

			if g.combo >= 3 {
				g.screenShake = 0.2
			}

			g.lastAction = time.Now()
			g.hintActive = false
		} else {
			g.combo = 0
		}

		if g.score >= g.targetScore {
			g.state = StateWin
			g.calculateStars()
		}

		if g.moves <= 0 && g.score < g.targetScore {
			g.state = StateLose
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StatePause
		}

		// Клавиша H для ручной подсказки
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			g.showHint()
		}
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.3
		p.Life -= 1.0 / 60.0
		
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
	
	if g.screenShake > 0 {
		g.screenShake -= 1.0 / 60.0
	}
}

func (g *Game) spawnParticles(tiles []*logic.Tile) {
	for _, tile := range tiles {
		x := float64(45 + tile.Col*50 + 25)
		y := float64(150 + tile.Row*50 + 25)
		
		for i := 0; i < 5; i++ {
			angle := float64(i) * math.Pi * 2 / 5
			g.particles = append(g.particles, Particle{
				X:       x,
				Y:       y,
				VX:      math.Cos(angle) * 3,
				VY:      math.Sin(angle) * 3 - 2,
				Life:    1.0,
				MaxLife: 1.0,
				Size:    4 + g.rng.Float64()*4,
				Color:   logic.GemColors[tile.Gem],
			})
		}
	}
}

func (g *Game) calculateStars() {
	ratio := float64(g.score) / float64(g.targetScore)
	stars := 0
	if ratio >= 0.5 {
		stars = 1
	}
	if ratio >= 0.8 {
		stars = 2
	}
	if ratio >= 1.0 {
		stars = 3
	}
	
	if stars > g.stars[g.level] {
		g.stars[g.level] = stars
	}
	
	g.coins += g.score / 10
	g.energy = min(g.energy+1, 5)
}

func (g *Game) startLevel(level int) {
	g.level = level
	g.state = StatePlaying
	g.score = 0
	g.moves = 20 + level*2
	g.combo = 0
	g.maxCombo = 0
	g.targetScore = 500 + level*200
	g.energy--
	g.particles = nil
	g.screenShake = 0
	
	g.board = logic.NewBoard(8, 8, level)
	g.lastAction = time.Now()
}

func (g *Game) nextLevel() {
	g.level = min(g.level+1, 50)
	g.startLevel(g.level)
}

func (g *Game) retryLevel() {
	if g.energy > 0 {
		g.startLevel(g.level)
	} else {
		g.state = StateMap
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ui.ColorDeepPurple)
	
	switch g.state {
	case StateMap:
		g.drawMap(screen)
	case StatePlaying:
		g.drawGame(screen)
	case StatePause:
		g.drawGame(screen)
		g.drawOverlay(screen, "ПАУЗА", "ESC - продолжить")
	case StateWin:
		g.drawGame(screen)
		g.drawWinScreen(screen)
	case StateLose:
		g.drawGame(screen)
		g.drawLoseScreen(screen)
	}
}

func (g *Game) drawMap(screen *ebiten.Image) {
	ui.DrawCenteredText(screen, screenWidth/2, 100, "FRUIT CRUSH", 28, ui.ColorHotPink)
	ui.DrawCenteredText(screen, screenWidth/2, 130, "SAGA", 20, ui.ColorCyan)
	
	// Энергия и монеты
	ui.DrawText(screen, 20, 20, fmt.Sprintf("Energy: %d/5", g.energy), 16, ui.ColorGold)
	ui.DrawText(screen, 20, 45, fmt.Sprintf("Coins: %d", g.coins), 16, ui.ColorGold)
	
	// Карта уровней
	startX := 60
	startY := 200
	spacing := 65
	
	for level := 1; level <= 50; level++ {
		col := (level - 1) % 5
		row := (level - 1) / 5
		
		x := startX + col*spacing
		y := startY + row*spacing
		
		// Фон
		var bgColor color.Color
		if level == g.level {
			bgColor = ui.ColorCyan
		} else if g.stars[level] > 0 {
			bgColor = ui.ColorGreen
		} else {
			bgColor = ui.ColorDarkGray
		}
		
		circle := g.createCircle(50, bgColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x-25), float64(y-25))
		screen.DrawImage(circle, op)
		
		ui.DrawCenteredText(screen, x, y+5, fmt.Sprintf("%d", level), 16, ui.ColorWhite)
		
		if g.stars[level] > 0 {
			stars := ""
			for i := 0; i < g.stars[level]; i++ {
				stars += "*"
			}
			ui.DrawCenteredText(screen, x, y+30, stars, 10, ui.ColorGold)
		}
	}
	
	ui.DrawCenteredText(screen, screenWidth/2, 700, fmt.Sprintf("Level %d", g.level), 18, ui.ColorWhite)
	ui.DrawCenteredText(screen, screenWidth/2, 730, "Click to play", 14, ui.ColorCyan)
	ui.DrawCenteredText(screen, screenWidth/2, 760, "< > to change level", 12, ui.ColorGray)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// HUD
	ui.DrawCenteredText(screen, screenWidth/2, 30, fmt.Sprintf("%d", g.score), 36, ui.ColorGold)
	ui.DrawCenteredText(screen, screenWidth/2, 65, fmt.Sprintf("/ %d", g.targetScore), 14, ui.ColorGray)
	
	// Прогресс-бар
	progress := float64(g.score) / float64(g.targetScore)
	if progress > 1.0 {
		progress = 1.0
	}
	
	bar := g.createProgressBar(300, 15, progress)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(75, 80)
	screen.DrawImage(bar, op)
	
	ui.DrawText(screen, 20, 115, fmt.Sprintf("Moves: %d", g.moves), 16, ui.ColorWhite)
	
	if g.combo > 1 {
		ui.DrawCenteredText(screen, screenWidth/2, 115, fmt.Sprintf("COMBO x%d!", g.combo), 18, ui.ColorOrange)
	}
	
	ui.DrawText(screen, 20, 15, fmt.Sprintf("Level %d", g.level), 14, ui.ColorCyan)
	
	// Доска
	if g.board != nil {
		g.board.Draw(screen)
	}

	// Подсказки
	if g.hintActive && g.hintTile1 != nil && g.hintTile2 != nil {
		g.drawHint(screen, g.hintTile1)
		g.drawHint(screen, g.hintTile2)
	}

	// Частицы
	g.drawParticles(screen)
}

// drawHint рисует подсветку для подсказки
func (g *Game) drawHint(screen *ebiten.Image, tile *logic.Tile) {
	if tile == nil {
		return
	}

	x := float64(tile.Col * 50 + 45)
	y := float64(tile.Row * 50 + 150)

	// Пульсирующая подсветка
	pulse := math.Sin(float64(time.Now().UnixMilli())*0.008) * 0.3 + 0.7
	alpha := uint8(200 * pulse)

	hint := ebiten.NewImage(54, 54)
	hint.Fill(color.RGBA{0, 255, 255, alpha})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x-2, y-2)
	screen.DrawImage(hint, op)
}

func (g *Game) drawOverlay(screen *ebiten.Image, title, subtitle string) {
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)
	
	ui.DrawCenteredText(screen, screenWidth/2, 350, title, 28, ui.ColorWhite)
	ui.DrawCenteredText(screen, screenWidth/2, 400, subtitle, 16, ui.ColorCyan)
}

func (g *Game) drawWinScreen(screen *ebiten.Image) {
	g.drawOverlay(screen, "LEVEL COMPLETE!", "ENTER - next level")
	
	// Звёзды
	starCount := g.stars[g.level]
	starStr := ""
	for i := 0; i < 3; i++ {
		if i < starCount {
			starStr += "* "
		} else {
			starStr += "o "
		}
	}
	ui.DrawCenteredText(screen, screenWidth/2, 450, starStr, 36, ui.ColorGold)
	
	ui.DrawCenteredText(screen, screenWidth/2, 520, fmt.Sprintf("Score: %d", g.score), 18, ui.ColorWhite)
	ui.DrawCenteredText(screen, screenWidth/2, 550, fmt.Sprintf("Max Combo: x%d", g.maxCombo), 16, ui.ColorOrange)
}

func (g *Game) drawLoseScreen(screen *ebiten.Image) {
	g.drawOverlay(screen, "NO MOVES LEFT", "ENTER - retry | ESC - menu")
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		alpha := uint8(255 * (p.Life / p.MaxLife))
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		
		circle := g.createCircle(int(p.Size*2), c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-p.Size, p.Y-p.Size)
		screen.DrawImage(circle, op)
	}
}

func (g *Game) createCircle(size int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	center := float64(size) / 2
	radius := float64(size) / 2
	
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			if math.Sqrt(dx*dx+dy*dy) <= radius {
				img.Set(x, y, c)
			}
		}
	}
	
	return img
}

func (g *Game) createProgressBar(width, height int, progress float64) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	
	// Фон
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, ui.ColorDarkerPurple)
		}
	}
	
	// Заполнение
	if progress > 0 {
		fillWidth := int(float64(width) * progress)
		var c color.RGBA
		if progress < 0.33 {
			c = color.RGBA{255, 100, 100, 255}
		} else if progress < 0.66 {
			c = color.RGBA{255, 215, 0, 255}
		} else {
			c = color.RGBA{100, 255, 100, 255}
		}
		
		for y := 0; y < height; y++ {
			for x := 0; x < fillWidth; x++ {
				img.Set(x, y, c)
			}
		}
	}
	
	return img
}

func (g *Game) inputClicked() bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		time.Sleep(50 * time.Millisecond) // Debounce
		return true
	}
	return false
}

// handleMouseInput обрабатывает клики мыши для выбора и обмена фишек
func (g *Game) handleMouseInput() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		tile := g.board.GetTileAt(x, y)
		if tile == nil {
			// Клик вне доски - сбросить выделение
			if g.selectedTile != nil {
				g.selectedTile.Selected = false
				g.selectedTile = nil
			}
			return
		}

		if g.selectedTile == nil {
			// Первая выбранная фишка
			g.selectedTile = tile
			tile.Selected = true
			g.lastAction = time.Now()
			g.hintActive = false
		} else if g.selectedTile == tile {
			// Клик на ту же фишку - снять выделение
			tile.Selected = false
			g.selectedTile = nil
		} else {
			// Клик на другую фишку - попытка обмена
			dr := absInt(g.selectedTile.Row - tile.Row)
			dc := absInt(g.selectedTile.Col - tile.Col)

			if dr+dc == 1 {
				// Соседние фишки - попытка обмена
				g.trySwap(g.selectedTile, tile)
			} else {
				// Не соседние - выбрать новую фишку
				g.selectedTile.Selected = false
				g.selectedTile = tile
				tile.Selected = true
				g.lastAction = time.Now()
				g.hintActive = false
			}
		}
	}
}

// trySwap пытается обменять две фишки и проверяет на наличие матчей
func (g *Game) trySwap(tile1, tile2 *logic.Tile) {
	// Временно обмениваем
	g.board.SwapTiles(tile1, tile2)

	// Проверяем наличие матчей
	matches := g.board.FindAllMatches()

	if len(matches) == 0 {
		// Нет матчей - отменяем обмен и показываем дрожание
		g.board.SwapTiles(tile1, tile2) // Обратно

		// Анимация дрожания
		tile1.Shake = 1.0
		tile2.Shake = 1.0

		// Сброс выделения
		tile1.Selected = false
		g.selectedTile = nil

		g.moves--
		g.lastAction = time.Now()
	} else {
		// Есть матчи - обмен успешен
		tile1.Selected = false
		g.selectedTile = nil
		g.moves--
		g.lastAction = time.Now()
		g.hintActive = false
	}
}

// showHint показывает подсказку - случайную валидную пару
func (g *Game) showHint() {
	if g.board == nil || g.board.IsAnimating {
		return
	}

	t1, t2 := g.board.FindHint()
	if t1 != nil && t2 != nil {
		g.hintTile1 = t1
		g.hintTile2 = t2
		g.hintActive = true
		// Убираем подсветку через 2 секунды
		go func() {
			time.Sleep(2 * time.Second)
			g.hintActive = false
		}()
	}
}

func (g *Game) saveProgress() {
	// TODO
}

func (g *Game) loadProgress() {
	g.level = 1
	g.coins = 1000
	g.energy = 5
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
