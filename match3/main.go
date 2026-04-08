// Три в ряд — Match-3 Puzzle
// Go365 Challenge — Day 102
// Современный дизайн с процедурной графикой
// 8 апреля 2026

package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
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
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ============================================================================
// КОНСТАНТЫ
// ============================================================================

const (
	ScreenW      = 500
	ScreenH      = 800
	Cols         = 6
	Rows         = 6
	CellSize     = 64
	CellGap      = 4
	BoardX       = (ScreenW - Cols*(CellSize+CellGap)) / 2
	BoardY       = 200
	AnimDuration = 0.2
)

//go:embed assets/gems/*.png
var gemFS embed.FS

// ============================================================================
// ЦВЕТА
// ============================================================================

var (
	bgTop        = color.RGBA{25, 20, 60, 255}
	bgBottom     = color.RGBA{15, 12, 40, 255}
	boardBg      = color.RGBA{20, 18, 50, 200}
	cellBg       = color.RGBA{35, 30, 70, 180}
	cellBorder   = color.RGBA{60, 55, 110, 120}
	panelBg      = color.RGBA{20, 18, 55, 220}
	panelBorder  = color.RGBA{80, 160, 255, 255}
	textWhite    = color.RGBA{240, 240, 255, 255}
	textGold     = color.RGBA{255, 210, 60, 255}
	textAccent   = color.RGBA{120, 200, 255, 255}
	highlight    = color.RGBA{255, 255, 120, 180}
	comboColor   = color.RGBA{255, 160, 40, 255}
)

var gemPalettes = []struct {
	name  string
	base, light, dark color.RGBA
}{
	{"peach", color.RGBA{255, 150, 80, 255}, color.RGBA{255, 200, 140, 255}, color.RGBA{200, 110, 50, 255}},
	{"blueberry", color.RGBA{60, 80, 200, 255}, color.RGBA{120, 140, 240, 255}, color.RGBA{30, 50, 150, 255}},
	{"plum", color.RGBA{160, 50, 180, 255}, color.RGBA{200, 120, 220, 255}, color.RGBA{110, 30, 130, 255}},
	{"raspberry", color.RGBA{220, 50, 80, 255}, color.RGBA{255, 120, 140, 255}, color.RGBA{170, 30, 50, 255}},
	{"grape", color.RGBA{120, 60, 200, 255}, color.RGBA{170, 120, 240, 255}, color.RGBA{80, 30, 150, 255}},
	{"kiwi", color.RGBA{80, 180, 50, 255}, color.RGBA{140, 220, 100, 255}, color.RGBA{50, 130, 30, 255}},
	{"cherry", color.RGBA{220, 30, 50, 255}, color.RGBA{255, 100, 100, 255}, color.RGBA{170, 20, 30, 255}},
	{"coconut", color.RGBA{200, 180, 140, 255}, color.RGBA{240, 220, 190, 255}, color.RGBA{150, 130, 100, 255}},
}

// ============================================================================
// ТИПЫ
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
	Type    GemType
	Special SpecialType
	Col     int
	Row     int
	X       float64
	Y       float64
	TargetX float64
	TargetY float64
	Alpha   float64
	Scale   float64
	Selected bool
	Removing bool
}

type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife float64
	Color        color.RGBA
	Size         float64
}

type FloatingText struct {
	X, Y    float64
	Text    string
	Life    float64
	MaxLife float64
	Color   color.RGBA
}

type Star struct {
	X, Y, Size float64
	Alpha      float64
	Speed      float64
}

type Board struct {
	Grid [][]*Gem
}

type Game struct {
	State      GameState
	Board      *Board
	GemImages  map[GemType]*ebiten.Image
	Particles  []*Particle
	FloatingTxt []*FloatingText
	Stars      []Star
	Score      int
	Combo      int
	MaxCombo   int
	Moves      int
	Level      int
	TargetScore int
	Selected   *Gem
	GameTime   float64
	StateTimer float64
	ShakeTimer float64
	ShakeX, ShakeY float64
	BestScore  int
}

// ============================================================================
// УТИЛИТЫ
// ============================================================================

func lerp(a, b, t float64) float64 {
	return a + (b - a) * t
}

func loadFoodSprite(foodName string, size int) *ebiten.Image {
	// Try embedded FS first
	data, err := gemFS.ReadFile("assets/gems/" + foodName + ".png")
	if err == nil {
		return decodeAndScale(data, size)
	}

	// Try filesystem
	execPath, _ := os.Getwd()
	path := filepath.Join(execPath, "assets", "gems", foodName+".png")
	data, err = os.ReadFile(path)
	if err != nil {
		return nil
	}
	return decodeAndScale(data, size)
}

func decodeAndScale(data []byte, targetSize int) *ebiten.Image {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()
	if srcW == 0 || srcH == 0 {
		return nil
	}
	scale := float64(targetSize) / float64(srcW)
	if float64(srcH)*scale > float64(targetSize) {
		scale = float64(targetSize) / float64(srcH)
	}
	resized := ebiten.NewImage(int(float64(srcW)*scale), int(float64(srcH)*scale))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	resized.DrawImage(ebiten.NewImageFromImage(img), op)
	return resized
}

// ============================================================================
// BOARD
// ============================================================================

func NewBoard() *Board {
	b := &Board{Grid: make([][]*Gem, Cols)}
	for c := 0; c < Cols; c++ {
		b.Grid[c] = make([]*Gem, Rows)
	}

	for c := 0; c < Cols; c++ {
		for r := 0; r < Rows; r++ {
			b.Grid[c][r] = b.createGemSafe(c, r)
		}
	}

	return b
}

func (b *Board) createGemSafe(c, r int) *Gem {
	var t GemType
	for i := 0; i < 50; i++ {
		t = GemType(rand.Intn(int(GemCount)))
		if !b.wouldMatch(c, r, t) {
			break
		}
	}

	x := float64(BoardX + c*(CellSize+CellGap))
	y := float64(BoardY + r*(CellSize+CellGap))

	return &Gem{
		Type: t, Col: c, Row: r,
		X: x, Y: y, TargetX: x, TargetY: y,
		Alpha: 1, Scale: 1,
	}
}

func (b *Board) wouldMatch(c, r int, t GemType) bool {
	// Horizontal
	count := 1
	for i := c - 1; i >= 0; i-- {
		g := b.Grid[i][r]
		if g == nil || g.Type != t { break }
		count++
	}
	for i := c + 1; i < Cols; i++ {
		g := b.Grid[i][r]
		if g == nil || g.Type != t { break }
		count++
	}
	if count >= 3 { return true }

	// Vertical
	count = 1
	for i := r - 1; i >= 0; i-- {
		g := b.Grid[c][i]
		if g == nil || g.Type != t { break }
		count++
	}
	for i := r + 1; i < Rows; i++ {
		g := b.Grid[c][i]
		if g == nil || g.Type != t { break }
		count++
	}
	return count >= 3
}

func (b *Board) Swap(c1, r1, c2, r2 int) bool {
	dc := c1 - c2
	dr := r1 - r2
	if !((dc == 1 || dc == -1) && dr == 0 || dc == 0 && (dr == 1 || dr == -1)) {
		return false
	}

	g1 := b.Grid[c1][r1]
	g2 := b.Grid[c2][r2]
	if g1 == nil || g2 == nil { return false }

	b.Grid[c1][r1] = g2
	b.Grid[c2][r2] = g1
	g1.Col, g1.Row = c2, r2
	g2.Col, g2.Row = c1, r1

	g1.TargetX = float64(BoardX + c2*(CellSize+CellGap))
	g1.TargetY = float64(BoardY + r2*(CellSize+CellGap))
	g2.TargetX = float64(BoardX + c1*(CellSize+CellGap))
	g2.TargetY = float64(BoardY + r1*(CellSize+CellGap))

	return true
}

func (b *Board) FindMatches() ([][][2]int, int) {
	matched := make(map[[2]int]bool)
	var result [][][2]int

	// Horizontal
	for r := 0; r < Rows; r++ {
		c := 0
		for c < Cols {
			g := b.Grid[c][r]
			if g == nil || g.Removing { c++; continue }

			match := [][2]int{{c, r}}
			for cc := c + 1; cc < Cols; cc++ {
				n := b.Grid[cc][r]
				if n != nil && !n.Removing && n.Type == g.Type {
					match = append(match, [2]int{cc, r})
				} else { break }
			}

			if len(match) >= 3 {
				// Special creation
				if len(match) == 4 {
					bc, br := match[1][0], match[1][1]
					if b.Grid[bc][br] != nil { b.Grid[bc][br].Special = SpecialBomb }
				} else if len(match) >= 5 {
					bc, br := match[2][0], match[2][1]
					if b.Grid[bc][br] != nil { b.Grid[bc][br].Special = SpecialRainbow }
				}

				for _, p := range match { matched[p] = true }
				result = append(result, match)
				c += len(match)
			} else { c++ }
		}
	}

	// Vertical
	for c := 0; c < Cols; c++ {
		r := 0
		for r < Rows {
			g := b.Grid[c][r]
			if g == nil || g.Removing { r++; continue }

			match := [][2]int{{c, r}}
			for rr := r + 1; rr < Rows; rr++ {
				n := b.Grid[c][rr]
				if n != nil && !n.Removing && n.Type == g.Type {
					match = append(match, [2]int{c, rr})
				} else { break }
			}

			if len(match) >= 3 {
				if len(match) == 4 {
					bc, br := match[1][0], match[1][1]
					if b.Grid[bc][br] != nil { b.Grid[bc][br].Special = SpecialCol }
				} else if len(match) >= 5 {
					bc, br := match[2][0], match[2][1]
					if b.Grid[bc][br] != nil { b.Grid[bc][br].Special = SpecialRainbow }
				}

				for _, p := range match { matched[p] = true }
				result = append(result, match)
				r += len(match)
			} else { r++ }
		}
	}

	return result, len(matched)
}

func (b *Board) ApplyGravity() bool {
	dropped := false

	for c := 0; c < Cols; c++ {
		// Remove marked gems
		for r := 0; r < Rows; r++ {
			if b.Grid[c][r] != nil && b.Grid[c][r].Removing {
				b.Grid[c][r] = nil
			}
		}

		// Drop down
		writePos := Rows - 1
		for r := Rows - 1; r >= 0; r-- {
			if b.Grid[c][r] != nil {
				if r != writePos {
					g := b.Grid[c][r]
					b.Grid[c][writePos] = g
					b.Grid[c][r] = nil
					g.Row = writePos
					g.TargetY = float64(BoardY + writePos*(CellSize+CellGap))
					dropped = true
				}
				writePos--
			}
		}

		// Fill from top
		for r := writePos; r >= 0; r-- {
			t := GemType(rand.Intn(int(GemCount)))
			x := float64(BoardX + c*(CellSize+CellGap))
			y := float64(BoardY - float64(writePos-r+1)*(CellSize+CellGap))
			targetY := float64(BoardY + r*(CellSize+CellGap))

			b.Grid[c][r] = &Gem{
				Type: t, Col: c, Row: r,
				X: x, Y: y, TargetX: x, TargetY: targetY,
				Alpha: 1, Scale: 1,
			}
			dropped = true
		}
	}

	return dropped
}

func (b *Board) IsAnimating() bool {
	for c := 0; c < Cols; c++ {
		for r := 0; r < Rows; r++ {
			g := b.Grid[c][r]
			if g == nil { continue }
			dx := g.TargetX - g.X
			dy := g.TargetY - g.Y
			if dx*dx+dy*dy > 0.5 { return true }
			if g.Removing && g.Alpha > 0.01 { return true }
		}
	}
	return false
}

func (b *Board) GemAt(mx, my float64) *Gem {
	for c := 0; c < Cols; c++ {
		for r := 0; r < Rows; r++ {
			g := b.Grid[c][r]
			if g != nil && !g.Removing {
				h := float64(CellSize) / 2
				gx := g.X + h
				gy := g.Y + h
				if mx >= gx-h && mx <= gx+h && my >= gy-h && my <= gy+h {
					return g
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
		GemImages:   make(map[GemType]*ebiten.Image),
		Particles:   []*Particle{},
		FloatingTxt: []*FloatingText{},
		Level:       1,
		TargetScore: 5000,
	}

	// Create gem sprites from food images
	for i := 0; i < int(GemCount); i++ {
		sprite := loadFoodSprite(gemPalettes[i].name, CellSize)
		if sprite == nil {
			// Fallback: create colored placeholder
			sprite = createPlaceholder(gemPalettes[i].light, CellSize)
		}
		g.GemImages[GemType(i)] = sprite
	}

	// Stars
	for i := 0; i < 80; i++ {
		g.Stars = append(g.Stars, Star{
			X: rand.Float64() * ScreenW,
			Y: rand.Float64() * ScreenH,
			Size: rand.Float64()*2 + 0.5,
			Alpha: rand.Float64(),
			Speed: rand.Float64()*2 + 0.5,
		})
	}

	// Load best score
	data, _ := os.ReadFile("bestscore.txt")
	fmt.Sscanf(string(data), "%d", &g.BestScore)

	return g
}

func (g *Game) Update() error {
	g.GameTime += 1.0 / 60.0

	// Shake
	if g.ShakeTimer > 0 {
		g.ShakeTimer -= 1.0 / 60.0
		g.ShakeX = (rand.Float64() - 0.5) * 5
		g.ShakeY = (rand.Float64() - 0.5) * 5
	} else {
		g.ShakeX = 0
		g.ShakeY = 0
	}

	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)

	// Stars twinkle
	for i := range g.Stars {
		g.Stars[i].Alpha = 0.4 + 0.6*math.Sin(g.GameTime*g.Stars[i].Speed)
	}

	// Particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := g.Particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.12
		p.Life -= 1.0 / 60.0
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}

	// Floating text
	for i := len(g.FloatingTxt) - 1; i >= 0; i-- {
		ft := g.FloatingTxt[i]
		ft.Y -= 1.2
		ft.Life -= 1.0 / 60.0
		if ft.Life <= 0 {
			g.FloatingTxt = append(g.FloatingTxt[:i], g.FloatingTxt[i+1:]...)
		}
	}

	// Gem animations
	if g.Board != nil {
		for c := 0; c < Cols; c++ {
			for r := 0; r < Rows; r++ {
				g := g.Board.Grid[c][r]
				if g != nil {
					g.X = lerp(g.X, g.TargetX, 0.2)
					g.Y = lerp(g.Y, g.TargetY, 0.2)
					if g.Removing {
						g.Scale *= 0.82
						g.Alpha *= 0.78
					}
				}
			}
		}
	}

	// Click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		switch g.State {
		case StateMenu:
			if fmx >= ScreenW/2-100 && fmx <= ScreenW/2+100 && fmy >= 500 && fmy <= 560 {
				g.startGame()
			}

		case StatePlaying:
			if g.Board != nil && !g.Board.IsAnimating() {
				gem := g.Board.GemAt(fmx, fmy)
				if gem != nil {
					if g.Selected == nil {
						g.Selected = gem
						gem.Selected = true
					} else if g.Selected == gem {
						g.Selected.Selected = false
						g.Selected = nil
					} else {
						dc := gem.Col - g.Selected.Col
						dr := gem.Row - g.Selected.Row
						if (dc == 1 || dc == -1) && dr == 0 || dc == 0 && (dr == 1 || dr == -1) {
							g.Board.Swap(gem.Col, gem.Row, g.Selected.Col, g.Selected.Row)
							g.Moves++
							g.Selected.Selected = false
							g.Selected = nil
							g.State = StateSwapping
							g.StateTimer = 0.5
						} else {
							g.Selected.Selected = false
							g.Selected = gem
							gem.Selected = true
						}
					}
				}
			}
		}
	}

	// State machine
	if g.State == StateSwapping {
		g.StateTimer -= 1.0 / 60.0
		if !g.Board.IsAnimating() || g.StateTimer <= 0 {
			_, count := g.Board.FindMatches()
			if count > 0 {
				g.State = StateRemoving
				g.StateTimer = 0.4
			} else {
				// Swap back
				for c := 0; c < Cols; c++ {
					for r := 0; r < Rows; r++ {
						gm := g.Board.Grid[c][r]
						if gm != nil && (gm.X != gm.TargetX || gm.Y != gm.TargetY) {
							// animate back handled by lerp
						}
					}
				}
				g.State = StatePlaying
			}
		}
	}

	if g.State == StateRemoving {
		g.StateTimer -= 1.0 / 60.0
		if g.StateTimer <= 0 {
			if g.Board.ApplyGravity() {
				g.State = StateDropping
				g.StateTimer = 0.8
			} else {
				g.State = StatePlaying
			}
		}
	}

	if g.State == StateDropping {
		g.StateTimer -= 1.0 / 60.0
		if !g.Board.IsAnimating() || g.StateTimer <= 0 {
			g.State = StatePlaying
		}
	}

	// Check matches in playing state
	if g.State == StatePlaying && g.Board != nil && !g.Board.IsAnimating() {
		matches, count := g.Board.FindMatches()
		if count > 0 {
			g.Combo++
			if g.Combo > g.MaxCombo { g.MaxCombo = g.Combo }

			points := count * 100 * (1 + g.Combo/2)
			g.Score += points

			for _, match := range matches {
				for _, pos := range match {
					gm := g.Board.Grid[pos[0]][pos[1]]
					if gm != nil {
						gm.Removing = true
						g.spawnParticles(gm.X+float64(CellSize)/2, gm.Y+float64(CellSize)/2, gm.Type, 8)
					}
				}
			}

			if len(matches) > 0 {
				pos := matches[0][0]
				gm := g.Board.Grid[pos[0]][pos[1]]
				if gm != nil {
					g.FloatingTxt = append(g.FloatingTxt, &FloatingText{
						X: gm.X + float64(CellSize)/2, Y: gm.Y,
						Text: fmt.Sprintf("+%d", points),
						Life: 1.5, MaxLife: 1.5,
						Color: textGold,
					})
				}
			}

			if g.Combo >= 3 {
				g.ShakeTimer = 0.25
			}

			g.State = StateRemoving
			g.StateTimer = 0.4
		} else {
			g.Combo = 0
		}
	}

	return nil
}

func createPlaceholder(c color.RGBA, size int) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := size / 2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			if dx*dx+dy*dy <= float64(center*center) {
				img.Set(x, y, c)
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

func (g *Game) startGame() {
	g.State = StatePlaying
	g.Board = NewBoard()
	g.Score = 0
	g.Combo = 0
	g.MaxCombo = 0
	g.Moves = 0
	g.Particles = []*Particle{}
	g.FloatingTxt = []*FloatingText{}
	g.Selected = nil
}

func (g *Game) spawnParticles(x, y float64, t GemType, count int) {
	pal := gemPalettes[t]
	for i := 0; i < count; i++ {
		a := float64(i)*6.2832/float64(count) + rand.Float64()*0.5
		s := 2 + rand.Float64()*3
		g.Particles = append(g.Particles, &Particle{
			X: x, Y: y,
			VX: math.Cos(a) * s, VY: math.Sin(a)*s - 1.5,
			Life: 0.8 + rand.Float64()*0.4, MaxLife: 1.2,
			Color: pal.light,
			Size: 3 + rand.Float64()*3,
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Gradient background
	for y := 0; y < ScreenH; y++ {
		t := float64(y) / float64(ScreenH)
		r := uint8(lerp(float64(bgTop.R), float64(bgBottom.R), t))
		gr := uint8(lerp(float64(bgTop.G), float64(bgBottom.G), t))
		b := uint8(lerp(float64(bgTop.B), float64(bgBottom.B), t))
		vector.DrawFilledRect(screen, 0, float32(y), ScreenW, 1, color.RGBA{r, gr, b, 255}, false)
	}

	// Stars
	for _, s := range g.Stars {
		alpha := uint8(s.Alpha * 180)
		sz := int(s.Size)
		if sz < 1 { sz = 1 }
		img := ebiten.NewImage(sz*2, sz*2)
		img.Fill(color.RGBA{255, 255, 255, alpha})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(s.X, s.Y)
		screen.DrawImage(img, op)
	}

	if g.State == StateMenu {
		g.drawMenu(screen)
	} else {
		g.drawBoard(screen)
		g.drawHUD(screen)
	}

	// Particles
	for _, p := range g.Particles {
		sz := int(p.Size * (p.Life / p.MaxLife))
		if sz < 1 { continue }
		alpha := uint8((p.Life / p.MaxLife) * 255)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		img := ebiten.NewImage(sz, sz)
		img.Fill(c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(sz)/2, p.Y-float64(sz)/2)
		screen.DrawImage(img, op)
	}

	// Floating text
	for _, ft := range g.FloatingTxt {
		alpha := uint8((ft.Life / ft.MaxLife) * 255)
		c := color.RGBA{ft.Color.R, ft.Color.G, ft.Color.B, alpha}
		ebitenutil.DebugPrintAt(screen, ft.Text, int(ft.X)-20, int(ft.Y))
		_ = c
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Title panel
	panel := ebiten.NewImage(380, 100)
	panel.Fill(panelBg)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ScreenW/2-190, 250)
	screen.DrawImage(panel, op)

	border := ebiten.NewImage(380, 4)
	border.Fill(panelBorder)
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(ScreenW/2-190, 248)
	screen.DrawImage(border, op2)

	ebitenutil.DebugPrintAt(screen, "💎  ТРИ В РЯД  💎", ScreenW/2-150, 285)
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge — Day 102", ScreenW/2-130, 330)

	// Decorative gems
	for i := 0; i < 5; i++ {
		gem := loadFoodSprite(gemPalettes[i].name, 40)
		if gem == nil {
			gem = createPlaceholder(gemPalettes[i].light, 40)
		}
		op3 := &ebiten.DrawImageOptions{}
		op3.GeoM.Translate(float64(ScreenW/2-120+i*60), 370)
		screen.DrawImage(gem, op3)
	}

	// Play button
	g.drawButton(screen, "▶  ИГРАТЬ", ScreenW/2-100, 500, 200, 60)

	// Info
	ebitenutil.DebugPrintAt(screen, "Соединяй 3+ кристалла", ScreenW/2-110, 600)
	ebitenutil.DebugPrintAt(screen, "Собирай комбо!", ScreenW/2-80, 625)
	if g.BestScore > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Лучший: %d", g.BestScore), ScreenW/2-75, 660)
	}
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	if g.Board == nil { return }

	// Board background
	bw := Cols*(CellSize+CellGap) + 12
	bh := Rows*(CellSize+CellGap) + 12
	bg := ebiten.NewImage(bw, bh)
	bg.Fill(boardBg)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(BoardX-6)+g.ShakeX, float64(BoardY-6)+g.ShakeY)
	screen.DrawImage(bg, op)

	// Border
	brd := ebiten.NewImage(bw, 4)
	brd.Fill(panelBorder)
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(BoardX-6)+g.ShakeX, float64(BoardY-8)+g.ShakeY)
	screen.DrawImage(brd, op2)

	// Cells
	for c := 0; c < Cols; c++ {
		for r := 0; r < Rows; r++ {
			cell := ebiten.NewImage(CellSize, CellSize)
			cell.Fill(cellBg)
			op3 := &ebiten.DrawImageOptions{}
			op3.GeoM.Translate(float64(BoardX+c*(CellSize+CellGap))+g.ShakeX, float64(BoardY+r*(CellSize+CellGap))+g.ShakeY)
			screen.DrawImage(cell, op3)
		}
	}

	// Gems
	for c := 0; c < Cols; c++ {
		for r := 0; r < Rows; r++ {
			gem := g.Board.Grid[c][r]
			if gem == nil || gem.Alpha < 0.01 { continue }

			img := g.GemImages[gem.Type]
			if img == nil { continue }

			op4 := &ebiten.DrawImageOptions{}
			op4.GeoM.Scale(gem.Scale, gem.Scale)
			op4.GeoM.Translate(gem.X+g.ShakeX, gem.Y+g.ShakeY)
			op4.ColorM.Scale(1, 1, 1, gem.Alpha)
			screen.DrawImage(img, op4)

			// Selected highlight
			if gem.Selected {
				hl := ebiten.NewImage(CellSize+6, CellSize+6)
				hl.Fill(highlight)
				op5 := &ebiten.DrawImageOptions{}
				op5.GeoM.Translate(gem.X-3+g.ShakeX, gem.Y-3+g.ShakeY)
				op5.ColorM.Scale(1, 1, 1, 0.3)
				screen.DrawImage(hl, op5)
			}

			// Special indicator
			if gem.Special != SpecialNone {
				sz := CellSize
				ind := ebiten.NewImage(sz, sz)
				var c color.RGBA
				switch gem.Special {
				case SpecialBomb:
					c = color.RGBA{255, 100, 50, 120}
				case SpecialRainbow:
					shimmer := uint8(160 + 80*math.Sin(g.GameTime*5))
					c = color.RGBA{shimmer, shimmer, 255, 120}
				default:
					c = color.RGBA{100, 255, 100, 120}
				}
				vector.DrawFilledRect(ind, 0, 0, float32(sz), float32(sz), c, false)
				op6 := &ebiten.DrawImageOptions{}
				op6.GeoM.Translate(gem.X+g.ShakeX, gem.Y+g.ShakeY)
				op6.ColorM.Scale(1, 1, 1, 0.35)
				screen.DrawImage(ind, op6)
			}
		}
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 60, panelBg, false)
	vector.StrokeLine(screen, 0, 60, ScreenW, 60, 2, panelBorder, false)

	ebitenutil.DebugPrintAt(screen, "💎 ТРИ В РЯД", ScreenW/2-90, 18)

	// Score panel
	sp := ebiten.NewImage(200, 140)
	sp.Fill(panelBg)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(15, 80)
	screen.DrawImage(sp, op)

	spBorder := ebiten.NewImage(200, 3)
	spBorder.Fill(panelBorder)
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(15, 80)
	screen.DrawImage(spBorder, op2)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), 30, 98)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ЦЕЛЬ: %d", g.TargetScore), 30, 122)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ХОДЫ: %d", g.Moves), 30, 146)
	if g.Combo > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("COMBO x%d", g.Combo), 30, 170)
	}

	// Level info
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Уровень: %d", g.Level), ScreenW-130, 80)

	if g.Score >= g.TargetScore {
		g.Level++
		g.TargetScore = int(float64(g.TargetScore) * 1.5)
	}
}

func (g *Game) drawButton(screen *ebiten.Image, text string, x, y, w, h int) {
	btn := ebiten.NewImage(w, h)
	mx, my := ebiten.CursorPosition()
	hover := float64(mx) >= float64(x) && float64(mx) <= float64(x+w) &&
		float64(my) >= float64(y) && float64(my) <= float64(y+h)

	if hover {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{50, 70, 120, 255}, false)
	} else {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{30, 40, 80, 255}, false)
	}

	border := ebiten.NewImage(w, 3)
	border.Fill(panelBorder)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(border, op2)

	ebitenutil.DebugPrintAt(screen, text, x+25, y+h/2-10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Три в ряд — Go365 Day 102")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
