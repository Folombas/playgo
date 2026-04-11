package logic

import (
	"testing"
)

// TestNewBoard tests board creation
func TestNewBoard(t *testing.T) {
	board := NewBoard(8, 8)
	
	if board == nil {
		t.Fatal("Board should not be nil")
	}
	
	if board.Rows != 8 {
		t.Errorf("Expected 8 rows, got %d", board.Rows)
	}
	
	if board.Cols != 8 {
		t.Errorf("Expected 8 cols, got %d", board.Cols)
	}
	
	if len(board.Tiles) != 8 {
		t.Errorf("Expected 8 rows in Tiles, got %d", len(board.Tiles))
	}
	
	if board.ImageCache == nil {
		t.Error("ImageCache should be initialized")
	}
}

// TestFindAllMatches tests match detection
func TestFindAllMatches(t *testing.T) {
	board := NewBoard(8, 8)
	
	// Create a horizontal match of 3
	board.Tiles[0][0].Gem = GemRed
	board.Tiles[0][1].Gem = GemRed
	board.Tiles[0][2].Gem = GemRed
	
	matches := board.FindAllMatches()
	
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}
}

// TestFindAllMatches_Vertical tests vertical match detection
func TestFindAllMatches_Vertical(t *testing.T) {
	board := NewBoard(8, 8)
	
	// Create a vertical match of 4
	board.Tiles[0][0].Gem = GemBlue
	board.Tiles[1][0].Gem = GemBlue
	board.Tiles[2][0].Gem = GemBlue
	board.Tiles[3][0].Gem = GemBlue
	
	matches := board.FindAllMatches()
	
	// Should find at least 4 matches (could find more due to random board)
	if len(matches) < 4 {
		t.Errorf("Expected at least 4 matches for vertical line, got %d", len(matches))
	}
}

// TestSwapTiles tests tile swapping
func TestSwapTiles(t *testing.T) {
	board := NewBoard(8, 8)
	
	tile1 := board.Tiles[0][0]
	tile2 := board.Tiles[0][1]
	
	gem1 := tile1.Gem
	gem2 := tile2.Gem
	
	success := board.SwapTiles(tile1, tile2)
	
	if !success {
		t.Error("Swap should succeed for adjacent tiles")
	}
	
	if board.Tiles[0][0].Gem != gem2 {
		t.Errorf("Expected tile1 to have gem2, got %d", board.Tiles[0][0].Gem)
	}
	
	if board.Tiles[0][1].Gem != gem1 {
		t.Errorf("Expected tile2 to have gem1, got %d", board.Tiles[0][1].Gem)
	}
}

// TestSwapTiles_NonAdjacent tests that non-adjacent tiles cannot be swapped
func TestSwapTiles_NonAdjacent(t *testing.T) {
	board := NewBoard(8, 8)
	
	tile1 := board.Tiles[0][0]
	tile2 := board.Tiles[2][2]
	
	success := board.SwapTiles(tile1, tile2)
	
	if success {
		t.Error("Swap should fail for non-adjacent tiles")
	}
}

// TestHasValidMoves tests move validation
func TestHasValidMoves(t *testing.T) {
	board := NewBoard(8, 8)
	
	// Board should have valid moves after initialization
	hasMoves := board.HasValidMoves()
	
	// This could be true or false depending on random generation
	// Just verify it doesn't panic
	t.Logf("HasValidMoves: %v", hasMoves)
}

// TestFindHint tests hint system
func TestFindHint(t *testing.T) {
	board := NewBoard(8, 8)
	
	tile1, tile2 := board.FindHint()
	
	// If hint is found, verify tiles are adjacent
	if tile1 != nil && tile2 != nil {
		dr := tile1.Row - tile2.Row
		dc := tile1.Col - tile2.Col
		
		if dr < 0 {
			dr = -dr
		}
		if dc < 0 {
			dc = -dc
		}
		
		if dr+dc != 1 {
			t.Errorf("Hint tiles should be adjacent, got dr=%d, dc=%d", dr, dc)
		}
	}
}

// TestRemoveMatches tests match removal
func TestRemoveMatches(t *testing.T) {
	board := NewBoard(8, 8)
	
	// Create a match
	board.Tiles[0][0].Gem = GemGreen
	board.Tiles[0][1].Gem = GemGreen
	board.Tiles[0][2].Gem = GemGreen
	
	score, bombCreated := board.RemoveMatches()
	
	if score == 0 {
		t.Error("Score should not be 0 for a match")
	}
	
	if bombCreated != nil {
		t.Error("Bomb should not be created for match of 3")
	}
}

// TestRemoveMatches_WithBomb tests bomb creation on match of 4+
func TestRemoveMatches_WithBomb(t *testing.T) {
	board := NewBoard(8, 8)
	
	// Create a match of 4
	board.Tiles[0][0].Gem = GemYellow
	board.Tiles[0][1].Gem = GemYellow
	board.Tiles[0][2].Gem = GemYellow
	board.Tiles[0][3].Gem = GemYellow
	
	score, bombCreated := board.RemoveMatches()
	
	if score == 0 {
		t.Error("Score should not be 0 for a match of 4")
	}
	
	if bombCreated == nil {
		t.Error("Bomb should be created for match of 4+")
	} else if !bombCreated.IsBomb {
		t.Error("Created tile should have IsBomb flag set")
	}
}

// TestApplyGravity tests gravity mechanics
func TestApplyGravity(t *testing.T) {
	board := NewBoard(8, 8)
	
	// Set some tiles to empty
	board.Tiles[7][0].Gem = GemType(-1)
	board.Tiles[6][0].Gem = GemRed
	
	board.ApplyGravity()
	
	// Tile should fall down
	if board.Tiles[7][0].Gem != GemRed {
		t.Errorf("Expected tile to fall to bottom, got gem type %d", board.Tiles[7][0].Gem)
	}
}

// TestBoardEdgeCases tests edge cases
func TestBoardEdgeCases(t *testing.T) {
	// Small board
	smallBoard := NewBoard(3, 3)
	if smallBoard.Rows != 3 || smallBoard.Cols != 3 {
		t.Error("Small board should have correct dimensions")
	}
	
	// GetTileAt edge cases
	tile := smallBoard.GetTileAt(-1, -1)
	if tile != nil {
		t.Error("GetTileAt should return nil for negative coordinates")
	}
	
	tile = smallBoard.GetTileAt(1000, 1000)
	if tile != nil {
		t.Error("GetTileAt should return nil for out of bounds coordinates")
	}
}
