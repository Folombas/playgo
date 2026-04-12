package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	_ "image/png"
)

// Constants
const (
	screenWidth  = 800
	screenHeight = 800
	gridRows     = 8
	gridCols     = 8
	tileSize     = 80
	gridOffsetX  = (screenWidth - gridCols*tileSize) / 2
	gridOffsetY  = 200
	numGemTypes  = 6
	gameDuration = 60 // seconds
)

// Gem colors for visual variety
var gemColors = []color.RGBA{
	{255, 59, 48, 255},   // Red
	{255, 149, 0, 255},   // Orange
	{255, 204, 0, 255},   // Yellow
	{52, 199, 89, 255},   // Green
	{0, 122, 255, 255},   // Blue
	{175, 82, 222, 255},  // Purple
}

// Gem names for sprite loading
var gemNames = []string{
	"red",
	"orange",
	"yellow",
	"green",
	"blue",
	"purple",
}

// GameState represents different screens
type GameState int

const (
	StatePlaying GameState = iota
	StateGameOver
)

// Tile represents a single gem on the board
type Tile struct {
	GemType   int
	Row, Col  int
	X, Y      float64
	TargetX   float64
	TargetY   float64
	Selected  bool
	Removing  bool
	Scale     float64
	Alpha     float64
	ShakeTime float64
}

// Board represents the game board
type Board struct {
	Tiles [][]*Tile
}

// Game is the main game structure
type Game struct {
	State        GameState
	Board        *Board
	Score        int
	Combo        int
	MaxCombo     int
	Timer        float64
	SelectedTile *Tile
	HintTile1    *Tile
	HintTile2    *Tile
	HintActive   bool
	LastAction   time.Time
	RNG          *rand.Rand
	GemImages    []*ebiten.Image
}

func main() {
	rand.Seed(time.Now().UnixNano())

	game := &Game{
		State:      StatePlaying,
		Timer:      gameDuration,
		LastAction: time.Now(),
		RNG:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Load gem sprites
	game.loadGemSprites()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("💎 Crystal Cascade - Match-3")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetTPS(60)

	if err := ebiten.RunGame(game); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Update() error {
	if g.State == StatePlaying {
		g.updatePlaying()
	} else if g.State == StateGameOver {
		g.updateGameOver()
	}
	return nil
}

func (g *Game) updatePlaying() {
	// Update timer
	g.Timer -= 1.0 / 60.0
	if g.Timer <= 0 {
		g.Timer = 0
		g.State = StateGameOver
		return
	}

	// Initialize board if needed
	if g.Board == nil {
		g.initBoard()
	}

	// Handle input
	g.handleInput()

	// Check for hint
	if time.Since(g.LastAction) > 5*time.Second && !g.HintActive {
		g.showHint()
	}

	// Update board animations
	g.Board.Update()

	// Process matches
	if !g.Board.IsAnimating() {
		matches := g.Board.FindAllMatches()
		if len(matches) > 0 {
			g.Combo++
			g.LastAction = time.Now()
			g.HintActive = false

			// Calculate score
			score := g.calculateMatchScore(matches)
			g.Score += score

			// Track max combo
			if g.Combo > g.MaxCombo {
				g.MaxCombo = g.Combo
			}

			// Remove matched tiles
			g.Board.RemoveTiles(matches)
		} else {
			g.Combo = 0
		}
	}

	// Apply gravity and fill
	if g.Board.HasEmptyTiles() {
		g.Board.ApplyGravity()
		g.Board.FillEmpty()
	}
}

func (g *Game) calculateMatchScore(matches []*Tile) int {
	// Group matches by connected components
	groups := g.Board.GroupMatches(matches)

	totalScore := 0
	for _, group := range groups {
		baseScore := len(group) * 10

		// Bonus for larger matches
		if len(group) == 4 {
			baseScore = 50
		} else if len(group) >= 5 {
			baseScore = 100
		}

		// Combo multiplier
		multiplier := 1.0 + float64(g.Combo-1)*0.5
		totalScore += int(float64(baseScore) * multiplier)
	}

	return totalScore
}

func (g *Game) updateGameOver() {
	// Check for restart
	if ebiten.IsKeyPressed(ebiten.KeyR) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		time.Sleep(200 * time.Millisecond) // Debounce
		g.restartGame()
	}
}

func (g *Game) handleInput() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.handleTileClick(x, y)
	}

	// Keyboard shortcuts
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.restartGame()
	}
}

func (g *Game) handleTileClick(x, y int) {
	col := (x - gridOffsetX) / tileSize
	row := (y - gridOffsetY) / tileSize

	if row < 0 || row >= gridRows || col < 0 || col >= gridCols {
		return
	}

	tile := g.Board.Tiles[row][col]
	if tile == nil || tile.Removing {
		return
	}

	g.LastAction = time.Now()
	g.HintActive = false

	if g.SelectedTile == nil {
		// Select first tile
		g.SelectedTile = tile
		tile.Selected = true
	} else if g.SelectedTile == tile {
		// Deselect
		tile.Selected = false
		g.SelectedTile = nil
	} else {
		// Try to swap
		dr := g.absInt(g.SelectedTile.Row - tile.Row)
		dc := g.absInt(g.SelectedTile.Col - tile.Col)

		if dr+dc == 1 {
			// Adjacent - try swap
			g.trySwap(g.SelectedTile, tile)
		} else {
			// Not adjacent - select new tile
			g.SelectedTile.Selected = false
			g.SelectedTile = tile
			tile.Selected = true
		}
	}
}

func (g *Game) trySwap(tile1, tile2 *Tile) {
	// Perform swap
	g.Board.SwapTiles(tile1, tile2)

	// Check for matches
	matches := g.Board.FindAllMatches()

	if len(matches) == 0 {
		// No matches - swap back and shake
		g.Board.SwapTiles(tile1, tile2)
		tile1.ShakeTime = 0.5
		tile2.ShakeTime = 0.5
	} else {
		// Valid swap
		g.LastAction = time.Now()
	}

	tile1.Selected = false
	g.SelectedTile = nil
}

func (g *Game) initBoard() {
	g.Board = NewBoard()
	g.Board.RemoveInitialMatches()
}

func (g *Game) restartGame() {
	g.State = StatePlaying
	g.Score = 0
	g.Combo = 0
	g.MaxCombo = 0
	g.Timer = gameDuration
	g.SelectedTile = nil
	g.HintActive = false
	g.HintTile1 = nil
	g.HintTile2 = nil
	g.LastAction = time.Now()
	g.Board = nil
	g.initBoard()
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear with gradient background
	g.drawBackground(screen)

	if g.State == StatePlaying {
		g.drawGame(screen)
	} else if g.State == StateGameOver {
		g.drawGameOver(screen)
	}
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	// Dark blue-purple gradient
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{25, 20, 50, 255}, false)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Draw HUD
	g.drawHUD(screen)

	// Draw board background
	g.drawBoardBackground(screen)

	// Draw tiles
	if g.Board != nil {
		g.Board.Draw(screen, g.GemImages)
	}

	// Draw hints
	if g.HintActive && g.HintTile1 != nil && g.HintTile2 != nil {
		g.drawHint(screen, g.HintTile1)
		g.drawHint(screen, g.HintTile2)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Score
	scoreText := fmt.Sprintf("Score: %d", g.Score)
	ebitenutil.DebugPrintAt(screen, scoreText, 20, 30)

	// Timer
	timerText := fmt.Sprintf("Time: %.1fs", g.Timer)
	ebitenutil.DebugPrintAt(screen, timerText, screenWidth-150, 30)

	// Combo
	if g.Combo > 1 {
		comboText := fmt.Sprintf("COMBO x%d!", g.Combo)
		ebitenutil.DebugPrintAt(screen, comboText, screenWidth/2-50, 60)
	}

	// Title
	ebitenutil.DebugPrintAt(screen, "Crystal Cascade", screenWidth/2-70, 100)

	// Instructions
	ebitenutil.DebugPrintAt(screen, "Press R to restart", 20, screenHeight-30)
}

func (g *Game) drawBoardBackground(screen *ebiten.Image) {
	// Board background
	boardWidth := gridCols * tileSize
	boardHeight := gridRows * tileSize
	vector.DrawFilledRect(screen,
		float32(gridOffsetX-5), float32(gridOffsetY-5),
		float32(boardWidth+10), float32(boardHeight+10),
		color.RGBA{40, 30, 70, 255}, false)

	// Grid cells
	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			x := gridOffsetX + col*tileSize
			y := gridOffsetY + row*tileSize

			cellColor := color.RGBA{50, 40, 80, 255}
			if (row+col)%2 == 0 {
				cellColor = color.RGBA{60, 50, 90, 255}
			}

			vector.DrawFilledRect(screen,
				float32(x+2), float32(y+2),
				float32(tileSize-4), float32(tileSize-4),
				cellColor, false)
		}
	}
}

func (g *Game) drawHint(screen *ebiten.Image, tile *Tile) {
	if tile == nil {
		return
	}

	x := float64(gridOffsetX + tile.Col*tileSize)
	y := float64(gridOffsetY + tile.Row*tileSize)

	// Pulsing highlight
	pulse := math.Sin(float64(time.Now().UnixMilli())*0.008)*0.3 + 0.7
	alpha := uint8(200 * pulse)

	highlight := ebiten.NewImage(tileSize, tileSize)
	highlight.Fill(color.RGBA{0, 255, 255, alpha})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(highlight, op)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Overlay
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 200}, false)

	// Game Over text
	ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-80, 250)

	// Final score
	scoreText := fmt.Sprintf("Final Score: %d", g.Score)
	ebitenutil.DebugPrintAt(screen, scoreText, screenWidth/2-100, 350)

	// Max combo
	if g.MaxCombo > 0 {
		comboText := fmt.Sprintf("Max Combo: x%d", g.MaxCombo)
		ebitenutil.DebugPrintAt(screen, comboText, screenWidth/2-80, 400)
	}

	// Restart instruction
	ebitenutil.DebugPrintAt(screen, "Press R or click to restart", screenWidth/2-120, 500)
}

func (g *Game) loadGemSprites() {
	g.GemImages = make([]*ebiten.Image, numGemTypes)

	// Try to load from sprites directory
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}

	// Get the directory of the executable
	execDir := filepath.Dir(execPath)

	// Try multiple paths for finding sprites
	spritePaths := []string{
		filepath.Join(execDir, "sprites"),
		filepath.Join(".", "sprites"),
		filepath.Join("..", "sprites"),
	}

	spriteDir := ""
	for _, path := range spritePaths {
		if _, err := os.Stat(path); err == nil {
			spriteDir = path
			break
		}
	}

	if spriteDir == "" {
		// Fall back to colored circles
		return
	}

	for i := 0; i < numGemTypes; i++ {
		spritePath := filepath.Join(spriteDir, gemNames[i]+".png")
		if img, err := loadPNG(spritePath); err == nil {
			g.GemImages[i] = img
		}
	}
}

func loadPNG(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

// ShowHint displays a hint to the player
func (g *Game) showHint() {
	if g.Board == nil {
		return
	}

	t1, t2 := g.Board.FindHint()
	if t1 != nil && t2 != nil {
		g.HintTile1 = t1
		g.HintTile2 = t2
		g.HintActive = true

		// Auto-hide hint after 2 seconds
		go func() {
			time.Sleep(2 * time.Second)
			g.HintActive = false
		}()
	}
}

func (g *Game) absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// NewBoard creates a new game board
func NewBoard() *Board {
	b := &Board{
		Tiles: make([][]*Tile, gridRows),
	}

	// Initialize tiles
	for row := 0; row < gridRows; row++ {
		b.Tiles[row] = make([]*Tile, gridCols)
		for col := 0; col < gridCols; col++ {
			b.Tiles[row][col] = b.createTile(row, col)
		}
	}

	return b
}

func (b *Board) createTile(row, col int) *Tile {
	return &Tile{
		GemType: rand.Intn(numGemTypes),
		Row:     row,
		Col:     col,
		X:       float64(gridOffsetX + col*tileSize),
		Y:       float64(gridOffsetY + row*tileSize),
		TargetX: float64(gridOffsetX + col*tileSize),
		TargetY: float64(gridOffsetY + row*tileSize),
		Scale:   1.0,
		Alpha:   1.0,
	}
}

// Update updates all tile animations
func (b *Board) Update() {
	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			tile := b.Tiles[row][col]
			if tile == nil {
				continue
			}

			// Smooth movement
			tile.X += (tile.TargetX - tile.X) * 0.2
			tile.Y += (tile.TargetY - tile.Y) * 0.2

			// Shake animation
			if tile.ShakeTime > 0 {
				tile.ShakeTime -= 1.0 / 60.0
				if tile.ShakeTime < 0 {
					tile.ShakeTime = 0
				}
			}

			// Removal animation
			if tile.Removing {
				tile.Scale *= 0.85
				tile.Alpha *= 0.85
			}
		}
	}
}

// Draw draws all tiles on the board
func (b *Board) Draw(screen *ebiten.Image, gemImages []*ebiten.Image) {
	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			tile := b.Tiles[row][col]
			if tile == nil || (tile.Removing && tile.Alpha < 0.01) {
				continue
			}

			b.drawTile(screen, tile, gemImages)
		}
	}
}

func (b *Board) drawTile(screen *ebiten.Image, tile *Tile, gemImages []*ebiten.Image) {
	x := tile.X
	y := tile.Y

	// Shake offset
	if tile.ShakeTime > 0 {
		shakeOffset := math.Sin(tile.ShakeTime*math.Pi*10) * 5 * tile.ShakeTime
		x += shakeOffset
	}

	// Selection highlight
	if tile.Selected {
		highlight := ebiten.NewImage(tileSize, tileSize)
		highlight.Fill(color.RGBA{255, 215, 0, 180})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x-2, y-2)
		screen.DrawImage(highlight, op)
	}

	// Draw gem
	if tile.Alpha > 0.01 {
		size := float64(tileSize - 8)
		offset := 4.0

		if tile.Scale < 0.99 {
			// Shrinking gem
			size *= tile.Scale
			offset = (float64(tileSize) - size) / 2
		}

		if gemImages[tile.GemType] != nil {
			// Use sprite
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x+offset, y+offset)
			op.ColorScale.ScaleAlpha(float32(tile.Alpha))
			screen.DrawImage(gemImages[tile.GemType], op)
		} else {
			// Fallback to colored circle
			b.drawGemCircle(screen, x+offset, y+offset, size/2, tile.GemType, tile.Alpha)
		}
	}
}

func (b *Board) drawGemCircle(screen *ebiten.Image, x, y, radius float64, gemType int, alpha float64) {
	// Draw a filled circle
	diameter := int(radius * 2)
	if diameter <= 0 {
		return
	}

	img := ebiten.NewImage(diameter, diameter)
	center := float64(diameter) / 2
	c := gemColors[gemType]

	for py := 0; py < diameter; py++ {
		for px := 0; px < diameter; px++ {
			dx := float64(px) - center
			dy := float64(py) - center
			if dx*dx+dy*dy <= radius*radius {
				img.Set(px, py, color.RGBA{c.R, c.G, c.B, uint8(float64(c.A) * alpha)})
			}
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

// RemoveInitialMatches removes any initial matches from the board
func (b *Board) RemoveInitialMatches() {
	for {
		matches := b.FindAllMatches()
		if len(matches) == 0 {
			break
		}
		for _, tile := range matches {
			tile.GemType = rand.Intn(numGemTypes)
		}
	}
}

// FindAllMatches finds all matched tiles (3+ in a row/column)
func (b *Board) FindAllMatches() []*Tile {
	matched := make(map[int]*Tile)

	// Horizontal matches
	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols-2; col++ {
			tile := b.Tiles[row][col]
			if tile == nil || tile.Removing {
				continue
			}

			gemType := tile.GemType
			matchLen := 1

			// Check how many consecutive tiles
			for nextCol := col + 1; nextCol < gridCols; nextCol++ {
				nextTile := b.Tiles[row][nextCol]
				if nextTile != nil && !nextTile.Removing && nextTile.GemType == gemType {
					matchLen++
				} else {
					break
				}
			}

			// If 3+ match, add all
			if matchLen >= 3 {
				for i := 0; i < matchLen; i++ {
					key := row*100 + (col + i)
					matched[key] = b.Tiles[row][col+i]
				}
			}
		}
	}

	// Vertical matches
	for col := 0; col < gridCols; col++ {
		for row := 0; row < gridRows-2; row++ {
			tile := b.Tiles[row][col]
			if tile == nil || tile.Removing {
				continue
			}

			gemType := tile.GemType
			matchLen := 1

			// Check how many consecutive tiles
			for nextRow := row + 1; nextRow < gridRows; nextRow++ {
				nextTile := b.Tiles[nextRow][col]
				if nextTile != nil && !nextTile.Removing && nextTile.GemType == gemType {
					matchLen++
				} else {
					break
				}
			}

			// If 3+ match, add all
			if matchLen >= 3 {
				for i := 0; i < matchLen; i++ {
					key := (row + i)*100 + col
					matched[key] = b.Tiles[row+i][col]
				}
			}
		}
	}

	// Convert map to slice
	result := make([]*Tile, 0, len(matched))
	for _, tile := range matched {
		result = append(result, tile)
	}

	return result
}

// GroupMatches groups matched tiles into connected components
func (b *Board) GroupMatches(matches []*Tile) [][]*Tile {
	if len(matches) == 0 {
		return nil
	}

	// Simple grouping by row and column
	groups := make([][]*Tile, 0)
	visited := make(map[int]bool)

	for _, tile := range matches {
		key := tile.Row*100 + tile.Col
		if visited[key] {
			continue
		}

		// Start a new group
		group := []*Tile{tile}
		visited[key] = true

		// Find connected tiles
		for _, other := range matches {
			otherKey := other.Row*100 + other.Col
			if visited[otherKey] {
				continue
			}

			// Check if adjacent
			dr := absInt(tile.Row - other.Row)
			dc := absInt(tile.Col - other.Col)
			if dr+dc <= 1 {
				group = append(group, other)
				visited[otherKey] = true
			}
		}

		groups = append(groups, group)
	}

	return groups
}

// RemoveTiles marks tiles for removal
func (b *Board) RemoveTiles(tiles []*Tile) {
	for _, tile := range tiles {
		tile.Removing = true
	}
}

// HasEmptyTiles checks if there are any empty tiles
func (b *Board) HasEmptyTiles() bool {
	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			if b.Tiles[row][col] == nil || b.Tiles[row][col].Removing {
				return true
			}
		}
	}
	return false
}

// IsAnimating checks if any tiles are still animating
func (b *Board) IsAnimating() bool {
	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			tile := b.Tiles[row][col]
			if tile == nil {
				continue
			}

			// Check if tile is moving
			dx := tile.TargetX - tile.X
			dy := tile.TargetY - tile.Y
			if dx*dx+dy*dy > 0.1 {
				return true
			}
		}
	}
	return false
}

// ApplyGravity moves tiles down to fill empty spaces
func (b *Board) ApplyGravity() {
	for col := 0; col < gridCols; col++ {
		writeRow := gridRows - 1

		for row := gridRows - 1; row >= 0; row-- {
			tile := b.Tiles[row][col]
			if tile != nil && !tile.Removing {
				if writeRow != row {
					b.Tiles[writeRow][col] = tile
					tile.Row = writeRow
					tile.TargetY = float64(gridOffsetY + writeRow*tileSize)
					b.Tiles[row][col] = nil
				}
				writeRow--
			}
		}
	}
}

// FillEmpty fills empty tiles with new ones
func (b *Board) FillEmpty() {
	for col := 0; col < gridCols; col++ {
		for row := 0; row < gridRows; row++ {
			if b.Tiles[row][col] == nil || b.Tiles[row][col].Removing {
				// Create new tile above the board
				newTile := b.createTile(row, col)
				newTile.Y = float64(gridOffsetY - (gridRows-row)*tileSize)
				newTile.TargetY = float64(gridOffsetY + row*tileSize)
				b.Tiles[row][col] = newTile
			}
		}
	}
}

// SwapTiles swaps two tiles
func (b *Board) SwapTiles(tile1, tile2 *Tile) {
	// Swap in array
	b.Tiles[tile1.Row][tile1.Col], b.Tiles[tile2.Row][tile2.Col] = b.Tiles[tile2.Row][tile2.Col], b.Tiles[tile1.Row][tile1.Col]

	// Update tile data
	tile1.Row, tile2.Row = tile2.Row, tile1.Row
	tile1.Col, tile2.Col = tile2.Col, tile1.Col

	// Update targets
	tile1.TargetX = float64(gridOffsetX + tile1.Col*tileSize)
	tile1.TargetY = float64(gridOffsetY + tile1.Row*tileSize)
	tile2.TargetX = float64(gridOffsetX + tile2.Col*tileSize)
	tile2.TargetY = float64(gridOffsetY + tile2.Row*tileSize)
}

// FindHint finds a valid hint move
func (b *Board) FindHint() (tile1, tile2 *Tile) {
	// Try all possible swaps
	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			// Try right
			if col < gridCols-1 {
				b.Tiles[row][col].GemType, b.Tiles[row][col+1].GemType = b.Tiles[row][col+1].GemType, b.Tiles[row][col].GemType
				if len(b.FindAllMatches()) > 0 {
					b.Tiles[row][col].GemType, b.Tiles[row][col+1].GemType = b.Tiles[row][col+1].GemType, b.Tiles[row][col].GemType
					return b.Tiles[row][col], b.Tiles[row][col+1]
				}
				b.Tiles[row][col].GemType, b.Tiles[row][col+1].GemType = b.Tiles[row][col+1].GemType, b.Tiles[row][col].GemType
			}
			// Try down
			if row < gridRows-1 {
				b.Tiles[row][col].GemType, b.Tiles[row+1][col].GemType = b.Tiles[row+1][col].GemType, b.Tiles[row][col].GemType
				if len(b.FindAllMatches()) > 0 {
					b.Tiles[row][col].GemType, b.Tiles[row+1][col].GemType = b.Tiles[row+1][col].GemType, b.Tiles[row][col].GemType
					return b.Tiles[row][col], b.Tiles[row+1][col]
				}
				b.Tiles[row][col].GemType, b.Tiles[row+1][col].GemType = b.Tiles[row+1][col].GemType, b.Tiles[row][col].GemType
			}
		}
	}
	return nil, nil
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
