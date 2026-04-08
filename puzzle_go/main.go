// Крестики-Нолики - Go365 Challenge Day 100
// Современная игра с красивыми эффектами и AI
// 8 апреля 2026

package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
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
	ScreenW = 800
	ScreenH = 900
	GridSize = 3
	CellSize = 160
	GridX = (ScreenW - GridSize*CellSize) / 2
	GridY = 200
	LineWidth = 6
)

// ============================================================================
// ТИПЫ
// ============================================================================

type Cell int

const (
	CellEmpty Cell = iota
	CellX
	CellO
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateWin
	StateDraw
)

type Particle struct {
	X, Y       float64
	VX, VY     float64
	Life       float64
	MaxLife    float64
	Color      color.RGBA
	Size       float64
}

type Star struct {
	X, Y  float64
	Size  float64
	Alpha float64
	TwinkleSpeed float64
}

// ============================================================================
// УТИЛИТЫ
// ============================================================================

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func createXImage(size int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	half := size / 2
	thickness := 12
	
	// Draw X shape
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - half
			dy := y - half
			
			// Diagonal 1: y = x
			dist1 := math.Abs(float64(dx - dy))
			// Diagonal 2: y = -x
			dist2 := math.Abs(float64(dx + dy))
			
			if dist1 < float64(thickness) || dist2 < float64(thickness) {
				// Glow effect
				dist := math.Min(dist1, dist2)
				alpha := 1.0 - dist/float64(thickness)
				img.Set(x, y, color.RGBA{
					R: c.R,
					G: c.G,
					B: c.B,
					A: uint8(float64(c.A) * alpha),
				})
			}
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

func createOImage(size int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	center := size / 2
	radius := size/2 - 20
	thickness := 12
	
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			dist := math.Sqrt(dx*dx + dy*dy)
			
			if math.Abs(dist-float64(radius)) < float64(thickness) {
				alpha := 1.0 - math.Abs(dist-float64(radius))/float64(thickness)
				img.Set(x, y, color.RGBA{
					R: c.R,
					G: c.G,
					B: c.B,
					A: uint8(float64(c.A) * alpha),
				})
			}
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

func createLineImage(w, h int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Edge glow
			edgeDist := math.Min(float64(x), float64(w-1-x))
			alpha := 0.6 + 0.4*(edgeDist/float64(w/2))
			img.Set(x, y, color.RGBA{
				R: c.R,
				G: c.G,
				B: c.B,
				A: uint8(float64(c.A) * alpha),
			})
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// ============================================================================
// ИГРОВОЕ ПОЛЕ
// ============================================================================

type Board struct {
	Grid [GridSize][GridSize]Cell
}

func NewBoard() *Board {
	return &Board{}
}

func (b *Board) Reset() {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			b.Grid[r][c] = CellEmpty
		}
	}
}

func (b *Board) IsFull() bool {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if b.Grid[r][c] == CellEmpty {
				return false
			}
		}
	}
	return true
}

func (b *Board) CheckWin(player Cell) (bool, [][2]int) {
	// Rows
	for r := 0; r < GridSize; r++ {
		if b.Grid[r][0] == player && b.Grid[r][1] == player && b.Grid[r][2] == player {
			return true, [][2]int{{r, 0}, {r, 1}, {r, 2}}
		}
	}
	
	// Cols
	for c := 0; c < GridSize; c++ {
		if b.Grid[0][c] == player && b.Grid[1][c] == player && b.Grid[2][c] == player {
			return true, [][2]int{{0, c}, {1, c}, {2, c}}
		}
	}
	
	// Diagonals
	if b.Grid[0][0] == player && b.Grid[1][1] == player && b.Grid[2][2] == player {
		return true, [][2]int{{0, 0}, {1, 1}, {2, 2}}
	}
	if b.Grid[0][2] == player && b.Grid[1][1] == player && b.Grid[2][0] == player {
		return true, [][2]int{{0, 2}, {1, 1}, {2, 0}}
	}
	
	return false, nil
}

func (b *Board) CellAt(mx, my float64) (int, int, bool) {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			x := GridX + c*CellSize
			y := GridY + r*CellSize
			
			if mx >= float64(x)+8 && mx <= float64(x+CellSize-8) &&
			   my >= float64(y)+8 && my <= float64(y+CellSize-8) {
				return r, c, true
			}
		}
	}
	return -1, -1, false
}

// AI - Minimax algorithm
func (b *Board) BestMove() (int, int) {
	bestScore := -999
	bestR, bestC := -1, -1
	
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if b.Grid[r][c] == CellEmpty {
				b.Grid[r][c] = CellO
				score := b.minimax(0, false)
				b.Grid[r][c] = CellEmpty
				
				if score > bestScore {
					bestScore = score
					bestR, bestC = r, c
				}
			}
		}
	}
	
	return bestR, bestC
}

func (b *Board) minimax(depth int, isMaximizing bool) int {
	if win, _ := b.CheckWin(CellO); win {
		return 10 - depth
	}
	if win, _ := b.CheckWin(CellX); win {
		return depth - 10
	}
	if b.IsFull() {
		return 0
	}
	
	if isMaximizing {
		best := -999
		for r := 0; r < GridSize; r++ {
			for c := 0; c < GridSize; c++ {
				if b.Grid[r][c] == CellEmpty {
					b.Grid[r][c] = CellO
					best = int(math.Max(float64(best), float64(b.minimax(depth+1, false))))
					b.Grid[r][c] = CellEmpty
				}
			}
		}
		return best
	} else {
		best := 999
		for r := 0; r < GridSize; r++ {
			for c := 0; c < GridSize; c++ {
				if b.Grid[r][c] == CellEmpty {
					b.Grid[r][c] = CellX
					best = int(math.Min(float64(best), float64(b.minimax(depth+1, true))))
					b.Grid[r][c] = CellEmpty
				}
			}
		}
		return best
	}
}

// ============================================================================
// GAME
// ============================================================================

type Game struct {
	State      GameState
	Board      *Board
	Turn       Cell // CellX or CellO
	
	ScoreX     int
	ScoreO     int
	Draws      int
	
	Particles  []Particle
	Stars      []Star
	WinCells   [][2]int
	WinLine    []struct{ X1, Y1, X2, Y2 float64 }
	
	CellAnims  [GridSize][GridSize]float64 // Scale animation
	
	GameTime   float64
	HoverRow   int
	HoverCol   int
	
	AI_Thinking bool
	AITimer    float64
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	
	g := &Game{
		State:    StateMenu,
		Board:    NewBoard(),
		Turn:     CellX,
		Particles: []Particle{},
		HoverRow: -1,
		HoverCol: -1,
	}
	
	// Init stars
	for i := 0; i < 60; i++ {
		g.Stars = append(g.Stars, Star{
			X:            rand.Float64() * ScreenW,
			Y:            rand.Float64() * ScreenH,
			Size:         rand.Float64()*2 + 0.5,
			Alpha:        rand.Float64(),
			TwinkleSpeed: rand.Float64()*2 + 1,
		})
	}
	
	return g
}

func (g *Game) Update() error {
	g.GameTime += 1.0 / 60.0
	
	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)
	
	// Update hover
	if g.State == StatePlaying {
		r, c, ok := g.Board.CellAt(fmx, fmy)
		if ok {
			g.HoverRow, g.HoverCol = r, c
		} else {
			g.HoverRow, g.HoverCol = -1, -1
		}
	}
	
	// Update particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := &g.Particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.15
		p.Life -= 1.0 / 60.0
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}
	
	// Update stars
	for i := range g.Stars {
		g.Stars[i].Alpha = 0.5 + 0.5*math.Sin(g.GameTime*g.Stars[i].TwinkleSpeed)
	}
	
	// Update cell animations
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			target := 1.0
			if g.Board.Grid[r][c] == CellEmpty {
				target = 1.0
			}
			g.CellAnims[r][c] = lerp(g.CellAnims[r][c], target, 0.15)
		}
	}
	
	// Handle click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		switch g.State {
		case StateMenu:
			// Start button
			btnX := ScreenW/2 - 100
			btnY := 500
			if fmx >= float64(btnX) && fmx <= float64(btnX+200) &&
			   fmy >= float64(btnY) && fmy <= float64(btnY+60) {
				g.startGame()
			}
			
		case StatePlaying:
			if g.Turn == CellX && !g.AI_Thinking {
				r, c, ok := g.Board.CellAt(fmx, fmy)
				if ok && g.Board.Grid[r][c] == CellEmpty {
					g.Board.Grid[r][c] = CellX
					g.CellAnims[r][c] = 0.3
					g.spawnCellParticles(r, c, CellX)
					
					if win, cells := g.Board.CheckWin(CellX); win {
						g.State = StateWin
						g.ScoreX++
						g.WinCells = cells
						g.spawnWinParticles(cells)
					} else if g.Board.IsFull() {
						g.State = StateDraw
						g.Draws++
					} else {
						g.Turn = CellO
						g.AI_Thinking = true
						g.AITimer = 0.5
					}
				}
			}
			
		case StateWin, StateDraw:
			// Restart button
			btnX := ScreenW/2 - 100
			btnY := 650
			if fmx >= float64(btnX) && fmx <= float64(btnX+200) &&
			   fmy >= float64(btnY) && fmy <= float64(btnY+60) {
				g.startGame()
			}
			
			// Menu button
			btnY2 := 730
			if fmx >= float64(btnX) && fmx <= float64(btnX+200) &&
			   fmy >= float64(btnY2) && fmy <= float64(btnY2+60) {
				g.State = StateMenu
			}
		}
	}
	
	// AI turn
	if g.AI_Thinking {
		g.AITimer -= 1.0 / 60.0
		if g.AITimer <= 0 {
			r, c := g.Board.BestMove()
			if r >= 0 && c >= 0 {
				g.Board.Grid[r][c] = CellO
				g.CellAnims[r][c] = 0.3
				g.spawnCellParticles(r, c, CellO)
				
				if win, cells := g.Board.CheckWin(CellO); win {
					g.State = StateWin
					g.ScoreO++
					g.WinCells = cells
					g.spawnWinParticles(cells)
				} else if g.Board.IsFull() {
					g.State = StateDraw
					g.Draws++
				} else {
					g.Turn = CellX
				}
			}
			g.AI_Thinking = false
		}
	}
	
	return nil
}

func (g *Game) startGame() {
	g.State = StatePlaying
	g.Board.Reset()
	g.Turn = CellX
	g.Particles = []Particle{}
	g.WinCells = nil
	g.AI_Thinking = false
	
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			g.CellAnims[r][c] = 0
		}
	}
}

func (g *Game) spawnCellParticles(row, col int, player Cell) {
	c := color.RGBA{100, 200, 255, 255}
	if player == CellO {
		c = color.RGBA{255, 100, 150, 255}
	}
	
	cx := float64(GridX + col*CellSize + CellSize/2)
	cy := float64(GridY + row*CellSize + CellSize/2)
	
	for i := 0; i < 12; i++ {
		angle := float64(i) * 6.2832 / 12
		speed := 2 + rand.Float64()*3
		
		g.Particles = append(g.Particles, Particle{
			X: cx,
			Y: cy,
			VX: math.Cos(angle) * speed,
			VY: math.Sin(angle)*speed - 1,
			Life: 0.6 + rand.Float64()*0.4,
			MaxLife: 1.0,
			Color: c,
			Size: 3 + rand.Float64()*3,
		})
	}
}

func (g *Game) spawnWinParticles(cells [][2]int) {
	for _, cell := range cells {
		cx := float64(GridX + cell[1]*CellSize + CellSize/2)
		cy := float64(GridY + cell[0]*CellSize + CellSize/2)
		
		for i := 0; i < 20; i++ {
			angle := rand.Float64() * 6.2832
			speed := 3 + rand.Float64()*5
			
			g.Particles = append(g.Particles, Particle{
				X: cx,
				Y: cy,
				VX: math.Cos(angle) * speed,
				VY: math.Sin(angle)*speed - 2,
				Life: 1.0 + rand.Float64()*0.5,
				MaxLife: 1.5,
				Color: color.RGBA{255, 215, 0, 255},
				Size: 4 + rand.Float64()*5,
			})
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background gradient
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{15, 18, 35, 255}, false)
	
	// Stars
	for _, star := range g.Stars {
		alpha := uint8(star.Alpha * 180)
		s := int(star.Size)
		if s < 1 {
			s = 1
		}
		starImg := ebiten.NewImage(s*2, s*2)
		starImg.Fill(color.RGBA{255, 255, 255, alpha})
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(star.X, star.Y)
		screen.DrawImage(starImg, op)
	}
	
	switch g.State {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawGame(screen)
	case StateWin, StateDraw:
		g.drawGame(screen)
		g.drawResult(screen)
	}
	
	// Particles
	for _, p := range g.Particles {
		size := int(p.Size * (p.Life / p.MaxLife))
		if size < 1 {
			continue
		}
		
		alpha := uint8((p.Life / p.MaxLife) * 255)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		
		pImg := ebiten.NewImage(size, size)
		pImg.Fill(c)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(size)/2, p.Y-float64(size)/2)
		screen.DrawImage(pImg, op)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Title
	ebitenutil.DebugPrintAt(screen, "КРЕСТИКИ-НОЛИКИ", ScreenW/2-140, 280)
	
	// Subtitle
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge - Day 100", ScreenW/2-110, 320)
	
	// Decorative line
	lineImg := createLineImage(300, 4, color.RGBA{100, 180, 255, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ScreenW/2-150), 350)
	screen.DrawImage(lineImg, op)
	
	// X and O decorations
	xImg := createXImage(80, color.RGBA{100, 200, 255, 200})
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(ScreenW/2-180), 380)
	screen.DrawImage(xImg, op2)
	
	oImg := createOImage(80, color.RGBA{255, 100, 150, 200})
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(float64(ScreenW/2+100), 380)
	screen.DrawImage(oImg, op3)
	
	// Start button
	g.drawButton(screen, "▶  ИГРАТЬ", ScreenW/2-100, 500, 200, 60)
	
	// Info
	ebitenutil.DebugPrintAt(screen, "Вы играете крестиками (X)", ScreenW/2-120, 600)
	ebitenutil.DebugPrintAt(screen, "AI играет ноликами (O)", ScreenW/2-110, 625)
	ebitenutil.DebugPrintAt(screen, "AI использует алгоритм Minimax", ScreenW/2-130, 650)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Grid lines
	lineH := createLineImage(GridSize*CellSize, LineWidth, color.RGBA{80, 160, 240, 200})
	lineV := createLineImage(LineWidth, GridSize*CellSize, color.RGBA{80, 160, 240, 200})
	
	// Horizontal lines
	for i := 1; i < GridSize; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(GridX), float64(GridY+i*CellSize-LineWidth/2))
		screen.DrawImage(lineH, op)
	}
	
	// Vertical lines
	for i := 1; i < GridSize; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(GridX+i*CellSize-LineWidth/2), float64(GridY))
		screen.DrawImage(lineV, op)
	}
	
	// Cells
	xImg := createXImage(CellSize-20, color.RGBA{100, 200, 255, 255})
	oImg := createOImage(CellSize-20, color.RGBA{255, 100, 150, 255})
	
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			cell := g.Board.Grid[r][c]
			
			// Hover highlight
			if r == g.HoverRow && c == g.HoverCol && cell == CellEmpty {
				hoverImg := ebiten.NewImage(CellSize, CellSize)
				vector.DrawFilledRect(hoverImg, 8, 8, float32(CellSize-16), float32(CellSize-16), color.RGBA{60, 80, 120, 100}, false)
				
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(GridX+c*CellSize), float64(GridY+r*CellSize))
				screen.DrawImage(hoverImg, op)
			}
			
			if cell == CellX {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(
					float64(GridX+c*CellSize+10),
					float64(GridY+r*CellSize+10),
				)
				op.GeoM.Scale(g.CellAnims[r][c], g.CellAnims[r][c])
				screen.DrawImage(xImg, op)
			} else if cell == CellO {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(
					float64(GridX+c*CellSize+10),
					float64(GridY+r*CellSize+10),
				)
				op.GeoM.Scale(g.CellAnims[r][c], g.CellAnims[r][c])
				screen.DrawImage(oImg, op)
			}
		}
	}
	
	// Win line highlight
	if g.State == StateWin && len(g.WinCells) > 0 {
		for _, cell := range g.WinCells {
			winImg := ebiten.NewImage(CellSize, CellSize)
			vector.DrawFilledRect(winImg, 8, 8, float32(CellSize-16), float32(CellSize-16), color.RGBA{255, 215, 0, 80}, false)
			
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(GridX+cell[1]*CellSize), float64(GridY+cell[0]*CellSize))
			screen.DrawImage(winImg, op)
		}
	}
	
	// Turn indicator
	turnText := "Ваш ход (X)"
	if g.Turn == CellO {
		turnText = "Ход AI (O)..."
	}
	if g.AI_Thinking {
		turnText = "AI думает..."
	}
	ebitenutil.DebugPrintAt(screen, turnText, ScreenW/2-80, 80)
	
	// Score
	ebitenutil.DebugPrintAt(screen, "Вы (X): ", 30, 40)
	ebitenutil.DebugPrintAt(screen, "AI (O): ", 30, 70)
	ebitenutil.DebugPrintAt(screen, "Ничьи: ", 30, 100)
}

func (g *Game) drawResult(screen *ebiten.Image) {
	// Overlay
	overlay := ebiten.NewImage(ScreenW, 300)
	overlay.Fill(color.RGBA{10, 12, 25, 220})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 450)
	screen.DrawImage(overlay, op)
	
	// Result text
	if g.State == StateWin {
		// Check who won
		wonX, _ := g.Board.CheckWin(CellX)
		if wonX {
			ebitenutil.DebugPrintAt(screen, "🎉 ВЫ ПОБЕДИЛИ!", ScreenW/2-120, 500)
		} else {
			ebitenutil.DebugPrintAt(screen, "😔 AI ПОБЕДИЛ", ScreenW/2-110, 500)
		}
	} else {
		ebitenutil.DebugPrintAt(screen, "🤝 НИЧЬЯ!", ScreenW/2-80, 500)
	}
	
	// Buttons
	g.drawButton(screen, "🔄  ЕЩЁ РАЗ", ScreenW/2-100, 650, 200, 60)
	g.drawButton(screen, "←  МЕНЮ", ScreenW/2-100, 730, 200, 60)
}

func (g *Game) drawButton(screen *ebiten.Image, text string, x, y, w, h int) {
	// Button background
	btn := ebiten.NewImage(w, h)
	
	// Check hover
	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)
	hover := fmx >= float64(x) && fmx <= float64(x+w) && fmy >= float64(y) && fmy <= float64(y+h)
	
	if hover {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{50, 70, 110, 255}, false)
	} else {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{30, 40, 70, 255}, false)
	}
	
	// Border
	border := ebiten.NewImage(w, 3)
	border.Fill(color.RGBA{100, 180, 255, 255})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(border, op2)
	
	// Text
	ebitenutil.DebugPrintAt(screen, text, x+20, y+h/2-10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Крестики-Нолики - Go365 Day 100")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
