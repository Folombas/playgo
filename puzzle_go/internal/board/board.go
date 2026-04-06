// Package board содержит логику игровой доски: поиск совпадений,
// гравитация, заполнение, проверка ходов, перемешивание.
package board

import (
	"math/rand"

	"github.com/playgo/puzzle_go/internal/config"
)

// WouldMatchAt проверяет, создаст ли размещение type в (r,c) совпадение.
func WouldMatchAt(b [config.Rows][config.Cols]int, r, c, t int) bool {
	cnt := 1
	for i := c - 1; i >= 0 && b[r][i] == t; i-- { cnt++ }
	for i := c + 1; i < config.Cols && b[r][i] == t; i++ { cnt++ }
	if cnt >= config.MatchMin { return true }
	cnt = 1
	for i := r - 1; i >= 0 && b[i][c] == t; i-- { cnt++ }
	for i := r + 1; i < config.Rows && b[i][c] == t; i++ { cnt++ }
	return cnt >= config.MatchMin
}

// FindMatches находит все совпадающие группы на доске.
// Возвращает map[[2]int]bool где ключ — позиция {row, col}.
func FindMatches(b [config.Rows][config.Cols]int) map[[2]int]bool {
	matched := make(map[[2]int]bool)

	// Horizontal
	for r := 0; r < config.Rows; r++ {
		for c := 0; c <= config.Cols-config.MatchMin; c++ {
			t := b[r][c]
			if t < 0 { continue }
			match := true
			for i := 1; i < config.MatchMin; i++ {
				if b[r][c+i] != t { match = false; break }
			}
			if match {
				for i := 0; i < config.MatchMin; i++ {
					matched[[2]int{r, c+i}] = true
				}
			}
		}
	}

	// Vertical
	for c := 0; c < config.Cols; c++ {
		for r := 0; r <= config.Rows-config.MatchMin; r++ {
			t := b[r][c]
			if t < 0 { continue }
			match := true
			for i := 1; i < config.MatchMin; i++ {
				if b[r+i][c] != t { match = false; break }
			}
			if match {
				for i := 0; i < config.MatchMin; i++ {
					matched[[2]int{r+i, c}] = true
				}
			}
		}
	}

	return matched
}

// ApplyGravity сдвигает все гемы вниз, возвращает список смещённых позиций.
func ApplyGravity(b *[config.Rows][config.Cols]int) [][2]int {
	moved := make([][2]int, 0)
	for c := 0; c < config.Cols; c++ {
		wr := config.Rows - 1
		for r := config.Rows - 1; r >= 0; r-- {
			if (*b)[r][c] >= 0 {
				if r != wr {
					(*b)[wr][c] = (*b)[r][c]
					(*b)[r][c] = -1
					moved = append(moved, [2]int{wr, c})
				}
				wr--
			}
		}
	}
	return moved
}

// FillEmpty заполняет пустые (-1) клетки случайными гемами.
func FillEmpty(b *[config.Rows][config.Cols]int) [][2]int {
	filled := make([][2]int, 0)
	for c := 0; c < config.Cols; c++ {
		for r := 0; r < config.Rows; r++ {
			if (*b)[r][c] < 0 {
				(*b)[r][c] = rand.Intn(config.GemTypes)
				filled = append(filled, [2]int{r, c})
			}
		}
	}
	return filled
}

// HasValidMoves проверяет есть ли возможные ходы (свапы создающие совпадения).
func HasValidMoves(b [config.Rows][config.Cols]int) bool {
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			// Swap right
			if c+1 < config.Cols {
				b[r][c], b[r][c+1] = b[r][c+1], b[r][c]
				if len(FindMatches(b)) > 0 { return true }
				b[r][c], b[r][c+1] = b[r][c+1], b[r][c]
			}
			// Swap down
			if r+1 < config.Rows {
				b[r][c], b[r+1][c] = b[r+1][c], b[r][c]
				if len(FindMatches(b)) > 0 { return true }
				b[r][c], b[r+1][c] = b[r+1][c], b[r][c]
			}
		}
	}
	return false
}

// ShuffleBoard перемешивает доску пока не будет совпадений и будут валидные ходы.
func ShuffleBoard(b *[config.Rows][config.Cols]int) {
	for {
		flat := make([]int, 0, config.Rows*config.Cols)
		for r := 0; r < config.Rows; r++ {
			for c := 0; c < config.Cols; c++ {
				flat = append(flat, (*b)[r][c])
			}
		}
		rand.Shuffle(len(flat), func(i, j int) { flat[i], flat[j] = flat[j], flat[i] })
		idx := 0
		for r := 0; r < config.Rows; r++ {
			for c := 0; c < config.Cols; c++ {
				(*b)[r][c] = flat[idx]
				idx++
			}
		}

		// Remove any existing matches
		matches := FindMatches(*b)
		if len(matches) == 0 && HasValidMoves(*b) {
			return
		}
		// Clear matches and try again
		for pos := range matches {
			(*b)[pos[0]][pos[1]] = rand.Intn(config.GemTypes)
		}
	}
}

// FillBoard заполняет доску без начальных совпадений.
func FillBoard(b *[config.Rows][config.Cols]int) {
	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			for {
				(*b)[r][c] = rand.Intn(config.GemTypes)
				if !WouldMatchAt(*b, r, c, (*b)[r][c]) {
					break
				}
			}
		}
	}
}

// Swap меняет местами два гема.
func Swap(b *[config.Rows][config.Cols]int, r1, c1, r2, c2 int) {
	(*b)[r1][c1], (*b)[r2][c2] = (*b)[r2][c2], (*b)[r1][c1]
}

// ClearMatches удаляет совпавшие гемы (устанавливает в -1).
func ClearMatches(b *[config.Rows][config.Cols]int, matches map[[2]int]bool) {
	for pos := range matches {
		(*b)[pos[0]][pos[1]] = -1
	}
}

// Pos2Coords преобразует позицию в map в row, col.
func Pos2Coords(pos [2]int) (int, int) {
	return pos[0], pos[1]
}
