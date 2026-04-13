package main

import (
	"testing"
)

func TestNewBoardNoInitialMatches(t *testing.T) {
	b := NewBoard()
	matches := b.findMatches()
	if len(matches) > 0 {
		t.Errorf("New board has %d initial matches, expected 0", len(matches))
	}
}

func TestBoardDimensions(t *testing.T) {
	b := NewBoard()
	if len(b.Grid) != BoardRows {
		t.Errorf("Board has %d rows, expected %d", len(b.Grid), BoardRows)
	}
	for r := 0; r < BoardRows; r++ {
		if len(b.Grid[r]) != BoardCols {
			t.Errorf("Row %d has %d cols, expected %d", r, len(b.Grid[r]), BoardCols)
		}
	}
}

func TestSwapTiles(t *testing.T) {
	b := NewBoard()
	// Swap two adjacent tiles
	b.swapTiles(0, 0, 0, 1)
	if b.Grid[0][0] == nil || b.Grid[0][1] == nil {
		t.Fatal("After swap, tiles should not be nil")
	}
	// Verify row/col updated
	if b.Grid[0][0].Row != 0 || b.Grid[0][0].Col != 0 {
		t.Errorf("Tile at (0,0) has wrong position: (%d,%d)", b.Grid[0][0].Row, b.Grid[0][0].Col)
	}
}

func TestAreAdjacent(t *testing.T) {
	tests := []struct {
		r1, c1, r2, c2 int
		expected       bool
	}{
		{0, 0, 0, 1, true},   // right
		{0, 0, 1, 0, true},   // down
		{0, 0, 0, 2, false},  // not adjacent
		{0, 0, 1, 1, false},  // diagonal
		{3, 4, 3, 3, true},   // left
		{5, 5, 4, 5, true},   // up
	}
	for _, tc := range tests {
		result := areAdjacent(tc.r1, tc.c1, tc.r2, tc.c2)
		if result != tc.expected {
			t.Errorf("areAdjacent(%d,%d,%d,%d) = %v, expected %v",
				tc.r1, tc.c1, tc.r2, tc.c2, result, tc.expected)
		}
	}
}

func TestFindMatchesHorizontal(t *testing.T) {
	b := &Board{
		Grid: make([][]*Tile, BoardRows),
	}
	for r := 0; r < BoardRows; r++ {
		b.Grid[r] = make([]*Tile, BoardCols)
	}

	// Create a horizontal match: row 0, cols 0-2 same color
	for c := 0; c < 3; c++ {
		b.Grid[0][c] = &Tile{Color: 0, Row: 0, Col: c}
	}
	matches := b.findMatches()
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}
}

func TestFindMatchesVertical(t *testing.T) {
	b := &Board{
		Grid: make([][]*Tile, BoardRows),
	}
	for r := 0; r < BoardRows; r++ {
		b.Grid[r] = make([]*Tile, BoardCols)
	}

	// Create a vertical match: col 0, rows 0-2 same color
	for r := 0; r < 3; r++ {
		b.Grid[r][0] = &Tile{Color: 1, Row: r, Col: 0}
	}
	matches := b.findMatches()
	if len(matches) != 3 {
		t.Errorf("Expected 3 vertical matches, got %d", len(matches))
	}
}

func TestGravity(t *testing.T) {
	b := &Board{
		Grid: make([][]*Tile, BoardRows),
	}
	for r := 0; r < BoardRows; r++ {
		b.Grid[r] = make([]*Tile, BoardCols)
	}

	// Place tile at bottom
	b.Grid[7][0] = &Tile{Color: 0, Row: 7, Col: 0}
	// Remove it and let gravity work
	b.Grid[7][0] = nil
	b.gravity()
	// After gravity, column 0 should have nil at top
	newTiles := b.gravity()
	_ = newTiles
}

func TestHasPossibleMoves(t *testing.T) {
	b := NewBoard()
	// A new board should usually have possible moves
	_ = b.hasPossibleMoves()
	// We can't guarantee true, but it shouldn't panic
}
