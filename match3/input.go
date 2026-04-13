package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputState tracks current input events.
type InputState struct {
	CurX, CurY       float64
	JustClicked      bool
	HoverRow, HoverCol int
	Hovering         bool
}

// UpdateInput reads mouse/touch state and updates input state.
func UpdateInput(g *Game) {
	// Check if any touch is active
	touchIDs := ebiten.TouchIDs()
	if len(touchIDs) > 0 {
		x, y := ebiten.TouchPosition(touchIDs[0])
		g.Input.CurX = float64(x)
		g.Input.CurY = float64(y)
		// Simplified touch handling
		g.Input.JustClicked = true
	} else {
		// Fallback to mouse
		mx, my := ebiten.CursorPosition()
		g.Input.CurX = float64(mx)
		g.Input.CurY = float64(my)
		g.Input.JustClicked = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	}

	// Convert pixel to grid position
	if g.Input.CurX >= g.BoardOffsetX && g.Input.CurX < g.BoardOffsetX+g.CellSize*BoardCols &&
		g.Input.CurY >= g.BoardOffsetY && g.Input.CurY < g.BoardOffsetY+g.CellSize*BoardRows {
		g.Input.HoverRow = int((g.Input.CurY - g.BoardOffsetY) / g.CellSize)
		g.Input.HoverCol = int((g.Input.CurX - g.BoardOffsetX) / g.CellSize)
		g.Input.Hovering = true
	} else {
		g.Input.Hovering = false
	}
}

// HandleClick processes a click on the board.
func HandleClick(g *Game) {
	if !g.Input.JustClicked || !g.Input.Hovering {
		return
	}

	r, c := g.Input.HoverRow, g.Input.HoverCol

	// Check if click is on the New Game button
	if g.isNewGameButtonClicked(g.Input.CurX, g.Input.CurY) {
		g.reset()
		return
	}

	// During animations or game over — ignore board clicks
	if g.Anim.IsAnimating() || g.GameOver {
		return
	}

	tile := g.Board.Grid[r][c]
	if tile == nil {
		return
	}

	// First selection
	if g.Selected == nil {
		g.Selected = tile
		g.SelectionTime = g.Elapsed
		return
	}

	// Clicked same tile — deselect
	if g.Selected == tile {
		g.Selected = nil
		return
	}

	// Check adjacency
	sr, sc := g.Selected.Row, g.Selected.Col
	if areAdjacent(sr, sc, r, c) {
		g.trySwap(sr, sc, r, c)
		g.Selected = nil
	} else {
		// Select new tile instead
		g.Selected = tile
		g.SelectionTime = g.Elapsed
	}
}

// trySwap attempts to swap two tiles and validates the move.
func (g *Game) trySwap(r1, c1, r2, c2 int) {
	g.Board.swapTiles(r1, c1, r2, c2)

	t1 := g.Board.Grid[r1][c1]
	t2 := g.Board.Grid[r2][c2]

	// Check if swap creates a match
	matches := g.Board.findMatches()
	if len(matches) > 0 {
		// Valid move!
		g.Anim.Start(NewSwapAnimation(t1, t2, swapDuration))
		g.PlaySound(SoundSwap)
		g.HintActive = false
		g.IdleTime = 0

		// After swap animation, resolve matches
		g.PendingResolve = true
	} else {
		// Invalid move — swap back with shake
		g.Board.swapTiles(r1, c1, r2, c2)
		g.Anim.Start(NewShakeAnimation([]*Tile{g.Board.Grid[r1][c1], g.Board.Grid[r2][c2]}, shakeDuration))
		g.PlaySound(SoundError)
	}
}
