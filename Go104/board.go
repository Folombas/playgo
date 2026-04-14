package main

import (
	"math/rand"
)

const (
	BoardSize  = 8
	NumColors  = 6
	CellSize   = 80
	BoardPixel = BoardSize * CellSize // 640
)

// Position представляет позицию фишки на доске
type Position struct {
	Row int
	Col int
}

// IsAdjacent проверяет, являются ли две позиции соседними (по горизонтали или вертикали)
func (p *Position) IsAdjacent(other *Position) bool {
	dr := abs(p.Row - other.Row)
	dc := abs(p.Col - other.Col)
	return (dr == 1 && dc == 0) || (dr == 0 && dc == 1)
}

// Tile представляет одну фишку на игровом поле
type Tile struct {
	Color    int     // 0-5, тип фишки
	Row      int     // текущая строка
	Col      int     // текущий столбец
	X        float64 // пиксельная позиция X
	Y        float64 // пиксельная позиция Y
	TargetY  float64 // целевая Y для анимации падения
	Scale    float64 // масштаб для анимации удаления
	Alpha    float64 // прозрачность для анимации удаления
	Selected bool    // выделена ли фишка
	Removing bool    // удаляется ли фишка
}

// Board представляет игровое поле 8x8
type Board struct {
	Grid [][]*Tile // двумерный массив фишек
}

// NewBoard создаёт новое игровое поле и заполняет его без начальных совпадений
func NewBoard() *Board {
	b := &Board{
		Grid: make([][]*Tile, BoardSize),
	}

	for r := 0; r < BoardSize; r++ {
		b.Grid[r] = make([]*Tile, BoardSize)
	}

	b.fillWithoutMatches()
	return b
}

// fillWithoutMatches заполняет поле случайными фишками, гарантируя отсутствие начальных совпадений
func (b *Board) fillWithoutMatches() {
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			var color int
			for {
				color = rand.Intn(NumColors)
				if !b.wouldCreateMatch(r, c, color) {
					break
				}
			}
			b.Grid[r][c] = b.newTile(color, r, c)
		}
	}
}

// wouldCreateMatch проверяет, создаст ли установка фишки данного цвета совпадение
func (b *Board) wouldCreateMatch(row, col, color int) bool {
	// Проверка горизонтали
	if col >= 2 {
		if b.Grid[row][col-1] != nil && b.Grid[row][col-1].Color == color &&
			b.Grid[row][col-2] != nil && b.Grid[row][col-2].Color == color {
			return true
		}
	}
	// Проверка вертикали
	if row >= 2 {
		if b.Grid[row-1][col] != nil && b.Grid[row-1][col].Color == color &&
			b.Grid[row-2][col] != nil && b.Grid[row-2][col].Color == color {
			return true
		}
	}
	return false
}

// newTile создаёт новую фишку с заданным цветом и позицией
func (b *Board) newTile(color, row, col int) *Tile {
	return &Tile{
		Color:   color,
		Row:     row,
		Col:     col,
		X:       float64(col * CellSize),
		Y:       float64(row * CellSize),
		TargetY: float64(row * CellSize),
		Scale:   1.0,
		Alpha:   1.0,
	}
}

// Swap меняет местами две фишки на доске
func (b *Board) Swap(p1, p2 *Position) {
	b.Grid[p1.Row][p1.Col], b.Grid[p2.Row][p2.Col] = b.Grid[p2.Row][p2.Col], b.Grid[p1.Row][p1.Col]

	// Обновляем позиции фишек
	if t := b.Grid[p1.Row][p1.Col]; t != nil {
		t.Row = p1.Row
		t.Col = p1.Col
		t.X = float64(p1.Col * CellSize)
		t.TargetY = float64(p1.Row * CellSize)
	}
	if t := b.Grid[p2.Row][p2.Col]; t != nil {
		t.Row = p2.Row
		t.Col = p2.Col
		t.X = float64(p2.Col * CellSize)
		t.TargetY = float64(p2.Row * CellSize)
	}
}

// FindMatches находит все совпадения (3+ в ряд) и возвращает список позиций
func (b *Board) FindMatches() map[string]*Position {
	matches := make(map[string]*Position)

	// Горизонтальные совпадения
	for r := 0; r < BoardSize; r++ {
		for c := 0; c <= BoardSize-3; c++ {
			if b.Grid[r][c] == nil || b.Grid[r][c].Removing {
				continue
			}
			color := b.Grid[r][c].Color
			length := 1
			for c+length < BoardSize && b.Grid[r][c+length] != nil && !b.Grid[r][c+length].Removing && b.Grid[r][c+length].Color == color {
				length++
			}
			if length >= 3 {
				for i := 0; i < length; i++ {
					key := positionKey(r, c+i)
					if _, exists := matches[key]; !exists {
						matches[key] = &Position{Row: r, Col: c + i}
					}
				}
			}
		}
	}

	// Вертикальные совпадения
	for c := 0; c < BoardSize; c++ {
		for r := 0; r <= BoardSize-3; r++ {
			if b.Grid[r][c] == nil || b.Grid[r][c].Removing {
				continue
			}
			color := b.Grid[r][c].Color
			length := 1
			for r+length < BoardSize && b.Grid[r+length][c] != nil && !b.Grid[r+length][c].Removing && b.Grid[r+length][c].Color == color {
				length++
			}
			if length >= 3 {
				for i := 0; i < length; i++ {
					key := positionKey(r+i, c)
					if _, exists := matches[key]; !exists {
						matches[key] = &Position{Row: r + i, Col: c}
					}
				}
			}
		}
	}

	return matches
}

// RemoveTiles помечает фишки на указанных позициях для удаления
func (b *Board) RemoveTiles(matches map[string]*Position) int {
	score := 0
	for _, pos := range matches {
		if b.Grid[pos.Row][pos.Col] != nil {
			b.Grid[pos.Row][pos.Col].Removing = true
			score += 10
		}
	}

	// Бонусы за длинные комбинации
	// Группируем по линиям для подсчёта бонусов
	score += b.calculateBonus(matches)

	return score
}

// calculateBonus подсчитывает бонусы за 4+ в ряд
func (b *Board) calculateBonus(matches map[string]*Position) int {
	bonus := 0

	// Проверяем горизонтали на длину 4+
	for r := 0; r < BoardSize; r++ {
		for c := 0; c <= BoardSize-4; c++ {
			if _, exists := matches[positionKey(r, c)]; !exists {
				continue
			}
			color := -1
			if b.Grid[r][c] != nil {
				color = b.Grid[r][c].Color
			}
			if color == -1 {
				continue
			}

			length := 1
			for c+length < BoardSize {
				if _, exists := matches[positionKey(r, c+length)]; !exists {
					break
				}
				if b.Grid[r][c+length] != nil && b.Grid[r][c+length].Color == color {
					length++
				} else {
					break
				}
			}

			if length == 4 {
				bonus += 50
			} else if length >= 5 {
				bonus += 100
			}
		}
	}

	// Проверяем вертикали на длину 4+
	for c := 0; c < BoardSize; c++ {
		for r := 0; r <= BoardSize-4; r++ {
			if _, exists := matches[positionKey(r, c)]; !exists {
				continue
			}
			color := -1
			if b.Grid[r][c] != nil {
				color = b.Grid[r][c].Color
			}
			if color == -1 {
				continue
			}

			length := 1
			for r+length < BoardSize {
				if _, exists := matches[positionKey(r+length, c)]; !exists {
					break
				}
				if b.Grid[r+length][c] != nil && b.Grid[r+length][c].Color == color {
					length++
				} else {
					break
				}
			}

			if length == 4 {
				bonus += 50
			} else if length >= 5 {
				bonus += 100
			}
		}
	}

	return bonus
}

// DropTiles сдвигает фишки вниз и заполняет пустоты новыми
func (b *Board) DropTiles() {
	for c := 0; c < BoardSize; c++ {
		emptyRows := 0
		for r := BoardSize - 1; r >= 0; r-- {
			if b.Grid[r][c] == nil || b.Grid[r][c].Removing {
				emptyRows++
				b.Grid[r][c] = nil
			} else if emptyRows > 0 {
				// Сдвигаем фишку вниз
				tile := b.Grid[r][c]
				newRow := r + emptyRows
				b.Grid[newRow][c] = tile
				b.Grid[r][c] = nil
				tile.Row = newRow
				tile.TargetY = float64(newRow * CellSize)
			}
		}

		// Заполняем пустые строки сверху новыми фишками
		for r := 0; r < emptyRows; r++ {
			color := rand.Intn(NumColors)
			tile := b.newTile(color, r, c)
			// Начинаем выше экрана для анимации падения
			tile.Y = float64(-(emptyRows - r) * CellSize)
			b.Grid[r][c] = tile
		}
	}
}

// HasValidMoves проверяет, есть ли на доске валидные ходы
func (b *Board) HasValidMoves() bool {
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			// Пробуем обмен вправо
			if c+1 < BoardSize {
				b.Swap(&Position{Row: r, Col: c}, &Position{Row: r, Col: c + 1})
				if len(b.FindMatches()) > 0 {
					b.Swap(&Position{Row: r, Col: c}, &Position{Row: r, Col: c + 1})
					return true
				}
				b.Swap(&Position{Row: r, Col: c}, &Position{Row: r, Col: c + 1})
			}
			// Пробуем обмен вниз
			if r+1 < BoardSize {
				b.Swap(&Position{Row: r, Col: c}, &Position{Row: r + 1, Col: c})
				if len(b.FindMatches()) > 0 {
					b.Swap(&Position{Row: r, Col: c}, &Position{Row: r + 1, Col: c})
					return true
				}
				b.Swap(&Position{Row: r, Col: c}, &Position{Row: r + 1, Col: c})
			}
		}
	}
	return false
}

// FindHint находит подсказку - валидный ход
func (b *Board) FindHint() (*Position, *Position) {
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			// Пробуем обмен вправо
			if c+1 < BoardSize {
				b.Swap(&Position{Row: r, Col: c}, &Position{Row: r, Col: c + 1})
				if len(b.FindMatches()) > 0 {
					b.Swap(&Position{Row: r, Col: c}, &Position{Row: r, Col: c + 1})
					return &Position{Row: r, Col: c}, &Position{Row: r, Col: c + 1}
				}
				b.Swap(&Position{Row: r, Col: c}, &Position{Row: r, Col: c + 1})
			}
			// Пробуем обмен вниз
			if r+1 < BoardSize {
				b.Swap(&Position{Row: r, Col: c}, &Position{Row: r + 1, Col: c})
				if len(b.FindMatches()) > 0 {
					b.Swap(&Position{Row: r, Col: c}, &Position{Row: r + 1, Col: c})
					return &Position{Row: r, Col: c}, &Position{Row: r + 1, Col: c}
				}
				b.Swap(&Position{Row: r, Col: c}, &Position{Row: r + 1, Col: c})
			}
		}
	}
	return nil, nil
}

// positionKey создаёт уникальный ключ для позиции
func positionKey(row, col int) string {
	return string(rune(row*BoardSize+col))
}

// abs возвращает абсолютное значение
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
