package main

import (
	"math/rand"
)

const (
	// BoardSize размер игрового поля (8x8)
	BoardSize = 8

	// NumTileTypes количество типов фишек (6 цветов)
	NumTileTypes = 6

	// MatchMin минимальное количество фишек для комбинации
	MatchMin = 3
)

// Board представляет игровое поле 8x8
type Board struct {
	grid [BoardSize][BoardSize]int // -1 = пусто, 0-5 = тип фишки
}

// NewBoard создаёт новое игровое поле без начальных комбинаций
func NewBoard() *Board {
	b := &Board{}
	b.fillNoMatches()
	return b
}

// fillNoMatches заполняет поле случайными фишками, гарантируя отсутствие готовых комбинаций
func (b *Board) fillNoMatches() {
	for {
		b.fillRandom()
		if !b.hasAnyMatches() {
			return
		}
		// Если есть комбинации - очищаем и заполняем заново
		b.clearMatches()
	}
}

// fillRandom случайное заполнение поля
func (b *Board) fillRandom() {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			b.grid[row][col] = rand.Intn(NumTileTypes)
		}
	}
}

// hasAnyMatches проверяет есть ли на поле комбинации 3+ подряд
func (b *Board) hasAnyMatches() bool {
	// Проверка горизонталей
	for row := 0; row < BoardSize; row++ {
		for col := 0; col <= BoardSize-MatchMin; col++ {
			t := b.grid[row][col]
			if t == -1 {
				continue
			}
			match := true
			for i := 1; i < MatchMin; i++ {
				if b.grid[row][col+i] != t {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}

	// Проверка вертикалей
	for col := 0; col < BoardSize; col++ {
		for row := 0; row <= BoardSize-MatchMin; row++ {
			t := b.grid[row][col]
			if t == -1 {
				continue
			}
			match := true
			for i := 1; i < MatchMin; i++ {
				if b.grid[row+i][col] != t {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}

	return false
}

// findAllMatches находит все комбинации на поле
// Возвращает карту позиций для удаления (row,col) -> true
func (b *Board) findAllMatches() map[[2]int]bool {
	matches := make(map[[2]int]bool)

	// Горизонтальные комбинации
	for row := 0; row < BoardSize; row++ {
		for col := 0; col <= BoardSize-MatchMin; col++ {
			t := b.grid[row][col]
			if t == -1 {
				continue
			}
			length := 1
			for col+length < BoardSize && b.grid[row][col+length] == t {
				length++
			}
			if length >= MatchMin {
				for i := 0; i < length; i++ {
					matches[[2]int{row, col + i}] = true
				}
			}
		}
	}

	// Вертикальные комбинации
	for col := 0; col < BoardSize; col++ {
		for row := 0; row <= BoardSize-MatchMin; row++ {
			t := b.grid[row][col]
			if t == -1 {
				continue
			}
			length := 1
			for row+length < BoardSize && b.grid[row+length][col] == t {
				length++
			}
			if length >= MatchMin {
				for i := 0; i < length; i++ {
					matches[[2]int{row + i, col}] = true
				}
			}
		}
	}

	return matches
}

// clearMatches удаляет фишки из найденных комбинаций
func (b *Board) clearMatches() {
	matches := b.findAllMatches()
	for pos := range matches {
		b.grid[pos[0]][pos[1]] = -1
	}
}

// dropDown сдвигает фишки вниз, заполняя пустоты новыми
// Возвращает карту новых позиций (row,col) -> true
func (b *Board) dropDown() map[[2]int]bool {
	newTiles := make(map[[2]int]bool)

	for col := 0; col < BoardSize; col++ {
		emptyRows := 0
		// Считаем снизу вверх
		for row := BoardSize - 1; row >= 0; row-- {
			if b.grid[row][col] == -1 {
				emptyRows++
			} else if emptyRows > 0 {
				// Сдвигаем фишку вниз
				b.grid[row+emptyRows][col] = b.grid[row][col]
				b.grid[row][col] = -1
			}
		}
		// Заполняем верхние пустые ячейки новыми фишками
		for row := 0; row < emptyRows; row++ {
			b.grid[row][col] = rand.Intn(NumTileTypes)
			newTiles[[2]int{row, col}] = true
		}
	}

	return newTiles
}

// Swap пытается обменять две соседние фишки
// Возвращает true если обмен привёл к комбинации
func (b *Board) Swap(r1, c1, r2, c2 int) bool {
	// Проверяем что позиции соседние
	if !b.isAdjacent(r1, c1, r2, c2) {
		return false
	}

	// Проверяем границы
	if !b.isValidPos(r1, c1) || !b.isValidPos(r2, c2) {
		return false
	}

	// Обмениваем
	b.grid[r1][c1], b.grid[r2][c2] = b.grid[r2][c2], b.grid[r1][c1]

	// Проверяем есть ли комбинации
	if b.hasAnyMatches() {
		return true
	}

	// Возвращаем обратно
	b.grid[r1][c1], b.grid[r2][c2] = b.grid[r2][c2], b.grid[r1][c1]
	return false
}

// PerformSwap выполняет обмен (без проверки валидности)
func (b *Board) PerformSwap(r1, c1, r2, c2 int) {
	if b.isValidPos(r1, c1) && b.isValidPos(r2, c2) {
		b.grid[r1][c1], b.grid[r2][c2] = b.grid[r2][c2], b.grid[r1][c1]
	}
}

// isAdjacent проверяет являются ли две позиции соседними (по горизонтали/вертикали)
func (b *Board) isAdjacent(r1, c1, r2, c2 int) bool {
	dr := abs(r1 - r2)
	dc := abs(c1 - c2)
	return (dr == 1 && dc == 0) || (dr == 0 && dc == 1)
}

// isValidPos проверяет корректность позиции
func (b *Board) isValidPos(row, col int) bool {
	return row >= 0 && row < BoardSize && col >= 0 && col < BoardSize
}

// GetTile возвращает тип фишки в позиции
func (b *Board) GetTile(row, col int) int {
	if b.isValidPos(row, col) {
		return b.grid[row][col]
	}
	return -1
}

// SetTile устанавливает тип фишки в позиции
func (b *Board) SetTile(row, col int, tileType int) {
	if b.isValidPos(row, col) {
		b.grid[row][col] = tileType
	}
}

// FindHint находит валидный ход (пару фишек для обмена)
// Возвращает две позиции или false если ходов нет
func (b *Board) FindHint() ([2]int, [2]int, bool) {
	// Проверяем все возможные горизонтальные обмены
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize-1; col++ {
			// Пробуем обмен
			b.grid[row][col], b.grid[row][col+1] = b.grid[row][col+1], b.grid[row][col]
			if b.hasAnyMatches() {
				// Возвращаем обратно
				b.grid[row][col], b.grid[row][col+1] = b.grid[row][col+1], b.grid[row][col]
				return [2]int{row, col}, [2]int{row, col + 1}, true
			}
			// Возвращаем обратно
			b.grid[row][col], b.grid[row][col+1] = b.grid[row][col+1], b.grid[row][col]
		}
	}

	// Проверяем все возможные вертикальные обмены
	for row := 0; row < BoardSize-1; row++ {
		for col := 0; col < BoardSize; col++ {
			// Пробуем обмен
			b.grid[row][col], b.grid[row+1][col] = b.grid[row+1][col], b.grid[row][col]
			if b.hasAnyMatches() {
				// Возвращаем обратно
				b.grid[row][col], b.grid[row+1][col] = b.grid[row+1][col], b.grid[row][col]
				return [2]int{row, col}, [2]int{row + 1, col}, true
			}
			// Возвращаем обратно
			b.grid[row][col], b.grid[row+1][col] = b.grid[row+1][col], b.grid[row][col]
		}
	}

	return [2]int{}, [2]int{}, false
}

// Grid возвращает ссылку на игровое поле (для отрисовки)
func (b *Board) Grid() *[BoardSize][BoardSize]int {
	return &b.grid
}

// abs возвращает абсолютное значение
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
