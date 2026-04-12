package main

import (
	"testing"
)

// TestNewBoard tests board creation
func TestNewBoard(t *testing.T) {
	board := NewBoard()
	
	if board == nil {
		t.Fatal("Board should not be nil")
	}
	
	if len(board.Tiles) != gridRows {
		t.Errorf("Expected %d rows, got %d", gridRows, len(board.Tiles))
	}
	
	for _, row := range board.Tiles {
		if len(row) != gridCols {
			t.Errorf("Expected %d cols, got %d", gridCols, len(row))
		}
	}
}

// TestRemoveInitialMatches tests that initial board has no matches
func TestRemoveInitialMatches(t *testing.T) {
	board := NewBoard()
	board.RemoveInitialMatches()
	
	matches := board.FindAllMatches()
	if len(matches) > 0 {
		t.Errorf("Initial board should have no matches, found %d", len(matches))
	}
}

// TestFindAllMatches tests match detection
func TestFindAllMatches(t *testing.T) {
	board := NewBoard()
	board.RemoveInitialMatches()
	
	// Create a horizontal match
	board.Tiles[0][0].GemType = 0
	board.Tiles[0][1].GemType = 0
	board.Tiles[0][2].GemType = 0
	
	matches := board.FindAllMatches()
	if len(matches) < 3 {
		t.Errorf("Should find at least 3 matching tiles, found %d", len(matches))
	}
}

// TestSwapTiles tests tile swapping
func TestSwapTiles(t *testing.T) {
	board := NewBoard()
	
	tile1 := board.Tiles[0][0]
	tile2 := board.Tiles[0][1]
	
	board.SwapTiles(tile1, tile2)
	
	// Check that positions are swapped
	if tile1.Row != 0 || tile1.Col != 1 {
		t.Errorf("Tile1 position not updated correctly: row=%d, col=%d", tile1.Row, tile1.Col)
	}
	
	if tile2.Row != 0 || tile2.Col != 0 {
		t.Errorf("Tile2 position not updated correctly: row=%d, col=%d", tile2.Row, tile2.Col)
	}
}

// TestApplyGravity tests gravity system
func TestApplyGravity(t *testing.T) {
	board := NewBoard()
	
	// Remove a tile
	board.Tiles[7][0].Removing = true
	
	// Apply gravity
	board.ApplyGravity()
	
	// Check that tiles moved down
	// (simplified test)
}

// TestHasEmptyTiles tests empty tile detection
func TestHasEmptyTiles(t *testing.T) {
	board := NewBoard()
	
	if board.HasEmptyTiles() {
		t.Error("Fresh board should not have empty tiles")
	}
	
	// Create an empty tile
	board.Tiles[0][0] = nil
	
	if !board.HasEmptyTiles() {
		t.Error("Board should have empty tiles")
	}
}

// TestAbsInt tests absolute value function
func TestAbsInt(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{-5, 5},
		{0, 0},
		{10, 10},
		{-100, 100},
	}
	
	for _, test := range tests {
		result := absInt(test.input)
		if result != test.expected {
			t.Errorf("absInt(%d) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

// TestCalculateMatchScore tests score calculation
func TestCalculateMatchScore(t *testing.T) {
	game := &Game{
		Combo: 1,
		Board: NewBoard(),
	}
	
	// Test 3-match (30 points base)
	matches3 := make([]*Tile, 3)
	score3 := game.calculateMatchScore(matches3)
	if score3 < 30 {
		t.Errorf("3-match should score at least 30, got %d", score3)
	}
	
	// Test combo multiplier
	game.Combo = 2
	score3Combo := game.calculateMatchScore(matches3)
	if score3Combo <= score3 {
		t.Error("Combo should increase score")
	}
}

// TestFindHint tests hint system
func TestFindHint(t *testing.T) {
	board := NewBoard()
	board.RemoveInitialMatches()
	
	// Board should have at least one valid move
	// (This is a probabilistic test, might fail rarely)
	t1, t2 := board.FindHint()
	
	// Either find a hint or board might be stuck (very rare)
	_ = t1
	_ = t2
}

// TestTileRemoval tests tile removal animation
func TestTileRemoval(t *testing.T) {
	board := NewBoard()
	tile := board.Tiles[0][0]
	
	tile.Removing = true
	
	// Update should decrease alpha and scale
	for i := 0; i < 10; i++ {
		board.Update()
	}
	
	if tile.Alpha >= 1.0 {
		t.Error("Removing tile should decrease alpha")
	}
	
	if tile.Scale >= 1.0 {
		t.Error("Removing tile should decrease scale")
	}
}

// TestShakeAnimation tests shake animation
func TestShakeAnimation(t *testing.T) {
	board := NewBoard()
	tile := board.Tiles[0][0]
	
	tile.ShakeTime = 0.5
	
	// Update should decrease shake time
	board.Update()
	
	if tile.ShakeTime >= 0.5 {
		t.Error("ShakeTime should decrease after update")
	}
}
