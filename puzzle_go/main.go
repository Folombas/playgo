// Crystal Cascade - Match-3 Game
// Go365 Challenge Day 99 - April 7, 2026
// A beautiful crystal matching game built with Ebitengine
package main

import (
	"bytes"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ─── Constants ───────────────────────────────────────────────────────────────

const (
	screenWidth  = 1280
	screenHeight = 800
	boardCols    = 8
	boardRows    = 8
	cellSize     = 70
	cellPad      = 4
	boardX       = 380
	boardY       = 60
	crystalCount = 6
)

var crystalNames = []string{"red", "blue", "green", "yellow", "violet", "orange"}

var crystalColors = []color.RGBA{
	{230, 50, 50, 255},    // red
	{50, 100, 230, 255},   // blue
	{50, 200, 80, 255},    // green
	{255, 210, 50, 255},   // yellow
	{180, 80, 230, 255},   // violet
	{255, 150, 50, 255},   // orange
}

var glowColors = [][]float64{
	{1.0, 0.2, 0.2, 1.0},   // red glow
	{0.2, 0.4, 1.0, 1.0},   // blue glow
	{0.2, 1.0, 0.3, 1.0},   // green glow
	{1.0, 0.8, 0.2, 1.0},   // yellow glow
	{0.7, 0.3, 1.0, 1.0},   // violet glow
	{1.0, 0.6, 0.2, 1.0},   // orange glow
}

// ─── Game States ─────────────────────────────────────────────────────────────

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

// ─── Easing Functions ────────────────────────────────────────────────────────

func easeOutBounce(t float64) float64 {
	const n1 = 7.5625
	const d1 = 2.75
	if t < 1/d1 {
		return n1 * t * t
	} else if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	} else if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	}
	t -= 2.625 / d1
	return n1*t*t + 0.984375
}

func easeOutElastic(t float64) float64 {
	const c4 = (2 * math.Pi) / 3
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	return math.Pow(2, -10*t)*math.Sin((t*10-0.75)*c4) + 1
}

func easeOutCubic(t float64) float64 {
	return 1 - math.Pow(1-t, 3)
}

func easeInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// ─── Crystal (Piece) ─────────────────────────────────────────────────────────

type Crystal struct {
	Type      int
	Col       int
	Row       int
	X         float64
	Y         float64
	TargetX   float64
	TargetY   float64
	Alpha     float64
	Scale     float64
	Rotation  float64
	Matched   bool
	Special   int // 0=none, 1=bomb, 2=rainbow, 3=horizontal-clear, 4=vertical-clear
	SpawnTime float64
	AnimProgress float64
	PulsePhase float64
}

func newCrystal(t, c, r int, fromAbove bool) *Crystal {
	x := float64(boardX + c*(cellSize+cellPad))
	y := float64(boardY + r*(cellSize+cellPad))
	if fromAbove {
		y = float64(boardY - (cellSize+cellPad)*5)
	}
	return &Crystal{
		Type: t, Col: c, Row: r,
		X: x, Y: y, TargetX: x, TargetY: y,
		Alpha: 1, Scale: 0.01, Rotation: 0,
		SpawnTime:    float64(time.Now().UnixNano()) / 1e9,
		AnimProgress: 0,
		PulsePhase:   rand.Float64() * 6.2832,
	}
}

func (c *Crystal) update(dt float64) bool {
	// Smooth movement
	dx := c.TargetX - c.X
	dy := c.TargetY - c.Y
	dist := dx*dx + dy*dy
	if dist > 0.25 {
		c.X += dx * 0.2
		c.Y += dy * 0.2
	} else {
		c.X = c.TargetX
		c.Y = c.TargetY
	}

	// Scale animation (spawn/match)
	if c.Scale < 1 && !c.Matched {
		c.AnimProgress += dt * 3
		if c.AnimProgress > 1 {
			c.AnimProgress = 1
		}
		c.Scale = easeOutBounce(c.AnimProgress)
	}

	// Pulse effect for special crystals
	if c.Special > 0 && !c.Matched {
		c.PulsePhase += dt * 3
	}

	// Match shrink animation
	if c.Matched {
		c.Scale *= 0.85
		c.Alpha *= 0.8
		c.Rotation += 0.1
		return c.Alpha > 0.02
	}

	return dist > 0.25
}

func (c *Crystal) contains(mx, my float64) bool {
	s := float64(cellSize) * c.Scale / 2
	cx := c.X + float64(cellSize)/2
	cy := c.Y + float64(cellSize)/2
	return mx >= cx-s && mx <= cx+s && my >= cy-s && my <= cy+s
}

// ─── Game Board ──────────────────────────────────────────────────────────────

type Board struct {
	grid [][]*Crystal
}

func newBoard() *Board {
	b := &Board{grid: make([][]*Crystal, boardCols)}
	for c := 0; c < boardCols; c++ {
		b.grid[c] = make([]*Crystal, boardRows)
	}
	b.fill()
	return b
}

func (b *Board) fill() {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			var t int
			for {
				t = rand.Intn(crystalCount)
				if !b.wouldMatch(c, r, t) {
					break
				}
			}
			b.grid[c][r] = newCrystal(t, c, r, false)
		}
	}
}

func (b *Board) wouldMatch(c, r, t int) bool {
	// Check horizontal
	n := 1
	for i := c - 1; i >= 0 && b.grid[i][r] != nil && b.grid[i][r].Type == t; i-- {
		n++
	}
	for i := c + 1; i < boardCols && b.grid[i][r] != nil && b.grid[i][r].Type == t; i++ {
		n++
	}
	if n >= 3 {
		return true
	}
	// Check vertical
	n = 1
	for i := r - 1; i >= 0 && b.grid[c][i] != nil && b.grid[c][i].Type == t; i-- {
		n++
	}
	for i := r + 1; i < boardRows && b.grid[c][i] != nil && b.grid[c][i].Type == t; i++ {
		n++
	}
	return n >= 3
}

func (b *Board) get(c, r int) *Crystal {
	if c < 0 || c >= boardCols || r < 0 || r >= boardRows {
		return nil
	}
	return b.grid[c][r]
}

func (b *Board) swap(c1, r1, c2, r2 int) bool {
	dc := c1 - c2
	dr := r1 - r2
	if dc < 0 {
		dc = -dc
	}
	if dr < 0 {
		dr = -dr
	}
	if !((dc == 1 && dr == 0) || (dc == 0 && dr == 1)) {
		return false
	}
	p1 := b.grid[c1][r1]
	p2 := b.grid[c2][r2]
	if p1 == nil || p2 == nil {
		return false
	}

	// Handle rainbow special
	if p1.Special == 2 {
		b.activateRainbow(c1, r1, p2.Type)
		p1.Matched = true
		p2.Matched = true
		return true
	}
	if p2.Special == 2 {
		b.activateRainbow(c2, r2, p1.Type)
		p1.Matched = true
		p2.Matched = true
		return true
	}

	b.grid[c1][r1] = p2
	b.grid[c2][r2] = p1
	p1.Col, p1.Row = c2, r2
	p2.Col, p2.Row = c1, r1
	p1.TargetX = float64(boardX + c2*(cellSize+cellPad))
	p1.TargetY = float64(boardY + r2*(cellSize+cellPad))
	p2.TargetX = float64(boardX + c1*(cellSize+cellPad))
	p2.TargetY = float64(boardY + r1*(cellSize+cellPad))
	return true
}

func (b *Board) findMatches() (int, bool) {
	matched := make(map[[2]int]bool)
	count := 0
	specialCreated := false

	// Horizontal matches
	for r := 0; r < boardRows; r++ {
		for c := 0; c <= boardCols-3; c++ {
			p := b.grid[c][r]
			if p == nil || p.Matched {
				continue
			}
			match := []*Crystal{p}
			for cc := c + 1; cc < boardCols && b.grid[cc][r] != nil && b.grid[cc][r].Type == p.Type; cc++ {
				match = append(match, b.grid[cc][r])
			}
			if len(match) >= 3 {
				// Create special pieces
				if len(match) == 4 {
					// Find middle piece to make special
					midIdx := len(match) / 2
					if match[midIdx] != nil {
						match[midIdx].Special = 1 // Bomb
						specialCreated = true
					}
				}
				if len(match) >= 5 {
					midIdx := len(match) / 2
					if match[midIdx] != nil {
						match[midIdx].Special = 2 // Rainbow
						specialCreated = true
					}
				}
				for _, m := range match {
					k := [2]int{m.Col, m.Row}
					if !matched[k] {
						matched[k] = true
						m.Matched = true
						count++
					}
				}
			}
		}
	}

	// Vertical matches
	for c := 0; c < boardCols; c++ {
		for r := 0; r <= boardRows-3; r++ {
			p := b.grid[c][r]
			if p == nil || p.Matched {
				continue
			}
			match := []*Crystal{p}
			for rr := r + 1; rr < boardRows && b.grid[c][rr] != nil && b.grid[c][rr].Type == p.Type; rr++ {
				match = append(match, b.grid[c][rr])
			}
			if len(match) >= 3 {
				if len(match) == 4 {
					midIdx := len(match) / 2
					if match[midIdx] != nil {
						match[midIdx].Special = 1 // Bomb
						specialCreated = true
					}
				}
				if len(match) >= 5 {
					midIdx := len(match) / 2
					if match[midIdx] != nil {
						match[midIdx].Special = 2 // Rainbow
						specialCreated = true
					}
				}
				for _, m := range match {
					k := [2]int{m.Col, m.Row}
					if !matched[k] {
						matched[k] = true
						m.Matched = true
						count++
					}
				}
			}
		}
	}
	return count, specialCreated
}

func (b *Board) explodeBomb(c, r int) {
	for dc := -1; dc <= 1; dc++ {
		for dr := -1; dr <= 1; dr++ {
			nc, nr := c+dc, r+dr
			if nc >= 0 && nc < boardCols && nr >= 0 && nr < boardRows && b.grid[nc][nr] != nil {
				b.grid[nc][nr].Matched = true
			}
		}
	}
}

func (b *Board) activateRainbow(c, r int, targetType int) {
	for cc := 0; cc < boardCols; cc++ {
		for rr := 0; rr < boardRows; rr++ {
			if b.grid[cc][rr] != nil && b.grid[cc][rr].Type == targetType {
				b.grid[cc][rr].Matched = true
			}
		}
	}
}

func (b *Board) remove() {
	// Activate specials before removing
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			if b.grid[c][r] != nil && b.grid[c][r].Matched {
				p := b.grid[c][r]
				if p.Special == 1 {
					b.explodeBomb(c, r)
				}
			}
		}
	}
	// Now remove
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			if b.grid[c][r] != nil && b.grid[c][r].Matched {
				b.grid[c][r] = nil
			}
		}
	}
}

func (b *Board) drop() bool {
	dropped := false
	for c := 0; c < boardCols; c++ {
		empty := -1
		for r := boardRows - 1; r >= 0; r-- {
			if b.grid[c][r] == nil {
				if empty == -1 {
					empty = r
				}
				continue
			}
			if empty != -1 {
				p := b.grid[c][r]
				b.grid[c][empty] = p
				b.grid[c][r] = nil
				p.Row = empty
				p.TargetY = float64(boardY + empty*(cellSize+cellPad))
				p.AnimProgress = 0
				empty--
				dropped = true
			}
		}
		for r := empty; r >= 0; r-- {
			t := rand.Intn(crystalCount)
			p := newCrystal(t, c, r, true)
			p.TargetY = float64(boardY + r*(cellSize+cellPad))
			b.grid[c][r] = p
			dropped = true
		}
	}
	return dropped
}

func (b *Board) isAnimating() bool {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			p := b.grid[c][r]
			if p == nil {
				continue
			}
			dx := p.TargetX - p.X
			dy := p.TargetY - p.Y
			if dx*dx+dy*dy > 0.5 {
				return true
			}
			if p.Matched && p.Alpha > 0.02 {
				return true
			}
			if p.Scale < 0.99 && !p.Matched {
				return true
			}
		}
	}
	return false
}

func (b *Board) at(mx, my float64) *Crystal {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			p := b.grid[c][r]
			if p != nil && p.contains(mx, my) {
				return p
			}
		}
	}
	return nil
}

// ─── Particle System ─────────────────────────────────────────────────────────

type Particle struct {
	X, Y, VX, VY float64
	Life         float64
	Color        color.RGBA
	Size         float64
	Rotation     float64
	RotSpeed     float64
	Type         int // 0=circle, 1=star, 2=sparkle
}

// ─── Background Star (parallax) ──────────────────────────────────────────────

type BGStar struct {
	X, Y   float64
	Size   float64
	Speed  float64
	Twinkle float64
	Phase  float64
}

// ─── Floating Text ───────────────────────────────────────────────────────────

type FloatingText struct {
	X, Y, VY float64
	Text     string
	Life     float64
	Color    color.RGBA
	Size     float64
}

func (ft *FloatingText) Update(dt float64) bool {
	ft.Y += ft.VY
	ft.VY *= 0.96
	ft.Life -= dt
	return ft.Life > 0
}

// ─── Main Game ───────────────────────────────────────────────────────────────

type Game struct {
	state GameState

	// Game data
	board   *Board
	score   int
	combo   int
	level   int
	moves   int
	maxMoves int

	// Assets
	crystals   []*ebiten.Image
	crystalGlow []*ebiten.Image
	bgStars    []BGStar
	bgTile     *ebiten.Image
	menuBg     *ebiten.Image
	menuStars  *ebiten.Image
	btnPlay    *ebiten.Image
	btnExit    *ebiten.Image
	particleStar *ebiten.Image
	particleSmallStar *ebiten.Image
	selector   *ebiten.Image

	// Effects
	particles    []*Particle
	floatingTexts []*FloatingText
	screenShakeX float64
	screenShakeY float64
	shakeTime    float64
	shakeIntensity float64

	// Input
	dragPiece  *Crystal
	dragStartX float64
	dragStartY float64
	isDragging bool

	// Timing
	gameTime   float64
	menuAnimT  float64

	// Audio (procedural)
	audio *AudioManager
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		state:     StateMenu,
		crystals:  make([]*ebiten.Image, crystalCount),
		bgStars:   make([]BGStar, 150),
		maxMoves:  30,
	}

	// Initialize background stars
	for i := range g.bgStars {
		g.bgStars[i] = BGStar{
			X:      rand.Float64() * screenWidth,
			Y:      rand.Float64() * screenHeight,
			Size:   rand.Float64()*2 + 0.5,
			Speed:  rand.Float64()*0.3 + 0.1,
			Phase:  rand.Float64() * 6.2832,
		}
	}

	g.loadAssets()
	g.audio = NewAudioManager()
	return g
}

func (g *Game) loadAssets() {
	execPath, _ := os.Getwd()
	spritesDir := filepath.Join(execPath, "assets", "sprites")
	menuDir := filepath.Join(execPath, "assets", "menu")

	// Load crystal sprites from gems
	gemFiles := []string{
		filepath.Join(execPath, "jewelred.png"),
		filepath.Join(execPath, "jewelblue_0.png"),
		filepath.Join(execPath, "jewelgreen.png"),
		filepath.Join(execPath, "jewelyellow.png"),
		filepath.Join(execPath, "jewelviolet.png"),
		filepath.Join(execPath, "gem5.png"), // orange-ish
	}

	// Try loading from sprites/gems first, fallback to root
	for i, path := range gemFiles {
		img, err := loadPNG(path)
		if err != nil {
			// Try alternative paths
			altPaths := []string{
				filepath.Join(spritesDir, filepath.Base(path)),
				filepath.Join(execPath, "assets", filepath.Base(path)),
			}
			for _, alt := range altPaths {
				img, err = loadPNG(alt)
				if err == nil {
					break
				}
			}
		}
		if img != nil {
			g.crystals[i] = img
		}
	}

	// If still no crystals, create procedural ones
	for i := 0; i < crystalCount; i++ {
		if g.crystals[i] == nil {
			g.crystals[i] = g.createProceduralCrystal(i)
		}
	}

	// Menu assets
	if img, err := loadPNG(filepath.Join(menuDir, "stars back.png")); err == nil {
		g.menuBg = img
	}
	if img, err := loadPNG(filepath.Join(menuDir, "stars.png")); err == nil {
		g.menuStars = img
	}
	if img, err := loadPNG(filepath.Join(menuDir, "play button.png")); err == nil {
		g.btnPlay = g.scaleImage(img, 200, 70)
	}
	if img, err := loadPNG(filepath.Join(menuDir, "Exit Button.png")); err == nil {
		g.btnExit = g.scaleImage(img, 200, 60)
	}

	// Particles
	if img, err := loadPNG(filepath.Join(spritesDir, "pack1", "particleStar.png")); err == nil {
		g.particleStar = img
	}
	if img, err := loadPNG(filepath.Join(spritesDir, "pack1", "particleSmallStar.png")); err == nil {
		g.particleSmallStar = img
	}

	// Selector
	if img, err := loadPNG(filepath.Join(spritesDir, "pack1", "selectorA.png")); err == nil {
		g.selector = img
	}

	// Background tile
	if img, err := loadPNG(filepath.Join(spritesDir, "backtiles", "BackTile_01.png")); err == nil {
		g.bgTile = g.scaleImage(img, screenWidth, screenHeight)
	}
}

func (g *Game) createProceduralCrystal(crystalType int) *ebiten.Image {
	img := ebiten.NewImage(cellSize, cellSize)
	col := crystalColors[crystalType]

	// Draw a diamond shape using rectangles
	center := cellSize / 2
	halfW := cellSize/2 - 6
	halfH := cellSize/2 - 6

	// Top triangle
	for y := center - halfH; y < center; y++ {
		if y < 0 || y >= cellSize {
			continue
		}
		progress := float64(y-(center-halfH)) / float64(halfH)
		w := int(float64(halfW) * progress)
		if w < 1 {
			continue
		}
		row := ebiten.NewImage(w*2, 1)
		row.Fill(col)
		rowOp := &ebiten.DrawImageOptions{}
		rowOp.GeoM.Translate(float64(center-w), float64(y))
		img.DrawImage(row, rowOp)
	}

	// Bottom triangle
	for y := center; y < center+halfH; y++ {
		if y < 0 || y >= cellSize {
			continue
		}
		progress := 1 - float64(y-center)/float64(halfH)
		w := int(float64(halfW) * progress)
		if w < 1 {
			continue
		}
		row := ebiten.NewImage(w*2, 1)
		row.Fill(col)
		rowOp := &ebiten.DrawImageOptions{}
		rowOp.GeoM.Translate(float64(center-w), float64(y))
		img.DrawImage(row, rowOp)
	}

	// Highlight
	hlCol := color.RGBA{255, 255, 255, 80}
	hlSize := 10
	hl := ebiten.NewImage(hlSize, hlSize/2)
	hl.Fill(hlCol)
	hlOp := &ebiten.DrawImageOptions{}
	hlOp.GeoM.Translate(float64(center-hlSize/2), float64(center-halfH+4))
	img.DrawImage(hl, hlOp)

	return img
}

func (g *Game) scaleImage(img *ebiten.Image, w, h int) *ebiten.Image {
	if img == nil {
		return nil
	}
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()
	if srcW == 0 || srcH == 0 {
		return nil
	}
	result := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(srcW), float64(h)/float64(srcH))
	result.DrawImage(img, op)
	return result
}

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

func (g *Game) Update() error {
	dt := 1.0 / 60.0
	g.gameTime += dt

	// Update bg stars
	for i := range g.bgStars {
		g.bgStars[i].Y += g.bgStars[i].Speed
		g.bgStars[i].Phase += dt * 2
		if g.bgStars[i].Y > screenHeight {
			g.bgStars[i].Y = -5
			g.bgStars[i].X = rand.Float64() * screenWidth
		}
	}

	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.15
		p.VX *= 0.98
		p.Life -= dt * 1.5
		p.Rotation += p.RotSpeed
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Update floating texts
	for i := len(g.floatingTexts) - 1; i >= 0; i-- {
		if !g.floatingTexts[i].Update(dt) {
			g.floatingTexts = append(g.floatingTexts[:i], g.floatingTexts[i+1:]...)
		}
	}

	// Screen shake decay
	if g.shakeTime > 0 {
		g.shakeTime -= dt
		progress := g.shakeTime / 0.3
		g.screenShakeX = math.Sin(g.gameTime*80) * g.shakeIntensity * progress
		g.screenShakeY = math.Cos(g.gameTime*60) * g.shakeIntensity * progress
	} else {
		g.screenShakeX = 0
		g.screenShakeY = 0
	}

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	// Menu animation
	if g.state == StateMenu {
		g.menuAnimT += dt * 1.5
		if g.menuAnimT > 1 {
			g.menuAnimT = 1
		}
	}

	// Click handling
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.state == StateMenu {
			// Play button area
			if fx >= 540 && fx <= 740 && fy >= 420 && fy <= 490 {
				g.playSound(600, 0.08, 0.3)
				g.startGame()
			}
			// Exit button area
			if fx >= 540 && fx <= 740 && fy >= 520 && fy <= 580 {
				os.Exit(0)
			}
		} else if g.state == StatePlaying {
			// Pause button
			if fx >= 1120 && fx <= 1240 && fy >= 20 && fy <= 60 {
				g.playSound(400, 0.06, 0.2)
				g.state = StatePaused
			}
		} else if g.state == StatePaused {
			// Resume
			if fx >= 540 && fx <= 740 && fy >= 350 && fy <= 420 {
				g.playSound(500, 0.06, 0.2)
				g.state = StatePlaying
			}
			// Quit to menu
			if fx >= 540 && fx <= 740 && fy >= 440 && fy <= 500 {
				g.playSound(300, 0.08, 0.2)
				g.state = StateMenu
				g.board = nil
			}
		}
	}

	// Game logic
	if g.state == StatePlaying {
		g.updateGame(fx, fy)
	}

	return nil
}

func (g *Game) startGame() {
	g.state = StatePlaying
	g.board = newBoard()
	g.score = 0
	g.combo = 0
	g.level = 1
	g.moves = 0
	g.particles = []*Particle{}
	g.floatingTexts = []*FloatingText{}
	g.playSound(523.25, 0.1, 0.3)
	g.playSound(659.25, 0.1, 0.25)
	g.playSound(783.99, 0.1, 0.2)
}

func (g *Game) updateGame(fx, fy float64) {
	if g.board == nil {
		return
	}

	// Update crystals
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			p := g.board.get(c, r)
			if p != nil {
				p.update(1.0 / 60.0)
			}
		}
	}

	// Check for matches when not animating
	if !g.board.isAnimating() {
		m, specialCreated := g.board.findMatches()
		if m > 0 {
			g.combo++
			g.processMatches(m, specialCreated)
		}
	}

	// Input handling
	g.handleInput(fx, fy)

	// Check game over
	if g.moves >= g.maxMoves {
		g.state = StateGameOver
	}
}

func (g *Game) processMatches(count int, specialCreated bool) {
	baseScore := count * 100
	comboMultiplier := 1 + float64(g.combo-1)*0.5
	totalScore := int(float64(baseScore) * comboMultiplier)
	g.score += totalScore

	// Spawn particles at match locations
	g.spawnMatchParticles(count)

	// Floating score text
	cx := float64(boardX + boardCols*(cellSize+cellPad)/2)
	cy := float64(boardY + boardRows*(cellSize+cellPad)/2)
	g.floatingTexts = append(g.floatingTexts, &FloatingText{
		X: cx - 40, Y: cy, VY: -2,
		Text: "+" + string(rune(totalScore)),
		Life: 1.5,
		Color: color.RGBA{255, 255, 255, 255},
		Size: 24,
	})

	// Remove matched crystals
	g.board.remove()

	// Screen shake for combos
	if g.combo >= 2 {
		g.shakeIntensity = float64(g.combo) * 3
		g.shakeTime = 0.3
	}

	// Play sounds
	if g.combo >= 3 {
		g.playComboSound(g.combo)
	} else {
		g.playMatchSound(g.combo)
	}

	// Drop new crystals
	if g.board.drop() {
		g.playSound(350, 0.05, 0.15)
	}

	g.moves++
}

func (g *Game) handleInput(fx, fy float64) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		p := g.board.at(fx, fy)
		if p != nil {
			g.dragPiece = p
			g.dragStartX = fx
			g.dragStartY = fy
			g.isDragging = false
			g.playSound(800, 0.05, 0.2)
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.dragPiece != nil {
		dx := fx - g.dragStartX
		dy := fy - g.dragStartY
		if !g.isDragging && (dx*dx+dy*dy > 100) {
			g.isDragging = true
		}
		if g.isDragging {
			g.dragPiece.X = g.dragPiece.TargetX + dx
			g.dragPiece.Y = g.dragPiece.TargetY + dy
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && g.dragPiece != nil {
		if g.isDragging {
			dx := g.dragPiece.X - g.dragPiece.TargetX
			dy := g.dragPiece.Y - g.dragPiece.TargetY
			tc, tr := g.dragPiece.Col, g.dragPiece.Row

			if dx*dx > dy*dy {
				if dx > 0 {
					tc++
				} else {
					tc--
				}
			} else {
				if dy > 0 {
					tr++
				} else {
					tr--
				}
			}

			if tc >= 0 && tc < boardCols && tr >= 0 && tr < boardRows {
				target := g.board.get(tc, tr)
				if target != nil {
					if g.board.swap(g.dragPiece.Col, g.dragPiece.Row, tc, tr) {
						m, _ := g.board.findMatches()
						if m == 0 {
							// Undo swap
							g.board.swap(tc, tr, g.dragPiece.Col, g.dragPiece.Row)
							g.combo = 0
							g.playSound(220, 0.1, 0.2)
						}
					}
				}
			}
		}
		g.dragPiece = nil
		g.isDragging = false
	}
}

// ─── Particle Effects ────────────────────────────────────────────────────────

func (g *Game) spawnMatchParticles(count int) {
	for i := 0; i < count*8; i++ {
		c := rand.Intn(boardCols)
		r := rand.Intn(boardRows)
		p := g.board.get(c, r)
		if p != nil {
			col := crystalColors[p.Type]
			g.particles = append(g.particles, &Particle{
				X:        float64(boardX+c*(cellSize+cellPad)+cellSize/2),
				Y:        float64(boardY+r*(cellSize+cellPad)+cellSize/2),
				VX:       float64(rand.Intn(9)-4) * 2,
				VY:       float64(rand.Intn(9)-4)*2 - 3,
				Life:     1,
				Color:    col,
				Size:     float64(4 + rand.Intn(6)),
				Rotation: rand.Float64() * 6.2832,
				RotSpeed: (rand.Float64() - 0.5) * 0.3,
				Type:     rand.Intn(3),
			})
		}
	}

	// Extra sparkle for special
	if count >= 4 {
		cx := float64(boardX + boardCols*(cellSize+cellPad)/2)
		cy := float64(boardY + boardRows*(cellSize+cellPad)/2)
		for i := 0; i < 30; i++ {
			angle := float64(i) * 6.2832 / 30
			speed := float64(3 + rand.Intn(5))
			g.particles = append(g.particles, &Particle{
				X:        cx,
				Y:        cy,
				VX:       math.Cos(angle) * speed,
				VY:       math.Sin(angle) * speed,
				Life:     1,
				Color:    color.RGBA{255, 255, 255, 255},
				Size:     float64(3 + rand.Intn(4)),
				Rotation: 0,
				RotSpeed: 0,
				Type:     2, // sparkle
			})
		}
	}
}

// ─── Sound System (Procedural) ───────────────────────────────────────────────

type AudioManager struct {
	volume float64
	ctx    *audio.Context
	mu     sync.Mutex
}

func NewAudioManager() *AudioManager {
	ctx := audio.NewContext(44100)
	return &AudioManager{volume: 0.4, ctx: ctx}
}

func (a *AudioManager) playTone(freq, dur, vol float64, wave string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sampleRate := 44100
	samples := int(float64(sampleRate) * dur)
	buf := make([]int16, samples)

	for i := 0; i < samples; i++ {
		t := float64(i) / float64(sampleRate)
		var val float64
		switch wave {
		case "sine":
			val = math.Sin(2 * math.Pi * freq * t)
		case "triangle":
			val = 2*math.Abs(2*(freq*t-math.Floor(freq*t+0.5))) - 1
		case "sawtooth":
			val = 2 * (freq*t - math.Floor(freq*t+0.5))
		default:
			val = math.Sin(2 * math.Pi * freq * t)
		}

		attack := 0.005
		release := 0.03
		env := 1.0
		if t < attack {
			env = t / attack
		} else if t > dur-release {
			env = (dur - t) / release
		}
		if env < 0 {
			env = 0
		}
		env *= vol * a.volume

		buf[i] = int16(val * env * 25000)
	}

	byteBuf := make([]byte, len(buf)*2)
	for i, s := range buf {
		byteBuf[i*2] = byte(s)
		byteBuf[i*2+1] = byte(s >> 8)
	}

	player := a.ctx.NewPlayerFromBytes(byteBuf)
	player.Play()
}

func (g *Game) playSound(freq, dur, vol float64) {
	if g.audio == nil {
		return
	}
	g.audio.playTone(freq, dur, vol, "sine")
}

func (g *Game) playMatchSound(combo int) {
	baseFreq := 440.0 + float64(combo)*60
	g.playSound(baseFreq, 0.12, 0.3)
	g.playSound(baseFreq*1.5, 0.08, 0.2)
}

func (g *Game) playComboSound(combo int) {
	freqs := []float64{523.25, 659.25, 783.99, 1046.50}
	for i := 0; i < combo && i < len(freqs); i++ {
		g.playSound(freqs[i%len(freqs)], 0.1, 0.25)
	}
}

// ─── Rendering ───────────────────────────────────────────────────────────────

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{10, 10, 25, 255})

	// Draw background stars
	for _, star := range g.bgStars {
		alpha := 0.5 + 0.5*math.Sin(star.Phase)
		s := int(star.Size * 2)
		if s < 1 {
			s = 1
		}
		starImg := ebiten.NewImage(s, s)
		starImg.Fill(color.RGBA{200, 220, 255, uint8(alpha * 255)})
		starOp := &ebiten.DrawImageOptions{}
		starOp.GeoM.Translate(star.X-float64(s)/2, star.Y-float64(s)/2)
		screen.DrawImage(starImg, starOp)
	}

	// Apply screen shake
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.screenShakeX, g.screenShakeY)

	if g.state == StateMenu {
		g.drawMenu(screen)
	} else if g.state == StatePlaying || g.state == StatePaused {
		g.drawGame(screen)
		if g.state == StatePaused {
			g.drawPause(screen)
		}
	} else if g.state == StateGameOver {
		g.drawGame(screen)
		g.drawGameOver(screen)
	}

	// Draw particles
	for _, p := range g.particles {
		if p.Life <= 0 {
			continue
		}
		alpha := uint8(p.Life * 255)
		s := int(p.Size * p.Life)
		if s < 1 {
			continue
		}

		particleImg := ebiten.NewImage(s, s)
		particleImg.Fill(color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha})

		drawOp := &ebiten.DrawImageOptions{}
		drawOp.GeoM.Translate(p.X-float64(s)/2, p.Y-float64(s)/2)
		drawOp.GeoM.Rotate(p.Rotation)
		drawOp.ColorM.Scale(1, 1, 1, p.Life)
		screen.DrawImage(particleImg, drawOp)
	}

	// Draw floating texts
	for _, ft := range g.floatingTexts {
		if ft.Life <= 0 {
			continue
		}
		alpha := uint8(ft.Life * 255)
		drawText(screen, ft.Text, ft.X, ft.Y, float64(ft.Size), color.RGBA{ft.Color.R, ft.Color.G, ft.Color.B, alpha})
	}
}

func drawText(screen *ebiten.Image, txt string, x, y, size float64, col color.RGBA) {
	if txt == "" {
		return
	}

	// Draw text as colored rectangle with background
	textW := float64(len(txt)) * size * 0.6
	textH := size

	// Background
	bg := ebiten.NewImage(int(textW)+10, int(textH)+6)
	bg.Fill(color.RGBA{0, 0, 0, uint8(float64(col.A) * 0.7)})
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Translate(x-5, y-3)
	screen.DrawImage(bg, bgOp)

	// Colored text block
	textBlock := ebiten.NewImage(int(textW), int(textH))
	textBlock.Fill(col)
	textOp := &ebiten.DrawImageOptions{}
	textOp.GeoM.Translate(x, y)
	screen.DrawImage(textBlock, textOp)
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Background
	if g.menuBg != nil {
		screen.DrawImage(g.menuBg, &ebiten.DrawImageOptions{})
	}
	if g.menuStars != nil {
		screen.DrawImage(g.menuStars, &ebiten.DrawImageOptions{})
	}

	// Title
	titleY := 180 + math.Sin(g.gameTime*2)*8
	titleW := 500
	titleH := 100

	// Title background
	titleBg := ebiten.NewImage(titleW, titleH)
	titleBg.Fill(color.RGBA{20, 30, 60, 220})
	titleBgOp := &ebiten.DrawImageOptions{}
	titleBgOp.GeoM.Translate(float64(screenWidth/2-titleW/2), titleY)
	screen.DrawImage(titleBg, titleBgOp)

	// Title border
	titleBorder := ebiten.NewImage(titleW, 4)
	titleBorder.Fill(color.RGBA{100, 180, 255, 255})
	titleBorderOp := &ebiten.DrawImageOptions{}
	titleBorderOp.GeoM.Translate(float64(screenWidth/2-titleW/2), titleY+float64(titleH))
	screen.DrawImage(titleBorder, titleBorderOp)

	// Title text placeholder
	titleText := ebiten.NewImage(400, 50)
	titleText.Fill(color.RGBA{150, 200, 255, 255})
	titleTextOp := &ebiten.DrawImageOptions{}
	titleTextOp.GeoM.Translate(float64(screenWidth/2-200), titleY+25)
	screen.DrawImage(titleText, titleTextOp)

	// Play button
	btnY := 420
	if g.btnPlay != nil {
		btnOp := &ebiten.DrawImageOptions{}
		btnOp.GeoM.Translate(540, float64(btnY))
		screen.DrawImage(g.btnPlay, btnOp)
	} else {
		btn := ebiten.NewImage(200, 70)
		btn.Fill(color.RGBA{50, 120, 180, 255})
		btnOp := &ebiten.DrawImageOptions{}
		btnOp.GeoM.Translate(540, float64(btnY))
		screen.DrawImage(btn, btnOp)
	}

	// Exit button
	exitY := 520
	if g.btnExit != nil {
		exitOp := &ebiten.DrawImageOptions{}
		exitOp.GeoM.Translate(540, float64(exitY))
		screen.DrawImage(g.btnExit, exitOp)
	} else {
		exitBtn := ebiten.NewImage(200, 60)
		exitBtn.Fill(color.RGBA{120, 50, 50, 255})
		exitOp := &ebiten.DrawImageOptions{}
		exitOp.GeoM.Translate(540, float64(exitY))
		screen.DrawImage(exitBtn, exitOp)
	}

	// Decorative crystals at bottom
	for i := 0; i < crystalCount; i++ {
		if g.crystals[i] != nil {
			crystalOp := &ebiten.DrawImageOptions{}
			crystalOp.GeoM.Scale(0.8, 0.8)
			crystalOp.GeoM.Translate(float64(200+i*100), 650+math.Sin(g.gameTime*2+float64(i))*10)
			screen.DrawImage(g.crystals[i], crystalOp)
		}
	}

	// Challenge badge
	badge := ebiten.NewImage(120, 40)
	badge.Fill(color.RGBA{40, 40, 80, 200})
	badgeOp := &ebiten.DrawImageOptions{}
	badgeOp.GeoM.Translate(20, screenHeight-60)
	screen.DrawImage(badge, badgeOp)

	ebitenutil.DebugPrintAt(screen, "Go365 Day 99", 30, screenHeight-50)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	if g.board == nil {
		return
	}

	// Board background
	bw := boardCols*(cellSize+cellPad) + 20
	bh := boardRows*(cellSize+cellPad) + 20
	boardBg := ebiten.NewImage(bw, bh)
	boardBg.Fill(color.RGBA{20, 25, 45, 240})
	boardBgOp := &ebiten.DrawImageOptions{}
	boardBgOp.GeoM.Translate(float64(boardX-10), float64(boardY-10))
	screen.DrawImage(boardBg, boardBgOp)

	// Board border
	boardBorder := ebiten.NewImage(bw, 4)
	boardBorder.Fill(color.RGBA{80, 160, 255, 255})
	boardBorderOp := &ebiten.DrawImageOptions{}
	boardBorderOp.GeoM.Translate(float64(boardX-10), float64(boardY-12))
	screen.DrawImage(boardBorder, boardBorderOp)

	// Grid lines
	for c := 0; c <= boardCols; c++ {
		x := float64(boardX + c*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, x, float64(boardY), x, float64(boardY+boardRows*(cellSize+cellPad)),
			color.RGBA{40, 50, 80, 100})
	}
	for r := 0; r <= boardRows; r++ {
		y := float64(boardY + r*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, float64(boardX), y, float64(boardX+boardCols*(cellSize+cellPad)), y,
			color.RGBA{40, 50, 80, 100})
	}

	// Draw crystals
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			p := g.board.get(c, r)
			if p == nil || p.Alpha < 0.02 || p == g.dragPiece {
				continue
			}
			g.drawCrystal(screen, p)
		}
	}
	if g.dragPiece != nil && g.dragPiece.Alpha > 0.02 {
		g.drawCrystal(screen, g.dragPiece)
	}

	// UI
	g.drawUI(screen)
}

func (g *Game) drawCrystal(screen *ebiten.Image, c *Crystal) {
	img := g.crystals[c.Type]
	if img == nil {
		return
	}

	s := int(float64(cellSize) * c.Scale)
	if s < 4 {
		return
	}

	sc := float64(s) / float64(img.Bounds().Dx())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sc, sc)
	op.GeoM.Rotate(c.Rotation)
	op.GeoM.Translate(c.X, c.Y)
	op.ColorM.Scale(1, 1, 1, c.Alpha)

	// Glow effect for special crystals
	if c.Special > 0 {
		glow := ebiten.NewImage(s+16, s+16)
		glowCol := glowColors[c.Type]
		glow.Fill(color.RGBA{uint8(glowCol[0]*255), uint8(glowCol[1]*255), uint8(glowCol[2]*255), 80})
		glowOp := &ebiten.DrawImageOptions{}
		glowOp.GeoM.Translate(c.X-8, c.Y-8)
		glowOp.ColorM.Scale(1, 1, 1, 0.3+0.2*math.Sin(c.PulsePhase))
		screen.DrawImage(glow, glowOp)
	}

	screen.DrawImage(img, op)

	// Special piece indicator
	if c.Special == 1 {
		// Bomb indicator - small circle
		bomb := ebiten.NewImage(s/3, s/3)
		bomb.Fill(color.RGBA{255, 100, 50, 200})
		bombOp := &ebiten.DrawImageOptions{}
		bombOp.GeoM.Translate(c.X+float64(s)/3, c.Y+float64(s)/3)
		screen.DrawImage(bomb, bombOp)
	} else if c.Special == 2 {
		// Rainbow indicator
		rainbow := ebiten.NewImage(s/3, s/3)
		rainbow.Fill(color.RGBA{200, 100, 255, 200})
		rainbowOp := &ebiten.DrawImageOptions{}
		rainbowOp.GeoM.Translate(c.X+float64(s)/3, c.Y+float64(s)/3)
		screen.DrawImage(rainbow, rainbowOp)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Score panel
	panelW := 280
	panelH := 300
	panel := ebiten.NewImage(panelW, panelH)
	panel.Fill(color.RGBA{20, 25, 45, 230})
	panelOp := &ebiten.DrawImageOptions{}
	panelOp.GeoM.Translate(20, 50)
	screen.DrawImage(panel, panelOp)

	// Panel border
	panelBorder := ebiten.NewImage(panelW, 3)
	panelBorder.Fill(color.RGBA{80, 160, 255, 255})
	panelBorderOp := &ebiten.DrawImageOptions{}
	panelBorderOp.GeoM.Translate(20, 50)
	screen.DrawImage(panelBorder, panelBorderOp)

	// Score display
	scoreW := 200
	scoreH := 40
	scoreBg := ebiten.NewImage(scoreW, scoreH)
	scoreBg.Fill(color.RGBA{40, 80, 120, 200})
	scoreBgOp := &ebiten.DrawImageOptions{}
	scoreBgOp.GeoM.Translate(60, 100)
	screen.DrawImage(scoreBg, scoreBgOp)

	// Combo display
	if g.combo > 1 {
		comboW := 80 + g.combo*15
		if comboW > 240 {
			comboW = 240
		}
		comboBg := ebiten.NewImage(comboW, 35)
		comboCol := color.RGBA{200, 180, 50, 200}
		if g.combo >= 5 {
			comboCol = color.RGBA{255, 100, 50, 220}
		}
		comboBg.Fill(comboCol)
		comboBgOp := &ebiten.DrawImageOptions{}
		comboBgOp.GeoM.Translate(60, 160)
		screen.DrawImage(comboBg, comboBgOp)
	}

	// Moves display
	movesW := 120
	movesH := 30
	movesBg := ebiten.NewImage(movesW, movesH)
	movesCol := color.RGBA{60, 140, 100, 200}
	if g.moves >= g.maxMoves-5 {
		movesCol = color.RGBA{180, 60, 60, 220}
	}
	movesBg.Fill(movesCol)
	movesBgOp := &ebiten.DrawImageOptions{}
	movesBgOp.GeoM.Translate(60, 220)
	screen.DrawImage(movesBg, movesBgOp)

	// Level display
	levelW := 80
	levelH := 30
	levelBg := ebiten.NewImage(levelW, levelH)
	levelBg.Fill(color.RGBA{100, 60, 180, 200})
	levelBgOp := &ebiten.DrawImageOptions{}
	levelBgOp.GeoM.Translate(60, 270)
	screen.DrawImage(levelBg, levelBgOp)

	// Pause button
	pauseBtn := ebiten.NewImage(120, 40)
	pauseBtn.Fill(color.RGBA{50, 50, 80, 200})
	pauseBtnOp := &ebiten.DrawImageOptions{}
	pauseBtnOp.GeoM.Translate(1120, 20)
	screen.DrawImage(pauseBtn, pauseBtnOp)

	// Game title at top
	titleText := ebiten.NewImage(400, 40)
	titleText.Fill(color.RGBA{100, 150, 200, 255})
	titleTextOp := &ebiten.DrawImageOptions{}
	titleTextOp.GeoM.Translate(440, 10)
	screen.DrawImage(titleText, titleTextOp)
}

func (g *Game) drawPause(screen *ebiten.Image) {
	// Dark overlay
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 150})
	overlayOp := &ebiten.DrawImageOptions{}
	overlayOp.GeoM.Translate(g.screenShakeX, g.screenShakeY)
	screen.DrawImage(overlay, overlayOp)

	// Pause panel
	panelW := 400
	panelH := 300
	panel := ebiten.NewImage(panelW, panelH)
	panel.Fill(color.RGBA{25, 30, 55, 245})
	panelOp := &ebiten.DrawImageOptions{}
	panelOp.GeoM.Translate(float64(screenWidth/2-panelW/2), float64(screenHeight/2-panelH/2))
	screen.DrawImage(panel, panelOp)

	// Panel border
	panelBorder := ebiten.NewImage(panelW, 4)
	panelBorder.Fill(color.RGBA{100, 180, 255, 255})
	panelBorderOp := &ebiten.DrawImageOptions{}
	panelBorderOp.GeoM.Translate(float64(screenWidth/2-panelW/2), float64(screenHeight/2-panelH/2))
	screen.DrawImage(panelBorder, panelBorderOp)

	// Resume button
	resumeBtn := ebiten.NewImage(200, 70)
	resumeBtn.Fill(color.RGBA{50, 120, 180, 255})
	resumeBtnOp := &ebiten.DrawImageOptions{}
	resumeBtnOp.GeoM.Translate(540, 350)
	screen.DrawImage(resumeBtn, resumeBtnOp)

	// Quit button
	quitBtn := ebiten.NewImage(200, 60)
	quitBtn.Fill(color.RGBA{120, 50, 50, 255})
	quitBtnOp := &ebiten.DrawImageOptions{}
	quitBtnOp.GeoM.Translate(540, 440)
	screen.DrawImage(quitBtn, quitBtnOp)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Dark overlay
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, &ebiten.DrawImageOptions{})

	// Game over panel
	panelW := 500
	panelH := 350
	panel := ebiten.NewImage(panelW, panelH)
	panel.Fill(color.RGBA{25, 30, 55, 250})
	panelOp := &ebiten.DrawImageOptions{}
	panelOp.GeoM.Translate(float64(screenWidth/2-panelW/2), float64(screenHeight/2-panelH/2))
	screen.DrawImage(panel, panelOp)

	// Panel border
	panelBorder := ebiten.NewImage(panelW, 4)
	panelBorder.Fill(color.RGBA{200, 100, 100, 255})
	panelBorderOp := &ebiten.DrawImageOptions{}
	panelBorderOp.GeoM.Translate(float64(screenWidth/2-panelW/2), float64(screenHeight/2-panelH/2))
	screen.DrawImage(panelBorder, panelBorderOp)

	// Score display
	scoreW := 300
	scoreH := 50
	scoreBg := ebiten.NewImage(scoreW, scoreH)
	scoreBg.Fill(color.RGBA{60, 100, 160, 220})
	scoreBgOp := &ebiten.DrawImageOptions{}
	scoreBgOp.GeoM.Translate(float64(screenWidth/2-scoreW/2), float64(screenHeight/2-30))
	screen.DrawImage(scoreBg, scoreBgOp)

	// Play again button
	againBtn := ebiten.NewImage(200, 70)
	againBtn.Fill(color.RGBA{50, 120, 180, 255})
	againBtnOp := &ebiten.DrawImageOptions{}
	againBtnOp.GeoM.Translate(540, 500)
	screen.DrawImage(againBtn, againBtnOp)
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Crystal Cascade - Go365 Day 99")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
