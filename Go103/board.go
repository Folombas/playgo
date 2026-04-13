package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	BoardRows    = 8
	BoardCols    = 8
	NumTileTypes = 6
)

// Tile represents a single gem on the board.
type Tile struct {
	Color     int     // 0..5
	Row       int     // grid row
	Col       int     // grid col
	X         float64 // pixel X (for animation)
	Y         float64 // pixel Y (for animation)
	Scale     float64 // animation scale (1.0 = normal)
	Alpha     float64 // animation alpha (1.0 = fully visible)
	Removing  bool    // flag: tile is being removed
	Falling   bool    // flag: tile is falling
	FallStart float64 // starting Y for fall animation
}

// Board represents the 8x8 game grid.
type Board struct {
	Grid [][]*Tile // Grid[row][col], nil = empty cell
}

// NewBoard creates and initializes a board with no pre-existing matches.
func NewBoard() *Board {
	b := &Board{
		Grid: make([][]*Tile, BoardRows),
	}
	for r := 0; r < BoardRows; r++ {
		b.Grid[r] = make([]*Tile, BoardCols)
	}
	b.fillNoMatches()
	return b
}

// fillNoMatches fills the board ensuring no initial matches exist.
func (b *Board) fillNoMatches() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for r := 0; r < BoardRows; r++ {
		for c := 0; c < BoardCols; c++ {
			var color int
			for {
				color = rng.Intn(NumTileTypes)
				b.Grid[r][c] = &Tile{Color: color, Row: r, Col: c, Scale: 1.0, Alpha: 1.0}
				if !b.wouldMatchAt(r, c) {
					break
				}
			}
		}
	}
}

// wouldMatchAt checks if placing a tile at (row, col) would create a match of 3+.
func (b *Board) wouldMatchAt(row, col int) bool {
	t := b.Grid[row][col]
	if t == nil {
		return false
	}

	// Check horizontal
	countH := 1
	// Left
	for c := col - 1; c >= 0; c-- {
		if b.Grid[row][c] != nil && b.Grid[row][c].Color == t.Color {
			countH++
		} else {
			break
		}
	}
	// Right
	for c := col + 1; c < BoardCols; c++ {
		if b.Grid[row][c] != nil && b.Grid[row][c].Color == t.Color {
			countH++
		} else {
			break
		}
	}
	if countH >= 3 {
		return true
	}

	// Check Vertical
	countV := 1
	// Up
	for r := row - 1; r >= 0; r-- {
		if b.Grid[r][col] != nil && b.Grid[r][col].Color == t.Color {
			countV++
		} else {
			break
		}
	}
	// Down
	for r := row + 1; r < BoardRows; r++ {
		if b.Grid[r][col] != nil && b.Grid[r][col].Color == t.Color {
			countV++
		} else {
			break
		}
	}
	return countV >= 3
}

// findMatches returns a list of all tiles that are part of a match (3+ in a row/col).
func (b *Board) findMatches() map[*Tile]bool {
	matched := make(map[*Tile]bool)

	// Horizontal matches
	for r := 0; r < BoardRows; r++ {
		for c := 0; c < BoardCols-2; c++ {
			t1 := b.Grid[r][c]
			t2 := b.Grid[r][c+1]
			t3 := b.Grid[r][c+2]
			if t1 != nil && t2 != nil && t3 != nil && !t1.Removing && !t2.Removing && !t3.Removing {
				if t1.Color == t2.Color && t2.Color == t3.Color {
					matched[t1] = true
					matched[t2] = true
					matched[t3] = true
					// Extend match further
					for cc := c + 3; cc < BoardCols; cc++ {
						t := b.Grid[r][cc]
						if t != nil && !t.Removing && t.Color == t1.Color {
							matched[t] = true
						} else {
							break
						}
					}
				}
			}
		}
	}

	// Vertical matches
	for c := 0; c < BoardCols; c++ {
		for r := 0; r < BoardRows-2; r++ {
			t1 := b.Grid[r][c]
			t2 := b.Grid[r+1][c]
			t3 := b.Grid[r+2][c]
			if t1 != nil && t2 != nil && t3 != nil && !t1.Removing && !t2.Removing && !t3.Removing {
				if t1.Color == t2.Color && t2.Color == t3.Color {
					matched[t1] = true
					matched[t2] = true
					matched[t3] = true
					for rr := r + 3; rr < BoardRows; rr++ {
						t := b.Grid[rr][c]
						if t != nil && !t.Removing && t.Color == t1.Color {
							matched[t] = true
						} else {
							break
						}
					}
				}
			}
		}
	}

	return matched
}

// swapTiles exchanges two tiles on the board (updates row/col references).
func (b *Board) swapTiles(row1, col1, row2, col2 int) {
	t1 := b.Grid[row1][col1]
	t2 := b.Grid[row2][col2]
	b.Grid[row1][col1] = t2
	b.Grid[row2][col2] = t1
	if t1 != nil {
		t1.Row, t1.Col = row2, col2
	}
	if t2 != nil {
		t2.Row, t2.Col = row1, col1
	}
}

// areAdjacent checks if two positions are horizontally or vertically adjacent.
func areAdjacent(r1, c1, r2, c2 int) bool {
	dr := r1 - r2
	dc := c1 - c2
	return (dr == 0 && (dc == 1 || dc == -1)) || (dc == 0 && (dr == 1 || dr == -1))
}

// gravity drops tiles down to fill empty spaces and returns the number of new tiles created.
func (b *Board) gravity() int {
	newTiles := 0
	for c := 0; c < BoardCols; c++ {
		emptySlots := 0
		for r := BoardRows - 1; r >= 0; r-- {
			if b.Grid[r][c] == nil {
				emptySlots++
			} else if emptySlots > 0 {
				t := b.Grid[r][c]
				b.Grid[r+emptySlots][c] = t
				b.Grid[r][c] = nil
				t.Row = r + emptySlots
				t.Falling = true
				t.FallStart = float64(r)
			}
		}
		// Fill top empty slots with new tiles
		for r := 0; r < emptySlots; r++ {
			if b.Grid[r][c] == nil {
				color := rand.Intn(NumTileTypes)
				t := &Tile{
					Color:     color,
					Row:       r,
					Col:       c,
					Scale:     1.0,
					Alpha:     1.0,
					Falling:   true,
					FallStart: float64(r - emptySlots),
				}
				b.Grid[r][c] = t
				newTiles++
			}
		}
	}
	return newTiles
}

// hasPossibleMoves checks if there's any valid swap that creates a match.
func (b *Board) hasPossibleMoves() bool {
	for r := 0; r < BoardRows; r++ {
		for c := 0; c < BoardCols; c++ {
			// Try swap right
			if c+1 < BoardCols {
				b.swapTiles(r, c, r, c+1)
				if len(b.findMatches()) > 0 {
					b.swapTiles(r, c, r, c+1) // undo
					return true
				}
				b.swapTiles(r, c, r, c+1) // undo
			}
			// Try swap down
			if r+1 < BoardRows {
				b.swapTiles(r, c, r+1, c)
				if len(b.findMatches()) > 0 {
					b.swapTiles(r, c, r+1, c) // undo
					return true
				}
				b.swapTiles(r, c, r+1, c) // undo
			}
		}
	}
	return false
}

// findHintMove returns a pair of positions that would create a match, or nil if none found.
func (b *Board) findHintMove() [][2]int {
	for r := 0; r < BoardRows; r++ {
		for c := 0; c < BoardCols; c++ {
			if c+1 < BoardCols {
				b.swapTiles(r, c, r, c+1)
				if len(b.findMatches()) > 0 {
					b.swapTiles(r, c, r, c+1)
					return [][2]int{{r, c}, {r, c + 1}}
				}
				b.swapTiles(r, c, r, c+1)
			}
			if r+1 < BoardRows {
				b.swapTiles(r, c, r+1, c)
				if len(b.findMatches()) > 0 {
					b.swapTiles(r, c, r+1, c)
					return [][2]int{{r, c}, {r + 1, c}}
				}
				b.swapTiles(r, c, r+1, c)
			}
		}
	}
	return nil
}

// String returns a debug representation of the board.
func (b *Board) String() string {
	s := ""
	for r := 0; r < BoardRows; r++ {
		for c := 0; c < BoardCols; c++ {
			if b.Grid[r][c] != nil {
				s += fmt.Sprintf("%d ", b.Grid[r][c].Color)
			} else {
				s += ". "
			}
		}
		s += "\n"
	}
	return s
}
