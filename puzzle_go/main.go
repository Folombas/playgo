// Crystal Cascade v2.0 - Go365 Challenge Day 100
// Полностью переписанная игра Match-3 с использованием Ebitengine
// Дата: 8 апреля 2026
// FIXED: исправлены все баги с индексацией и созданием поля

package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ============================================================================
// КОНСТАНТЫ И КОНФИГУРАЦИЯ
// ============================================================================

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	BoardCols    = 8
	BoardRows    = 8
	CellSize     = 56
	CellPadding  = 4
	BoardOffsetX = (ScreenWidth - BoardCols*(CellSize+CellPadding)) / 2
	BoardOffsetY = 120
	AnimSpeed    = 0.18
	TargetFPS    = 60
)

// Цветовая палитра
var (
	ColorBgDark     = color.RGBA{18, 22, 40, 255}
	ColorBgPanel    = color.RGBA{28, 33, 58, 235}
	ColorBorder     = color.RGBA{90, 170, 245, 255}
	ColorTextWhite  = color.RGBA{235, 235, 245, 255}
	ColorTextGold   = color.RGBA{255, 210, 0, 255}
	ColorTextAccent = color.RGBA{110, 190, 250, 255}
	ColorCellBg     = color.RGBA{35, 40, 65, 190}
	ColorHighlight  = color.RGBA{255, 255, 90, 180}
	ColorCombo      = color.RGBA{255, 175, 45, 255}
)

// Цвета кристаллов
var gemColors = []color.RGBA{
	{255, 60, 60, 255},   // Red
	{60, 120, 255, 255},  // Blue
	{60, 250, 60, 255},   // Green
	{255, 250, 55, 255},  // Yellow
	{175, 55, 255, 255},  // Purple
	{255, 155, 55, 255},  // Orange
	{55, 250, 250, 255},  // Cyan
	{255, 115, 175, 255}, // Pink
}

var gemNames = []string{
	"Красный", "Синий", "Зелёный", "Жёлтый",
	"Фиолетовый", "Оранжевый", "Голубой", "Розовый",
}

// ============================================================================
// ТИПЫ ДАННЫХ
// ============================================================================

type GemType int

const (
	GemRed GemType = iota
	GemBlue
	GemGreen
	GemYellow
	GemPurple
	GemOrange
	GemCyan
	GemPink
	GemCount
)

type SpecialType int

const (
	SpecialNone SpecialType = iota
	SpecialBomb
	SpecialRow
	SpecialCol
	SpecialRainbow
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateSwapping
	StateRemoving
	StateDropping
)

type Gem struct {
	Type      GemType
	Special   SpecialType
	Col       int
	Row       int
	X         float64
	Y         float64
	TargetX   float64
	TargetY   float64
	Alpha     float64
	Scale     float64
	Selected  bool
	Removing  bool
}

type Particle struct {
	X, Y         float64
	VX, VY       float64
	Life         float64
	MaxLife      float64
	Color        color.RGBA
	Size         float64
	Acceleration float64
}

type FloatingText struct {
	X, Y    float64
	Text    string
	Life    float64
	MaxLife float64
	Color   color.RGBA
}

type Button struct {
	X, Y, W, H float64
	Text       string
	Hover      bool
	Action     func()
}

type Board struct {
	Grid [][]*Gem
}

type Game struct {
	State         GameState
	Board         *Board
	Particles     []*Particle
	FloatingTxt   []*FloatingText
	Score         int
	Combo         int
	MaxCombo      int
	Moves         int
	TargetScore   int
	Level         int
	SelectedGem   *Gem
	Buttons       map[string]*Button
	GemImages     []*ebiten.Image
	GameTime      float64
	StateTimer    float64
	ShakeTimer    float64
	ShakeDuration float64
	ShakeIntensity float64
	ShakeX        float64
	ShakeY        float64
}

// ============================================================================
// УТИЛИТЫ
// ============================================================================

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func createGemImage(gemType GemType, size int) *ebiten.Image {
	// Создаём изображение в памяти (standard library)
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	c := gemColors[gemType]

	half := size / 2

	// Рисуем форму кристалла (ромб)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := math.Abs(float64(x)-float64(half)) / float64(half)
			dy := math.Abs(float64(y)-float64(half)) / float64(half)

			if dx+dy <= 1.0 {
				img.Set(x, y, c)
			}
		}
	}

	// Блик
	shineX := half / 2
	shineY := half / 2
	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			if dx*dx+dy*dy <= 16 {
				sx, sy := shineX+dx, shineY+dy
				if sx >= 0 && sx < size && sy >= 0 && sy < size {
					px := img.Pix[(sy*img.Stride + sx*4)]
					if px > 0 {
						img.Set(sx, sy, color.RGBA{255, 255, 255, 160})
					}
				}
			}
		}
	}

	// Обводка
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := math.Abs(float64(x)-float64(half)) / float64(half)
			dy := math.Abs(float64(y)-float64(half)) / float64(half)

			if dx+dy >= 0.88 && dx+dy <= 1.0 {
				img.Set(x, y, color.RGBA{255, 255, 255, 180})
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// ============================================================================
// BOARD - ИСПРАВЛЕННАЯ ВЕРСИЯ
// ============================================================================

func NewBoard() *Board {
	b := &Board{
		Grid: make([][]*Gem, BoardCols),
	}
	for c := 0; c < BoardCols; c++ {
		b.Grid[c] = make([]*Gem, BoardRows)
	}
	
	// Заполняем без совпадений
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			b.Grid[c][r] = b.createGemSafe(c, r)
		}
	}
	
	return b
}

// createGemSafe создаёт кристалл без совпадений (БЕЗОПАСНАЯ ВЕРСИЯ)
func (b *Board) createGemSafe(c, r int) *Gem {
	var gemType GemType
	
	// Пробуем найти тип без совпадений
	for attempts := 0; attempts < 50; attempts++ {
		gemType = GemType(rand.Intn(int(GemCount)))
		if !b.wouldMatch(c, r, gemType) {
			break
		}
	}
	
	x := float64(BoardOffsetX + c*(CellSize+CellPadding))
	y := float64(BoardOffsetY + r*(CellSize+CellPadding))
	
	return &Gem{
		Type:      gemType,
		Special:   SpecialNone,
		Col:       c,
		Row:       r,
		X:         x,
		Y:         y,
		TargetX:   x,
		TargetY:   y,
		Alpha:     1,
		Scale:     1,
		Selected:  false,
		Removing:  false,
	}
}

// wouldMatch проверяет только УЖЕ СУЩЕСТВУЮЩИЕ соседи
func (b *Board) wouldMatch(c, r int, t GemType) bool {
	// Horizontal - проверяем только заполненные ячейки
	count := 1
	for i := c - 1; i >= 0; i-- {
		g := b.Grid[i][r]
		if g == nil || g.Type != t {
			break
		}
		count++
	}
	for i := c + 1; i < BoardCols; i++ {
		g := b.Grid[i][r]
		if g == nil || g.Type != t {
			break
		}
		count++
	}
	if count >= 3 {
		return true
	}
	
	// Vertical - проверяем только заполненные ячейки
	count = 1
	for i := r - 1; i >= 0; i-- {
		g := b.Grid[c][i]
		if g == nil || g.Type != t {
			break
		}
		count++
	}
	for i := r + 1; i < BoardRows; i++ {
		g := b.Grid[c][i]
		if g == nil || g.Type != t {
			break
		}
		count++
	}
	return count >= 3
}

func (b *Board) Get(c, r int) *Gem {
	if c < 0 || c >= BoardCols || r < 0 || r >= BoardRows {
		return nil
	}
	return b.Grid[c][r]
}

func (b *Board) Swap(c1, r1, c2, r2 int) bool {
	dc := c1 - c2
	dr := r1 - r2
	if !((dc == 1 || dc == -1) && dr == 0 || dc == 0 && (dr == 1 || dr == -1)) {
		return false
	}
	
	g1 := b.Grid[c1][r1]
	g2 := b.Grid[c2][r2]
	if g1 == nil || g2 == nil {
		return false
	}
	
	b.Grid[c1][r1] = g2
	b.Grid[c2][r2] = g1
	
	g1.Col, g1.Row = c2, r2
	g2.Col, g2.Row = c1, r1
	
	g1.TargetX = float64(BoardOffsetX + c2*(CellSize+CellPadding))
	g1.TargetY = float64(BoardOffsetY + r2*(CellSize+CellPadding))
	g2.TargetX = float64(BoardOffsetX + c1*(CellSize+CellPadding))
	g2.TargetY = float64(BoardOffsetY + r1*(CellSize+CellPadding))
	
	return true
}

func (b *Board) FindMatches() ([][][2]int, int) {
	matched := make(map[[2]int]bool)
	var matches [][][2]int
	
	// Horizontal
	for r := 0; r < BoardRows; r++ {
		c := 0
		for c < BoardCols {
			gem := b.Grid[c][r]
			if gem == nil || gem.Removing {
				c++
				continue
			}
			
			match := [][2]int{{c, r}}
			for cc := c + 1; cc < BoardCols; cc++ {
				g := b.Grid[cc][r]
				if g != nil && !g.Removing && g.Type == gem.Type {
					match = append(match, [2]int{cc, r})
				} else {
					break
				}
			}
			
			if len(match) >= 3 {
				// Special creation
				if len(match) == 4 {
					bc, br := match[1][0], match[1][1]
					if b.Grid[bc][br] != nil {
						b.Grid[bc][br].Special = SpecialBomb
					}
				} else if len(match) >= 5 {
					bc, br := match[2][0], match[2][1]
					if b.Grid[bc][br] != nil {
						b.Grid[bc][br].Special = SpecialRainbow
					}
				}
				
				for _, pos := range match {
					matched[pos] = true
				}
				matches = append(matches, match)
				c += len(match)
			} else {
				c++
			}
		}
	}
	
	// Vertical
	for c := 0; c < BoardCols; c++ {
		r := 0
		for r < BoardRows {
			gem := b.Grid[c][r]
			if gem == nil || gem.Removing {
				r++
				continue
			}
			
			match := [][2]int{{c, r}}
			for rr := r + 1; rr < BoardRows; rr++ {
				g := b.Grid[c][rr]
				if g != nil && !g.Removing && g.Type == gem.Type {
					match = append(match, [2]int{c, rr})
				} else {
					break
				}
			}
			
			if len(match) >= 3 {
				if len(match) == 4 {
					bc, br := match[1][0], match[1][1]
					if b.Grid[bc][br] != nil {
						b.Grid[bc][br].Special = SpecialCol
					}
				} else if len(match) >= 5 {
					bc, br := match[2][0], match[2][1]
					if b.Grid[bc][br] != nil {
						b.Grid[bc][br].Special = SpecialRainbow
					}
				}
				
				for _, pos := range match {
					matched[pos] = true
				}
				matches = append(matches, match)
				r += len(match)
			} else {
				r++
			}
		}
	}
	
	return matches, len(matched)
}

func (b *Board) ApplyGravity() bool {
	dropped := false
	
	for c := 0; c < BoardCols; c++ {
		// Удаляем помеченные
		for r := 0; r < BoardRows; r++ {
			if b.Grid[c][r] != nil && b.Grid[c][r].Removing {
				b.Grid[c][r] = nil
			}
		}
		
		// Сдвигаем вниз
		writePos := BoardRows - 1
		for r := BoardRows - 1; r >= 0; r-- {
			if b.Grid[c][r] != nil {
				if r != writePos {
					gem := b.Grid[c][r]
					b.Grid[c][writePos] = gem
					b.Grid[c][r] = nil
					
					gem.Row = writePos
					gem.TargetY = float64(BoardOffsetY + writePos*(CellSize+CellPadding))
					dropped = true
				}
				writePos--
			}
		}
		
		// Создаём новые сверху
		for r := writePos; r >= 0; r-- {
			gemType := GemType(rand.Intn(int(GemCount)))
			x := float64(BoardOffsetX + c*(CellSize+CellPadding))
			startY := float64(BoardOffsetY - (writePos-r+1)*(CellSize+CellPadding))
			targetY := float64(BoardOffsetY + r*(CellSize+CellPadding))
			
			gem := &Gem{
				Type:      gemType,
				Special:   SpecialNone,
				Col:       c,
				Row:       r,
				X:         x,
				Y:         startY,
				TargetX:   x,
				TargetY:   targetY,
				Alpha:     1,
				Scale:     1,
				Selected:  false,
				Removing:  false,
			}
			
			b.Grid[c][r] = gem
			dropped = true
		}
	}
	
	return dropped
}

func (b *Board) IsAnimating() bool {
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			gem := b.Grid[c][r]
			if gem == nil {
				continue
			}
			
			dx := gem.TargetX - gem.X
			dy := gem.TargetY - gem.Y
			if dx*dx+dy*dy > 0.5 {
				return true
			}
			
			if gem.Removing && gem.Alpha > 0.01 {
				return true
			}
		}
	}
	return false
}

func (b *Board) GetGemAt(mx, my float64) *Gem {
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			gem := b.Grid[c][r]
			if gem != nil && !gem.Removing {
				half := float64(CellSize) / 2
				gx := gem.X + half
				gy := gem.Y + half
				
				if mx >= gx-half && mx <= gx+half && my >= gy-half && my <= gy+half {
					return gem
				}
			}
		}
	}
	return nil
}

// ============================================================================
// GAME
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	
	g := &Game{
		State:       StateMenu,
		Score:       0,
		Combo:       0,
		MaxCombo:    0,
		Moves:       0,
		Level:       1,
		TargetScore: 3000,
		Buttons:     make(map[string]*Button),
		GemImages:   make([]*ebiten.Image, GemCount),
		Particles:   []*Particle{},
		FloatingTxt: []*FloatingText{},
	}
	
	// Создаём изображения кристаллов
	for i := 0; i < int(GemCount); i++ {
		g.GemImages[i] = createGemImage(GemType(i), CellSize)
	}
	
	g.createButtons()
	
	return g
}

func (g *Game) createButtons() {
	g.Buttons["play"] = &Button{
		X: ScreenWidth/2 - 100, Y: 420, W: 200, H: 55,
		Text: "▶  ИГРАТЬ",
		Action: func() {
			g.startGame()
		},
	}
	
	g.Buttons["exit"] = &Button{
		X: ScreenWidth/2 - 100, Y: 495, W: 200, H: 55,
		Text: "✕  ВЫЙТИ",
		Action: func() {
			os.Exit(0)
		},
	}
	
	g.Buttons["pause"] = &Button{
		X: ScreenWidth - 130, Y: 15, W: 115, H: 45,
		Text: "⏸ Пауза",
		Action: func() {},
	}
	
	g.Buttons["back"] = &Button{
		X: 15, Y: ScreenHeight - 70, W: 140, H: 45,
		Text: "← Назад",
		Action: func() {
			g.State = StateMenu
			g.Board = nil
		},
	}
}

func (g *Game) startGame() {
	g.State = StatePlaying
	g.Board = NewBoard()
	g.Score = 0
	g.Combo = 0
	g.MaxCombo = 0
	g.Moves = 0
	g.Level = 1
	g.TargetScore = 3000
	g.Particles = []*Particle{}
	g.FloatingTxt = []*FloatingText{}
	g.SelectedGem = nil
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.GameTime += 1.0 / TargetFPS
	
	// Update shake
	if g.ShakeTimer > 0 {
		g.ShakeTimer -= 1.0 / TargetFPS
		if g.ShakeTimer > 0 {
			g.ShakeX = (rand.Float64() - 0.5) * g.ShakeIntensity
			g.ShakeY = (rand.Float64() - 0.5) * g.ShakeIntensity
		} else {
			g.ShakeX = 0
			g.ShakeY = 0
		}
	}
	
	// Mouse position
	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)
	
	// Button hover
	for _, btn := range g.Buttons {
		btn.Hover = fmx >= btn.X && fmx <= btn.X+btn.W && fmy >= btn.Y && fmy <= btn.Y+btn.H
	}
	
	// Button click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for _, btn := range g.Buttons {
			if btn.Hover && btn.Action != nil {
				btn.Action()
				return nil
			}
		}
	}
	
	// Game state update
	switch g.State {
	case StatePlaying:
		g.updatePlaying(fmx, fmy)
	case StateSwapping:
		g.updateSwapping()
	case StateRemoving:
		g.updateRemoving()
	case StateDropping:
		g.updateDropping()
	}
	
	// Update particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := g.Particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += p.Acceleration
		p.Life -= 1.0 / TargetFPS
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}
	
	// Update floating texts
	for i := len(g.FloatingTxt) - 1; i >= 0; i-- {
		ft := g.FloatingTxt[i]
		ft.Y -= 1.2
		ft.Life -= 1.0 / TargetFPS
		if ft.Life <= 0 {
			g.FloatingTxt = append(g.FloatingTxt[:i], g.FloatingTxt[i+1:]...)
		}
	}
	
	return nil
}

func (g *Game) updatePlaying(mx, my float64) {
	if g.Board == nil {
		return
	}
	
	// Animate gems
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			gem := g.Board.Grid[c][r]
			if gem != nil {
				gem.X = lerp(gem.X, gem.TargetX, AnimSpeed)
				gem.Y = lerp(gem.Y, gem.TargetY, AnimSpeed)
				if gem.Removing {
					gem.Scale *= 0.82
					gem.Alpha *= 0.78
				}
			}
		}
	}
	
	// Check if board is stable
	if g.Board.IsAnimating() {
		return
	}
	
	// Check for matches
	matches, count := g.Board.FindMatches()
	if count > 0 {
		g.Combo++
		if g.Combo > g.MaxCombo {
			g.MaxCombo = g.Combo
		}
		
		points := count * 100 * (1 + g.Combo/2)
		g.Score += points
		
		// Mark for removal
		for _, match := range matches {
			for _, pos := range match {
				gem := g.Board.Get(pos[0], pos[1])
				if gem != nil {
					gem.Removing = true
					g.spawnParticles(gem.X+float64(CellSize)/2, gem.Y+float64(CellSize)/2, gem.Type, 6)
				}
			}
		}
		
		// Floating score
		if len(matches) > 0 {
			pos := matches[0][0]
			gem := g.Board.Get(pos[0], pos[1])
			if gem != nil {
				g.FloatingTxt = append(g.FloatingTxt, &FloatingText{
					X:       gem.X + float64(CellSize)/2,
					Y:       gem.Y,
					Text:    fmt.Sprintf("+%d", points),
					Life:    1.5,
					MaxLife: 1.5,
					Color:   ColorTextGold,
				})
			}
		}
		
		// Screen shake
		if g.Combo >= 3 {
			g.ShakeDuration = 0.25
			g.ShakeTimer = g.ShakeDuration
			g.ShakeIntensity = math.Min(float64(g.Combo)*1.5, 8)
		}
		
		g.State = StateRemoving
		g.StateTimer = 0.4
		return
	}
	
	// Reset combo if no matches
	if g.Combo > 0 {
		g.Combo = 0
	}
	
	// Handle input
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		gem := g.Board.GetGemAt(mx, my)
		if gem != nil {
			if g.SelectedGem == nil {
				g.SelectedGem = gem
				gem.Selected = true
			} else if g.SelectedGem == gem {
				g.SelectedGem.Selected = false
				g.SelectedGem = nil
			} else {
				// Check adjacency
				dc := gem.Col - g.SelectedGem.Col
				dr := gem.Row - g.SelectedGem.Row
				if (dc == 1 || dc == -1) && dr == 0 || dc == 0 && (dr == 1 || dr == -1) {
					// Try swap
					g.Board.Swap(gem.Col, gem.Row, g.SelectedGem.Col, g.SelectedGem.Row)
					g.Moves++
					g.SelectedGem.Selected = false
					g.SelectedGem = nil
					g.State = StateSwapping
					g.StateTimer = 0.5
				} else {
					g.SelectedGem.Selected = false
					g.SelectedGem = gem
					gem.Selected = true
				}
			}
		}
	}
}

func (g *Game) updateSwapping() {
	g.StateTimer -= 1.0 / TargetFPS
	
	if !g.Board.IsAnimating() || g.StateTimer <= 0 {
		// Check if swap created matches
		_, count := g.Board.FindMatches()
		if count > 0 {
			g.State = StateRemoving
			g.StateTimer = 0.4
		} else {
			// Swap back
			// Find the two gems that were swapped
			for c := 0; c < BoardCols; c++ {
				for r := 0; r < BoardRows; r++ {
					gem := g.Board.Grid[c][r]
					if gem != nil && (gem.X != gem.TargetX || gem.Y != gem.TargetY) {
						// This shouldn't happen, just go back to playing
					}
				}
			}
			g.State = StatePlaying
		}
	}
}

func (g *Game) updateRemoving() {
	g.StateTimer -= 1.0 / TargetFPS
	
	if g.StateTimer <= 0 {
		// Apply gravity
		if g.Board.ApplyGravity() {
			g.State = StateDropping
			g.StateTimer = 0.8
		} else {
			g.State = StatePlaying
		}
	}
}

func (g *Game) updateDropping() {
	g.StateTimer -= 1.0 / TargetFPS
	
	if !g.Board.IsAnimating() || g.StateTimer <= 0 {
		g.State = StatePlaying
	}
}

func (g *Game) spawnParticles(x, y float64, gemType GemType, count int) {
	c := gemColors[gemType]
	
	for i := 0; i < count; i++ {
		angle := float64(i)*6.2832/float64(count) + rand.Float64()*0.5
		speed := 2 + rand.Float64()*3
		
		g.Particles = append(g.Particles, &Particle{
			X:            x,
			Y:            y,
			VX:           math.Cos(angle) * speed,
			VY:           math.Sin(angle)*speed - 1.5,
			Life:         0.8 + rand.Float64()*0.4,
			MaxLife:      1.0,
			Color:        c,
			Size:         3 + rand.Float64()*3,
			Acceleration: 0.12,
		})
	}
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ColorBgDark)
	
	if g.State == StateMenu {
		g.drawMenu(screen)
	} else {
		g.drawGame(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Stars background
	rand.Seed(42)
	for i := 0; i < 80; i++ {
		x := rand.Float64() * ScreenWidth
		y := rand.Float64() * ScreenHeight
		size := rand.Float64()*2 + 0.5
		alpha := uint8(rand.Float64()*120 + 80)
		
		star := ebiten.NewImage(int(size*2), int(size*2))
		star.Fill(color.RGBA{255, 255, 255, alpha})
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		screen.DrawImage(star, op)
	}
	rand.Seed(time.Now().UnixNano())
	
	// Title panel
	panel := ebiten.NewImage(450, 90)
	panel.Fill(ColorBgPanel)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ScreenWidth/2-225), 220)
	screen.DrawImage(panel, op)
	
	// Title border
	border := ebiten.NewImage(450, 3)
	border.Fill(ColorBorder)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(ScreenWidth/2-225), 218)
	screen.DrawImage(border, op2)
	
	// Title text
	text.Draw(screen, "💎 CRYSTAL CASCADE 💎", nil, ScreenWidth/2-170, 275, ColorTextWhite)
	text.Draw(screen, "Go365 Challenge - Day 100", nil, ScreenWidth/2-110, 300, ColorTextAccent)
	
	// Buttons
	g.drawButton(screen, g.Buttons["play"])
	g.drawButton(screen, g.Buttons["exit"])
	
	// Footer
	text.Draw(screen, "Собери 3+ кристалла • Создавай комбо • Бей рекорды!", nil, ScreenWidth/2-220, ScreenHeight-40, ColorTextAccent)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	if g.Board == nil {
		return
	}
	
	// Apply shake offset
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.ShakeX, g.ShakeY)
	
	// Board background
	boardW := BoardCols*(CellSize+CellPadding) + 12
	boardH := BoardRows*(CellSize+CellPadding) + 12
	
	boardBg := ebiten.NewImage(boardW, boardH)
	boardBg.Fill(ColorBgPanel)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(BoardOffsetX-6), float64(BoardOffsetY-6))
	op2.GeoM.Translate(g.ShakeX, g.ShakeY)
	screen.DrawImage(boardBg, op2)
	
	// Board border
	border := ebiten.NewImage(boardW, 3)
	border.Fill(ColorBorder)
	
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(float64(BoardOffsetX-6), float64(BoardOffsetY-8))
	op3.GeoM.Translate(g.ShakeX, g.ShakeY)
	screen.DrawImage(border, op3)
	
	// Cells
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			cellX := float64(BoardOffsetX + c*(CellSize+CellPadding))
			cellY := float64(BoardOffsetY + r*(CellSize+CellPadding))
			
			cell := ebiten.NewImage(CellSize, CellSize)
			vector.DrawFilledRect(cell, 0, 0, float32(CellSize), float32(CellSize), ColorCellBg, false)
			
			op4 := &ebiten.DrawImageOptions{}
			op4.GeoM.Translate(cellX, cellY)
			op4.GeoM.Translate(g.ShakeX, g.ShakeY)
			screen.DrawImage(cell, op4)
		}
	}
	
	// Gems
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			gem := g.Board.Grid[c][r]
			if gem != nil && gem.Alpha > 0.01 {
				g.drawGem(screen, gem)
			}
		}
	}
	
	// Particles
	for _, p := range g.Particles {
		size := int(p.Size * (p.Life / p.MaxLife))
		if size < 1 {
			continue
		}
		
		alpha := uint8((p.Life / p.MaxLife) * 255)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		
		particle := ebiten.NewImage(size, size)
		particle.Fill(c)
		
		op5 := &ebiten.DrawImageOptions{}
		op5.GeoM.Translate(p.X-float64(size)/2, p.Y-float64(size)/2)
		screen.DrawImage(particle, op5)
	}
	
	// Floating texts
	for _, ft := range g.FloatingTxt {
		alpha := uint8((ft.Life / ft.MaxLife) * 255)
		c := color.RGBA{ft.Color.R, ft.Color.G, ft.Color.B, alpha}
		text.Draw(screen, ft.Text, nil, int(ft.X), int(ft.Y), c)
	}
	
	// UI
	g.drawGameUI(screen)
}

func (g *Game) drawGem(screen *ebiten.Image, gem *Gem) {
	img := g.GemImages[gem.Type]
	if img == nil {
		return
	}
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(gem.Scale, gem.Scale)
	op.GeoM.Translate(gem.X, gem.Y)
	op.ColorM.Scale(1, 1, 1, gem.Alpha)
	op.GeoM.Translate(g.ShakeX, g.ShakeY)
	
	screen.DrawImage(img, op)
	
	// Selection highlight
	if gem.Selected {
		highlight := ebiten.NewImage(CellSize+6, CellSize+6)
		vector.DrawFilledRect(highlight, 0, 0, float32(CellSize+6), float32(CellSize+6), ColorHighlight, false)
		
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(gem.X-3, gem.Y-3)
		op2.ColorM.Scale(1, 1, 1, 0.35)
		screen.DrawImage(highlight, op2)
	}
	
	// Special indicators
	if gem.Special != SpecialNone {
		indicator := ebiten.NewImage(CellSize, CellSize)
		var indColor color.RGBA
		
		switch gem.Special {
		case SpecialBomb:
			indColor = color.RGBA{255, 95, 45, 140}
		case SpecialRainbow:
			shimmer := uint8(math.Sin(g.GameTime*5)*50 + 195)
			indColor = color.RGBA{shimmer, shimmer, 255, 140}
		case SpecialRow, SpecialCol:
			indColor = color.RGBA{95, 250, 95, 140}
		}
		
		vector.DrawFilledRect(indicator, 0, 0, float32(CellSize), float32(CellSize), indColor, false)
		
		op3 := &ebiten.DrawImageOptions{}
		op3.GeoM.Translate(gem.X, gem.Y)
		op3.ColorM.Scale(1, 1, 1, 0.35)
		screen.DrawImage(indicator, op3)
	}
}

func (g *Game) drawGameUI(screen *ebiten.Image) {
	// Score panel
	panel := ebiten.NewImage(220, 160)
	vector.DrawFilledRect(panel, 0, 0, 220, 160, ColorBgPanel, false)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(15, 35)
	screen.DrawImage(panel, op)
	
	// Border
	border := ebiten.NewImage(220, 3)
	border.Fill(ColorBorder)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(15, 35)
	screen.DrawImage(border, op2)
	
	// Score
	text.Draw(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), nil, 35, 68, ColorTextWhite)
	text.Draw(screen, fmt.Sprintf("ЦЕЛЬ: %d", g.TargetScore), nil, 35, 92, ColorTextAccent)
	
	if g.Combo > 1 {
		text.Draw(screen, fmt.Sprintf("COMBO x%d", g.Combo), nil, 35, 116, ColorCombo)
	}
	
	text.Draw(screen, fmt.Sprintf("ХОДЫ: %d", g.Moves), nil, 35, 140, ColorTextAccent)
	
	// Title bar
	titleBar := ebiten.NewImage(ScreenWidth, 50)
	vector.DrawFilledRect(titleBar, 0, 0, float32(ScreenWidth), 50, ColorBgPanel, false)
	
	op3 := &ebiten.DrawImageOptions{}
	screen.DrawImage(titleBar, op3)
	
	titleBorder := ebiten.NewImage(ScreenWidth, 3)
	titleBorder.Fill(ColorBorder)
	
	op4 := &ebiten.DrawImageOptions{}
	op4.GeoM.Translate(0, 50)
	screen.DrawImage(titleBorder, op4)
	
	text.Draw(screen, "💎 CRYSTAL CASCADE", nil, ScreenWidth/2-130, 33, ColorTextWhite)
	
	// Pause button
	g.drawButton(screen, g.Buttons["pause"])
}

func (g *Game) drawButton(screen *ebiten.Image, btn *Button) {
	if btn == nil {
		return
	}
	
	// Button background
	btnImg := ebiten.NewImage(int(btn.W), int(btn.H))
	
	if btn.Hover {
		vector.DrawFilledRect(btnImg, 0, 0, float32(btn.W), float32(btn.H), color.RGBA{55, 75, 115, 255}, false)
	} else {
		vector.DrawFilledRect(btnImg, 0, 0, float32(btn.W), float32(btn.H), color.RGBA{35, 45, 75, 255}, false)
	}
	
	// Border
	border := ebiten.NewImage(int(btn.W), 3)
	border.Fill(ColorBorder)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(btn.X, btn.Y)
	screen.DrawImage(btnImg, op)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(btn.X, btn.Y)
	screen.DrawImage(border, op2)
	
	// Text
	text.Draw(screen, btn.Text, nil, int(btn.X+btn.W/2)-len(btn.Text)*5, int(btn.Y+btn.H/2+6), ColorTextWhite)
}

// ============================================================================
// EBITENGINE
// ============================================================================

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Crystal Cascade - Go365 Day 100")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
