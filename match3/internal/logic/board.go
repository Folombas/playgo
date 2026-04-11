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
	GemBomb GemType = -1 // Специальная бомба (при матче 4+)
)

// Tile представляет одну ячейку на доске
type Tile struct {
	Gem       GemType
	Row       int
	Col       int
	Selected  bool
	Removing  bool
	Falling   bool
	OffsetY   float64
	IsBomb    bool    // Является ли этот камень бомбой
	IsFire    bool    // Огненный камень - уничтожает весь ряд
	IsIce     bool    // Ледяной камень - требует двойного клика
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
	SwapAnim   *SwapAnimation // Текущая анимация обмена
	ImageCache *ImageCache    // Кэш изображений
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
		Tiles:      make([][]*Tile, rows),
		Rows:       rows,
		Cols:       cols,
		TileSize:   60,
		OffsetX:    40,
		OffsetY:    150,
		ImageCache: NewImageCache(),
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
// Оптимизированная версия без аллокаций строк
func (b *Board) FindAllMatches() []*Tile {
	// Используем map[int64]*Tile для избежания строковых аллокаций
	// Ключ: (row << 32) | col
	matched := make(map[int64]*Tile)

	// Проверка горизонтальных матчей
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols-2; c++ {
			gem := b.Tiles[r][c].Gem
			if gem == b.Tiles[r][c+1].Gem && gem == b.Tiles[r][c+2].Gem {
				// Нашли матч - добавляем все 3 камня
				key1 := (int64(r) << 32) | int64(c)
				key2 := (int64(r) << 32) | int64(c+1)
				key3 := (int64(r) << 32) | int64(c+2)
				matched[key1] = b.Tiles[r][c]
				matched[key2] = b.Tiles[r][c+1]
				matched[key3] = b.Tiles[r][c+2]
				
				// Проверяем, есть ли больше 3 в ряд
				for extra := c + 3; extra < b.Cols; extra++ {
					if gem == b.Tiles[r][extra].Gem {
						keyExtra := (int64(r) << 32) | int64(extra)
						matched[keyExtra] = b.Tiles[r][extra]
					} else {
						break
					}
				}
			}
		}
	}

	// Проверка вертикальных матчей
	for c := 0; c < b.Cols; c++ {
		for r := 0; r < b.Rows-2; r++ {
			gem := b.Tiles[r][c].Gem
			if gem == b.Tiles[r+1][c].Gem && gem == b.Tiles[r+2][c].Gem {
				// Нашли матч - добавляем все 3 камня
				key1 := (int64(r) << 32) | int64(c)
				key2 := (int64(r+1) << 32) | int64(c)
				key3 := (int64(r+2) << 32) | int64(c)
				matched[key1] = b.Tiles[r][c]
				matched[key2] = b.Tiles[r+1][c]
				matched[key3] = b.Tiles[r+2][c]
				
				// Проверяем, есть ли больше 3 в ряд
				for extra := r + 3; extra < b.Rows; extra++ {
					if gem == b.Tiles[extra][c].Gem {
						keyExtra := (int64(extra) << 32) | int64(c)
						matched[keyExtra] = b.Tiles[extra][c]
					} else {
						break
					}
				}
			}
		}
	}

	// Преобразование в слайс
	result := make([]*Tile, 0, len(matched))
	for _, t := range matched {
		result = append(result, t)
	}

	return result
}

// Update обновляет состояние доски каждый кадр
func (b *Board) Update() {
	// Обновление анимации обмена
	if b.SwapAnim != nil && !b.SwapAnim.IsComplete() {
		b.SwapAnim.Update(1.0 / 60.0) // Предполагаем 60 FPS
	}
	
	// Анимация падения с ускорением
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			tile := b.Tiles[r][c]
			if tile.Falling && tile.OffsetY > 0 {
				// Ускорение падения (гравитация)
				tile.OffsetY -= 8 + tile.OffsetY*0.05
				if tile.OffsetY <= 0 {
					tile.OffsetY = 0
					tile.Falling = false
					// Эффект приземления
					b.SpawnLandEffect(r, c)
				}
			}
		}
	}
}

// SpawnLandEffect создаёт эффект приземления камня
func (b *Board) SpawnLandEffect(row, col int) {
	// Можно добавить мелкие эффекты приземления
	// Пока оставим пустым для будущей реализации
}

// Draw отрисовывает доску
func (b *Board) Draw(screen *ebiten.Image) {
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			tile := b.Tiles[r][c]
			x := b.OffsetX + c*b.TileSize
			y := b.OffsetY + r*b.TileSize + int(tile.OffsetY)
			
			// Применяем смещение анимации, если есть
			if b.SwapAnim != nil && !b.SwapAnim.IsComplete() {
				if tile == b.SwapAnim.Tile1 {
					offX, offY := b.SwapAnim.GetOffset1()
					x += int(offX)
					y += int(offY)
				} else if tile == b.SwapAnim.Tile2 {
					offX, offY := b.SwapAnim.GetOffset2()
					x += int(offX)
					y += int(offY)
				}
			}

			// Отрисовка фона ячейки (из кэша)
			rect := b.ImageCache.GetTileBackground(b.TileSize)

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

						// Выделение выбранного камня (из кэша)
						if tile.Selected {
							highlight := b.ImageCache.GetHighlight(b.TileSize)
							hlOp := &ebiten.DrawImageOptions{}
							hlOp.GeoM.Translate(float64(x), float64(y))
							screen.DrawImage(highlight, hlOp)
						}

						screen.DrawImage(sprite, op)
						
						// Если это бомба, рисуем индикатор
						if tile.IsBomb {
							b.drawBombIndicator(screen, x, y, b.TileSize)
						}
					} else {
						// Fallback на цвета
						b.drawGemFallback(screen, x, y, tile.Gem, tile.Selected)
						if tile.IsBomb {
							b.drawBombIndicator(screen, x, y, b.TileSize)
						}
					}
				} else {
					// Fallback на цвета
					b.drawGemFallback(screen, x, y, tile.Gem, tile.Selected)
					if tile.IsBomb {
						b.drawBombIndicator(screen, x, y, b.TileSize)
					}
				}
			}
		}
	}
}

// drawBombIndicator рисует маленький значок бомбы в углу камня
func (b *Board) drawBombIndicator(screen *ebiten.Image, x, y, tileSize int) {
	// Маленький кружок в правом верхнем углу (из кэша)
	bombSize := 12
	bombX := x + tileSize - bombSize - 2
	bombY := y + 2
	
	bomb := b.ImageCache.GetBombIndicator(bombSize)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(bombX), float64(bombY))
	screen.DrawImage(bomb, op)
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

// SwapTiles меняет местами два соседних камня и создаёт анимацию
func (b *Board) SwapTiles(t1, t2 *Tile) bool {
	// Проверка на соседство
	dr := abs(t1.Row - t2.Row)
	dc := abs(t1.Col - t2.Col)

	if dr+dc != 1 {
		return false
	}

	// Создаём анимацию
	b.SwapAnim = NewSwapAnimation(t1, t2, b.TileSize)

	// Обмен камнями (визуально будет через анимацию)
	b.Tiles[t1.Row][t1.Col].Gem, b.Tiles[t2.Row][t2.Col].Gem =
		b.Tiles[t2.Row][t2.Col].Gem, b.Tiles[t1.Row][t1.Col].Gem

	return true
}

// RemoveMatches удаляет найденные совпадения и возвращает очки
// При матче 4+ создаёт специальную бомбу
func (b *Board) RemoveMatches() (score int, bombCreated *Tile) {
	matches := b.FindAllMatches()
	if len(matches) == 0 {
		return 0, nil
	}

	// Бонус за большее количество
	score = len(matches) * 10
	if len(matches) >= 4 {
		score *= 2
		
		// Создаём бомбу на месте центрального камня матча
		bombCreated = b.createBomb(matches)
		// createBomb уже пометил камни для удаления
		b.ApplyGravity()
		return score, bombCreated
	} else if len(matches) == 3 {
		// Обычный матч 3
		score = 30
	}

	// Удаление камней (обычный матч 3)
	for _, t := range matches {
		t.Gem = GemType(-1) // Пустая ячейка
		t.Removing = true
	}

	// Падение камней сверху
	b.ApplyGravity()

	return score, nil
}

// createBomb создаёт бомбу на месте центрального камня из матча
func (b *Board) createBomb(matches []*Tile) *Tile {
	// Выбираем центральный камень
	centerIdx := len(matches) / 2
	bombTile := matches[centerIdx]
	
	// Превращаем в бомбу
	bombTile.Gem = GemType(rng.Intn(int(GemCount))) // Обычный камень для отображения
	bombTile.IsBomb = true
	bombTile.Removing = false // Не удалять!
	
	// При матче 5+ создаём огненный камень вместо бомбы
	if len(matches) >= 5 {
		bombTile.IsBomb = false
		bombTile.IsFire = true
		fmt.Printf("🔥 Огненный камень создан на (%d, %d)!\n", bombTile.Row, bombTile.Col)
	} else {
		fmt.Printf("💣 Бомба создана на (%d, %d)!\n", bombTile.Row, bombTile.Col)
	}
	
	// Помечаем все остальные камни матча для удаления
	for _, t := range matches {
		if t != bombTile {
			t.Gem = GemType(-1)
			t.Removing = true
		}
	}
	
	return bombTile
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
	tile1, _ := b.FindHint()
	return tile1 != nil
}

// FindHint находит один возможный ход и возвращает два камня для обмена
// Возвращает nil, если ходов нет
func (b *Board) FindHint() (tile1, tile2 *Tile) {
	// Проверка всех возможных обменов
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			// Проверка обмена вправо
			if c < b.Cols-1 {
				b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
				if len(b.FindAllMatches()) > 0 {
					b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
					return b.Tiles[r][c], b.Tiles[r][c+1]
				}
				b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
			}

			// Проверка обмена вниз
			if r < b.Rows-1 {
				b.Tiles[r][c].Gem, b.Tiles[r+1][c].Gem = b.Tiles[r+1][c].Gem, b.Tiles[r][c].Gem
				if len(b.FindAllMatches()) > 0 {
					b.Tiles[r][c].Gem, b.Tiles[r+1][c].Gem = b.Tiles[r+1][c].Gem, b.Tiles[r][c].Gem
					return b.Tiles[r][c], b.Tiles[r+1][c]
				}
				b.Tiles[r][c].Gem, b.Tiles[r+1][c].Gem = b.Tiles[r+1][c].Gem, b.Tiles[r][c].Gem
			}
		}
	}
	return nil, nil
}

// Shuffle перемешивает доску случайным образом
func (b *Board) Shuffle() {
	// Собираем все камни
	allGems := make([]GemType, 0, b.Rows*b.Cols)
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			if b.Tiles[r][c].Gem != GemType(-1) {
				allGems = append(allGems, b.Tiles[r][c].Gem)
			}
		}
	}
	
	// Перемешиваем
	rng.Shuffle(len(allGems), func(i, j int) {
		allGems[i], allGems[j] = allGems[j], allGems[i]
	})
	
	// Возвращаем на доску
	idx := 0
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			if idx < len(allGems) {
				b.Tiles[r][c].Gem = allGems[idx]
				b.Tiles[r][c].IsBomb = false // Сбрасываем бомбы
				b.Tiles[r][c].Falling = true
				b.Tiles[r][c].OffsetY = float64(rng.Intn(b.TileSize * 2))
				idx++
			}
		}
	}
	
	// Убираем начальные матчи
	b.RemoveInitialMatches()
	
	fmt.Println("🔀 Доска перемешана!")
}

// abs возвращает абсолютное значение
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
