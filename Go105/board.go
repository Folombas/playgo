package main

import "image/color"

// Board представляет игровое поле
type Board struct {
	Grid [boardRows][boardCols]color.Color
}

// NewBoard создаёт новое пустое поле
func NewBoard() *Board {
	return &Board{}
}

// Collides проверяет, коллайдит ли фигура с границами или другими блоками
func (b *Board) Collides(t *Tetromino) bool {
	for y, row := range t.Shape() {
		for x, cell := range row {
			if cell == 0 {
				continue
			}
			boardX := t.X + x
			boardY := t.Y + y

			// Проверка границ
			if boardX < 0 || boardX >= boardCols || boardY >= boardRows {
				return true
			}
			// Не проверять выше поля (фигура ещё входит)
			if boardY < 0 {
				continue
			}
			// Проверка занятости
			if b.Grid[boardY][boardX] != nil {
				return true
			}
		}
	}
	return false
}

// Place фиксирует фигуру на поле
func (b *Board) Place(t *Tetromino) {
	for y, row := range t.Shape() {
		for x, cell := range row {
			if cell == 0 {
				continue
			}
			boardX := t.X + x
			boardY := t.Y + y
			if boardY >= 0 && boardY < boardRows && boardX >= 0 && boardX < boardCols {
				b.Grid[boardY][boardX] = t.Color()
			}
		}
	}
}

// ClearLines удаляет заполненные линии и возвращает их количество
func (b *Board) ClearLines() int {
	cleared := 0
	rowsToClear := []int{}

	for y := 0; y < boardRows; y++ {
		full := true
		for x := 0; x < boardCols; x++ {
			if b.Grid[y][x] == nil {
				full = false
				break
			}
		}
		if full {
			rowsToClear = append(rowsToClear, y)
		}
	}

	// Удалить строки снизу вверх
	for _, rowIdx := range rowsToClear {
		b.clearRow(rowIdx)
		cleared++
	}

	return cleared
}

// clearRow очищает одну строку и сдвигает всё вниз
func (b *Board) clearRow(rowIdx int) {
	// Сдвинуть все строки выше вниз
	for y := rowIdx; y > 0; y-- {
		for x := 0; x < boardCols; x++ {
			b.Grid[y][x] = b.Grid[y-1][x]
		}
	}
	// Очистить верхнюю строку
	for x := 0; x < boardCols; x++ {
		b.Grid[0][x] = nil
	}
}

// GetDropY возвращает Y-координату, куда упадёт фигура (для тени)
func (b *Board) GetDropY(t *Tetromino) int {
	shadow := &Tetromino{
		Type:       t.Type,
		Rotation:   t.Rotation,
		X:          t.X,
		Y:          t.Y,
	}

	for shadow.Y < boardRows {
		shadow.Y++
		if b.Collides(shadow) {
			shadow.Y--
			break
		}
	}
	return shadow.Y
}
