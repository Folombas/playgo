package board

import (
	"math/rand"

	"github.com/playgo/puzzle_go/internal/config"
	"github.com/playgo/puzzle_go/internal/entity"
)

// Board представляет игровое поле
type Board struct {
	Crystals [][]*entity.Crystal
	Width    int
	Height   int
}

// NewBoard создает новое игровое поле
func NewBoard() *Board {
	b := &Board{
		Crystals: make([][]*entity.Crystal, config.BoardCols),
		Width:    config.BoardCols,
		Height:   config.BoardRows,
	}

	// Инициализация столбцов
	for col := 0; col < config.BoardCols; col++ {
		b.Crystals[col] = make([]*entity.Crystal, config.BoardRows)
	}

	// Заполнение без начальных совпадений
	b.fillWithoutMatches()

	return b
}

// fillWithoutMatches заполняет поле без начальных совпадений
func (b *Board) fillWithoutMatches() {
	for col := 0; col < b.Width; col++ {
		for row := 0; row < b.Height; row++ {
			var crystalType entity.CrystalType
			for {
				crystalType = entity.CrystalType(rand.Intn(config.CrystalTypes))
				
				// Проверка на совпадения с соседями
				if !b.wouldCreateMatch(col, row, crystalType) {
					break
				}
			}

			b.Crystals[col][row] = entity.NewCrystal(crystalType, col, row)
		}
	}
}

// wouldCreateMatch проверяет, создаст ли данный тип совпадение
func (b *Board) wouldCreateMatch(col, row int, crystalType entity.CrystalType) bool {
	// Проверка горизонтали
	matchCount := 1
	
	// Влево
	for c := col - 1; c >= 0 && b.Crystals[c][row] != nil && b.Crystals[c][row].Type == crystalType; c-- {
		matchCount++
	}
	
	// Вправо
	for c := col + 1; c < b.Width && b.Crystals[c][row] != nil && b.Crystals[c][row].Type == crystalType; c++ {
		matchCount++
	}

	if matchCount >= 3 {
		return true
	}

	// Проверка вертикали
	matchCount = 1
	
	// Вверх
	for r := row - 1; r >= 0 && b.Crystals[col][r] != nil && b.Crystals[col][r].Type == crystalType; r-- {
		matchCount++
	}
	
	// Вниз
	for r := row + 1; r < b.Height && b.Crystals[col][r] != nil && b.Crystals[col][r].Type == crystalType; r++ {
		matchCount++
	}

	return matchCount >= 3
}

// GetCrystal возвращает кристалл по координатам
func (b *Board) GetCrystal(col, row int) *entity.Crystal {
	if col < 0 || col >= b.Width || row < 0 || row >= b.Height {
		return nil
	}
	return b.Crystals[col][row]
}

// SetCrystal устанавливает кристалл по координатам
func (b *Board) SetCrystal(col, row int, crystal *entity.Crystal) {
	if col >= 0 && col < b.Width && row >= 0 && row < b.Height {
		b.Crystals[col][row] = crystal
	}
}

// Swap меняет местами два кристалла
func (b *Board) Swap(col1, row1, col2, row2 int) bool {
	// Проверка на соседство
	if !b.areAdjacent(col1, row1, col2, row2) {
		return false
	}

	// Обмен в массиве
	crystal1 := b.Crystals[col1][row1]
	crystal2 := b.Crystals[col2][row2]

	b.Crystals[col1][row1] = crystal2
	b.Crystals[col2][row2] = crystal1

	// Обновление позиций
	crystal1.Col, crystal1.Row = col2, row2
	crystal2.Col, crystal2.Row = col1, row1

	crystal1.SetTarget(col2, row2)
	crystal2.SetTarget(col1, row1)

	return true
}

// areAdjacent проверяет, являются ли ячейки соседними
func (b *Board) areAdjacent(col1, row1, col2, row2 int) bool {
	colDiff := abs(col1 - col2)
	rowDiff := abs(row1 - row2)
	return (colDiff == 1 && rowDiff == 0) || (colDiff == 0 && rowDiff == 1)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// FindAllMatches находит все совпадения на поле
func (b *Board) FindAllMatches() [][]*entity.Crystal {
	var allMatches [][]*entity.Crystal
	matched := make(map[[2]int]bool)

	// Поиск горизонтальных совпадений
	for row := 0; row < b.Height; row++ {
		for col := 0; col <= b.Width-3; col++ {
			crystal := b.Crystals[col][row]
			if crystal == nil || crystal.IsMatched {
				continue
			}

			match := b.findHorizontalMatch(col, row)
			if len(match) >= 3 {
				for _, c := range match {
					key := [2]int{c.Col, c.Row}
					if !matched[key] {
						matched[key] = true
						c.IsMatched = true
					}
				}
				allMatches = append(allMatches, match)
			}
		}
	}

	// Поиск вертикальных совпадений
	for col := 0; col < b.Width; col++ {
		for row := 0; row <= b.Height-3; row++ {
			crystal := b.Crystals[col][row]
			if crystal == nil || crystal.IsMatched {
				continue
			}

			match := b.findVerticalMatch(col, row)
			if len(match) >= 3 {
				for _, c := range match {
					key := [2]int{c.Col, c.Row}
					if !matched[key] {
						matched[key] = true
						c.IsMatched = true
					}
				}
				allMatches = append(allMatches, match)
			}
		}
	}

	return allMatches
}

// findHorizontalMatch находит горизонтальное совпадение
func (b *Board) findHorizontalMatch(col, row int) []*entity.Crystal {
	crystal := b.Crystals[col][row]
	if crystal == nil {
		return nil
	}

	match := []*entity.Crystal{crystal}

	// Влево
	for c := col - 1; c >= 0 && b.Crystals[c][row] != nil && b.Crystals[c][row].Type == crystal.Type; c-- {
		match = append([]*entity.Crystal{b.Crystals[c][row]}, match...)
	}

	// Вправо
	for c := col + 1; c < b.Width && b.Crystals[c][row] != nil && b.Crystals[c][row].Type == crystal.Type; c++ {
		match = append(match, b.Crystals[c][row])
	}

	return match
}

// findVerticalMatch находит вертикальное совпадение
func (b *Board) findVerticalMatch(col, row int) []*entity.Crystal {
	crystal := b.Crystals[col][row]
	if crystal == nil {
		return nil
	}

	match := []*entity.Crystal{crystal}

	// Вверх
	for r := row - 1; r >= 0 && b.Crystals[col][r] != nil && b.Crystals[col][r].Type == crystal.Type; r-- {
		match = append([]*entity.Crystal{b.Crystals[col][r]}, match...)
	}

	// Вниз
	for r := row + 1; r < b.Height && b.Crystals[col][r] != nil && b.Crystals[col][r].Type == crystal.Type; r++ {
		match = append(match, b.Crystals[col][r])
	}

	return match
}

// ApplyGravity применяет гравитацию к кристаллам
func (b *Board) ApplyGravity() []*entity.Crystal {
	var fallingCrystals []*entity.Crystal

	for col := 0; col < b.Width; col++ {
		emptyRow := -1

		// Идем снизу вверх
		for row := b.Height - 1; row >= 0; row-- {
			if b.Crystals[col][row] == nil {
				if emptyRow == -1 {
					emptyRow = row
				}
				continue
			}

			if b.Crystals[col][row].IsMatched {
				b.Crystals[col][row] = nil
				if emptyRow == -1 {
					emptyRow = row
				}
				continue
			}

			// Если есть пустые ячейки ниже, сдвигаем кристалл вниз
			if emptyRow != -1 {
				crystal := b.Crystals[col][row]
				b.Crystals[col][emptyRow] = crystal
				b.Crystals[col][row] = nil

				crystal.Row = emptyRow
				crystal.SetTarget(col, emptyRow)
				crystal.IsFalling = true

				fallingCrystals = append(fallingCrystals, crystal)

				emptyRow--
			}
		}

		// Заполняем пустые ячейки новыми кристаллами сверху
		for row := emptyRow; row >= 0; row-- {
			crystalType := entity.CrystalType(rand.Intn(config.CrystalTypes))
			newCrystal := entity.NewCrystal(crystalType, col, row)
			
			// Начальная позиция выше экрана
			newCrystal.Y = float64(config.BoardOffsetY + (row - 3)*(config.CellSize+config.CellPadding))
			newCrystal.IsFalling = true
			
			b.Crystals[col][row] = newCrystal
			fallingCrystals = append(fallingCrystals, newCrystal)
		}
	}

	return fallingCrystals
}

// RemoveMatched удаляет все совпавшие кристаллы
func (b *Board) RemoveMatched() int {
	count := 0

	for col := 0; col < b.Width; col++ {
		for row := 0; row < b.Height; row++ {
			if b.Crystals[col][row] != nil && b.Crystals[col][row].IsMatched {
				b.Crystals[col][row] = nil
				count++
			}
		}
	}

	return count
}

// GetCrystalAtPosition возвращает кристалл по экранным координатам
func (b *Board) GetCrystalAtPosition(mx, my float64) *entity.Crystal {
	for col := 0; col < b.Width; col++ {
		for row := 0; row < b.Height; row++ {
			crystal := b.Crystals[col][row]
			if crystal != nil && crystal.Contains(mx, my) {
				return crystal
			}
		}
	}
	return nil
}

// IsAnimating проверяет, есть ли активные анимации
func (b *Board) IsAnimating() bool {
	for col := 0; col < b.Width; col++ {
		for row := 0; row < b.Height; row++ {
			crystal := b.Crystals[col][row]
			if crystal != nil {
				// Проверяем, достиг ли кристалл целевой позиции
				dx := crystal.X - crystal.TargetX
				dy := crystal.Y - crystal.TargetY
				if dx*dx+dy*dy > 0.1 {
					return true
				}
				if crystal.IsMatched && crystal.Alpha > 0.01 {
					return true
				}
			}
		}
	}
	return false
}
