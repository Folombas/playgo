// Крестики-Нолики - Go365 Challenge Day 101
// Улучшенная версия: ЗВУКИ + УРОВНИ СЛОЖНОСТИ AI
// 8 апреля 2026

package main

import (
	"bytes"
	"embed"
	"encoding/json"
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
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

//go:embed assets/sounds/*.ogg assets/sprites/*.png
var assetFS embed.FS

// ============================================================================
// ДОСТИЖЕНИЯ
// ============================================================================

type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Unlocked    bool   `json:"unlocked"`
	Icon        string `json:"icon"`
}

type PlayerStats struct {
	TotalGames   int  `json:"total_games"`
	Wins         int  `json:"wins"`
	Losses       int  `json:"losses"`
	Draws        int  `json:"draws"`
	WinStreak    int  `json:"win_streak"`
	BestStreak   int  `json:"best_streak"`
	FirstWin     bool `json:"first_win"`
	TenWins      bool `json:"ten_wins"`
	FirstLoss    bool `json:"first_loss"`
	Unbeatable   bool `json:"unbeatable"` // 5 побед подряд
	Achievements []Achievement `json:"achievements"`
}

func DefaultStats() *PlayerStats {
	return &PlayerStats{
		Achievements: []Achievement{
			{"first_game", "Первая игра", "Сыграть первую игру", false, "🎮"},
			{"first_win", "Первая победа", "Выиграть первую игру", false, "🏆"},
			{"ten_wins", "Десять побед", "Выиграть 10 игр", false, "⭐"},
			{"win_streak_3", "Серия 3", "Выиграть 3 игры подряд", false, "🔥"},
			{"win_streak_5", "Серия 5", "Выиграть 5 игр подряд", false, "💎"},
			{"first_draw", "Ничья!", "Сыграть вничью", false, "🤝"},
			{"ten_draws", "Мастер ничьих", "10 ничьих", false, "🎯"},
			{"beat_hard", "Победить сложного AI", "Выиграть на сложном уровне", false, "👑"},
		},
	}
}

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
	StateAchievements
)

type AIDifficulty int

const (
	AIEasy AIDifficulty = iota
	AIMedium
	AIHard
)

var aiNames = []string{"Лёгкий", "Средний", "Сложный"}

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

// AI - Minimax algorithm с уровнями сложности
func (b *Board) BestMove(difficulty AIDifficulty) (int, int) {
	// Easy: 40% random moves
	if difficulty == AIEasy && rand.Float64() < 0.4 {
		return b.RandomMove()
	}
	
	// Medium: 20% random moves
	if difficulty == AIMedium && rand.Float64() < 0.2 {
		return b.RandomMove()
	}
	
	// Hard: always best move (minimax)
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

func (b *Board) RandomMove() (int, int) {
	var moves [][2]int
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if b.Grid[r][c] == CellEmpty {
				moves = append(moves, [2]int{r, c})
			}
		}
	}
	if len(moves) == 0 {
		return -1, -1
	}
	move := moves[rand.Intn(len(moves))]
	return move[0], move[1]
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
	AIDifficulty AIDifficulty
	
	// Audio
	audioContext  *audio.Context
	clickPlayer   *audio.Player
	winPlayer     *audio.Player
	losePlayer    *audio.Player
	
	// Sprites
	SpriteX      *ebiten.Image
	SpriteO      *ebiten.Image
	
	// Stats & Achievements
	Stats        *PlayerStats
	NewAchievement *Achievement
	AchieveTimer float64
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
		AIDifficulty: AIHard,
		Stats:    DefaultStats(),
	}
	
	// Init audio
	g.initAudio()
	
	// Load sprites
	g.loadSprites()
	
	// Load stats
	g.loadStats()
	
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

func (g *Game) initAudio() {
	// Create audio context
	g.audioContext = audio.NewContext(44100)
	
	// Load sounds
	g.clickPlayer = g.loadSound("click.ogg")
	g.winPlayer = g.loadSound("win.ogg")
	g.losePlayer = g.loadSound("lose.ogg")
}

func (g *Game) loadSound(filename string) *audio.Player {
	// Try embedded FS first
	data, err := assetFS.ReadFile("assets/sounds/" + filename)
	if err != nil {
		// Fallback to file system
		execPath, _ := os.Getwd()
		soundPath := filepath.Join(execPath, "assets", "sounds", filename)
		data, err = os.ReadFile(soundPath)
		if err != nil {
			log.Printf("Warning: Could not load sound %s: %v", filename, err)
			return nil
		}
	}
	
	// Decode OGG
	stream, err := vorbis.DecodeWithSampleRate(44100, bytes.NewReader(data))
	if err != nil {
		log.Printf("Warning: Could not decode sound %s: %v", filename, err)
		return nil
	}
	
	player, err := g.audioContext.NewPlayer(stream)
	if err != nil {
		log.Printf("Warning: Could not create player for %s: %v", filename, err)
		return nil
	}
	
	return player
}

func (g *Game) playSound(name string) {
	switch name {
	case "click":
		if g.clickPlayer != nil {
			g.clickPlayer.Rewind()
			g.clickPlayer.Play()
		}
	case "win":
		if g.winPlayer != nil {
			g.winPlayer.Rewind()
			g.winPlayer.Play()
		}
	case "lose":
		if g.losePlayer != nil {
			g.losePlayer.Rewind()
			g.losePlayer.Play()
		}
	}
}

func (g *Game) loadSprites() {
	// Load X sprite (red gem)
	if data, err := assetFS.ReadFile("assets/sprites/gem_red.png"); err == nil {
		img, _ := pngDecode(bytes.NewReader(data))
		g.SpriteX = img
	}
	
	// Load O sprite (blue gem)
	if data, err := assetFS.ReadFile("assets/sprites/gem_blue.png"); err == nil {
		img, _ := pngDecode(bytes.NewReader(data))
		g.SpriteO = img
	}
}

func pngDecode(r *bytes.Reader) (*ebiten.Image, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func (g *Game) loadStats() {
	execPath, _ := os.Getwd()
	path := filepath.Join(execPath, "stats.json")
	
	data, err := os.ReadFile(path)
	if err != nil {
		return // File doesn't exist, use defaults
	}
	
	if err := json.Unmarshal(data, g.Stats); err != nil {
		log.Printf("Warning: Could not parse stats: %v", err)
	}
	
	// Ensure achievements exist
	if len(g.Stats.Achievements) == 0 {
		g.Stats = DefaultStats()
	}
}

func (g *Game) saveStats() {
	execPath, _ := os.Getwd()
	path := filepath.Join(execPath, "stats.json")
	
	data, err := json.MarshalIndent(g.Stats, "", "  ")
	if err != nil {
		return
	}
	
	os.WriteFile(path, data, 0644)
}

func (g *Game) checkAchievements() {
	s := g.Stats
	
	// First game
	if !g.getAchievement("first_game").Unlocked {
		g.unlockAchievement("first_game")
	}
	
	// First win
	if s.Wins >= 1 && !g.getAchievement("first_win").Unlocked {
		g.unlockAchievement("first_win")
	}
	
	// Ten wins
	if s.Wins >= 10 && !g.getAchievement("ten_wins").Unlocked {
		g.unlockAchievement("ten_wins")
	}
	
	// Win streak 3
	if s.WinStreak >= 3 && !g.getAchievement("win_streak_3").Unlocked {
		g.unlockAchievement("win_streak_3")
	}
	
	// Win streak 5
	if s.WinStreak >= 5 && !g.getAchievement("win_streak_5").Unlocked {
		g.unlockAchievement("win_streak_5")
	}
	
	// First draw
	if s.Draws >= 1 && !g.getAchievement("first_draw").Unlocked {
		g.unlockAchievement("first_draw")
	}
	
	// Ten draws
	if s.Draws >= 10 && !g.getAchievement("ten_draws").Unlocked {
		g.unlockAchievement("ten_draws")
	}
}

func (g *Game) getAchievement(id string) *Achievement {
	for i := range g.Stats.Achievements {
		if g.Stats.Achievements[i].ID == id {
			return &g.Stats.Achievements[i]
		}
	}
	return nil
}

func (g *Game) unlockAchievement(id string) {
	ach := g.getAchievement(id)
	if ach != nil && !ach.Unlocked {
		ach.Unlocked = true
		g.NewAchievement = ach
		g.AchieveTimer = 3.0
	}
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
	
	// Update achievement display timer
	if g.AchieveTimer > 0 {
		g.AchieveTimer -= 1.0 / 60.0
		if g.AchieveTimer <= 0 {
			g.NewAchievement = nil
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
				g.playSound("click")
				g.startGame()
			}
			
			// Difficulty buttons
			for i := 0; i < 3; i++ {
				dbtnX := ScreenW/2 - 160 + i*110
				dbtnY := 580
				if fmx >= float64(dbtnX) && fmx <= float64(dbtnX+100) &&
				   fmy >= float64(dbtnY) && fmy <= float64(dbtnY+45) {
					g.playSound("click")
					g.AIDifficulty = AIDifficulty(i)
				}
			}
			
		case StatePlaying:
			if g.Turn == CellX && !g.AI_Thinking {
				r, c, ok := g.Board.CellAt(fmx, fmy)
				if ok && g.Board.Grid[r][c] == CellEmpty {
					g.playSound("click")
					g.Board.Grid[r][c] = CellX
					g.CellAnims[r][c] = 0.3
					g.spawnCellParticles(r, c, CellX)
					
					if win, cells := g.Board.CheckWin(CellX); win {
						g.State = StateWin
						g.ScoreX++
						g.WinCells = cells
						g.playSound("win")
						g.spawnWinParticles(cells)
						
						// Update stats
						g.Stats.TotalGames++
						g.Stats.Wins++
						g.Stats.WinStreak++
						if g.Stats.WinStreak > g.Stats.BestStreak {
							g.Stats.BestStreak = g.Stats.WinStreak
						}
						if g.AIDifficulty == AIHard {
							g.Stats.Unbeatable = true
						}
						g.checkAchievements()
						g.saveStats()
					} else if g.Board.IsFull() {
						g.State = StateDraw
						g.Draws++
						g.Stats.TotalGames++
						g.Stats.Draws++
						g.Stats.WinStreak = 0
						g.checkAchievements()
						g.saveStats()
					} else {
						g.Turn = CellO
						g.AI_Thinking = true
						g.AITimer = 0.5
					}
				}
			}
			
		case StateAchievements:
			// Back button
			btnX := ScreenW/2 - 100
			btnY := 750
			if fmx >= float64(btnX) && fmx <= float64(btnX+200) &&
			   fmy >= float64(btnY) && fmy <= float64(btnY+60) {
				g.playSound("click")
				g.State = StateMenu
			}
			
		case StateWin, StateDraw:
			// Restart button
			btnX := ScreenW/2 - 100
			btnY := 530
			if fmx >= float64(btnX) && fmx <= float64(btnX+200) &&
			   fmy >= float64(btnY) && fmy <= float64(btnY+60) {
				g.playSound("click")
				g.startGame()
			}
			
			// Achievements button
			btnY2 := 610
			if fmx >= float64(btnX) && fmx <= float64(btnX+200) &&
			   fmy >= float64(btnY2) && fmy <= float64(btnY2+60) {
				g.playSound("click")
				g.State = StateAchievements
			}
			
			// Menu button
			btnY3 := 690
			if fmx >= float64(btnX) && fmx <= float64(btnX+200) &&
			   fmy >= float64(btnY3) && fmy <= float64(btnY3+60) {
				g.playSound("click")
				g.State = StateMenu
			}
		}
	}
	
	// AI turn
	if g.AI_Thinking {
		g.AITimer -= 1.0 / 60.0
		if g.AITimer <= 0 {
			r, c := g.Board.BestMove(g.AIDifficulty)
			if r >= 0 && c >= 0 {
				g.Board.Grid[r][c] = CellO
				g.CellAnims[r][c] = 0.3
				g.spawnCellParticles(r, c, CellO)
				
				if win, cells := g.Board.CheckWin(CellO); win {
					g.State = StateWin
					g.ScoreO++
					g.WinCells = cells
					g.playSound("lose")
					g.spawnWinParticles(cells)
					
					// Update stats
					g.Stats.TotalGames++
					g.Stats.Losses++
					g.Stats.WinStreak = 0
					g.checkAchievements()
					g.saveStats()
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
	case StateAchievements:
		g.drawAchievements(screen)
	}
	
	// Achievement notification
	if g.NewAchievement != nil && g.AchieveTimer > 0 {
		g.drawAchievementNotification(screen)
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
	ebitenutil.DebugPrintAt(screen, "КРЕСТИКИ-НОЛИКИ", ScreenW/2-140, 250)
	
	// Subtitle
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge - Day 101", ScreenW/2-110, 290)
	
	// Decorative line
	lineImg := createLineImage(300, 4, color.RGBA{100, 180, 255, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ScreenW/2-150), 320)
	screen.DrawImage(lineImg, op)
	
	// X and O decorations
	xImg := createXImage(80, color.RGBA{100, 200, 255, 200})
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(ScreenW/2-180), 340)
	screen.DrawImage(xImg, op2)
	
	oImg := createOImage(80, color.RGBA{255, 100, 150, 200})
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(float64(ScreenW/2+100), 340)
	screen.DrawImage(oImg, op3)
	
	// Start button
	g.drawButton(screen, "▶  ИГРАТЬ", ScreenW/2-100, 500, 200, 60)
	
	// Difficulty selection
	ebitenutil.DebugPrintAt(screen, "Сложность AI:", ScreenW/2-95, 560)
	
	for i := 0; i < 3; i++ {
		dbtnX := ScreenW/2 - 160 + i*110
		dbtnY := 580
		label := aiNames[i]
		
		// Highlight selected
		if AIDifficulty(i) == g.AIDifficulty {
			g.drawButtonGreen(screen, label, dbtnX, dbtnY, 100, 45)
		} else {
			g.drawButton(screen, label, dbtnX, dbtnY, 100, 45)
		}
	}
	
	// Info
	ebitenutil.DebugPrintAt(screen, "Вы: X  |  AI: O", ScreenW/2-90, 650)
	ebitenutil.DebugPrintAt(screen, "Minimax algorithm", ScreenW/2-100, 670)
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
	
	// Gems - использу спрайты если есть, иначе процедурные
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
			
			if cell == CellX || cell == CellO {
				sprite := g.SpriteX
				if cell == CellO {
					sprite = g.SpriteO
				}
				
				if sprite != nil {
					// Используем спрайт
					op := &ebiten.DrawImageOptions{}
					scale := float64(CellSize-10) / float64(sprite.Bounds().Dx()) * g.CellAnims[r][c]
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(
						float64(GridX+c*CellSize+5),
						float64(GridY+r*CellSize+5),
					)
					screen.DrawImage(sprite, op)
				} else {
					// Fallback на процедурную графику
					img := createXImage(CellSize-20, color.RGBA{100, 200, 255, 255})
					if cell == CellO {
						img = createOImage(CellSize-20, color.RGBA{255, 100, 150, 255})
					}
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(
						float64(GridX+c*CellSize+10),
						float64(GridY+r*CellSize+10),
					)
					op.GeoM.Scale(g.CellAnims[r][c], g.CellAnims[r][c])
					screen.DrawImage(img, op)
				}
			}
		}
	}
	
	// Win line highlight
	if g.State == StateWin && len(g.WinCells) > 0 {
		// Animated glow
		glowAlpha := uint8(100 + 80*math.Sin(g.GameTime*5))
		
		// Draw line through winning cells
		if len(g.WinCells) == 3 {
			x1 := float64(GridX + g.WinCells[0][1]*CellSize + CellSize/2)
			y1 := float64(GridY + g.WinCells[0][0]*CellSize + CellSize/2)
			x2 := float64(GridX + g.WinCells[2][1]*CellSize + CellSize/2)
			y2 := float64(GridY + g.WinCells[2][0]*CellSize + CellSize/2)
			
			// Thick glowing line
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 8, color.RGBA{255, 215, 0, glowAlpha}, false)
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 4, color.RGBA{255, 255, 255, 200}, false)
		}
		
		// Confetti particles for winning cells
		if rand.Float64() < 0.3 {
			cell := g.WinCells[rand.Intn(len(g.WinCells))]
			cx := float64(GridX + cell[1]*CellSize + CellSize/2)
			cy := float64(GridY + cell[0]*CellSize + CellSize/2)
			
			g.Particles = append(g.Particles, Particle{
				X: cx + (rand.Float64()-0.5)*float64(CellSize),
				Y: cy,
				VX: (rand.Float64() - 0.5) * 4,
				VY: -2 - rand.Float64()*3,
				Life: 1.0,
				MaxLife: 1.0,
				Color: []color.RGBA{
					{255, 215, 0, 255},
					{255, 100, 150, 255},
					{100, 200, 255, 255},
					{100, 255, 100, 255},
				}[rand.Intn(4)],
				Size: 3 + rand.Float64()*4,
			})
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
	overlay := ebiten.NewImage(ScreenW, 400)
	overlay.Fill(color.RGBA{10, 12, 25, 220})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 380)
	screen.DrawImage(overlay, op)
	
	// Result text
	if g.State == StateWin {
		wonX, _ := g.Board.CheckWin(CellX)
		if wonX {
			ebitenutil.DebugPrintAt(screen, "🎉 ВЫ ПОБЕДИЛИ!", ScreenW/2-120, 420)
		} else {
			ebitenutil.DebugPrintAt(screen, "😔 AI ПОБЕДИЛ", ScreenW/2-110, 420)
		}
	} else {
		ebitenutil.DebugPrintAt(screen, "🤝 НИЧЬЯ!", ScreenW/2-80, 420)
	}
	
	// Stats
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Игр: %d | Побед: %d | Поражений: %d | Ничьих: %d",
		g.Stats.TotalGames, g.Stats.Wins, g.Stats.Losses, g.Stats.Draws), 60, 460)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Лучшая серия: %d | Winrate: %.0f%%",
		g.Stats.BestStreak, winrate(g.Stats)), 120, 485)
	
	// Buttons
	g.drawButton(screen, "🔄  ЕЩЁ РАЗ", ScreenW/2-100, 530, 200, 60)
	g.drawButton(screen, "🏆 ДОСТИЖЕНИЯ", ScreenW/2-100, 610, 200, 60)
	g.drawButton(screen, "←  МЕНЮ", ScreenW/2-100, 690, 200, 60)
}

func winrate(s *PlayerStats) float64 {
	if s.TotalGames == 0 {
		return 0
	}
	return float64(s.Wins) / float64(s.TotalGames) * 100
}

func (g *Game) drawAchievements(screen *ebiten.Image) {
	// Background
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{15, 18, 35, 255}, false)
	
	// Title
	ebitenutil.DebugPrintAt(screen, "🏆 ДОСТИЖЕНИЯ", ScreenW/2-110, 40)
	
	// Stats summary
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Игр: %d | Побед: %d | Winrate: %.0f%% | Серия: %d",
		g.Stats.TotalGames, g.Stats.Wins, winrate(g.Stats), g.Stats.BestStreak), 100, 80)
	
	// Achievements list
	unlocked := 0
	for i, ach := range g.Stats.Achievements {
		y := 130 + i*70
		
		// Background
		bgColor := color.RGBA{30, 35, 60, 200}
		if ach.Unlocked {
			bgColor = color.RGBA{40, 60, 40, 220}
			unlocked++
		}
		
		panel := ebiten.NewImage(ScreenW-40, 60)
		vector.DrawFilledRect(panel, 0, 0, float32(ScreenW-40), 60, bgColor, false)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(20, float64(y))
		screen.DrawImage(panel, op)
		
		// Icon and text
		icon := "🔒"
		if ach.Unlocked {
			icon = ach.Icon
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s %s", icon, ach.Name), 40, y+8)
		ebitenutil.DebugPrintAt(screen, ach.Description, 80, y+32)
	}
	
	// Counter
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/%d разблокировано", unlocked, len(g.Stats.Achievements)), ScreenW/2-100, 130+len(g.Stats.Achievements)*70+20)
	
	// Back button
	g.drawButton(screen, "← НАЗАД", ScreenW/2-100, 750, 200, 60)
}

func (g *Game) drawAchievementNotification(screen *ebiten.Image) {
	if g.NewAchievement == nil {
		return
	}
	
	alpha := 255
	if g.AchieveTimer < 0.5 {
		alpha = int(g.AchieveTimer / 0.5 * 255)
	}
	
	// Notification panel
	panel := ebiten.NewImage(350, 70)
	panel.Fill(color.RGBA{40, 80, 40, uint8(alpha)})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ScreenW/2-175), 20)
	screen.DrawImage(panel, op)
	
	// Border
	border := ebiten.NewImage(350, 3)
	border.Fill(color.RGBA{100, 255, 100, uint8(alpha)})
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(ScreenW/2-175), 20)
	screen.DrawImage(border, op2)
	
	// Text
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🏆 %s: %s", g.NewAchievement.Icon, g.NewAchievement.Name), ScreenW/2-150, 28)
	ebitenutil.DebugPrintAt(screen, g.NewAchievement.Description, ScreenW/2-150, 52)
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

func (g *Game) drawButtonGreen(screen *ebiten.Image, text string, x, y, w, h int) {
	btn := ebiten.NewImage(w, h)
	
	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)
	hover := fmx >= float64(x) && fmx <= float64(x+w) && fmy >= float64(y) && fmy <= float64(y+h)
	
	if hover {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{60, 140, 80, 255}, false)
	} else {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{40, 110, 60, 255}, false)
	}
	
	border := ebiten.NewImage(w, 3)
	border.Fill(color.RGBA{100, 255, 140, 255})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(border, op2)
	
	ebitenutil.DebugPrintAt(screen, text, x+15, y+h/2-10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Крестики-Нолики - Go365 Day 101")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
