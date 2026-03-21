package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Sprite - структура для работы со спрайтами
type Sprite struct {
	image  *ebiten.Image
	width  int
	height int
}

// Tileset - набор тайлов
type Tileset struct {
	tiles    map[BlockType]*Sprite
	tileSize int
}

// InitTileset инициализирует набор тайлов
func InitTileset() *Tileset {
	ts := &Tileset{
		tiles:    make(map[BlockType]*Sprite),
		tileSize: blockSize,
	}
	ts.generateProceduralTiles()
	return ts
}

// generateProceduralTiles генерирует текстуры процедурно
func (ts *Tileset) generateProceduralTiles() {
	ts.tiles[Dirt] = ts.createDirtTile()
	ts.tiles[Grass] = ts.createGrassTile()
	ts.tiles[Stone] = ts.createStoneTile()
	ts.tiles[Wood] = ts.createWoodTile()
	ts.tiles[Leaves] = ts.createLeavesTile()
	ts.tiles[Bricks] = ts.createBricksTile()
	ts.tiles[Coal_Ore] = ts.createCoalOreTile()
	ts.tiles[Iron_Ore] = ts.createIronOreTile()
	ts.tiles[Gold_Ore] = ts.createGoldOreTile()
	ts.tiles[Diamond_Ore] = ts.createDiamondOreTile()
	ts.tiles[Plank] = ts.createPlankTile()
	ts.tiles[Crafting_Table] = ts.createCraftingTableTile()
}

// DrawTile отрисовывает тайл
func (ts *Tileset) DrawTile(screen *ebiten.Image, blockType BlockType, screenX, screenY float32) {
	sprite, exists := ts.tiles[blockType]
	if !exists || sprite == nil {
		vector.DrawFilledRect(screen, screenX, screenY, float32(ts.tileSize), float32(ts.tileSize), color.RGBA{255, 0, 255, 255}, false)
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(screenX), float64(screenY))
	screen.DrawImage(sprite.image, op)
}

// createDirtTile создаёт текстуру земли
func (ts *Tileset) createDirtTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*7 + y*13) % 20
			img.Set(x, y, color.RGBA{uint8(100 + noise), uint8(60 + noise/2), uint8(30 + noise/4), 255})
		}
	}
	for i := 0; i < 5; i++ {
		px := (i * 37) % ts.tileSize
		py := (i*23 + 10) % ts.tileSize
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if px+dx >= 0 && px+dx < ts.tileSize && py+dy >= 0 && py+dy < ts.tileSize {
					img.Set(px+dx, py+dy, color.RGBA{80, 50, 30, 255})
				}
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createGrassTile создаёт текстуру травы
func (ts *Tileset) createGrassTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := ts.tileSize / 3; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*7 + y*13) % 15
			img.Set(x, y, color.RGBA{uint8(100 + noise), uint8(60 + noise/2), 30, 255})
		}
	}
	for y := 0; y < ts.tileSize/3; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*5 + y*7) % 20
			img.Set(x, y, color.RGBA{uint8(34 + noise/2), uint8(120 + noise), uint8(34 + noise/3), 255})
		}
	}
	for x := 0; x < ts.tileSize; x += 4 {
		bladeHeight := 3 + (x % 4)
		for by := 0; by < bladeHeight; by++ {
			if by < ts.tileSize/3 {
				img.Set(x, by, color.RGBA{30, 100, 30, 255})
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createStoneTile создаёт текстуру камня
func (ts *Tileset) createStoneTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*11 + y*17) % 30
			gray := uint8(100 + noise)
			img.Set(x, y, color.RGBA{gray, gray, gray + 5, 255})
		}
	}
	for i := 0; i < 8; i++ {
		cx := (i * 43) % ts.tileSize
		cy := (i * 29) % ts.tileSize
		for c := 0; c < 5; c++ {
			if cx+c < ts.tileSize {
				img.Set(cx+c, cy+c/2, color.RGBA{60, 60, 65, 255})
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createWoodTile создаёт текстуру дерева
func (ts *Tileset) createWoodTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*3 + y*7) % 25
			img.Set(x, y, color.RGBA{uint8(80 + noise), uint8(50 + noise/3), uint8(25 + noise/5), 255})
		}
	}
	for ring := 5; ring < ts.tileSize; ring += 15 {
		for x := 0; x < ts.tileSize; x++ {
			if x%2 == 0 {
				img.Set(x, ring, color.RGBA{60, 35, 20, 255})
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createLeavesTile создаёт текстуру листвы
func (ts *Tileset) createLeavesTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*9 + y*11) % 30
			img.Set(x, y, color.RGBA{uint8(20 + noise/3), uint8(50 + noise), uint8(20 + noise/4), 255})
		}
	}
	for i := 0; i < 10; i++ {
		lx := (i * 31) % (ts.tileSize - 4)
		ly := (i * 27) % (ts.tileSize - 4)
		for dy := 0; dy < 3; dy++ {
			for dx := 0; dx < 3; dx++ {
				if lx+dx < ts.tileSize && ly+dy < ts.tileSize {
					img.Set(lx+dx, ly+dy, color.RGBA{25, 90, 25, 255})
				}
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createBricksTile создаёт текстуру кирпичей
func (ts *Tileset) createBricksTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	brickHeight := ts.tileSize / 4
	for row := 0; row < 4; row++ {
		offset := (row % 2) * (ts.tileSize / 2)
		for brick := 0; brick < 3; brick++ {
			bx := offset + brick*(ts.tileSize/2)
			by := row * brickHeight
			for y := by + 1; y < by+brickHeight-1 && y < ts.tileSize; y++ {
				for x := bx + 1; x < bx+ts.tileSize/2-1 && x < ts.tileSize; x++ {
					noise := (x*5 + y*7) % 20
					img.Set(x, y, color.RGBA{uint8(150 + noise), uint8(40 + noise/2), 40, 255})
				}
			}
			for x := bx; x < bx+ts.tileSize/2 && x < ts.tileSize; x++ {
				img.Set(x, by, color.RGBA{100, 30, 30, 255})
				if by+brickHeight < ts.tileSize {
					img.Set(x, by+brickHeight, color.RGBA{80, 25, 25, 255})
				}
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createCoalOreTile создаёт текстуру угольной руды
func (ts *Tileset) createCoalOreTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*11 + y*17) % 25
			gray := uint8(110 + noise)
			img.Set(x, y, color.RGBA{gray, gray, gray + 5, 255})
		}
	}
	coalSpots := [][2]int{{10, 10}, {25, 15}, {15, 28}, {28, 28}}
	for _, spot := range coalSpots {
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if dx*dx+dy*dy <= 4 {
					x := spot[0] + dx
					y := spot[1] + dy
					if x >= 0 && x < ts.tileSize && y >= 0 && y < ts.tileSize {
						img.Set(x, y, color.RGBA{30, 30, 35, 255})
					}
				}
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createIronOreTile создаёт текстуру железной руды
func (ts *Tileset) createIronOreTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*11 + y*17) % 25
			gray := uint8(110 + noise)
			img.Set(x, y, color.RGBA{gray, gray, gray + 5, 255})
		}
	}
	ironSpots := [][2]int{{12, 12}, {28, 18}, {18, 30}}
	for _, spot := range ironSpots {
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if dx*dx+dy*dy <= 5 {
					x := spot[0] + dx
					y := spot[1] + dy
					if x >= 0 && x < ts.tileSize && y >= 0 && y < ts.tileSize {
						img.Set(x, y, color.RGBA{180, 120, 80, 255})
					}
				}
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createGoldOreTile создаёт текстуру золотой руды
func (ts *Tileset) createGoldOreTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*11 + y*17) % 25
			gray := uint8(110 + noise)
			img.Set(x, y, color.RGBA{gray, gray, gray + 5, 255})
		}
	}
	goldSpots := [][2]int{{10, 15}, {25, 12}, {15, 28}, {30, 25}}
	for _, spot := range goldSpots {
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if dx*dx+dy*dy <= 4 {
					x := spot[0] + dx
					y := spot[1] + dy
					if x >= 0 && x < ts.tileSize && y >= 0 && y < ts.tileSize {
						img.Set(x, y, color.RGBA{255, 200, 50, 255})
					}
				}
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createDiamondOreTile создаёт текстуру алмазной руды
func (ts *Tileset) createDiamondOreTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*11 + y*17) % 25
			gray := uint8(110 + noise)
			img.Set(x, y, color.RGBA{gray, gray, gray + 5, 255})
		}
	}
	diamondSpots := [][2]int{{12, 12}, {28, 20}, {20, 30}}
	for _, spot := range diamondSpots {
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if dx*dx+dy*dy <= 4 {
					x := spot[0] + dx
					y := spot[1] + dy
					if x >= 0 && x < ts.tileSize && y >= 0 && y < ts.tileSize {
						img.Set(x, y, color.RGBA{50, 255, 255, 255})
					}
				}
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createPlankTile создаёт текстуру досок
func (ts *Tileset) createPlankTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	plankHeight := ts.tileSize / 4
	for row := 0; row < 4; row++ {
		by := row * plankHeight
		for y := by; y < by+plankHeight && y < ts.tileSize; y++ {
			for x := 0; x < ts.tileSize; x++ {
				noise := (x*3 + y*7) % 20
				img.Set(x, y, color.RGBA{uint8(180 + noise), uint8(140 + noise/2), uint8(100 + noise/3), 255})
			}
		}
		if row < 3 {
			for x := 0; x < ts.tileSize; x++ {
				img.Set(x, by+plankHeight-1, color.RGBA{100, 70, 40, 255})
			}
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// createCraftingTableTile создаёт текстуру верстака
func (ts *Tileset) createCraftingTableTile() *Sprite {
	img := ebiten.NewImage(ts.tileSize, ts.tileSize)
	for y := 0; y < ts.tileSize; y++ {
		for x := 0; x < ts.tileSize; x++ {
			noise := (x*3 + y*7) % 25
			img.Set(x, y, color.RGBA{uint8(100 + noise), uint8(60 + noise/3), uint8(30 + noise/5), 255})
		}
	}
	for y := 8; y < 18; y++ {
		for x := 12; x < 28; x++ {
			img.Set(x, y, color.RGBA{100, 100, 110, 255})
		}
	}
	for y := 18; y < 35; y++ {
		for x := 17; x < 23; x++ {
			img.Set(x, y, color.RGBA{120, 80, 50, 255})
		}
	}
	return &Sprite{image: img, width: ts.tileSize, height: ts.tileSize}
}

// getBlockColor возвращает цвет блока (для инвентаря)
func getBlockColor(block BlockType) color.RGBA {
	switch block {
	case Dirt:
		return color.RGBA{139, 69, 19, 255}
	case Grass:
		return color.RGBA{34, 139, 34, 255}
	case Stone:
		return color.RGBA{128, 128, 128, 255}
	case Wood:
		return color.RGBA{101, 67, 33, 255}
	case Leaves:
		return color.RGBA{34, 100, 34, 255}
	case Bricks:
		return color.RGBA{178, 34, 34, 255}
	default:
		return color.RGBA{255, 0, 255, 255}
	}
}
