package board

import (
	"testing"

	"github.com/playgo/puzzle_go/internal/config"
)

// ======================== WouldMatchAt ========================

func TestWouldMatchAt_Horizontal3(t *testing.T) {
	var b [config.Rows][config.Cols]int
	// Создаём 2 одинаковых гема рядом
	b[0][0] = 1
	b[0][1] = 1
	// Размещаем третий — должно совпасть
	if !WouldMatchAt(b, 0, 2, 1) {
		t.Error("Expected match for horizontal 3 in a row")
	}
}

func TestWouldMatchAt_Vertical3(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 2
	b[1][0] = 2
	if !WouldMatchAt(b, 2, 0, 2) {
		t.Error("Expected match for vertical 3 in a column")
	}
}

func TestWouldMatchAt_NoMatch(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 1
	b[0][1] = 2
	b[0][2] = 3
	if WouldMatchAt(b, 0, 3, 4) {
		t.Error("Expected no match with different types")
	}
}

func TestWouldMatchAt_Match4(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 3
	b[0][1] = 3
	b[0][2] = 3
	if !WouldMatchAt(b, 0, 3, 3) {
		t.Error("Expected match for 4 in a row")
	}
}

func TestWouldMatchAt_Match5(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 5
	b[0][1] = 5
	b[0][2] = 5
	b[0][3] = 5
	if !WouldMatchAt(b, 0, 4, 5) {
		t.Error("Expected match for 5 in a row")
	}
}

func TestWouldMatchAt_Only2Adjacent(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 1
	b[0][1] = 1
	// Разные типы с другой стороны
	b[0][2] = 2
	if WouldMatchAt(b, 0, 2, 2) {
		t.Error("2 adjacent + different type should not match")
	}
}

func TestWouldMatchAt_BothDirections(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[1][0] = 4
	b[1][1] = 4
	b[0][1] = 4
	b[2][1] = 4
	// Размещаем в (1,1) уже занято, проверяем (1,2) — крестообразно
	if !WouldMatchAt(b, 1, 2, 4) {
		t.Error("Expected match when both directions have same type")
	}
}

// ======================== FindMatches ========================

func TestFindMatches_NoMatches(t *testing.T) {
	b := [config.Rows][config.Cols]int{}
	// Заполняем -1 чтобы не было случайных совпадений
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			b[r][c] = -1
		}
	}
	// Ставим разные типы — без совпадений
	b[0][0] = 0; b[0][1] = 1; b[0][2] = 2; b[0][3] = 3
	b[1][0] = 1; b[1][1] = 2; b[1][2] = 3; b[1][3] = 0
	matches := FindMatches(b)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

func TestFindMatches_Horizontal3(t *testing.T) {
	b := [config.Rows][config.Cols]int{}
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			b[r][c] = -1
		}
	}
	b[0][0] = 1; b[0][1] = 1; b[0][2] = 1
	matches := FindMatches(b)
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}
}

func makeEmptyBoard() [config.Rows][config.Cols]int {
	var b [config.Rows][config.Cols]int
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			b[r][c] = -1
		}
	}
	return b
}

func TestFindMatches_Vertical3(t *testing.T) {
	b := makeEmptyBoard()
	b[0][0] = 2; b[1][0] = 2; b[2][0] = 2
	matches := FindMatches(b)
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}
}

func TestFindMatches_MultipleMatches(t *testing.T) {
	b := makeEmptyBoard()
	b[0][0] = 1; b[0][1] = 1; b[0][2] = 1
	b[2][0] = 3; b[2][1] = 3; b[2][2] = 3
	matches := FindMatches(b)
	if len(matches) != 6 {
		t.Errorf("Expected 6 matches (2 horizontal 3s), got %d", len(matches))
	}
}

func TestFindMatches_LShape(t *testing.T) {
	b := makeEmptyBoard()
	b[0][0] = 1; b[0][1] = 1; b[0][2] = 1
	b[1][0] = 1; b[2][0] = 1
	matches := FindMatches(b)
	// L-shape: (0,0),(0,1),(0,2) horizontal + (0,0),(1,0),(2,0) vertical
	// Unique: 5 cells
	if len(matches) != 5 {
		t.Errorf("Expected 5 unique matches for L-shape, got %d", len(matches))
	}
}

func TestFindMatches_EmptyCells(t *testing.T) {
	b := makeEmptyBoard()
	b[0][3] = 1; b[0][4] = 1; b[0][5] = 1
	matches := FindMatches(b)
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches (ignoring -1 cells), got %d", len(matches))
	}
}

func TestFindMatches_FullBoard(t *testing.T) {
	var b [config.Rows][config.Cols]int
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			b[r][c] = 1 // All same — every row and col matches
		}
	}
	matches := FindMatches(b)
	// 8 rows x 8 cells + 8 cols x 8 cells - overlap = all 64
	if len(matches) != config.Rows*config.Cols {
		t.Errorf("Expected %d matches (all cells), got %d", config.Rows*config.Cols, len(matches))
	}
}

// ======================== ApplyGravity ========================

func TestApplyGravity_SingleColumn(t *testing.T) {
	b := makeEmptyBoard()
	// Fill all columns with -1 except column 0
	for c := 1; c < config.Cols; c++ {
		for r := 0; r < config.Rows; r++ {
			b[r][c] = -1
		}
	}
	// Column 0: some gems with gaps
	b[1][0] = 1
	b[3][0] = 2
	moved := ApplyGravity(&b)
	// Column 0 should have: 1 at row 6, 2 at row 7
	if b[6][0] != 1 {
		t.Errorf("Expected 1 at row 6, got %d", b[6][0])
	}
	if b[7][0] != 2 {
		t.Errorf("Expected 2 at row 7, got %d", b[7][0])
	}
	if len(moved) != 2 {
		t.Errorf("Expected 2 moved positions, got %d", len(moved))
	}
}

func TestApplyGravity_NoEmpty(t *testing.T) {
	var b [config.Rows][config.Cols]int
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			b[r][c] = 1
		}
	}
	moved := ApplyGravity(&b)
	if len(moved) != 0 {
		t.Errorf("Expected 0 moved, got %d", len(moved))
	}
}

func TestApplyGravity_FullColumn(t *testing.T) {
	b := [config.Rows][config.Cols]int{
		{1, 0, 0, 0, 0, 0, 0, 0},
		{2, 0, 0, 0, 0, 0, 0, 0},
		{3, 0, 0, 0, 0, 0, 0, 0},
		{4, 0, 0, 0, 0, 0, 0, 0},
		{5, 0, 0, 0, 0, 0, 0, 0},
		{6, 0, 0, 0, 0, 0, 0, 0},
		{7, 0, 0, 0, 0, 0, 0, 0},
		{8, 0, 0, 0, 0, 0, 0, 0},
	}
	moved := ApplyGravity(&b)
	if len(moved) != 0 {
		t.Errorf("Full column should have 0 moved, got %d", len(moved))
	}
	// Values should remain unchanged
	if b[0][0] != 1 || b[7][0] != 8 {
		t.Error("Values should not change order in full column")
	}
}

// ======================== FillEmpty ========================

func TestFillEmpty_SomeEmpty(t *testing.T) {
	b := [config.Rows][config.Cols]int{
		{-1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{-1, 0, 0, 0, 0, 0, 0, 0},
		{2, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	filled := FillEmpty(&b)
	if len(filled) != 2 {
		t.Errorf("Expected 2 filled, got %d", len(filled))
	}
	if b[0][0] < 0 || b[0][0] >= config.GemTypes {
		t.Errorf("Filled cell should be valid gem type, got %d", b[0][0])
	}
}

func TestFillEmpty_NoEmpty(t *testing.T) {
	var b [config.Rows][config.Cols]int
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			b[r][c] = 1
		}
	}
	filled := FillEmpty(&b)
	if len(filled) != 0 {
		t.Errorf("Expected 0 filled, got %d", len(filled))
	}
}

func TestFillEmpty_AllEmpty(t *testing.T) {
	b := makeEmptyBoard() // All -1
	filled := FillEmpty(&b)
	if len(filled) != config.Rows*config.Cols {
		t.Errorf("Expected %d filled (all empty), got %d", config.Rows*config.Cols, len(filled))
	}
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			if b[r][c] < 0 || b[r][c] >= config.GemTypes {
				t.Errorf("Invalid gem type at [%d][%d]: %d", r, c, b[r][c])
			}
		}
	}
}

// ======================== HasValidMoves ========================

func TestHasValidMoves_SwapCreatesMatch(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 1; b[0][1] = 2; b[0][2] = 1; b[0][3] = 1
	// Swap (0,1) with (0,2) → 1,1,1 match
	if !HasValidMoves(b) {
		t.Error("Expected valid move: swapping creates horizontal match")
	}
}

func TestHasValidMoves_NoValidMoves(t *testing.T) {
	// Создаём доску без возможных ходов (чередование)
	b := [config.Rows][config.Cols]int{
		{0, 1, 2, 3, 0, 1, 2, 3},
		{1, 2, 3, 0, 1, 2, 3, 0},
		{2, 3, 0, 1, 2, 3, 0, 1},
		{3, 0, 1, 2, 3, 0, 1, 2},
		{0, 1, 2, 3, 0, 1, 2, 3},
		{1, 2, 3, 0, 1, 2, 3, 0},
		{2, 3, 0, 1, 2, 3, 0, 1},
		{3, 0, 1, 2, 3, 0, 1, 2},
	}
	if HasValidMoves(b) {
		t.Error("Expected no valid moves on this board")
	}
}

// ======================== ShuffleBoard ========================

func TestShuffleBoard_NoMatchesAfter(t *testing.T) {
	var b [config.Rows][config.Cols]int
	// Заполняем доску гарантированными совпадениями
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			b[r][c] = 1
		}
	}
	ShuffleBoard(&b)
	matches := FindMatches(b)
	if len(matches) > 0 {
		t.Errorf("Shuffled board should have no matches, got %d", len(matches))
	}
}

func TestShuffleBoard_HasValidMoves(t *testing.T) {
	var b [config.Rows][config.Cols]int
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			b[r][c] = 1
		}
	}
	ShuffleBoard(&b)
	if !HasValidMoves(b) {
		t.Error("Shuffled board should have valid moves")
	}
}

// ======================== ClearMatches ========================

func TestClearMatches(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 1; b[0][1] = 1; b[0][2] = 1
	matches := map[[2]int]bool{{0, 0}: true, {0, 1}: true, {0, 2}: true}
	ClearMatches(&b, matches)
	if b[0][0] != -1 || b[0][1] != -1 || b[0][2] != -1 {
		t.Error("ClearMatches should set matched cells to -1")
	}
}

func TestClearMatches_UnmatchedUnchanged(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 1; b[0][1] = 2; b[0][2] = 3
	matches := map[[2]int]bool{{0, 0}: true}
	ClearMatches(&b, matches)
	if b[0][1] != 2 || b[0][2] != 3 {
		t.Error("Unmatched cells should remain unchanged")
	}
}

// ======================== Swap ========================

func TestSwap(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 1
	b[0][1] = 2
	Swap(&b, 0, 0, 0, 1)
	if b[0][0] != 2 || b[0][1] != 1 {
		t.Errorf("Swap failed: expected (2,1), got (%d,%d)", b[0][0], b[0][1])
	}
}

func TestSwap_SamePosition(t *testing.T) {
	var b [config.Rows][config.Cols]int
	b[0][0] = 5
	Swap(&b, 0, 0, 0, 0)
	if b[0][0] != 5 {
		t.Error("Swap same position should not change anything")
	}
}

// ======================== FillBoard ========================

func TestFillBoard_NoInitialMatches(t *testing.T) {
	var b [config.Rows][config.Cols]int
	FillBoard(&b)
	matches := FindMatches(b)
	if len(matches) > 0 {
		t.Errorf("Filled board should have no initial matches, got %d", len(matches))
	}
}

func TestFillBoard_AllValidTypes(t *testing.T) {
	var b [config.Rows][config.Cols]int
	FillBoard(&b)
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			if b[r][c] < 0 || b[r][c] >= config.GemTypes {
				t.Errorf("Invalid gem type at [%d][%d]: %d", r, c, b[r][c])
			}
		}
	}
}
