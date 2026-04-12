package logic

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type GemType int

const (
	GemApple GemType = iota
	GemOrange
	GemLemon
	GemGrape
	GemBerry
	GemMelon
	GemCount
)

var GemColors = []color.RGBA{
	{255, 50, 50, 255},    // Apple - Red
	{255, 165, 0, 255},    // Orange
	{255, 255, 0, 255},    // Lemon - Yellow
	{50, 205, 50, 255},    // Grape - Green
	{0, 191, 255, 255},    // Berry - Blue
	{148, 0, 211, 255},    // Melon - Purple
}

var GemEmojis = []string{"🍎", "🍊", "🍋", "🍇", "🫐", "🍈"}

type Tile struct {
	Gem         GemType
	Row, Col    int
	X, Y        float64
	TargetX     float64
	TargetY     float64
	Selected    bool
	Removing    bool
	Falling     bool
	Scale       float64
	Alpha       float64
	Shake       float64 // Анимация дрожания при невалидном обмене
	IsBomb      bool
	IsRocketV   bool
	IsRocketH   bool
	IsRainbow   bool
}

type Board struct {
	Tiles       [][]*Tile
	Rows        int
	Cols        int
	TileSize    int
	OffsetX     int
	OffsetY     int
	IsAnimating bool
	rng         *rand.Rand
	selected    *Tile
}

func NewBoard(rows, cols, level int) *Board {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	b := &Board{
		Tiles:    make([][]*Tile, rows),
		Rows:     rows,
		Cols:     cols,
		TileSize: 50,
		OffsetX:  45,
		OffsetY:  150,
		rng:      rng,
	}
	
	// Создание доски
	for r := 0; r < rows; r++ {
		b.Tiles[r] = make([]*Tile, cols)
		for c := 0; c < cols; c++ {
			b.Tiles[r][c] = b.createTile(r, c)
		}
	}
	
	// Убрать начальные матчи
	b.RemoveInitialMatches()
	
	return b
}

func (b *Board) createTile(row, col int) *Tile {
	return &Tile{
		Gem:      GemType(b.rng.Intn(int(GemCount))),
		Row:      row,
		Col:      col,
		X:        float64(b.OffsetX + col*b.TileSize),
		Y:        float64(b.OffsetY + row*b.TileSize),
		TargetX:  float64(b.OffsetX + col*b.TileSize),
		TargetY:  float64(b.OffsetY + row*b.TileSize),
		Scale:    1.0,
		Alpha:    1.0,
	}
}

func (b *Board) RemoveInitialMatches() {
	for {
		matches := b.FindAllMatches()
		if len(matches) == 0 {
			break
		}
		for _, t := range matches {
			t.Gem = GemType(b.rng.Intn(int(GemCount)))
		}
	}
}

func (b *Board) Update() {
	// Анимация перемещения
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			tile := b.Tiles[r][c]
			if tile != nil {
				// Плавное перемещение
				tile.X += (tile.TargetX - tile.X) * 0.2
				tile.Y += (tile.TargetY - tile.Y) * 0.2

				// Анимация дрожания
				if tile.Shake > 0 {
					tile.Shake -= 0.05
					if tile.Shake < 0 {
						tile.Shake = 0
					}
				}

				// Анимация удаления
				if tile.Removing {
					tile.Scale *= 0.8
					tile.Alpha *= 0.8
				}
			}
		}
	}
}

func (b *Board) Draw(screen *ebiten.Image) {
	// Фон доски
	bg := ebiten.NewImage(b.Cols*b.TileSize+10, b.Rows*b.TileSize+10)
	bg.Fill(color.RGBA{20, 10, 40, 255})
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Translate(float64(b.OffsetX-5), float64(b.OffsetY-5))
	screen.DrawImage(bg, bgOp)
	
	// Ячейки
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			x := b.OffsetX + c*b.TileSize
			y := b.OffsetY + r*b.TileSize
			
			// Фон ячейки
			cellColor := color.RGBA{50, 40, 70, 255}
			if (r+c)%2 == 0 {
				cellColor = color.RGBA{60, 50, 80, 255}
			}
			
			cell := ebiten.NewImage(b.TileSize-4, b.TileSize-4)
			cell.Fill(cellColor)
			
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x+2), float64(y+2))
			screen.DrawImage(cell, op)
		}
	}
	
	// Фрукты
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			tile := b.Tiles[r][c]
			if tile != nil && !tile.Removing {
				b.drawTile(screen, tile)
			}
		}
	}
}

func (b *Board) drawTile(screen *ebiten.Image, tile *Tile) {
	x := int(tile.X)
	y := int(tile.Y)
	size := b.TileSize - 4

	// Анимация дрожания - смещение туда-сюда
	if tile.Shake > 0 {
		shakeOffset := math.Sin(tile.Shake*math.Pi*10) * 5 * tile.Shake
		x += int(shakeOffset)
	}

	// Выделение
	if tile.Selected {
		highlight := ebiten.NewImage(size+4, size+4)
		highlight.Fill(color.RGBA{255, 215, 0, 200})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x-2), float64(y-2))
		screen.DrawImage(highlight, op)
	}

	// Круглый фрукт
	centerX := x + size/2
	centerY := y + size/2
	radius := size / 2

	circleColor := GemColors[tile.Gem]

	circle := b.createCircle(radius*2, circleColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(centerX-radius), float64(centerY-radius))
	screen.DrawImage(circle, op)

	// Блик
	highlight := b.createCircle(radius/2, color.RGBA{255, 255, 255, 100})
	hlOp := &ebiten.DrawImageOptions{}
	hlOp.GeoM.Translate(float64(centerX-radius/3-radius/4), float64(centerY-radius/3-radius/4))
	screen.DrawImage(highlight, hlOp)

	// Индикаторы бонусов
	if tile.IsBomb {
		bomb := ebiten.NewImage(12, 12)
		bomb.Fill(color.RGBA{0, 0, 0, 255})
		bombOp := &ebiten.DrawImageOptions{}
		bombOp.GeoM.Translate(float64(x+size-10), float64(y+6))
		screen.DrawImage(bomb, bombOp)
	}
}

func (b *Board) createCircle(diameter int, c color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(diameter, diameter)
	center := float64(diameter) / 2
	radius := float64(diameter) / 2
	
	for y := 0; y < diameter; y++ {
		for x := 0; x < diameter; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			if math.Sqrt(dx*dx+dy*dy) <= radius {
				img.Set(x, y, c)
			}
		}
	}
	
	return img
}

func (b *Board) FindAllMatches() []*Tile {
	matched := make(map[int]*Tile)
	
	// Горизонтальные
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols-2; c++ {
			gem := b.Tiles[r][c].Gem
			if gem == b.Tiles[r][c+1].Gem && gem == b.Tiles[r][c+2].Gem {
				matched[r*b.Cols+c] = b.Tiles[r][c]
				matched[r*b.Cols+c+1] = b.Tiles[r][c+1]
				matched[r*b.Cols+c+2] = b.Tiles[r][c+2]
				
				// 4+ в ряд
				for extra := c + 3; extra < b.Cols; extra++ {
					if gem == b.Tiles[r][extra].Gem {
						matched[r*b.Cols+extra] = b.Tiles[r][extra]
					} else {
						break
					}
				}
			}
		}
	}
	
	// Вертикальные
	for c := 0; c < b.Cols; c++ {
		for r := 0; r < b.Rows-2; r++ {
			gem := b.Tiles[r][c].Gem
			if gem == b.Tiles[r+1][c].Gem && gem == b.Tiles[r+2][c].Gem {
				matched[r*b.Cols+c] = b.Tiles[r][c]
				matched[(r+1)*b.Cols+c] = b.Tiles[r+1][c]
				matched[(r+2)*b.Cols+c] = b.Tiles[r+2][c]
				
				// 4+ в ряд
				for extra := r + 3; extra < b.Rows; extra++ {
					if gem == b.Tiles[extra][c].Gem {
						matched[extra*b.Cols+c] = b.Tiles[extra][c]
					} else {
						break
					}
				}
			}
		}
	}
	
	result := make([]*Tile, 0, len(matched))
	for _, t := range matched {
		result = append(result, t)
	}
	
	return result
}

func (b *Board) ProcessMatches() []*Tile {
	matches := b.FindAllMatches()
	
	if len(matches) == 0 {
		return nil
	}
	
	// Проверка на 4+ и 5+ для бонусов
	b.checkSpecialCreation(matches)
	
	// Удаление
	for _, t := range matches {
		t.Removing = true
	}
	
	// Гравитация и заполнение
	time.Sleep(100 * time.Millisecond)
	b.ApplyGravity()
	b.FillEmpty()
	
	return matches
}

func (b *Board) checkSpecialCreation(matches []*Tile) {
	// Группировка по рядам и колонкам
	rowGroups := make(map[int]int)
	colGroups := make(map[int]int)
	
	for _, t := range matches {
		rowGroups[t.Row]++
		colGroups[t.Col]++
	}
	
	// 5 в ряд = радужный шар
	for _, count := range rowGroups {
		if count >= 5 {
			// Создать радужный шар
			// (будет добавлен позже)
		}
	}
	
	for _, count := range colGroups {
		if count >= 5 {
			// Создать радужный шар
		}
	}
	
	// 4 в ряд = бомба/ракета
	for _, count := range rowGroups {
		if count == 4 {
			// Создать ракету
		}
	}
	
	for _, count := range colGroups {
		if count == 4 {
			// Создать ракету
		}
	}
}

func (b *Board) ApplyGravity() {
	for c := 0; c < b.Cols; c++ {
		emptyRow := b.Rows - 1
		
		for r := b.Rows - 1; r >= 0; r-- {
			if !b.Tiles[r][c].Removing {
				if r != emptyRow {
					b.Tiles[emptyRow][c] = b.Tiles[r][c]
					b.Tiles[emptyRow][c].Row = emptyRow
					b.Tiles[emptyRow][c].TargetY = float64(b.OffsetY + emptyRow*b.TileSize)
					b.Tiles[r][c] = nil
				}
				emptyRow--
			}
		}
	}
}

func (b *Board) FillEmpty() {
	for c := 0; c < b.Cols; c++ {
		for r := 0; r < b.Rows; r++ {
			if b.Tiles[r][c] == nil {
				b.Tiles[r][c] = b.createTile(r, c)
				b.Tiles[r][c].Y = float64(b.OffsetY - (b.Rows-r)*b.TileSize)
				b.Tiles[r][c].TargetY = float64(b.OffsetY + r*b.TileSize)
			}
		}
	}
}

func (b *Board) SwapTiles(t1, t2 *Tile) bool {
	// Проверка соседства
	dr := absInt(t1.Row - t2.Row)
	dc := absInt(t1.Col - t2.Col)
	
	if dr+dc != 1 {
		return false
	}
	
	// Обмен
	b.Tiles[t1.Row][t1.Col], b.Tiles[t2.Row][t2.Col] = b.Tiles[t2.Row][t2.Col], b.Tiles[t1.Row][t1.Col]
	t1.Row, t2.Row = t2.Row, t1.Row
	t1.Col, t2.Col = t2.Col, t1.Col
	
	t1.TargetX = float64(b.OffsetX + t1.Col*b.TileSize)
	t1.TargetY = float64(b.OffsetY + t1.Row*b.TileSize)
	t2.TargetX = float64(b.OffsetX + t2.Col*b.TileSize)
	t2.TargetY = float64(b.OffsetY + t2.Row*b.TileSize)
	
	return true
}

func (b *Board) GetTileAt(x, y int) *Tile {
	col := (x - b.OffsetX) / b.TileSize
	row := (y - b.OffsetY) / b.TileSize
	
	if row >= 0 && row < b.Rows && col >= 0 && col < b.Cols {
		return b.Tiles[row][col]
	}
	return nil
}

func (b *Board) HasValidMoves() bool {
	// TODO: Проверить наличие возможных ходов
	return true
}

func (b *Board) FindHint() (tile1, tile2 *Tile) {
	// Проверка всех возможных обменов
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			// Вправо
			if c < b.Cols-1 {
				b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
				if len(b.FindAllMatches()) > 0 {
					b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
					return b.Tiles[r][c], b.Tiles[r][c+1]
				}
				b.Tiles[r][c].Gem, b.Tiles[r][c+1].Gem = b.Tiles[r][c+1].Gem, b.Tiles[r][c].Gem
			}
			// Вниз
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

func absInt(x int) int {
	return int(math.Abs(float64(x)))
}
