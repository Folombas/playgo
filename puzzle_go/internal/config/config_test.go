package config

import (
	"testing"
)

func TestLayoutConstants(t *testing.T) {
	// WinW should accommodate all columns with offsets
	expectedW := Cols*Tile + BoardOffX*2
	if WinW != expectedW {
		t.Errorf("WinW = %d, want %d", WinW, expectedW)
	}

	// WinH should accommodate all rows with offsets + HUD
	expectedH := Rows*Tile + BoardOffY + HUD
	if WinH != expectedH {
		t.Errorf("WinH = %d, want %d", WinH, expectedH)
	}
}

func TestBoardDimensions(t *testing.T) {
	if Cols != 8 {
		t.Errorf("Cols = %d, want 8", Cols)
	}
	if Rows != 8 {
		t.Errorf("Rows = %d, want 8", Rows)
	}
	if Tile != 64 {
		t.Errorf("Tile = %d, want 64", Tile)
	}
}

func TestGameplayConstants(t *testing.T) {
	if MatchMin != 3 {
		t.Errorf("MatchMin = %d, want 3", MatchMin)
	}
	if GemTypes != 6 {
		t.Errorf("GemTypes = %d, want 6", GemTypes)
	}
	if TargetScore <= 0 {
		t.Errorf("TargetScore should be positive, got %d", TargetScore)
	}
	if CascadeDelay <= 0 {
		t.Errorf("CascadeDelay should be positive, got %d", CascadeDelay)
	}
}

func TestStateValues(t *testing.T) {
	// States should be sequential starting from 0
	if StateMenu != 0 {
		t.Errorf("StateMenu = %d, want 0", StateMenu)
	}
	if StatePlay != 1 {
		t.Errorf("StatePlay = %d, want 1", StatePlay)
	}
	if StatePause != 2 {
		t.Errorf("StatePause = %d, want 2", StatePause)
	}
	if StateOptions != 3 {
		t.Errorf("StateOptions = %d, want 3", StateOptions)
	}
	if StateWin != 4 {
		t.Errorf("StateWin = %d, want 4", StateWin)
	}
	if StateNoMoves != 5 {
		t.Errorf("StateNoMoves = %d, want 5", StateNoMoves)
	}
}

func TestWindowFitsBoard(t *testing.T) {
	// Board area should fit within window
	boardW := Cols*Tile + BoardOffX*2
	boardH := Rows*Tile + BoardOffY + HUD
	if WinW < boardW {
		t.Errorf("Window width %d < board width %d", WinW, boardW)
	}
	if WinH < boardH {
		t.Errorf("Window height %d < board height %d", WinH, boardH)
	}
}
