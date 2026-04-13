package main

import (
	"time"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 800
	GameDuration = 60 // seconds

	// Animation durations
	swapDuration  = 150 * time.Millisecond
	shakeDuration = 150 * time.Millisecond
	removeDuration = 150 * time.Millisecond
	fallDuration  = 200 * time.Millisecond

	// Hint idle time (5 seconds)
	HintIdleTime = 5 * time.Second
)

// Game holds all game state.
type Game struct {
	Board          *Board
	Score          int
	HighScore      int
	TimeRemaining  float64 // seconds
	Elapsed        time.Duration
	Running        bool
	GameOver       bool
	Selected       *Tile
	SelectionTime  time.Duration
	IdleTime       time.Duration
	HintActive     bool
	HintPositions  [][2]int
	Anim           AnimationManager
	Input          InputState
	Sound          *SoundManager

	// Layout
	CellSize    float64
	BoardOffsetX float64
	BoardOffsetY float64

	// UI button bounds
	BtnX, BtnY, BtnW, BtnH int

	// Cascade resolution
	PendingResolve   bool
	CascadeDepth     int
	MaxCascadeDepth  int

	// Keyboard
	RPressed bool
	PPressed bool
	Paused   bool
}

// NewGame creates a new game instance.
func NewGame() *Game {
	g := &Game{
		TimeRemaining:  GameDuration,
		Running:        true,
		MaxCascadeDepth: 10,
		Sound:          NewSoundManager(),
	}
	g.reset()
	return g
}

// reset initializes or re-initializes the game.
func (g *Game) reset() {
	g.Board = NewBoard()
	g.Score = 0
	g.TimeRemaining = GameDuration
	g.Elapsed = 0
	g.Running = true
	g.GameOver = false
	g.Selected = nil
	g.IdleTime = 0
	g.HintActive = false
	g.Anim = AnimationManager{}
	g.PendingResolve = false
	g.CascadeDepth = 0
	g.Paused = false

	// Calculate board layout
	g.CellSize = 60.0
	boardPixelSize := g.CellSize * BoardCols
	g.BoardOffsetX = (ScreenWidth - boardPixelSize) / 2
	g.BoardOffsetY = 80.0 // Leave room for UI at top

	// Load high score (simplified — from file)
	g.loadHighScore()
}

// Update is called every frame (1/60s by default).
func (g *Game) Update() error {
	if g.GameOver {
		// Allow restart
		if g.RPressed {
			g.reset()
			g.RPressed = false
		}
		return nil
	}

	if g.Paused {
		return nil
	}

	dt := 1.0 / 60.0 // Assume 60 FPS
	g.Elapsed += time.Duration(dt * float64(time.Second))

	// Timer countdown
	g.TimeRemaining -= dt
	if g.TimeRemaining <= 0 {
		g.TimeRemaining = 0
		g.gameOver()
		return nil
	}

	// Update input
	UpdateInput(g)

	// Handle keyboard shortcuts
	if g.RPressed {
		g.reset()
		g.RPressed = false
		return nil
	}
	if g.PPressed {
		g.Paused = !g.Paused
		g.PPressed = false
	}

	// Handle clicks
	if g.Input.JustClicked && !g.Anim.IsAnimating() {
		HandleClick(g)
	}

	// Idle time for hints
	if g.Selected == nil && !g.Anim.IsAnimating() {
		g.IdleTime += time.Duration(dt * float64(time.Second))
		if g.IdleTime >= HintIdleTime && !g.HintActive {
			g.showHint()
		}
	} else {
		g.IdleTime = 0
		g.HintActive = false
	}

	// Update hint pulse animation
	if g.HintActive {
		for _, pos := range g.HintPositions {
			tile := g.Board.Grid[pos[0]][pos[1]]
			if tile != nil {
				pulse := 1.0 + 0.1*float64(time.Now().UnixNano()%1000000)/1000000
				tile.Scale = pulse
			}
		}
	}

	// Update animations
	animDone := g.Anim.Update()

	// Resolve cascade after swap animation
	if g.PendingResolve && animDone {
		g.resolveCascade()
		g.PendingResolve = false
	}

	return nil
}

// resolveCascade handles the match-remove-drop cycle.
func (g *Game) resolveCascade() {
	g.CascadeDepth = 0

	for g.CascadeDepth < g.MaxCascadeDepth {
		matches := g.Board.findMatches()
		if len(matches) == 0 {
			break
		}

		// Calculate score
		chainBonus := g.CascadeDepth * 5
		for range matches {
			g.Score += 10
		}

		// Bonus for 4+ tiles in a single match group
		if len(matches) >= 4 {
			g.Score += 50 + chainBonus
		}
		if len(matches) >= 5 {
			g.Score += 100 + chainBonus
		}

		g.PlaySound(SoundMatch)

		// Animate removal
		matchTiles := make([]*Tile, 0, len(matches))
		for tile := range matches {
			matchTiles = append(matchTiles, tile)
		}

		g.Anim.Start(NewRemoveAnimation(matchTiles, removeDuration))
		// Wait for removal animation (simplified: just continue)
		for _, t := range matchTiles {
			t.Removing = true
		}

		// Remove tiles from grid
		for _, t := range matchTiles {
			g.Board.Grid[t.Row][t.Col] = nil
		}

		// Gravity: drop tiles
		g.Board.gravity()

		// Animate fall
		fallingTiles := make([]*Tile, 0)
		for r := 0; r < BoardRows; r++ {
			for c := 0; c < BoardCols; c++ {
				t := g.Board.Grid[r][c]
				if t != nil && t.Falling {
					fallingTiles = append(fallingTiles, t)
				}
			}
		}

		if len(fallingTiles) > 0 {
			g.Anim.Start(NewFallAnimation(fallingTiles, fallDuration))
		}

		g.CascadeDepth++
	}

	// Check if there are possible moves; if not, reshuffle
	if !g.Board.hasPossibleMoves() {
		g.Board = NewBoard()
	}
}

// showHint finds and highlights a valid move.
func (g *Game) showHint() {
	move := g.Board.findHintMove()
	if len(move) >= 2 {
		g.HintPositions = [][2]int{move[0], move[1]}
		g.HintActive = true
	}
}

// gameOver handles game end.
func (g *Game) gameOver() {
	g.GameOver = true
	g.PlaySound(SoundGameOver)
	if g.Score > g.HighScore {
		g.HighScore = g.Score
		g.saveHighScore()
	}
}

// PlaySound plays a sound effect.
func (g *Game) PlaySound(st SoundType) {
	if g.Sound != nil {
		g.Sound.Play(st)
	}
}

// loadHighScore loads the high score from a file.
func (g *Game) loadHighScore() {
	// Simplified — in production use JSON file or local storage
	g.HighScore = 0
}

// saveHighScore saves the high score to a file.
func (g *Game) saveHighScore() {
	// Simplified — in production use JSON file or local storage
}
