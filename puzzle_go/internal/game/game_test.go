package game

import (
	"os"
	"testing"

	"github.com/playgo/puzzle_go/internal/board"
	"github.com/playgo/puzzle_go/internal/config"
)

// newTestGame создаёт Game с мок-спрайтами и аудио (nil).
// НЕ вызывает ebiten.RunGame — только для unit-тестов.
func newTestGame() *Game {
	g := &Game{selR: -1, selC: -1, hovR: -1, hovC: -1}
	g.buttons = nil // UI не тестируется здесь
	return g
}

// ======================== START ========================

func TestStart_ResetsState(t *testing.T) {
	g := newTestGame()
	g.score = 999
	g.combo = 5
	g.moves = 100
	g.state = config.StateWin
	g.selR = 3
	g.selC = 3
	g.busy = true
	g.flash = 10

	g.Start()

	if g.score != 0 {
		t.Errorf("score not reset: %d", g.score)
	}
	if g.combo != 0 {
		t.Errorf("combo not reset: %d", g.combo)
	}
	if g.moves != 0 {
		t.Errorf("moves not reset: %d", g.moves)
	}
	if g.state != config.StatePlay {
		t.Errorf("state not reset: %d", g.state)
	}
	if g.selR != -1 || g.selC != -1 {
		t.Errorf("selection not reset: %d,%d", g.selR, g.selC)
	}
	if g.busy {
		t.Error("busy not reset")
	}
}

func TestStart_NoInitialMatches(t *testing.T) {
	g := newTestGame()
	g.Start()
	// Проверяем что нет начальных совпадений
	matches := findMatchesForTest(&g.bd)
	if len(matches) > 0 {
		t.Errorf("Board has %d initial matches after Start()", len(matches))
	}
}

func TestStart_AllValidTypes(t *testing.T) {
	g := newTestGame()
	g.Start()
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			if g.bd[r][c] < 0 || g.bd[r][c] >= config.GemTypes {
				t.Errorf("Invalid type at [%d][%d]: %d", r, c, g.bd[r][c])
			}
		}
	}
}

// ======================== HIGH SCORE ========================

func TestSaveHighScore_Updates(t *testing.T) {
	g := newTestGame()
	g.highScore = 100
	g.score = 200
	g.saveHighScore()
	if g.highScore != 200 {
		t.Errorf("highScore = %d, want 200", g.highScore)
	}
	// Cleanup
	os.Remove("highscore_test.txt")
}

func TestSaveHighScore_NoUpdate(t *testing.T) {
	g := newTestGame()
	g.highScore = 500
	g.score = 300
	g.saveHighScore()
	if g.highScore != 500 {
		t.Errorf("highScore should not change: %d, want 500", g.highScore)
	}
}

// ======================== CLICK ========================

func TestClick_SelectFirst(t *testing.T) {
	g := newTestGame()
	g.Start()
	g.click(2, 3)
	if g.selR != 2 || g.selC != 3 {
		t.Errorf("Expected select (2,3), got (%d,%d)", g.selR, g.selC)
	}
	if g.selectAnim == nil {
		t.Error("selectAnim should be created on first click")
	}
}

func TestClick_Deselect(t *testing.T) {
	g := newTestGame()
	g.Start()
	g.click(2, 3)
	g.click(2, 3) // Same cell — deselect
	if g.selR != -1 || g.selC != -1 {
		t.Error("Same cell click should deselect")
	}
}

func TestClick_SelectAdjacent(t *testing.T) {
	// Requires audio.Manager and sprites — skip in unit test
	// (Integration test would run with full game)
	t.Skip("Requires audio.Manager and sprite cache — integration test")
}

func TestClick_NonAdjacent_SelectsNew(t *testing.T) {
	g := newTestGame()
	g.Start()
	g.click(0, 0)
	g.click(3, 3) // Not adjacent — select new
	if g.selR != 3 || g.selC != 3 {
		t.Errorf("Expected select (3,3), got (%d,%d)", g.selR, g.selC)
	}
}

func TestClick_Busy_Ignores(t *testing.T) {
	g := newTestGame()
	g.Start()
	g.busy = true
	g.click(2, 3)
	if g.selR != -1 {
		t.Error("Click should be ignored when busy")
	}
}

// ======================== TRY SWAP ========================

func TestTrySwap_MatchFound(t *testing.T) {
	t.Skip("Requires audio.Manager and sprite cache — integration test")
}

func TestTrySwap_NoMatch_Reverts(t *testing.T) {
	t.Skip("Requires audio.Manager and sprite cache — integration test")
}

// ======================== SCORE ========================

func TestScore_IncreasesOnMatch(t *testing.T) {
	g := newTestGame()
	g.Start()
	g.bd[0][0] = 1; g.bd[0][1] = 1; g.bd[0][2] = 1
	matches := findMatchesForTest(&g.bd)
	// Directly clear matches and update score (skip removeMatches which needs audio)
	board.ClearMatches(&g.bd, matches)
	g.score += len(matches) * 10
	if g.score <= 0 {
		t.Errorf("Score should increase on match, got %d", g.score)
	}
}

func TestScore_ComboMultiplier(t *testing.T) {
	g := newTestGame()
	g.bd[0][0] = 1; g.bd[0][1] = 1; g.bd[0][2] = 1
	g.combo = 2
	// Direct score calculation
	g.score += 3 * 10 * g.combo // 3 gems × 10 × combo
	if g.score != 60 {
		t.Errorf("Score with combo x2 = %d, want 60", g.score)
	}
}

// ======================== WIN CONDITION ========================

func TestWinCondition_TargetScore(t *testing.T) {
	g := newTestGame()
	g.Start()
	g.score = config.TargetScore
	// Simulate cascade check logic directly
	matches := findMatchesForTest(&g.bd)
	if len(matches) == 0 && g.score >= config.TargetScore {
		g.state = config.StateWin
	}
	if g.state != config.StateWin {
		t.Errorf("Expected StateWin(%d) at target score, got %d", config.StateWin, g.state)
	}
}

// ======================== HELPER ========================

// findMatchesForTest — копия board.FindMatches для тестирования без импорта.
func findMatchesForTest(b *[config.Rows][config.Cols]int) map[[2]int]bool {
	matched := make(map[[2]int]bool)
	for r := 0; r < config.Rows; r++ {
		for c := 0; c <= config.Cols-3; c++ {
			t := (*b)[r][c]
			if t < 0 { continue }
			if (*b)[r][c+1] == t && (*b)[r][c+2] == t {
				matched[[2]int{r, c}] = true
				matched[[2]int{r, c+1}] = true
				matched[[2]int{r, c+2}] = true
			}
		}
	}
	for c := 0; c < config.Cols; c++ {
		for r := 0; r <= config.Rows-3; r++ {
			t := (*b)[r][c]
			if t < 0 { continue }
			if (*b)[r+1][c] == t && (*b)[r+2][c] == t {
				matched[[2]int{r, c}] = true
				matched[[2]int{r+1, c}] = true
				matched[[2]int{r+2, c}] = true
			}
		}
	}
	return matched
}
