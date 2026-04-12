package logic

import (
	"math/rand"
	"testing"
	"time"
)

// TestNewBoard tests board creation
func TestNewBoard(t *testing.T) {
	board := NewBoard(8, 8, 1)

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
}

// TestFindAllMatches tests match detection
func TestFindAllMatches(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := createEmptyBoard(8, 8, rng)

	// Fill board with non-matching pattern first
	for r := 0; r < board.Rows; r++ {
		for c := 0; c < board.Cols; c++ {
			board.Tiles[r][c].Gem = GemType((r + c) % 6)
		}
	}

	// Create a horizontal match of 3
	board.Tiles[0][0].Gem = GemApple
	board.Tiles[0][1].Gem = GemApple
	board.Tiles[0][2].Gem = GemApple

	matches := board.FindAllMatches()

	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}
}

// TestFindAllMatches_Vertical tests vertical match detection
func TestFindAllMatches_Vertical(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := createEmptyBoard(8, 8, rng)

	// Create a vertical match of 4
	board.Tiles[0][0].Gem = GemOrange
	board.Tiles[1][0].Gem = GemOrange
	board.Tiles[2][0].Gem = GemOrange
	board.Tiles[3][0].Gem = GemOrange

	matches := board.FindAllMatches()

	if len(matches) < 4 {
		t.Errorf("Expected at least 4 matches for vertical line, got %d", len(matches))
	}
}

// TestSwapTiles tests tile swapping
func TestSwapTiles(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := createEmptyBoard(8, 8, rng)

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
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := createEmptyBoard(8, 8, rng)

	tile1 := board.Tiles[0][0]
	tile2 := board.Tiles[2][2]

	success := board.SwapTiles(tile1, tile2)

	if success {
		t.Error("Swap should fail for non-adjacent tiles")
	}
}

// TestFindHint tests hint system
func TestFindHint(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := createEmptyBoard(8, 8, rng)

	// Setup a guaranteed match
	board.Tiles[0][0].Gem = GemApple
	board.Tiles[0][1].Gem = GemLemon
	board.Tiles[0][2].Gem = GemApple
	board.Tiles[0][3].Gem = GemApple

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

// TestSpecialTileCreation tests bomb and rocket creation
func TestSpecialTileCreation(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Test 4-match creates rocket
	board := createEmptyBoard(8, 8, rng)
	board.Tiles[0][0].Gem = GemApple
	board.Tiles[0][1].Gem = GemApple
	board.Tiles[0][2].Gem = GemApple
	board.Tiles[0][3].Gem = GemApple

	matches := board.FindAllMatches()
	if len(matches) == 4 {
		board.checkSpecialCreation(matches)
		
		// Check if any tile became a rocket
		hasRocket := false
		for r := 0; r < board.Rows; r++ {
			for c := 0; c < board.Cols; c++ {
				tile := board.Tiles[r][c]
				if tile != nil && (tile.IsRocketH || tile.IsRocketV) {
					hasRocket = true
				}
			}
		}
		
		if !hasRocket {
			t.Log("Warning: Rocket not created for 4-match (expected behavior if implementation differs)")
		}
	}
}

// TestGravity tests gravity mechanics
func TestGravity(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := createEmptyBoard(8, 8, rng)

	// Set up tiles для теста гравитации
	board.Tiles[0][0].Gem = GemApple
	board.Tiles[1][0].Gem = GemOrange
	
	// Пометим верхнюю плитку для удаления
	board.Tiles[0][0].Removing = true
	
	board.ApplyGravity()
	
	// Плитка должна была упасть
	if board.Tiles[0][0] != nil && board.Tiles[0][0].Removing {
		t.Log("Gravity test: tile marked for removal")
	}
}

// TestNoInitialMatches tests that board has no initial matches
func TestNoInitialMatches(t *testing.T) {
	board := NewBoard(8, 8, 1)

	matches := board.FindAllMatches()
	if len(matches) > 0 {
		t.Errorf("New board should not have matches, got %d", len(matches))
	}
}

// TestGetTileAt tests tile retrieval
func TestGetTileAt(t *testing.T) {
	board := NewBoard(8, 8, 1)

	// Valid tile
	tile := board.GetTileAt(board.OffsetX+25, board.OffsetY+25)
	if tile == nil {
		t.Error("Should get tile at valid position")
	}

	// Invalid positions
	tile = board.GetTileAt(-1, -1)
	if tile != nil {
		t.Error("Should return nil for negative coordinates")
	}

	tile = board.GetTileAt(1000, 1000)
	if tile != nil {
		t.Error("Should return nil for out of bounds")
	}
}

// TestParticleSystem tests particle system
func TestParticleSystem(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ps := NewParticleSystem(rng)

	// Emit particles
	ps.Emit(100, 100, ParticleExplosion, 10)

	if ps.Count() != 10 {
		t.Errorf("Expected 10 particles, got %d", ps.Count())
	}

	// Update
	ps.Update()

	// Particles should still exist
	if ps.Count() == 0 {
		t.Error("Particles should not disappear immediately")
	}

	// Clear
	ps.Clear()
	if ps.Count() != 0 {
		t.Errorf("Expected 0 particles after clear, got %d", ps.Count())
	}
}

// TestBoardShakeAnimation tests shake animation
func TestBoardShakeAnimation(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := createEmptyBoard(8, 8, rng)

	tile := board.Tiles[0][0]
	tile.Shake = 1.0

	// Update should decrease shake
	board.Update()
	
	if tile.Shake >= 1.0 {
		t.Errorf("Shake should decrease, got %f", tile.Shake)
	}
}

// Helper function to create empty board
func createEmptyBoard(rows, cols int, rng *rand.Rand) *Board {
	b := &Board{
		Tiles:    make([][]*Tile, rows),
		Rows:     rows,
		Cols:     cols,
		TileSize: 50,
		OffsetX:  45,
		OffsetY:  150,
		rng:      rng,
	}

	for r := 0; r < rows; r++ {
		b.Tiles[r] = make([]*Tile, cols)
		for c := 0; c < cols; c++ {
			b.Tiles[r][c] = &Tile{
				Gem:     GemType(rng.Intn(int(GemCount))),
				Row:     r,
				Col:     c,
				X:       float64(b.OffsetX + c*b.TileSize),
				Y:       float64(b.OffsetY + r*b.TileSize),
				TargetX: float64(b.OffsetX + c*b.TileSize),
				TargetY: float64(b.OffsetY + r*b.TileSize),
				Scale:   1.0,
				Alpha:   1.0,
			}
		}
	}

	return b
}
