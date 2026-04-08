// Crystal Cascade v2.0 - Go365 Challenge Day 100
// Полностью переписанная игра Match-3 с использованием Ebitengine
// Дата: 8 апреля 2026

package main

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// ============================================================================
// КОНСТАНТЫ И КОНФИГУРАЦИЯ
// ============================================================================

const (
	ScreenWidth  = 1280
	ScreenHeight = 960
	BoardCols    = 8
	BoardRows    = 10
	CellSize     = 64
	CellPadding  = 4
	BoardOffsetX = (ScreenWidth - BoardCols*(CellSize+CellPadding)) / 2
	BoardOffsetY = 140
	AnimSpeed    = 0.2
	TargetFPS    = 60
)

// Цветовая палитра
var (
	ColorBgDark      = color.RGBA{20, 25, 45, 255}
	ColorBgPanel     = color.RGBA{30, 35, 60, 240}
	ColorBorder      = color.RGBA{100, 180, 255, 255}
	ColorTextWhite   = color.RGBA{240, 240, 250, 255}
	ColorTextGold    = color.RGBA{255, 215, 0, 255}
	ColorTextAccent  = color.RGBA{120, 200, 255, 255}
	ColorCellBg      = color.RGBA{40, 45, 70, 200}
	ColorCellBorder  = color.RGBA{60, 70, 100, 150}
	ColorCombo       = color.RGBA{255, 180, 50, 255}
	ColorHighlight   = color.RGBA{255, 255, 100, 200}
)

// ============================================================================
// ТИПЫ ДАННЫХ
// ============================================================================

// GemType определяет тип кристалла
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

// SpecialType определяет тип специального элемента
type SpecialType int

const (
	SpecialNone SpecialType = iota
	SpecialBomb       // Взрыв 3x3
	SpecialRow        // Очистка строки
	SpecialCol        // Очистка колонки
	SpecialRainbow    // Уничтожает все одного типа
)

// GameState определяет состояние игры
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateSwapping
	StateRemoving
	StateDropping
	StateGameOver
)

// Gem - игровой кристалл
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
	Rotation  float64
	Selected  bool
	Removing  bool
	SwapTarget *Gem
}

// Particle - частица для эффектов
type Particle struct {
	X, Y       float64
	VX, VY     float64
	Life       float64
	MaxLife    float64
	Color      color.RGBA
	Size       float64
	Acceleration float64
}

// FloatingText - всплывающий текст с очками
type FloatingText struct {
	X, Y     float64
	Text     string
	Life     float64
	MaxLife  float64
	Color    color.RGBA
	Size     float64
}

// Button - UI кнопка
type Button struct {
	X, Y, W, H float64
	Text       string
	Image      *ebiten.Image
	Hover      bool
	Pressed    bool
	Action     func()
}

// Board - игровое поле
type Board struct {
	Grid [][]*Gem
}

// Game - основная структура игры
type Game struct {
	State       GameState
	Board       *Board
	GemImages   map[GemType]*ebiten.Image
	Particles   []*Particle
	FloatingTxt []*FloatingText
	Score       int
	Combo       int
	MaxCombo    int
	Moves       int
	TargetScore int
	Level       int

	SelectedGem *Gem
	DragStartX  float64
	DragStartY  float64
	Dragging    bool

	Buttons     map[string]*Button
	Font        font.Face
	SmallFont   font.Face

	BackgroundImg  *ebiten.Image
	PanelImg       *ebiten.Image

	GameTime     float64
	StateTimer   float64
	SwapGem1     *Gem
	SwapGem2     *Gem
	RemoveList   [][]int
	DropAnimDone bool

	ShakeX, ShakeY   float64
	ShakeDuration    float64
	ShakeTimer       float64
	ShakeIntensity   float64
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// loadPNG загружает PNG изображение в ebiten.Image
func loadPNG(path string) (*ebiten.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// createRoundedRect создаёт изображение со скруглёнными углами
func createRoundedRect(w, h int, radius int, fillColor color.RGBA, borderColor color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	
	// Fill main area
	mainRect := ebiten.NewImage(w-h, h)
	mainRect.Fill(fillColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(radius), 0)
	img.DrawImage(mainRect, op)
	
	// Border
	imgBorder := ebiten.NewImage(w, 3)
	imgBorder.Fill(borderColor)
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(0, 0)
	img.DrawImage(imgBorder, op2)
	
	return img
}

// lerp линейная интерполяция
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// clamp ограничивает значение
func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// ============================================================================
// РЕАЛИЗАЦИЯ Board
// ============================================================================

func NewBoard() *Board {
	b := &Board{
		Grid: make([][]*Gem, BoardCols),
	}
	for c := 0; c < BoardCols; c++ {
		b.Grid[c] = make([]*Gem, BoardRows)
	}
	b.Fill()
	return b
}

// Fill заполняет поле без начальных совпадений
func (b *Board) Fill() {
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			b.Grid[c][r] = b.createGemNoMatch(c, r)
		}
	}
}

// createGemNoMatch создаёт кристалл без совпадений
func (b *Board) createGemNoMatch(c, r int) *Gem {
	var gemType GemType
	attempts := 0
	for attempts < 100 {
		gemType = GemType(rand.Intn(int(GemCount)))
		if !b.wouldMatch(c, r, gemType) {
			break
		}
		attempts++
	}
	
	x := float64(BoardOffsetX + c*(CellSize+CellPadding))
	y := float64(BoardOffsetY + r*(CellSize+CellPadding))
	
	return &Gem{
		Type:    gemType,
		Special: SpecialNone,
		Col:     c,
		Row:     r,
		X:       x,
		Y:       y,
		TargetX: x,
		TargetY: y,
		Alpha:   1,
		Scale:   1,
	}
}

// wouldMatch проверяет, создаст ли кристалл совпадение
func (b *Board) wouldMatch(c, r int, t GemType) bool {
	// Horizontal
	count := 1
	for i := c - 1; i >= 0 && b.Grid[i][r] != nil && b.Grid[i][r].Type == t; i-- {
		count++
	}
	for i := c + 1; i < BoardCols && b.Grid[i][r] != nil && b.Grid[i][r].Type == t; i++ {
		count++
	}
	if count >= 3 {
		return true
	}
	
	// Vertical
	count = 1
	for i := r - 1; i >= 0 && b.Grid[c][i] != nil && b.Grid[c][i].Type == t; i-- {
		count++
	}
	for i := r + 1; r < BoardRows && b.Grid[c][i] != nil && b.Grid[c][i].Type == t; i++ {
		count++
	}
	return count >= 3
}

// Get возвращает кристалл по координатам
func (b *Board) Get(c, r int) *Gem {
	if c < 0 || c >= BoardCols || r < 0 || r >= BoardRows {
		return nil
	}
	return b.Grid[c][r]
}

// Set устанавливает кристалл по координатам
func (b *Board) Set(c, r int, gem *Gem) {
	if c >= 0 && c < BoardCols && r >= 0 && r < BoardRows {
		b.Grid[c][r] = gem
	}
}

// Swap меняет местами два кристалла
func (b *Board) Swap(c1, r1, c2, r2 int) bool {
	dc := c1 - c2
	dr := r1 - r2
	if (dc == 1 || dc == -1) && dr == 0 || dc == 0 && (dr == 1 || dr == -1) {
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
	return false
}

// FindMatches находит все совпадения
func (b *Board) FindMatches() ([][][2]int, int) {
	matched := make(map[[2]int]bool)
	var matches [][][2]int
	
	// Horizontal
	for r := 0; r < BoardRows; r++ {
		for c := 0; c <= BoardCols-3; c++ {
			gem := b.Grid[c][r]
			if gem == nil || gem.Removing {
				continue
			}
			
			match := [][2]int{{c, r}}
			for cc := c + 1; cc < BoardCols && b.Grid[cc][r] != nil && b.Grid[cc][r].Type == gem.Type && !b.Grid[cc][r].Removing; cc++ {
				match = append(match, [2]int{cc, r})
			}
			
			if len(match) >= 3 {
				// Check for special creation
				if len(match) == 4 {
					// Create bomb at center
					centerIdx := len(match) / 2
					bc, br := match[centerIdx][0], match[centerIdx][1]
					if b.Grid[bc][br] != nil {
						b.Grid[bc][br].Special = SpecialBomb
					}
				} else if len(match) >= 5 {
					// Create rainbow at center
					centerIdx := len(match) / 2
					bc, br := match[centerIdx][0], match[centerIdx][1]
					if b.Grid[bc][br] != nil {
						b.Grid[bc][br].Special = SpecialRainbow
					}
				}
				
				for _, pos := range match {
					key := [2]int{pos[0], pos[1]}
					if !matched[key] {
						matched[key] = true
					}
				}
				matches = append(matches, match)
				c += len(match) - 1
			}
		}
	}
	
	// Vertical
	for c := 0; c < BoardCols; c++ {
		for r := 0; r <= BoardRows-3; r++ {
			gem := b.Grid[c][r]
			if gem == nil || gem.Removing {
				continue
			}
			
			match := [][2]int{{c, r}}
			for rr := r + 1; rr < BoardRows && b.Grid[c][rr] != nil && b.Grid[c][rr].Type == gem.Type && !b.Grid[c][rr].Removing; rr++ {
				match = append(match, [2]int{c, rr})
			}
			
			if len(match) >= 3 {
				// Check for special creation
				if len(match) == 4 {
					centerIdx := len(match) / 2
					bc, br := match[centerIdx][0], match[centerIdx][1]
					if b.Grid[bc][br] != nil {
						b.Grid[bc][br].Special = SpecialCol
					}
				} else if len(match) >= 5 {
					centerIdx := len(match) / 2
					bc, br := match[centerIdx][0], match[centerIdx][1]
					if b.Grid[bc][br] != nil {
						b.Grid[bc][br].Special = SpecialRainbow
					}
				}
				
				for _, pos := range match {
					key := [2]int{pos[0], pos[1]}
					if !matched[key] {
						matched[key] = true
					}
				}
				matches = append(matches, match)
				r += len(match) - 1
			}
		}
	}
	
	totalMatched := 0
	for range matched {
		totalMatched++
	}
	
	return matches, totalMatched
}

// RemoveGem помечает кристалл на удаление
func (b *Board) RemoveGem(c, r int) {
	if c >= 0 && c < BoardCols && r >= 0 && r < BoardRows && b.Grid[c][r] != nil {
		b.Grid[c][r].Removing = true
	}
}

// ApplyGravity применяет гравитацию и создаёт новые кристаллы
func (b *Board) ApplyGravity() bool {
	dropped := false
	
	for c := 0; c < BoardCols; c++ {
		emptyRow := -1
		
		// Move existing gems down
		for r := BoardRows - 1; r >= 0; r-- {
			if b.Grid[c][r] == nil || b.Grid[c][r].Removing {
				if emptyRow == -1 {
					emptyRow = r
				}
				continue
			}
			
			if emptyRow != -1 {
				gem := b.Grid[c][r]
				b.Grid[c][emptyRow] = gem
				b.Grid[c][r] = nil
				
				gem.Row = emptyRow
				gem.TargetY = float64(BoardOffsetY + emptyRow*(CellSize+CellPadding))
				emptyRow--
				dropped = true
			}
		}
		
		// Create new gems at top
		for r := emptyRow; r >= 0; r-- {
			gemType := GemType(rand.Intn(int(GemCount)))
			x := float64(BoardOffsetX + c*(CellSize+CellPadding))
			y := float64(BoardOffsetY - (emptyRow-r+1)*(CellSize+CellPadding))
			targetY := float64(BoardOffsetY + r*(CellSize+CellPadding))
			
			gem := &Gem{
				Type:    gemType,
				Special: SpecialNone,
				Col:     c,
				Row:     r,
				X:       x,
				Y:       y,
				TargetX: x,
				TargetY: targetY,
				Alpha:   1,
				Scale:   1,
			}
			
			b.Grid[c][r] = gem
			dropped = true
		}
		
		// Clear removed gems
		for r := 0; r < BoardRows; r++ {
			if b.Grid[c][r] != nil && b.Grid[c][r].Removing {
				b.Grid[c][r] = nil
			}
		}
	}
	
	return dropped
}

// IsAnimating проверяет, есть ли активные анимации
func (b *Board) IsAnimating() bool {
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			gem := b.Grid[c][r]
			if gem == nil {
				continue
			}
			
			dx := gem.TargetX - gem.X
			dy := gem.TargetY - gem.Y
			if dx*dx+dy*dy > 1 {
				return true
			}
			
			if gem.Removing && gem.Alpha > 0.01 {
				return true
			}
		}
	}
	return false
}

// GetGemAt возвращает кристалл по экранным координатам
func (b *Board) GetGemAt(mx, my float64) *Gem {
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			gem := b.Grid[c][r]
			if gem != nil && !gem.Removing {
				gx := gem.X + float64(CellSize)/2
				gy := gem.Y + float64(CellSize)/2
				halfSize := float64(CellSize) / 2
				
				if mx >= gx-halfSize && mx <= gx+halfSize && my >= gy-halfSize && my <= gy+halfSize {
					return gem
				}
			}
		}
	}
	return nil
}

// ============================================================================
// РЕАЛИЗАЦИЯ Game
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
		TargetScore: 5000,
		Buttons:     make(map[string]*Button),
		GemImages:   make(map[GemType]*ebiten.Image),
		Particles:   []*Particle{},
		FloatingTxt: []*FloatingText{},
	}
	
	g.loadAssets()
	g.createButtons()
	g.initFonts()
	
	return g
}

func (g *Game) loadAssets() {
	// Get executable path
	execPath, _ := os.Getwd()
	spritesPath := filepath.Join(execPath, "..", "..", "sprites")
	
	// Try to load gem sprites from puzzle/pack1
	pack1Path := filepath.Join(spritesPath, "puzzle", "pack1", "PNG", "Default")
	
	// Map gem types to colors
	colorMap := map[GemType]string{
		GemRed:    "red",
		GemBlue:   "blue",
		GemGreen:  "green",
		GemYellow: "yellow",
		GemPurple: "purple",
		GemOrange: "orange",
		GemCyan:   "blue", // fallback
		GemPink:   "red",  // fallback
	}
	
	shapeMap := map[GemType]string{
		GemRed:    "diamond",
		GemBlue:   "polygon",
		GemGreen:  "rectangle",
		GemYellow: "square",
		GemPurple: "cube",
		GemOrange: "diamond",
		GemCyan:   "polygon",
		GemPink:   "rectangle",
	}
	
	loaded := 0
	for gemType, shape := range shapeMap {
		colorName := colorMap[gemType]
		
		// Try different filename patterns
		possibleNames := []string{
			fmt.Sprintf("element_%s_%s.png", colorName, shape),
			fmt.Sprintf("element_%s_%s_glossy.png", colorName, shape),
			fmt.Sprintf("gem%s.png", colorName),
			fmt.Sprintf("jewel%s.png", colorName),
		}
		
		for _, name := range possibleNames {
			path := filepath.Join(pack1Path, name)
			if img, err := loadPNG(path); err == nil {
				g.GemImages[gemType] = img
				loaded++
				break
			}
		}
	}
	
	// If no gem sprites found, create colored placeholders
	if loaded == 0 {
		log.Println("Warning: No gem sprites found, using colored placeholders")
		g.GemImages[GemRed] = createGemPlaceholder(64, 64, color.RGBA{255, 60, 60, 255})
		g.GemImages[GemBlue] = createGemPlaceholder(64, 64, color.RGBA{60, 120, 255, 255})
		g.GemImages[GemGreen] = createGemPlaceholder(64, 64, color.RGBA{60, 255, 60, 255})
		g.GemImages[GemYellow] = createGemPlaceholder(64, 64, color.RGBA{255, 255, 60, 255})
		g.GemImages[GemPurple] = createGemPlaceholder(64, 64, color.RGBA{180, 60, 255, 255})
		g.GemImages[GemOrange] = createGemPlaceholder(64, 64, color.RGBA{255, 160, 60, 255})
		g.GemImages[GemCyan] = createGemPlaceholder(64, 64, color.RGBA{60, 255, 255, 255})
		g.GemImages[GemPink] = createGemPlaceholder(64, 64, color.RGBA{255, 120, 180, 255})
	}
	
	// Try to load background
	bgPath := filepath.Join(spritesPath, "puzzle", "pack1", "PNG", "Default", "BackTile_01.png")
	if img, err := loadPNG(bgPath); err == nil {
		g.BackgroundImg = img
	}
}

func createGemPlaceholder(w, h int, fillColor color.RGBA) *ebiten.Image {
	// Draw diamond shape
	centerX := float64(w) / 2
	centerY := float64(h) / 2
	
	// Create diamond polygon
	diamond := ebiten.NewImage(w, h)
	
	// Fill with color
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := math.Abs(float64(x)-centerX) / centerX
			dy := math.Abs(float64(y)-centerY) / centerY
			if dx+dy <= 1 {
				diamond.Set(x, y, fillColor)
			}
		}
	}
	
	// Add border
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := math.Abs(float64(x)-centerX) / centerX
			dy := math.Abs(float64(y)-centerY) / centerY
			if dx+dy >= 0.9 && dx+dy <= 1.0 {
				diamond.Set(x, y, color.RGBA{255, 255, 255, 200})
			}
		}
	}
	
	// Add shine
	shineX := int(centerX * 0.7)
	shineY := int(centerY * 0.7)
	for dy := -3; dy <= 3; dy++ {
		for dx := -3; dx <= 3; dx++ {
			if dx*dx+dy*dy <= 9 {
				if shineX+dx >= 0 && shineX+dx < w && shineY+dy >= 0 && shineY+dy < h {
					c, _ := diamond.At(shineX+dx, shineY+dy).(color.RGBA)
					if c.A > 0 {
						diamond.Set(shineX+dx, shineY+dy, color.RGBA{255, 255, 255, 180})
					}
				}
			}
		}
	}
	
	return diamond
}

func (g *Game) initFonts() {
	// Try to load BoldPixels font
	execPath, _ := os.Getwd()
	fontPath := filepath.Join(execPath, "..", "..", "sprites", "08_Fonts", "webfontkit-boldpixels", "BoldPixels.ttf")
	
	fontData, err := os.ReadFile(fontPath)
	if err == nil {
		tt, parseErr := opentype.Parse(fontData)
		if parseErr == nil {
			g.Font, err = opentype.NewFace(tt, &opentype.FaceOptions{
				Size:    24,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if err != nil {
				log.Printf("Warning: Could not create font face: %v", err)
				g.Font = nil
			}
			
			g.SmallFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
				Size:    16,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if err != nil {
				g.SmallFont = nil
			}
		}
	}
	
	if g.Font == nil {
		// Fallback to default
		g.Font = nil
		g.SmallFont = nil
	}
}

func (g *Game) createButtons() {
	g.Buttons["play"] = &Button{
		X: ScreenWidth/2 - 100, Y: 450, W: 200, H: 60,
		Text: "▶ ИГРАТЬ",
		Action: func() {
			g.startGame()
		},
	}
	
	g.Buttons["exit"] = &Button{
		X: ScreenWidth/2 - 100, Y: 530, W: 200, H: 60,
		Text: "✕ ВЫЙТИ",
		Action: func() {
			os.Exit(0)
		},
	}
	
	g.Buttons["pause"] = &Button{
		X: ScreenWidth - 140, Y: 20, W: 120, H: 50,
		Text: "⏸ ПАУЗА",
		Action: func() {
			if g.State == StatePlaying {
				g.State = StatePlaying // Will toggle pause
			}
		},
	}
	
	g.Buttons["back"] = &Button{
		X: 20, Y: ScreenHeight - 80, W: 150, H: 50,
		Text: "← НАЗАД",
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
	g.TargetScore = 5000
	g.Particles = []*Particle{}
	g.FloatingTxt = []*FloatingText{}
	g.SelectedGem = nil
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.GameTime += 1.0 / TargetFPS
	
	// Update buttons hover state
	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)
	
	for _, btn := range g.Buttons {
		btn.Hover = fmx >= btn.X && fmx <= btn.X+btn.W && fmy >= btn.Y && fmy <= btn.Y+btn.H
	}
	
	// Handle button clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for _, btn := range g.Buttons {
			if btn.Hover && btn.Action != nil {
				btn.Action()
				return nil
			}
		}
	}
	
	// Update game state
	switch g.State {
	case StatePlaying:
		g.updatePlaying(fmx, fmy)
	case StateSwapping, StateRemoving, StateDropping:
		g.updateAnimationState()
	}
	
	// Update particles
	g.updateParticles()
	
	// Update floating texts
	g.updateFloatingTexts()
	
	return nil
}

func (g *Game) updatePlaying(mx, my float64) {
	if g.Board == nil {
		return
	}
	
	// Update gem animations
	g.updateGemAnimations()
	
	// Check if board is stable
	if g.Board.IsAnimating() {
		return
	}
	
	// Find matches
	matches, count := g.Board.FindMatches()
	if count > 0 {
		g.Combo++
		if g.Combo > g.MaxCombo {
			g.MaxCombo = g.Combo
		}
		
		points := count * 100 * (1 + g.Combo/2)
		g.Score += points
		
		// Mark gems for removal
		for _, match := range matches {
			for _, pos := range match {
				gem := g.Board.Get(pos[0], pos[1])
				if gem != nil {
					gem.Removing = true
					// Spawn particles
					g.spawnParticles(gem.X+float64(CellSize)/2, gem.Y+float64(CellSize)/2, gem.Type, 8)
				}
			}
		}
		
		// Add floating score
		if len(matches) > 0 {
			firstMatch := matches[0]
			if len(firstMatch) > 0 {
				pos := firstMatch[0]
				gem := g.Board.Get(pos[0], pos[1])
				if gem != nil {
					g.FloatingTxt = append(g.FloatingTxt, &FloatingText{
						X:     gem.X + float64(CellSize)/2,
						Y:     gem.Y,
						Text:  fmt.Sprintf("+%d", points),
						Life:  1.5,
						MaxLife: 1.5,
						Color: ColorTextGold,
						Size:  24,
					})
				}
			}
		}
		
		// Screen shake for combos
		if g.Combo >= 3 {
			g.ShakeDuration = 0.3
			g.ShakeIntensity = float64(g.Combo) * 2
		}
		
		g.State = StateRemoving
		g.StateTimer = 0.5
		return
	}
	
	// No matches, reset combo
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
					g.State = StateSwapping
					g.SwapGem1 = g.SelectedGem
					g.SwapGem2 = gem
					g.SwapGem1.Selected = false
					g.Board.Swap(g.SwapGem1.Col, g.SwapGem1.Row, g.SwapGem2.Col, g.SwapGem2.Row)
					g.Moves++
					g.SelectedGem = nil
				} else {
					g.SelectedGem.Selected = false
					g.SelectedGem = gem
					gem.Selected = true
				}
			}
		}
	}
}

func (g *Game) updateAnimationState() {
	g.StateTimer -= 1.0 / TargetFPS
	
	if g.State == StateSwapping {
		if !g.Board.IsAnimating() || g.StateTimer <= 0 {
			// Check if swap created matches
			_, count := g.Board.FindMatches()
			if count > 0 {
				g.State = StateRemoving
				g.StateTimer = 0.5
			} else {
				// Swap back
				g.Board.Swap(g.SwapGem1.Col, g.SwapGem1.Row, g.SwapGem2.Col, g.SwapGem2.Row)
				g.State = StatePlaying
			}
			g.SwapGem1 = nil
			g.SwapGem2 = nil
		}
	} else if g.State == StateRemoving {
		if g.StateTimer <= 0 {
			// Clear removed gems
			for c := 0; c < BoardCols; c++ {
				for r := 0; r < BoardRows; r++ {
					if g.Board.Grid[c][r] != nil && g.Board.Grid[c][r].Removing {
						g.Board.Grid[c][r] = nil
					}
				}
			}
			
			// Apply gravity
			if g.Board.ApplyGravity() {
				g.State = StateDropping
				g.StateTimer = 1.0
			} else {
				g.State = StatePlaying
			}
		}
	} else if g.State == StateDropping {
		if !g.Board.IsAnimating() || g.StateTimer <= 0 {
			g.State = StatePlaying
		}
	}
}

func (g *Game) updateGemAnimations() {
	if g.Board == nil {
		return
	}
	
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			gem := g.Board.Grid[c][r]
			if gem != nil {
				gem.X = lerp(gem.X, gem.TargetX, AnimSpeed)
				gem.Y = lerp(gem.Y, gem.TargetY, AnimSpeed)
				
				if gem.Removing {
					gem.Scale *= 0.85
					gem.Alpha *= 0.8
				}
			}
		}
	}
}

func (g *Game) updateParticles() {
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
}

func (g *Game) updateFloatingTexts() {
	for i := len(g.FloatingTxt) - 1; i >= 0; i-- {
		ft := g.FloatingTxt[i]
		ft.Y -= 1.5
		ft.Life -= 1.0 / TargetFPS
		
		if ft.Life <= 0 {
			g.FloatingTxt = append(g.FloatingTxt[:i], g.FloatingTxt[i+1:]...)
		}
	}
}

func (g *Game) spawnParticles(x, y float64, gemType GemType, count int) {
	colors := map[GemType]color.RGBA{
		GemRed:    {255, 60, 60, 255},
		GemBlue:   {60, 120, 255, 255},
		GemGreen:  {60, 255, 60, 255},
		GemYellow: {255, 255, 60, 255},
		GemPurple: {180, 60, 255, 255},
		GemOrange: {255, 160, 60, 255},
		GemCyan:   {60, 255, 255, 255},
		GemPink:   {255, 120, 180, 255},
	}
	
	col := colors[gemType]
	
	for i := 0; i < count; i++ {
		angle := float64(i)*6.2832/float64(count) + rand.Float64()*0.5
		speed := 2 + rand.Float64()*4
		
		g.Particles = append(g.Particles, &Particle{
			X:  x,
			Y:  y,
			VX: math.Cos(angle) * speed,
			VY: math.Sin(angle)*speed - 2,
			Life:  1.0,
			MaxLife: 1.0,
			Color: col,
			Size:  3 + rand.Float64()*4,
			Acceleration: 0.15,
		})
	}
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ColorBgDark)
	
	// Screen shake
	if g.ShakeDuration > 0 {
		g.ShakeTimer += 1.0 / TargetFPS
		if g.ShakeTimer <= g.ShakeDuration {
			g.ShakeX = (rand.Float64() - 0.5) * g.ShakeIntensity
			g.ShakeY = (rand.Float64() - 0.5) * g.ShakeIntensity
		} else {
			g.ShakeX = 0
			g.ShakeY = 0
			g.ShakeDuration = 0
		}
	}
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.ShakeX, g.ShakeY)
	
	switch g.State {
	case StateMenu:
		g.drawMenu(screen)
	default:
		g.drawGame(screen)
	}
	
	// Draw UI layer (not affected by shake)
	g.drawUI(screen)
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Draw starfield background
	g.drawStarfield(screen)
	
	// Title panel
	titlePanel := ebiten.NewImage(500, 100)
	titlePanel.Fill(ColorBgPanel)
	
	// Border
	border := ebiten.NewImage(500, 4)
	border.Fill(ColorBorder)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ScreenWidth/2-250), 250)
	screen.DrawImage(titlePanel, op)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(ScreenWidth/2-250), 248)
	screen.DrawImage(border, op2)
	
	// Title text
	title := "💎 CRYSTAL CASCADE 💎"
	if g.Font != nil {
		text.Draw(screen, title, g.Font, ScreenWidth/2-180, 315, ColorTextWhite)
	} else {
		ebitenutil.DebugPrintAt(screen, title, ScreenWidth/2-180, 310)
	}
	
	// Subtitle
	subtitle := "Go365 Challenge - Day 100"
	if g.SmallFont != nil {
		text.Draw(screen, subtitle, g.SmallFont, ScreenWidth/2-120, 340, ColorTextAccent)
	} else {
		ebitenutil.DebugPrintAt(screen, subtitle, ScreenWidth/2-120, 335)
	}
	
	// Draw buttons
	for name, btn := range g.Buttons {
		if name == "play" || name == "exit" {
			g.drawButton(screen, btn)
		}
	}
	
	// Footer
	footer := "Match 3+ crystals • Create combos • Beat the target score!"
	if g.SmallFont != nil {
		text.Draw(screen, footer, g.SmallFont, ScreenWidth/2-220, ScreenHeight-50, ColorTextAccent)
	} else {
		ebitenutil.DebugPrintAt(screen, footer, ScreenWidth/2-220, ScreenHeight-55)
	}
}

func (g *Game) drawStarfield(screen *ebiten.Image) {
	// Generate deterministic starfield
	rand.Seed(42)
	for i := 0; i < 100; i++ {
		x := rand.Float64() * ScreenWidth
		y := rand.Float64() * ScreenHeight
		size := rand.Float64()*2 + 1
		alpha := uint8(rand.Float64()*155 + 100)
		
		star := ebiten.NewImage(int(size), int(size))
		star.Fill(color.RGBA{255, 255, 255, alpha})
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		screen.DrawImage(star, op)
	}
	rand.Seed(time.Now().UnixNano())
}

func (g *Game) drawGame(screen *ebiten.Image) {
	if g.Board == nil {
		return
	}
	
	// Draw background
	if g.BackgroundImg != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(g.BackgroundImg, op)
	}
	
	// Draw board background
	boardW := BoardCols*(CellSize+CellPadding) + 16
	boardH := BoardRows*(CellSize+CellPadding) + 16
	
	boardBg := ebiten.NewImage(boardW, boardH)
	boardBg.Fill(ColorBgPanel)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(BoardOffsetX-8), float64(BoardOffsetY-8))
	screen.DrawImage(boardBg, op)
	
	// Border
	boardBorder := ebiten.NewImage(boardW, 4)
	boardBorder.Fill(ColorBorder)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(BoardOffsetX-8), float64(BoardOffsetY-10))
	screen.DrawImage(boardBorder, op2)
	
	// Draw cells
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			cellX := BoardOffsetX + c*(CellSize+CellPadding)
			cellY := BoardOffsetY + r*(CellSize+CellPadding)
			
			cell := ebiten.NewImage(CellSize, CellSize)
			cell.Fill(ColorCellBg)
			
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(cellX), float64(cellY))
			screen.DrawImage(cell, op)
		}
	}
	
	// Draw gems
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows; r++ {
			gem := g.Board.Grid[c][r]
			if gem != nil && gem.Alpha > 0.01 {
				g.drawGem(screen, gem)
			}
		}
	}
	
	// Draw particles
	for _, p := range g.Particles {
		size := int(p.Size * (p.Life / p.MaxLife))
		if size < 1 {
			continue
		}
		
		alpha := uint8((p.Life / p.MaxLife) * 255)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		
		particle := ebiten.NewImage(size, size)
		particle.Fill(c)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(size)/2, p.Y-float64(size)/2)
		screen.DrawImage(particle, op)
	}
	
	// Draw floating texts
	for _, ft := range g.FloatingTxt {
		alpha := uint8((ft.Life / ft.MaxLife) * 255)
		c := color.RGBA{ft.Color.R, ft.Color.G, ft.Color.B, alpha}
		
		if g.Font != nil {
			text.Draw(screen, ft.Text, g.Font, int(ft.X), int(ft.Y), c)
		} else {
			ebitenutil.DebugPrintAt(screen, ft.Text, int(ft.X), int(ft.Y))
		}
	}
}

func (g *Game) drawGem(screen *ebiten.Image, gem *Gem) {
	img := g.GemImages[gem.Type]
	if img == nil {
		return
	}
	
	op := &ebiten.DrawImageOptions{}
	
	// Scale
	scale := gem.Scale * float64(CellSize) / float64(img.Bounds().Dx())
	op.GeoM.Scale(scale, scale)
	
	// Translate
	op.GeoM.Translate(gem.X, gem.Y)
	
	// Alpha
	op.ColorM.Scale(1, 1, 1, gem.Alpha)
	
	screen.DrawImage(img, op)
	
	// Selection highlight
	if gem.Selected {
		highlight := ebiten.NewImage(CellSize+4, CellSize+4)
		highlight.Fill(ColorHighlight)
		
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(gem.X-2, gem.Y-2)
		screen.DrawImage(highlight, op2)
	}
	
	// Special gem indicators
	if gem.Special != SpecialNone {
		g.drawSpecialIndicator(screen, gem)
	}
}

func (g *Game) drawSpecialIndicator(screen *ebiten.Image, gem *Gem) {
	indicator := ebiten.NewImage(CellSize, CellSize)
	
	switch gem.Special {
	case SpecialBomb:
		// Bomb glow
		for i := 0; i < 3; i++ {
			size := CellSize - i*8
			if size > 0 {
				c := color.RGBA{255, 100, 50, uint8(100 - i*30)}
				inner := ebiten.NewImage(size, size)
				inner.Fill(c)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(i*4), float64(i*4))
				indicator.DrawImage(inner, op)
			}
		}
	case SpecialRainbow:
		// Rainbow shimmer
		shimmer := uint8(math.Sin(g.GameTime*5)*50 + 200)
		c := color.RGBA{shimmer, shimmer, 255, 150}
		inner := ebiten.NewImage(CellSize-8, CellSize-8)
		inner.Fill(c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(4, 4)
		indicator.Fill(c)
	case SpecialRow, SpecialCol:
		// Line indicator
		c := color.RGBA{100, 255, 100, 180}
		indicator.Fill(c)
	}
	
	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(1, 1, 1, 0.4)
	op.GeoM.Translate(gem.X, gem.Y)
	screen.DrawImage(indicator, op)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	if g.State == StatePlaying || g.State == StateSwapping || g.State == StateRemoving || g.State == StateDropping {
		g.drawGameUI(screen)
	}
}

func (g *Game) drawGameUI(screen *ebiten.Image) {
	// Score panel
	scorePanel := ebiten.NewImage(240, 180)
	scorePanel.Fill(ColorBgPanel)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(20, 40)
	screen.DrawImage(scorePanel, op)
	
	// Border
	scoreBorder := ebiten.NewImage(240, 3)
	scoreBorder.Fill(ColorBorder)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(20, 40)
	screen.DrawImage(scoreBorder, op2)
	
	// Score text
	scoreText := fmt.Sprintf("СЧЁТ: %d", g.Score)
	if g.Font != nil {
		text.Draw(screen, scoreText, g.Font, 40, 75, ColorTextWhite)
	} else {
		ebitenutil.DebugPrintAt(screen, scoreText, 40, 70)
	}
	
	// Target
	targetText := fmt.Sprintf("ЦЕЛЬ: %d", g.TargetScore)
	if g.SmallFont != nil {
		text.Draw(screen, targetText, g.SmallFont, 40, 100, ColorTextAccent)
	} else {
		ebitenutil.DebugPrintAt(screen, targetText, 40, 95)
	}
	
	// Combo
	if g.Combo > 1 {
		comboText := fmt.Sprintf("COMBO x%d", g.Combo)
		if g.SmallFont != nil {
			text.Draw(screen, comboText, g.SmallFont, 40, 125, ColorCombo)
		} else {
			ebitenutil.DebugPrintAt(screen, comboText, 40, 120)
		}
	}
	
	// Moves
	movesText := fmt.Sprintf("ХОДЫ: %d", g.Moves)
	if g.SmallFont != nil {
		text.Draw(screen, movesText, g.SmallFont, 40, 150, ColorTextAccent)
	} else {
		ebitenutil.DebugPrintAt(screen, movesText, 40, 145)
	}
	
	// Title bar at top
	titleBar := ebiten.NewImage(ScreenWidth, 60)
	titleBar.Fill(ColorBgPanel)
	
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(0, 0)
	screen.DrawImage(titleBar, op3)
	
	titleBorder := ebiten.NewImage(ScreenWidth, 3)
	titleBorder.Fill(ColorBorder)
	
	op4 := &ebiten.DrawImageOptions{}
	op4.GeoM.Translate(0, 60)
	screen.DrawImage(titleBorder, op4)
	
	// Title
	titleText := "💎 CRYSTAL CASCADE"
	if g.Font != nil {
		text.Draw(screen, titleText, g.Font, ScreenWidth/2-140, 40, ColorTextWhite)
	} else {
		ebitenutil.DebugPrintAt(screen, titleText, ScreenWidth/2-140, 35)
	}
	
	// Pause button
	if btn, ok := g.Buttons["pause"]; ok {
		g.drawButton(screen, btn)
	}
}

func (g *Game) drawButton(screen *ebiten.Image, btn *Button) {
	// Button background
	btnImg := ebiten.NewImage(int(btn.W), int(btn.H))
	
	if btn.Pressed {
		btnImg.Fill(color.RGBA{80, 100, 140, 255})
	} else if btn.Hover {
		btnImg.Fill(color.RGBA{60, 80, 120, 255})
	} else {
		btnImg.Fill(color.RGBA{40, 50, 80, 255})
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
	
	// Button text
	textX := int(btn.X + btn.W/2) - len(btn.Text)*4
	textY := int(btn.Y + btn.H/2 + 6)
	
	if g.SmallFont != nil {
		text.Draw(screen, btn.Text, g.SmallFont, textX, textY, ColorTextWhite)
	} else {
		ebitenutil.DebugPrintAt(screen, btn.Text, int(btn.X+10), int(btn.Y+btn.H/2-8))
	}
}

// ============================================================================
// EBITENGINE ENTRY POINT
// ============================================================================

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Crystal Cascade - Go365 Day 100")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
