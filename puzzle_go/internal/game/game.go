package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/puzzle_go/internal/board"
	"github.com/playgo/puzzle_go/internal/config"
	"github.com/playgo/puzzle_go/internal/entity"
	"github.com/playgo/puzzle_go/internal/render"
	"github.com/playgo/puzzle_go/internal/ui"
)

// GameState определяет состояние игры
type GameState int

const (
	StateIdle GameState = iota
	StateSelected
	StateSwapping
	StateChecking
	StateRemoving
	StateFalling
	StateCombo
	StateGameOver
)

// Game представляет основную игру
type Game struct {
	board        *board.Board
	spriteMgr    *render.SpriteManager
	ui           *ui.UI
	state        GameState
	score        int
	combo        int
	selected     *entity.Crystal
	hoverCrystal *entity.Crystal
	particles    []*Particle
	animationT   float64
}

// Particle представляет частицу для эффектов
type Particle struct {
	X      float64
	Y      float64
	VX     float64
	VY     float64
	Life   float64
	MaxLife float64
	Color  []uint8
	Size   float64
}

// NewGame создает новую игру
func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		state:     StateIdle,
		score:     0,
		combo:     0,
		particles: make([]*Particle, 0),
	}

	g.board = board.NewBoard()
	g.spriteMgr = render.NewSpriteManager()
	
	if err := g.spriteMgr.Load(); err != nil {
		fmt.Printf("Warning: sprite loading error: %v\n", err)
	}

	g.ui = ui.NewUI()

	return g
}

// Update обновляет игру
func (g *Game) Update() error {
	// Обновление частиц
	g.updateParticles()

	// Обновление кристаллов
	g.updateCrystals()

	// Обработка состояний
	switch g.state {
	case StateIdle:
		g.handleInput()
	case StateSelected:
		g.handleInput()
	case StateSwapping:
		g.animationT += 0.05
		if g.animationT >= 1.0 {
			g.state = StateChecking
			g.animationT = 0
		}
	case StateChecking:
		matches := g.board.FindAllMatches()
		if len(matches) > 0 {
			g.combo++
			g.state = StateRemoving
			g.animationT = 0
		} else {
			// Нет совпадений - отмена обмена
			g.state = StateIdle
			g.selected = nil
		}
	case StateRemoving:
		g.animationT += 0.05
		if g.animationT >= 1.0 {
			count := g.board.RemoveMatched()
			if count > 0 {
				baseScore := count * config.BaseMatchScore
				comboBonus := int(float64(baseScore) * (1.0 + float64(g.combo-1)*0.5))
				g.score += comboBonus

				// Создание частиц
				g.spawnParticles(count)
			}
			g.state = StateFalling
			g.animationT = 0
		}
	case StateFalling:
		g.board.ApplyGravity()
		g.state = StateChecking
		g.animationT = 0
	}

	return nil
}

// updateCrystals обновляет все кристаллы
func (g *Game) updateCrystals() {
	for col := 0; col < config.BoardCols; col++ {
		for row := 0; row < config.BoardRows; row++ {
			crystal := g.board.GetCrystal(col, row)
			if crystal != nil {
				crystal.Update(0.2)
			}
		}
	}
}

// handleInput обрабатывает ввод
func (g *Game) handleInput() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		
		crystal := g.board.GetCrystalAtPosition(float64(mx), float64(my))
		
		if crystal != nil {
			if g.state == StateIdle {
				// Выделение кристалла
				g.selected = crystal
				crystal.IsSelected = true
				g.state = StateSelected
			} else if g.state == StateSelected {
				// Попытка обмена
				if crystal != g.selected {
					if g.board.Swap(g.selected.Col, g.selected.Row, crystal.Col, crystal.Row) {
						g.selected.IsSelected = false
						g.state = StateSwapping
						g.animationT = 0
					}
				}
			}
		}
	}

	// Отмена выделения правой кнопкой
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		if g.selected != nil {
			g.selected.IsSelected = false
			g.selected = nil
			g.state = StateIdle
		}
	}
}

// spawnParticles создает частицы
func (g *Game) spawnParticles(count int) {
	for i := 0; i < count*10; i++ {
		particle := &Particle{
			X:       float64(config.BoardOffsetX + rand.Intn(config.BoardCols*config.CellSize)),
			Y:       float64(config.BoardOffsetY + rand.Intn(config.BoardRows*config.CellSize)),
			VX:      (rand.Float64() - 0.5) * 10,
			VY:      (rand.Float64() - 0.5) * 10,
			Life:    1.0,
			MaxLife: 1.0,
			Color:   []uint8{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255},
			Size:    float64(2 + rand.Intn(4)),
		}
		g.particles = append(g.particles, particle)
	}
}

// updateParticles обновляет частицы
func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.5 // гравитация
		p.Life -= 0.02

		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка экрана
	screen.Fill(color.RGBA{20, 20, 40, 255})

	// Отрисовка фона
	g.drawBackground(screen)

	// Отрисовка игрового поля
	g.drawBoard(screen)

	// Отрисовка частиц
	g.drawParticles(screen)

	// Отрисовка UI
	g.ui.Draw(screen, g.score, g.combo)
}

// drawBackground рисует фон
func (g *Game) drawBackground(screen *ebiten.Image) {
	// Градиентный фон
	for y := 0; y < config.ScreenHeight; y++ {
		ratio := float64(y) / float64(config.ScreenHeight)
		r := uint8(20 + float64(40)*ratio)
		g := uint8(20 + float64(30)*ratio)
		b := uint8(60 + float64(80)*ratio)
		
		line := ebiten.NewImage(config.ScreenWidth, 1)
		line.Fill(color.RGBA{r, g, b, 255})
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(y))
		screen.DrawImage(line, op)
	}
}

// drawBoard рисует игровое поле
func (g *Game) drawBoard(screen *ebiten.Image) {
	// Фон игрового поля
	boardWidth := config.BoardCols * (config.CellSize + config.CellPadding)
	boardHeight := config.BoardRows * (config.CellSize + config.CellPadding)
	
	boardBg := ebiten.NewImage(boardWidth+20, boardHeight+20)
	boardBg.Fill(color.RGBA{30, 30, 50, 200})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(config.BoardOffsetX-10), float64(config.BoardOffsetY-10))
	screen.DrawImage(boardBg, op)

	// Отрисовка кристаллов
	for col := 0; col < config.BoardCols; col++ {
		for row := 0; row < config.BoardRows; row++ {
			crystal := g.board.GetCrystal(col, row)
			if crystal != nil && crystal.Alpha > 0.01 {
				g.drawCrystal(screen, crystal)
			}
		}
	}
}

// drawCrystal рисует кристалл
func (g *Game) drawCrystal(screen *ebiten.Image, crystal *entity.Crystal) {
	img := g.spriteMgr.GetCrystalImage(crystal.Type)
	if img == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	
	// Масштабирование от центра
	centerX := float64(config.CellSize) / 2
	centerY := float64(config.CellSize) / 2
	
	op.GeoM.Translate(-centerX, -centerY)
	op.GeoM.Scale(crystal.Scale, crystal.Scale)
	op.GeoM.Translate(centerX, centerY)
	
	op.GeoM.Translate(crystal.X, crystal.Y)
	
	// Прозрачность
	colorMatrix := createAlphaMatrix(crystal.Alpha)
	op.ColorM = colorMatrix

	screen.DrawImage(img, op)
}

func createAlphaMatrix(alpha float64) ebiten.ColorM {
	var cm ebiten.ColorM
	cm.Scale(1, 1, 1, alpha)
	return cm
}

// Suppress unused import warning
var _ = color.RGBA{}

// drawParticles рисует частицы
func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		size := p.Size * p.Life
		if size < 0.5 {
			continue
		}

		particle := ebiten.NewImage(int(size), int(size))
		particle.Fill(color.RGBA{p.Color[0], p.Color[1], p.Color[2], uint8(p.Life * 255)})

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		
		screen.DrawImage(particle, op)
	}
}

// Layout возвращает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}
