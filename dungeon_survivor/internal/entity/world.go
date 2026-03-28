// Package entity содержит игровые сущности
package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// World представляет игровой мир
type World struct {
	Width      int
	Height     int
	TileSize   int
	Tiles      [][]Tile
	Objects    [][]WorldObject
	Buildings  []Building
	Trees      []Tree
	Flowers    []Flower
	Fences     []Fence
	Paths      []Path
	Houses     []House
	Castle     *Castle
}

// TileType определяет тип тайла
type TileType int

const (
	TileGrass TileType = iota
	TileDirt
	TileStone
	TileWater
	TileSand
	TileSnow
)

// Tile представляет тайл мира
type Tile struct {
	Type TileType
	X    int
	Y    int
}

// WorldObject представляет объект в мире
type WorldObject struct {
	Type   string
	X      float64
	Y      float64
	Width  float64
	Height float64
	Color  color.Color
}

// Tree представляет дерево
type Tree struct {
	X     float64
	Y     float64
	Type  string // "oak", "pine", "cherry"
	Size  float64
	Color color.Color
}

// Flower представляет цветок
type Flower struct {
	X     float64
	Y     float64
	Color string // "pink", "yellow", "blue", "red", "purple"
	Size  float64
}

// Building представляет постройку
type Building struct {
	X      float64
	Y      float64
	Type   string
	Width  float64
	Height float64
	Color  color.Color
}

// Fence представляет забор
type Fence struct {
	X     float64
	Y     float64
	Width float64
}

// Path представляет дорожку
type Path struct {
	X     float64
	Y     float64
	Size  float64
	Color color.Color
}

// House представляет домик
type House struct {
	X         float64
	Y         float64
	Width     float64
	Height    float64
	RoofColor color.Color
	WallColor color.Color
}

// Castle представляет замок
type Castle struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NewWorld создаёт новый мир
func NewWorld(config *Config) *World {
	width := config.ScreenWidth / config.TileSize * 2
	height := config.ScreenHeight / config.TileSize * 2

	w := &World{
		Width:    width,
		Height:   height,
		TileSize: config.TileSize,
		Tiles:    make([][]Tile, width),
		Objects:  make([][]WorldObject, width),
		Trees:    make([]Tree, 0),
		Flowers:  make([]Flower, 0),
		Fences:   make([]Fence, 0),
		Paths:    make([]Path, 0),
		Houses:   make([]House, 0),
	}

	// Инициализация тайлов
	for x := 0; x < width; x++ {
		w.Tiles[x] = make([]Tile, height)
		w.Objects[x] = make([]WorldObject, height)
		for y := 0; y < height; y++ {
			// Генерация ландшафта
			tileType := TileGrass

			// Вода по краям
			if x < 2 || x > width-3 || y < 2 || y > height-3 {
				tileType = TileWater
			}

			// Случайные островки
			if rand.Float64() < 0.1 {
				tileType = TileGrass
			}

			w.Tiles[x][y] = Tile{Type: tileType, X: x, Y: y}
		}
	}

	return w
}

// PlaceTree размещает дерево
func (w *World) PlaceTree(x, y float64) {
	types := []string{"oak", "pine", "cherry"}
	colors := []color.Color{
		color.RGBA{50, 150, 50, 255},
		color.RGBA{30, 100, 30, 255},
		color.RGBA{255, 150, 180, 255},
	}
	idx := int(x+y) % len(types)

	w.Trees = append(w.Trees, Tree{
		X:     x,
		Y:     y,
		Type:  types[idx],
		Size:  40 + rand.Float64()*20,
		Color: colors[idx],
	})
}

// PlaceFlower размещает цветок
func (w *World) PlaceFlower(x, y float64, flowerColor string) {
	w.Flowers = append(w.Flowers, Flower{
		X:     x,
		Y:     y,
		Color: flowerColor,
		Size:  15 + rand.Float64()*10,
	})
}

// PlaceBush размещает куст
func (w *World) PlaceBush(x, y float64) {
	w.Trees = append(w.Trees, Tree{
		X:     x,
		Y:     y,
		Type:  "bush",
		Size:  30 + rand.Float64()*15,
		Color: color.RGBA{80, 180, 80, 255},
	})
}

// PlaceFence размещает забор
func (w *World) PlaceFence(x, y float64) {
	w.Fences = append(w.Fences, Fence{
		X:     x,
		Y:     y,
		Width: 48,
	})
}

// PlacePath размещает дорожку
func (w *World) PlacePath(x, y float64) {
	w.Paths = append(w.Paths, Path{
		X:     x,
		Y:     y,
		Size:  48,
		Color: color.RGBA{180, 160, 140, 255},
	})
}

// PlaceHouse размещает домик
func (w *World) PlaceHouse(x, y float64) {
	w.Houses = append(w.Houses, House{
		X:         x,
		Y:         y,
		Width:     100,
		Height:    80,
		RoofColor: color.RGBA{180, 50, 50, 255},
		WallColor: color.RGBA{240, 220, 180, 255},
	})
}

// PlaceCastle размещает замок
func (w *World) PlaceCastle(x, y float64) {
	w.Castle = &Castle{
		X:      x,
		Y:      y,
		Width:  200,
		Height: 180,
	}
}

// Draw отрисовывает мир
func (w *World) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	// Отрисовка тайлов
	startX := int(cameraX / float64(w.TileSize))
	startY := int(cameraY / float64(w.TileSize))
	endX := startX + (1280/w.TileSize + 2)
	endY := startY + (720/w.TileSize + 2)

	for x := startX; x < endX && x < w.Width; x++ {
		for y := startY; y < endY && y < w.Height; y++ {
			tile := w.Tiles[x][y]
			screenX := float32(x*w.TileSize) - float32(cameraX)
			screenY := float32(y*w.TileSize) - float32(cameraY)

			var tileColor color.Color
			switch tile.Type {
			case TileGrass:
				tileColor = color.RGBA{100, 200, 100, 255}
			case TileDirt:
				tileColor = color.RGBA{180, 140, 100, 255}
			case TileWater:
				tileColor = color.RGBA{80, 150, 200, 255}
			case TileSand:
				tileColor = color.RGBA{220, 200, 150, 255}
			default:
				tileColor = color.RGBA{100, 200, 100, 255}
			}

			vector.DrawFilledRect(screen, screenX, screenY, float32(w.TileSize), float32(w.TileSize), tileColor, true)
		}
	}

	// Отрисовка дорожек
	for _, path := range w.Paths {
		screenX := path.X - cameraX
		screenY := path.Y - cameraY
		vector.DrawFilledRect(screen, float32(screenX-path.Size/2), float32(screenY-path.Size/2), float32(path.Size), float32(path.Size), path.Color, true)
	}

	// Отрисовка цветов
	for _, flower := range w.Flowers {
		screenX := flower.X - cameraX
		screenY := flower.Y - cameraY
		w.drawFlower(screen, screenX, screenY, flower)
	}

	// Отрисовка деревьев
	for _, tree := range w.Trees {
		screenX := tree.X - cameraX
		screenY := tree.Y - cameraY
		w.drawTree(screen, screenX, screenY, tree)
	}

	// Отрисовка заборов
	for _, fence := range w.Fences {
		screenX := fence.X - cameraX
		screenY := fence.Y - cameraY
		vector.DrawFilledRect(screen, float32(screenX-24), float32(screenY-10), float32(fence.Width), 20, color.RGBA{140, 100, 60, 255}, true)
	}

	// Отрисовка домов
	for _, house := range w.Houses {
		screenX := house.X - cameraX
		screenY := house.Y - cameraY
		w.drawHouse(screen, screenX, screenY, house)
	}

	// Отрисовка замка
	if w.Castle != nil {
		screenX := w.Castle.X - cameraX
		screenY := w.Castle.Y - cameraY
		w.drawCastle(screen, screenX, screenY)
	}
}

// drawFlower отрисовывает цветок
func (w *World) drawFlower(screen *ebiten.Image, x, y float64, flower Flower) {
	var petalColor color.Color
	switch flower.Color {
	case "pink":
		petalColor = color.RGBA{255, 150, 200, 255}
	case "yellow":
		petalColor = color.RGBA{255, 220, 50, 255}
	case "blue":
		petalColor = color.RGBA{100, 150, 255, 255}
	case "red":
		petalColor = color.RGBA{255, 80, 80, 255}
	case "purple":
		petalColor = color.RGBA{180, 100, 220, 255}
	default:
		petalColor = color.RGBA{255, 150, 200, 255}
	}

	// Стебель
	vector.StrokeLine(screen, float32(x), float32(y), float32(x), float32(y+flower.Size), 2, color.RGBA{50, 150, 50, 255}, false)

	// Лепестки
	for i := 0; i < 5; i++ {
		angle := float64(i) * math.Pi * 2 / 5
		px := x + math.Cos(angle)*flower.Size*0.4
		py := y + math.Sin(angle)*flower.Size*0.4
		vector.DrawFilledCircle(screen, float32(px), float32(py), float32(flower.Size*0.25), petalColor, true)
	}

	// Центр
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(flower.Size*0.2), color.RGBA{255, 255, 100, 255}, true)
}

// drawTree отрисовывает дерево
func (w *World) drawTree(screen *ebiten.Image, x, y float64, tree Tree) {
	// Ствол
	trunkWidth := tree.Size * 0.15
	vector.DrawFilledRect(screen, float32(x-trunkWidth/2), float32(y), float32(trunkWidth), float32(tree.Size*0.4), color.RGBA{100, 60, 30, 255}, true)

	// Крона
	if tree.Type == "cherry" {
		// Сакура - розовая
		vector.DrawFilledCircle(screen, float32(x), float32(y-tree.Size*0.3), float32(tree.Size*0.5), color.RGBA{255, 180, 200, 255}, true)
		vector.DrawFilledCircle(screen, float32(x-tree.Size*0.2), float32(y-tree.Size*0.2), float32(tree.Size*0.35), color.RGBA{255, 150, 180, 255}, true)
		vector.DrawFilledCircle(screen, float32(x+tree.Size*0.2), float32(y-tree.Size*0.2), float32(tree.Size*0.35), color.RGBA{255, 150, 180, 255}, true)
	} else if tree.Type == "pine" {
		// Сосна - треугольная
		vector.StrokeLine(screen, float32(x), float32(y-tree.Size*0.8), float32(x-tree.Size*0.3), float32(y), 3, tree.Color, false)
		vector.StrokeLine(screen, float32(x), float32(y-tree.Size*0.8), float32(x+tree.Size*0.3), float32(y), 3, tree.Color, false)
		vector.StrokeLine(screen, float32(x-tree.Size*0.3), float32(y), float32(x+tree.Size*0.3), float32(y), 3, tree.Color, false)
	} else {
		// Дуб - круглая крона
		vector.DrawFilledCircle(screen, float32(x), float32(y-tree.Size*0.3), float32(tree.Size*0.5), tree.Color, true)
		vector.DrawFilledCircle(screen, float32(x-tree.Size*0.25), float32(y-tree.Size*0.15), float32(tree.Size*0.3), tree.Color, true)
		vector.DrawFilledCircle(screen, float32(x+tree.Size*0.25), float32(y-tree.Size*0.15), float32(tree.Size*0.3), tree.Color, true)
	}
}

// drawHouse отрисовывает домик
func (w *World) drawHouse(screen *ebiten.Image, x, y float64, house House) {
	// Стены
	vector.DrawFilledRect(screen, float32(x-house.Width/2), float32(y-house.Height/2), float32(house.Width), float32(house.Height*0.7), house.WallColor, true)

	// Крыша
	vector.StrokeLine(screen, float32(x-house.Width/2-10), float32(y-house.Height/2), float32(x), float32(y-house.Height*0.9), 3, house.RoofColor, false)
	vector.StrokeLine(screen, float32(x), float32(y-house.Height*0.9), float32(x+house.Width/2+10), float32(y-house.Height/2), 3, house.RoofColor, false)
	vector.StrokeLine(screen, float32(x-house.Width/2-10), float32(y-house.Height/2), float32(x+house.Width/2+10), float32(y-house.Height/2), 3, house.RoofColor, false)

	// Дверь
	vector.DrawFilledRect(screen, float32(x-15), float32(y), 30, float32(house.Height*0.35), color.RGBA{100, 60, 30, 255}, true)

	// Окно
	vector.DrawFilledRect(screen, float32(x-40), float32(y-house.Height*0.2), 20, 20, color.RGBA{150, 200, 255, 255}, true)
	vector.DrawFilledRect(screen, float32(x+20), float32(y-house.Height*0.2), 20, 20, color.RGBA{150, 200, 255, 255}, true)
}

// drawCastle отрисовывает замок
func (w *World) drawCastle(screen *ebiten.Image, x, y float64) {
	castleColor := color.RGBA{200, 200, 200, 255}
	roofColor := color.RGBA{180, 50, 50, 255}

	// Основное здание
	vector.DrawFilledRect(screen, float32(x-80), float32(y-60), 160, 100, castleColor, true)

	// Башни
	vector.DrawFilledRect(screen, float32(x-90), float32(y-100), 40, 140, castleColor, true)
	vector.DrawFilledRect(screen, float32(x+50), float32(y-100), 40, 140, castleColor, true)

	// Крыши башен
	vector.StrokeLine(screen, float32(x-90), float32(y-100), float32(x-70), float32(y-130), 3, roofColor, false)
	vector.StrokeLine(screen, float32(x-70), float32(y-130), float32(x-50), float32(y-100), 3, roofColor, false)

	vector.StrokeLine(screen, float32(x+50), float32(y-100), float32(x+70), float32(y-130), 3, roofColor, false)
	vector.StrokeLine(screen, float32(x+70), float32(y-130), float32(x+90), float32(y-100), 3, roofColor, false)

	// Флаг
	vector.StrokeLine(screen, float32(x), float32(y-100), float32(x), float32(y-140), 2, color.RGBA{100, 100, 100, 255}, false)
	vector.StrokeLine(screen, float32(x), float32(y-140), float32(x+25), float32(y-130), 2, roofColor, false)
	vector.StrokeLine(screen, float32(x), float32(y-135), float32(x+25), float32(y-125), 2, roofColor, false)

	// Ворота
	vector.DrawFilledRect(screen, float32(x-25), float32(y), 50, 40, color.RGBA{80, 60, 40, 255}, true)
	// Арка ворот (полуверхность)
	for i := -25; i <= 25; i += 3 {
		h := math.Sqrt(float64(25*25 - i*i))
		vector.DrawFilledRect(screen, float32(x+float64(i)), float32(y-20-float64(h)/2), 3, float32(h/2), color.RGBA{80, 60, 40, 255}, true)
	}

	// Окна
	vector.DrawFilledRect(screen, float32(x-50), float32(y-40), 20, 25, color.RGBA{150, 200, 255, 200}, true)
	vector.DrawFilledRect(screen, float32(x+30), float32(y-40), 20, 25, color.RGBA{150, 200, 255, 200}, true)
}

// BuildingCount возвращает количество построек
func (w *World) BuildingCount() int {
	return len(w.Houses) + len(w.Fences) + len(w.Paths)
}

// TreeCount возвращает количество деревьев
func (w *World) TreeCount() int {
	return len(w.Trees)
}

// FlowerCount возвращает количество цветов
func (w *World) FlowerCount() int {
	return len(w.Flowers)
}
