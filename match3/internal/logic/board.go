package logic

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// rng - глобальный генератор случайных чисел (потокобезопасный в Go 1.20+)
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetRNG возвращает генератор случайных чисел
func GetRNG() *rand.Rand {
	return rng
}

// GemType определяет тип драгоценного камня
type GemType int

const (
	GemRed GemType = iota
	GemBlue
	GemGreen
	GemYellow
	GemPurple
	GemOrange
	GemCount
)

// Tile представляет одну ячейку на доске
type Tile struct {
	Gem      GemType
	Row      int
	Col      int
	Selected bool
	Removing bool
	Falling  bool
	OffsetY  float64
}

// Board представляет игровое поле
type Board struct {
	Tiles      [][]*Tile
	Rows       int
	Cols       int
	TileSize   int
	OffsetX    int
	OffsetY    int
	GemSprites map[int]*ebiten.Image
}

// GemColors содержит цвета для разных типов камней
var GemColors = []color.Color{
	color.RGBA{255, 50, 50, 255},    // Red
	color.RGBA{50, 100, 255, 255},   // Blue
	color.RGBA{50, 200, 50, 255},    // Green
	color.RGBA{255, 255, 50, 255},   // Yellow
	color.RGBA{180, 50, 255, 255},   // Purple
	color.RGBA{255, 165, 0, 255},    // Orange
}

// Цвета для разных типов камней (внутренняя)
var gemColors = GemColors

// SetGemSprites устанавливает спрайты камней
func (b *Board) SetGemSprites(sprites map[int]*ebiten.Image) {
	b.GemSprites = sprites
}

// drawGemFallback отрисовывает камень цветом (fallback)
func (b *Board) drawGemFallback(screen *ebiten.Image, x, y int, gemType GemType, selected bool) {
	gemImg := ebiten.NewImage(b.TileSize-4, b.TileSize-4)
	gemImg.Fill(gemColors[gemType])
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x+2), float64(y+2))
	
	if selected {
		op.GeoM.Translate(-2, -2)
		gemOut := ebiten.NewImage(b.TileSize+2, b.TileSize+2)
		gemOut.Fill(color.White)
		gemOut.DrawImage(gemImg, nil)
		screen.DrawImage(gemOut, op)
	} else {
		screen.DrawImage(gemImg, op)
	}
}

// NewBoard создаёт новую игровую доску заданного размера
func NewBoard(rows, cols int) *Board {
	b := &Board{
		Tiles:    make([][]*Tile, rows),
		Rows:     rows,
		Cols:     cols,
		TileSize: 60,
		OffsetX:  40,
		OffsetY:  150,
	}

	// Инициализация доски случайными камнями
	for r := 0; r < rows; r++ {
		b.Tiles[r] = make([]*Tile, cols)
		for c := 0; c < cols; c++ {
			b.Tiles[r][c] = b.createRandomTile(r, c)
		}
	}

	// Убрать начальные матчи
	b.RemoveInitialMatches()

	return b
}

// createRandomTile создаёт случайный камень
func (b *Board) createRandomTile(row, col int) *Tile {
	return &Tile{
		Gem:     GemType(rng.Intn(int(GemCount))),
		Row:     row,
		Col:     col,
		OffsetY: 0,
	}
}

// RemoveInitialMatches убирает начальные совпадения при генерации
func (b *Board) RemoveInitialMatches() {
	for {
		matches := b.FindAllMatches()
		if len(matches) == 0 {
			break
		}
		for _, m := range matches {
			b.Tiles[m.Row][m.Col].Gem = GemType(rng.Intn(int(GemCount)))
		}
	}
}

// FindAllMatches находит все текущие совпадения на доске
func (b *Board) FindAllMatches() []*Tile {
	matched := make(map[string]*Tile)

	// Проверка горизонтальных матчей
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols-2; c++ {
			gem := b.Tiles[r][c].Gem
			if gem == b.Tiles[r][c+1].Gem && gem == b.Tiles[r][c+2].Gem {
				matched[fmt.Sprintf("%d-%d", r, c)] = b.Tiles[r][c]
				matched[fmt.Sprintf("%d-%d", r, c+1)] = b.Tiles[r][c+1]
				matched[fmt.Sprintf("%d-%d", r, c+2)] = b.Tiles[r][c+2]
			}
		}
	}

	// Проверка вертикальных матчей
	for c := 0; c < b.Cols; c++ {
		for r := 0; r < b.Rows-2; r++ {
			gem := b.Tiles[r][c].Gem
			if gem == b.Tiles[r+1][c].Gem && gem == b.Tiles[r+2][c].Gem {
				matched[fmt.Sprintf("%d-%d", r, c)] = b.Tiles[r][c]
				matched[fmt.Sprintf("%d-%d", r+1, c)] = b.Tiles[r+1][c]
				matched[fmt.Sprintf("%d-%d", r+2, c)] = b.Tiles[r+2][c]
			}
		}
	}

	// Преобразование в слайс
	var result []*Tile
	for _, t := range matched {
		result = append(result, t)
	}

	return result
}

// Update обновляет состояние доски каждый кадр
func (b *Board) Update() {
	// Анимация падения
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			tile := b.Tiles[r][c]
			if tile.Falling && tile.OffsetY > 0 {
				tile.OffsetY -= 5
				if tile.OffsetY <= 0 {
					tile.OffsetY = 0
					tile.Falling = false
				}
			}
		}
	}
}

// Draw отрисовывает доску
func (b *Board) Draw(screen *ebiten.Image) {
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			tile := b.Tiles[r][c]
			x := b.OffsetX + c*b.TileSize
			y := b.OffsetY + r*b.TileSize + int(tile.OffsetY)

			// Отрисовка фона ячейки
			rect := ebiten.NewImage(b.TileSize, b.TileSize)
			rect.Fill(color.RGBA{200, 200, 200, 255})
			
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			screen.DrawImage(rect, op)

			// Отрисовка камня
			if tile.Gem >= 0 && tile.Gem < GemCount {
				// Используем спрайт если доступен
				if b.GemSprites != nil {
					if sprite, ok := b.GemSprites[int(tile.Gem)]; ok && sprite != nil {
						op := &ebiten.DrawImageOptions{}
						op.GeoM.Translate(float64(x+2), float64(y+2))
						
						// Масштабирование до размера ячейки
						scale := float64(b.TileSize-4) / 32.0
						op.GeoM.Scale(scale, scale)
						
						// Выделение выбранного камня
						if tile.Selected {
							// Рисуем белую рамку
							highlight := ebiten.NewImage(b.TileSize, b.TileSize)
							highlight.Fill(color.White)
							hlOp := &ebiten.DrawImageOptions{}
							hlOp.GeoM.Translate(float64(x), float64(y))
							screen.DrawImage(highlight, hlOp)
						}
						
						screen.DrawImage(sprite, op)
					} else {
						// Fallback на цвета
						b.drawGemFallback(screen, x, y, tile.Gem, tile.Selected)
					}
				} else {
					// Fallback на цвета
					b.drawGemFallback(screen, x, y, tile.Gem, tile.Selected)
				}
			}
		}
	}
}

// GetTileAt возвращает камень по экранным координатам
func (b *Board) GetTileAt(x, y int) *Tile {
	col := (x - b.OffsetX) / b.TileSize
	row := (y - b.OffsetY) / b.TileSize

	if row >= 0 && row < b.Rows && col >= 0 && col < b.Cols {
		return b.Tiles[row][col]
	}
	return nil
}

// SwapTiles меняет местами два соседних камня
func (b *Board) SwapTiles(t1, t2 *Tile) bool {
	// Проверка на соседство
	dr := abs(t1.Row - t2.Row)
	dc := abs(t1.Col - t2.Col)
	
	if dr+dc != 1 {
		return false
	}

	// Обмен камнями
	b.Tiles[t1.Row][t1.Col].Gem, b.Tiles[t2.Row][t2.Col].Gem = 
		b.Tiles[t2.Row][t2.Col].Gem, b.Tiles[t1.Row][t1.Col].Gem

	return true
}

// RemoveMatches удаляет найденные совпадения и возвращает очки
func (b *Board) RemoveMatches() int {
	matches := b.FindAllMatches()
	if len(matches) == 0 {
		return 0
	}

	// Бонус за большее количество
	score := len(matches) * 10
	if len(matches) > 3 {
		score *= 2
	}

	// Удаление камней
	for _, t := range matches {
		t.Gem = GemType(-1) // Пустая ячейка
		t.Removing = true
	}

	// Падение камней сверху
	b.ApplyGravity()

	return score
}

// ApplyGravity применяет гравитацию к камням
func (b *Board) ApplyGravity() {
	for c := 0; c < b.Cols; c++ {
		emptyRow := b.Rows - 1
		
		for r := b.Rows - 1; r >= 0; r-- {
			if b.Tiles[r][c].Gem != GemType(-1) {
				if r != emptyRow {
					b.Tiles[emptyRow][c].Gem = b.Tiles[r][c].Gem
					b.Tiles[emptyRow][c].Falling = true
					b.Tiles[emptyRow][c].OffsetY = float64((emptyRow - r) * b.TileSize)
					b.Tiles[r][c].Gem = GemType(-1)
				}
				emptyRow--
			}
		}

		// Заполнение пустых ячеек сверху новыми камнями
		for r := emptyRow; r >= 0; r-- {
			b.Tiles[r][c].Gem = GemType(rng.Intn(int(GemCount)))
			b.Tiles[r][c].Falling = true
			b.Tiles[r][c].OffsetY = float64((emptyRow - r + 1) * b.TileSize)
		}
	}
}

// HasValidModes проверяет, есть ли допустимые ходы
func (b *Board) HasValidMoves() bool {
	// Проверка всех возможных обменов
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			// Проверка обмена вправо
			if c < b.Cols-1 {
				b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
				if len(b.FindAllMatches()) > 0 {
					b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
					return true
				}
				b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
			}
			
			// Проверка обмена вниз
			if r < b.Rows-1 {
				b.Tiles[r][c].Gem, b.Tiles[r+1][c].Gem = b.Tiles[r+1][c].Gem, b.Tiles[r][c].Gem
				if len(b.FindAllMatches()) > 0 {
					b.Tiles[r][c].Gem, b.Tiles[r+1][c].Gem = b.Tiles[r+1][c].Gem, b.Tiles[r][c].Gem
					return true
				}
				b.Tiles[r][c].Gem, b.Tiles[r+1][c].Gem = b.Tiles[r+1][c].Gem, b.Tiles[r][c].Gem
			}
		}
	}
	return false
}

// abs возвращает абсолютное значение
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
