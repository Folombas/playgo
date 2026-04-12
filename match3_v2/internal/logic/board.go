package logic

import (
	"image"
	"math/rand"
)

const numGemTypes = 5

// Board игровое поле
type Board struct {
	grid [][]int // grid[row][col] = тип гема
	rows int
	cols int
	rng  *rand.Rand
}

// NewBoard создаёт новое поле
func NewBoard(size int, rng *rand.Rand) *Board {
	b := &Board{
		grid: make([][]int, size),
		rows: size,
		cols: size,
		rng:  rng,
	}

	// Заполнить
	for r := 0; r < size; r++ {
		b.grid[r] = make([]int, size)
		for c := 0; c < size; c++ {
			b.grid[r][c] = b.randomGem(r, c)
		}
	}

	return b
}

// randomGem генерирует случайный гем без начальных матчей
func (b *Board) randomGem(row, col int) int {
	for {
		gem := b.rng.Intn(numGemTypes)

		// Проверить горизонтальный матч
		if col >= 2 &&
			b.grid[row][col-1] == gem &&
			b.grid[row][col-2] == gem {
			continue
		}

		// Проверить вертикальный матч
		if row >= 2 &&
			b.grid[row-1][col] == gem &&
			b.grid[row-2][col] == gem {
			continue
		}

		return gem
	}
}

// Get возвращает тип гема
func (b *Board) Get(row, col int) int {
	if row < 0 || row >= b.rows || col < 0 || col >= b.cols {
		return -1
	}
	return b.grid[row][col]
}

// Set устанавливает тип гема
func (b *Board) Set(row, col int, gem int) {
	if row >= 0 && row < b.rows && col >= 0 && col < b.cols {
		b.grid[row][col] = gem
	}
}

// Swap меняет местами два гема
func (b *Board) Swap(r1, c1, r2, c2 int) {
	b.grid[r1][c1], b.grid[r2][c2] = b.grid[r2][c2], b.grid[r1][c1]
}

// Shake - заглушка для анимации тряски
func (b *Board) Shake(row, col int) {
	// Можно добавить анимацию позже
}

// FindMatches находит все комбинации 3+
func (b *Board) FindMatches() []image.Point {
	matched := make(map[image.Point]bool)

	// Горизонтальные
	for r := 0; r < b.rows; r++ {
		for c := 0; c < b.cols-2; c++ {
			gem := b.grid[r][c]
			if gem == -1 {
				continue
			}
			if b.grid[r][c+1] == gem && b.grid[r][c+2] == gem {
				matched[image.Point{r, c}] = true
				matched[image.Point{r, c + 1}] = true
				matched[image.Point{r, c + 2}] = true

				// Проверить 4+
				for extra := c + 3; extra < b.cols; extra++ {
					if b.grid[r][extra] == gem {
						matched[image.Point{r, extra}] = true
					} else {
						break
					}
				}
			}
		}
	}

	// Вертикальные
	for c := 0; c < b.cols; c++ {
		for r := 0; r < b.rows-2; r++ {
			gem := b.grid[r][c]
			if gem == -1 {
				continue
			}
			if b.grid[r+1][c] == gem && b.grid[r+2][c] == gem {
				matched[image.Point{r, c}] = true
				matched[image.Point{r + 1, c}] = true
				matched[image.Point{r + 2, c}] = true

				// Проверить 4+
				for extra := r + 3; extra < b.rows; extra++ {
					if b.grid[extra][c] == gem {
						matched[image.Point{extra, c}] = true
					} else {
						break
					}
				}
			}
		}
	}

	result := make([]image.Point, 0, len(matched))
	for p := range matched {
		result = append(result, p)
	}

	return result
}

// RemoveMatches удаляет найденные комбинации
func (b *Board) RemoveMatches(matches []image.Point) {
	for _, m := range matches {
		b.grid[m.X][m.Y] = -1
	}
}

// FillEmpty заполняет пустые ячейки (гравитация + новые)
func (b *Board) FillEmpty() {
	for c := 0; c < b.cols; c++ {
		// Сдвинуть вниз
		writeRow := b.rows - 1
		for r := b.rows - 1; r >= 0; r-- {
			if b.grid[r][c] != -1 {
				b.grid[writeRow][c] = b.grid[r][c]
				if writeRow != r {
					b.grid[r][c] = -1
				}
				writeRow--
			}
		}

		// Заполнить сверху новыми
		for r := writeRow; r >= 0; r-- {
			b.grid[r][c] = b.rng.Intn(numGemTypes)
		}
	}
}

// FindHint находит возможный ход
func (b *Board) FindHint() [2]image.Point {
	for r := 0; r < b.rows; r++ {
		for c := 0; c < b.cols; c++ {
			// Попробовать交换 вправо
			if c < b.cols-1 {
				b.Swap(r, c, r, c+1)
				if len(b.FindMatches()) > 0 {
					b.Swap(r, c, r, c+1) // вернуть обратно
					return [2]image.Point{{r, c}, {r, c + 1}}
				}
				b.Swap(r, c, r, c+1) // вернуть обратно
			}
			// Попробовать交换 вниз
			if r < b.rows-1 {
				b.Swap(r, c, r+1, c)
				if len(b.FindMatches()) > 0 {
					b.Swap(r, c, r+1, c) // вернуть обратно
					return [2]image.Point{{r, c}, {r + 1, c}}
				}
				b.Swap(r, c, r+1, c) // вернуть обратно
			}
		}
	}
	return [2]image.Point{}
}
