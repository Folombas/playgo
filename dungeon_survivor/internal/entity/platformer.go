// Package entity содержит игровые сущности для платформера
package entity

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TerrainSprite хранит загруженный спрайт ландшафта
var TerrainSprite *ebiten.Image

// ObjectsSprite хранит загруженный спрайт объектов
var ObjectsSprite *ebiten.Image

// PlatformerPlayer представляет игрока в платформере
type PlatformerPlayer struct {
	X         float64
	Y         float64
	Width     float64
	Height    float64
	VX        float64
	VY        float64
	Speed     float64
	JumpForce float64
	OnGround  bool
	Facing    int // 1 = вправо, -1 = влево
	AnimFrame float64
}

// NewPlatformerPlayer создаёт нового игрока
func NewPlatformerPlayer(x, y float64) *PlatformerPlayer {
	return &PlatformerPlayer{
		X:         x,
		Y:         y,
		Width:     32,
		Height:    40,
		VX:        0,
		VY:        0,
		Speed:     5.0,
		JumpForce: -12.0,
		OnGround:  false,
		Facing:    1,
		AnimFrame: 0,
	}
}

// Update обновляет игрока
func (p *PlatformerPlayer) Update() {
	p.AnimFrame += 0.15
}

// MoveLeft двигает влево
func (p *PlatformerPlayer) MoveLeft() {
	p.VX = -p.Speed
	p.Facing = -1
}

// MoveRight двигает вправо
func (p *PlatformerPlayer) MoveRight() {
	p.VX = p.Speed
	p.Facing = 1
}

// Jump прыгает
func (p *PlatformerPlayer) Jump() {
	p.VY = p.JumpForce
	p.OnGround = false
}

// CanJump проверяет, можно ли прыгать
func (p *PlatformerPlayer) CanJump() bool {
	return p.OnGround
}

// Draw отрисовывает игрока
func (p *PlatformerPlayer) Draw(screen *ebiten.Image, cameraX float64) {
	screenX := p.X - cameraX

	// Тело (зелёная рубашка, синие штаны)
	vector.DrawFilledRect(screen, float32(screenX), float32(p.Y), float32(p.Width), float32(p.Height/2), color.RGBA{50, 200, 50, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX), float32(p.Y+p.Height/2), float32(p.Width), float32(p.Height/2), color.RGBA{50, 50, 200, 255}, true)

	// Голова
	vector.DrawFilledCircle(screen, float32(screenX+p.Width/2), float32(p.Y-10), 12, color.RGBA{255, 200, 150, 255}, true)

	// Глаза
	eyeOffset := float64(p.Facing) * 3.0
	vector.DrawFilledCircle(screen, float32(screenX+p.Width/2+eyeOffset-3), float32(p.Y-12), 3, color.Black, true)

	// Ноги (анимация бега)
	if math.Abs(p.VX) > 0.5 && p.OnGround {
		legOffset := math.Sin(p.AnimFrame*2) * 8
		vector.DrawFilledRect(screen, float32(screenX+8), float32(p.Y+p.Height+legOffset), 8, 8, color.RGBA{100, 100, 100, 255}, true)
		vector.DrawFilledRect(screen, float32(screenX+16), float32(p.Y+p.Height-legOffset), 8, 8, color.RGBA{100, 100, 100, 255}, true)
	} else {
		vector.DrawFilledRect(screen, float32(screenX+8), float32(p.Y+p.Height), 8, 8, color.RGBA{100, 100, 100, 255}, true)
		vector.DrawFilledRect(screen, float32(screenX+16), float32(p.Y+p.Height), 8, 8, color.RGBA{100, 100, 100, 255}, true)
	}

	// Руки
	armOffset := math.Sin(p.AnimFrame) * 5.0 * float64(p.Facing)
	vector.DrawFilledRect(screen, float32(screenX+8+armOffset), float32(p.Y+10), 8, 8, color.RGBA{255, 200, 150, 255}, true)
}

// Platform представляет платформу
type Platform struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Type   string // "grass", "dirt", "stone", "wood"
}

// Coin представляет монетку
type Coin struct {
	X        float64
	Y        float64
	Size     float64
	AnimFrame float64
}

// Flag представляет флаг победы
type Flag struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// House представляет домик
type House struct {
	X         float64
	Y         float64
	Width     float64
	Height    float64
	RoofType  string // "gable" - двускатная
	HasChimney bool
	Color     color.Color
}

// Tree представляет дерево
type Tree struct {
	X     float64
	Y     float64
	Type  string // "oak", "pine", "birch"
	Size  float64
	Color color.Color
}

// Hill представляет холм
type Hill struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Color  color.Color
}

// Cloud представляет облако
type Cloud struct {
	X     float64
	Y     float64
	Size  float64
	Speed float64
}

// PlatformerWorld представляет мир платформера
type PlatformerWorld struct {
	Width      float64
	Height     float64
	Platforms  []Platform
	Coins      []Coin
	Flag       Flag
	Houses     []House
	Trees      []Tree
	Hills      []Hill
	Clouds     []Cloud
	GroundY    float64
}

// NewPlatformerWorld создаёт новый мир
func NewPlatformerWorld(screenWidth, screenHeight int) *PlatformerWorld {
	return &PlatformerWorld{
		Width:     float64(screenWidth * 5), // Уровень в 5 раз шире экрана
		Height:    float64(screenHeight),
		Platforms: make([]Platform, 0),
		Coins:     make([]Coin, 0),
		Houses:    make([]House, 0),
		Trees:     make([]Tree, 0),
		Hills:     make([]Hill, 0),
		Clouds:    make([]Cloud, 0),
		GroundY:   float64(screenHeight) - 50,
	}
}

// GenerateLevel генерирует уровень
func (w *PlatformerWorld) GenerateLevel(level int) {
	w.Platforms = make([]Platform, 0)
	w.Coins = make([]Coin, 0)
	w.Houses = make([]House, 0)
	w.Trees = make([]Tree, 0)
	w.Hills = make([]Hill, 0)
	w.Clouds = make([]Cloud, 0)

	rng := rand.New(rand.NewSource(int64(level * 1000)))

	// Добавляем холмы
	for i := 0; i < 5; i++ {
		x := float64(i*400 + 100 + rng.Intn(200))
		w.Hills = append(w.Hills, Hill{
			X:      x,
			Y:      w.GroundY,
			Width:  200 + rng.Float64()*100,
			Height: 50 + rng.Float64()*50,
			Color:  color.RGBA{100, 180, 100, 255},
		})
	}

	// Земля (базовая платформа)
	w.Platforms = append(w.Platforms, Platform{
		X:      0,
		Y:      w.GroundY,
		Width:  w.Width,
		Height: 50,
		Type:   "grass",
	})

	// Платформы
	for i := 0; i < 15+level*3; i++ {
		x := 200 + rng.Float64()*(w.Width-400)
		y := w.GroundY - 80 - rng.Float64()*200
		width := 80 + rng.Float64()*120

		w.Platforms = append(w.Platforms, Platform{
			X:      x,
			Y:      y,
			Width:  width,
			Height: 20,
			Type:   "grass",
		})

		// Монетки на платформах
		if rng.Float64() < 0.7 {
			w.Coins = append(w.Coins, Coin{
				X:         x + width/2 - 10,
				Y:         y - 30,
				Size:      20,
				AnimFrame: rng.Float64() * 10,
			})
		}
	}

	// Домики
	for i := 0; i < 3+level; i++ {
		x := 300 + float64(i)*500 + rng.Float64()*200
		w.Houses = append(w.Houses, House{
			X:          x,
			Y:          w.GroundY - 120,
			Width:      100,
			Height:     120,
			RoofType:   "gable",
			HasChimney: true,
			Color:      color.RGBA{200, 150, 100, 255},
		})
	}

	// Деревья
	for i := 0; i < 10+level*2; i++ {
		x := 100 + rng.Float64()*(w.Width-200)
		treeType := "oak"
		treeColor := color.RGBA{50, 150, 50, 255}

		r := rng.Float64()
		if r < 0.3 {
			treeType = "pine"
			treeColor = color.RGBA{30, 100, 30, 255}
		} else if r < 0.5 {
			treeType = "birch"
			treeColor = color.RGBA{100, 180, 100, 255}
		}

		w.Trees = append(w.Trees, Tree{
			X:     x,
			Y:     w.GroundY,
			Type:  treeType,
			Size:  60 + rng.Float64()*40,
			Color: treeColor,
		})
	}

	// Облака - больше и разнообразнее
	for i := 0; i < 15; i++ {
		w.Clouds = append(w.Clouds, Cloud{
			X:     rng.Float64() * w.Width,
			Y:     30 + rng.Float64()*150,
			Size:  50 + rng.Float64()*60,
			Speed: 0.05 + rng.Float64()*0.15,
		})
	}

	// Флаг победы (в конце уровня)
	w.Flag = Flag{
		X:      w.Width - 150,
		Y:      w.GroundY - 150,
		Width:  10,
		Height: 150,
	}
}

// Draw отрисовывает мир
func (w *PlatformerWorld) Draw(screen *ebiten.Image, cameraX float64) {
	// Яркое синее небо с градиентом
	for iy := 0; iy < int(w.Height); iy++ {
		ratio := float64(iy) / w.Height
		// От ярко-синего вверху к более светлому внизу
		r := uint8(float64(135) * (1 - ratio*0.3))
		g := uint8(float64(206) * (1 - ratio*0.2))
		b := uint8(235)
		vector.DrawFilledRect(screen, 0, float32(iy), float32(w.Width), 1, color.RGBA{r, g, b, 255}, true)
	}

	// Солнце с сиянием
	w.drawSun(screen, 80, 60)

	// Облака (передний план, параллакс)
	for i := range w.Clouds {
		cloud := &w.Clouds[i]
		cloud.X += cloud.Speed
		if cloud.X > w.Width+200 {
			cloud.X = -200
			cloud.Y = 30 + rand.Float64()*150
		}
		screenX := cloud.X - cameraX*0.3 // Лёгкий параллакс
		if screenX > -200 && screenX < w.Width+200 {
			w.drawCloud(screen, screenX, cloud.Y, cloud.Size)
		}
	}

	// Холмы (задний план)
	for _, hill := range w.Hills {
		screenX := hill.X - cameraX*0.5 // Параллакс
		w.drawHill(screen, screenX, hill)
	}

	// Деревья (задний план)
	for _, tree := range w.Trees {
		screenX := tree.X - cameraX*0.8
		if screenX > -100 && screenX < 1380 {
			w.drawTree(screen, screenX, tree)
		}
	}

	// Домики
	for _, house := range w.Houses {
		screenX := house.X - cameraX
		if screenX > -150 && screenX < 1380 {
			w.drawHouse(screen, screenX, house)
		}
	}

	// Платформы
	for _, platform := range w.Platforms {
		screenX := platform.X - cameraX
		if screenX > -200 && screenX < 1380 {
			w.drawPlatform(screen, screenX, platform)
		}
	}

	// Земля внизу экрана
	w.drawGround(screen, cameraX)

	// Монетки
	for i := range w.Coins {
		coin := &w.Coins[i]
		coin.AnimFrame += 0.1
		screenX := coin.X - cameraX
		w.drawCoin(screen, screenX, coin)
	}

	// Флаг
	w.drawFlag(screen, w.Flag.X-cameraX, w.Flag)
}

// drawCloud рисует облако
func (w *PlatformerWorld) drawCloud(screen *ebiten.Image, x, y, size float64) {
	// Пушистое облако из нескольких кругов
	// Основной слой
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(size), color.RGBA{255, 255, 255, 240}, true)
	vector.DrawFilledCircle(screen, float32(x+size*0.6), float32(y+size*0.1), float32(size*0.8), color.RGBA{255, 255, 255, 240}, true)
	vector.DrawFilledCircle(screen, float32(x-size*0.6), float32(y+size*0.1), float32(size*0.8), color.RGBA{255, 255, 255, 240}, true)
	vector.DrawFilledCircle(screen, float32(x), float32(y+size*0.4), float32(size*0.9), color.RGBA{255, 255, 255, 240}, true)
	vector.DrawFilledCircle(screen, float32(x+size*0.3), float32(y+size*0.3), float32(size*0.7), color.RGBA{255, 255, 255, 240}, true)
	vector.DrawFilledCircle(screen, float32(x-size*0.3), float32(y+size*0.3), float32(size*0.7), color.RGBA{255, 255, 255, 240}, true)
	
	// Блик (светлая середина)
	vector.DrawFilledCircle(screen, float32(x), float32(y-size*0.1), float32(size*0.5), color.RGBA{255, 255, 255, 255}, true)
}

// drawSun рисует солнце с сиянием
func (w *PlatformerWorld) drawSun(screen *ebiten.Image, x, y float64) {
	// Сияние (большой полупрозрачный круг)
	vector.DrawFilledCircle(screen, float32(x), float32(y), 60, color.RGBA{255, 255, 200, 100}, true)
	vector.DrawFilledCircle(screen, float32(x), float32(y), 50, color.RGBA{255, 255, 150, 150}, true)
	
	// Основное солнце
	vector.DrawFilledCircle(screen, float32(x), float32(y), 40, color.RGBA{255, 255, 100, 255}, true)
	vector.DrawFilledCircle(screen, float32(x), float32(y), 30, color.RGBA{255, 255, 150, 255}, true)
	
	// Лучи солнца (8 направлений)
	for i := 0; i < 8; i++ {
		angle := float64(i)*math.Pi/4 + math.Pi/8
		innerRadius := 45.0
		outerRadius := 65.0
		x1 := x + math.Cos(angle)*innerRadius
		y1 := y + math.Sin(angle)*innerRadius
		x2 := x + math.Cos(angle)*outerRadius
		y2 := y + math.Sin(angle)*outerRadius
		
		// Рисуем луч как линию
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 4, color.RGBA{255, 255, 100, 200}, false)
	}
	
	// Яркий центр
	vector.DrawFilledCircle(screen, float32(x), float32(y), 15, color.RGBA{255, 255, 255, 255}, true)
}

// drawHill рисует холм
func (w *PlatformerWorld) drawHill(screen *ebiten.Image, x float64, hill Hill) {
	// Рисуем эллипс как холм
	for dy := 0.0; dy < hill.Height; dy += 1 {
		ratio := dy / hill.Height
		width := hill.Width * math.Sqrt(1-ratio*ratio)
		vector.DrawFilledRect(screen, float32(x-width/2), float32(hill.Y-dy), float32(width), 1, hill.Color, true)
	}
}

// drawTree рисует дерево
func (w *PlatformerWorld) drawTree(screen *ebiten.Image, x float64, tree Tree) {
	y := tree.Y

	if tree.Type == "pine" {
		// Ёлка (треугольная)
		trunkWidth := tree.Size * 0.15
		vector.DrawFilledRect(screen, float32(x-trunkWidth/2), float32(y-tree.Size*0.3), float32(trunkWidth), float32(tree.Size*0.3), color.RGBA{100, 60, 30, 255}, true)

		// Кроны (три треугольника)
		for i := 0; i < 3; i++ {
			size := tree.Size * (0.5 - float64(i)*0.1)
			yOffset := float64(i) * tree.Size * 0.2
			vector.StrokeLine(screen, float32(x), float32(y-tree.Size*0.8+yOffset), float32(x-size), float32(y-tree.Size*0.3+yOffset), 3, tree.Color, false)
			vector.StrokeLine(screen, float32(x), float32(y-tree.Size*0.8+yOffset), float32(x+size), float32(y-tree.Size*0.3+yOffset), 3, tree.Color, false)
			vector.StrokeLine(screen, float32(x-size), float32(y-tree.Size*0.3+yOffset), float32(x+size), float32(y-tree.Size*0.3+yOffset), 3, tree.Color, false)
		}
	} else if tree.Type == "birch" {
		// Берёза (светлый ствол)
		trunkWidth := tree.Size * 0.12
		vector.DrawFilledRect(screen, float32(x-trunkWidth/2), float32(y-tree.Size*0.4), float32(trunkWidth), float32(tree.Size*0.4), color.RGBA{240, 240, 220, 255}, true)
		// Чёрные полоски на стволе
		for i := 0; i < 4; i++ {
			vector.StrokeLine(screen, float32(x-trunkWidth/2), float32(y-tree.Size*0.35+float64(i)*tree.Size*0.08), float32(x), float32(y-tree.Size*0.35+float64(i)*tree.Size*0.08), 1, color.Black, false)
		}

		// Крона (овальная)
		vector.DrawFilledCircle(screen, float32(x), float32(y-tree.Size*0.5), float32(tree.Size*0.35), tree.Color, true)
	} else {
		// Дуб (круглая крона)
		trunkWidth := tree.Size * 0.15
		vector.DrawFilledRect(screen, float32(x-trunkWidth/2), float32(y-tree.Size*0.4), float32(trunkWidth), float32(tree.Size*0.4), color.RGBA{100, 60, 30, 255}, true)

		// Крона (несколько кругов)
		vector.DrawFilledCircle(screen, float32(x), float32(y-tree.Size*0.5), float32(tree.Size*0.4), tree.Color, true)
		vector.DrawFilledCircle(screen, float32(x-tree.Size*0.25), float32(y-tree.Size*0.4), float32(tree.Size*0.25), tree.Color, true)
		vector.DrawFilledCircle(screen, float32(x+tree.Size*0.25), float32(y-tree.Size*0.4), float32(tree.Size*0.25), tree.Color, true)
		vector.DrawFilledCircle(screen, float32(x), float32(y-tree.Size*0.3), float32(tree.Size*0.3), tree.Color, true)
	}
}

// drawHouse рисует домик
func (w *PlatformerWorld) drawHouse(screen *ebiten.Image, x float64, house House) {
	y := house.Y

	// Стены
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(house.Width), float32(house.Height), house.Color, true)

	// Двускатная крыша
	roofHeight := house.Height * 0.6
	vector.StrokeLine(screen, float32(x-10), float32(y), float32(x+house.Width/2), float32(y-roofHeight), 8, color.RGBA{150, 50, 50, 255}, false)
	vector.StrokeLine(screen, float32(x+house.Width/2), float32(y-roofHeight), float32(x+house.Width+10), float32(y), 8, color.RGBA{150, 50, 50, 255}, false)

	// Заполнение крыши
	for i := 0; i < int(roofHeight); i++ {
		ratio := float64(i) / roofHeight
		lineWidth := float32(float64(house.Width)+20) * float32(1-ratio)
		vector.DrawFilledRect(screen, float32(x)+float32(house.Width)/2-lineWidth/2, float32(y-float64(i)), lineWidth, 1, color.RGBA{150, 50, 50, 255}, true)
	}

	// Труба
	if house.HasChimney {
		chimneyX := x + house.Width*0.7
		vector.DrawFilledRect(screen, float32(chimneyX), float32(y-roofHeight*0.5), 20, float32(roofHeight*0.8), color.RGBA{150, 50, 50, 255}, true)

		// Дым из трубы
		for i := 0; i < 5; i++ {
			smokeY := y - roofHeight*0.5 - float64(i)*15
			smokeX := chimneyX + 10 + math.Sin(float64(i)*0.5)*5
			smokeSize := 8 + float64(i)*3
			alpha := uint8(150 - i*25)
			vector.DrawFilledCircle(screen, float32(smokeX), float32(smokeY), float32(smokeSize), color.RGBA{200, 200, 200, alpha}, true)
		}
	}

	// Дверь
	doorWidth := house.Width * 0.2
	vector.DrawFilledRect(screen, float32(x+house.Width/2-doorWidth/2), float32(y+house.Height*0.4), float32(doorWidth), float32(house.Height*0.6), color.RGBA{100, 60, 30, 255}, true)

	// Окна
	windowSize := house.Width * 0.15
	vector.DrawFilledRect(screen, float32(x+house.Width*0.2), float32(y+house.Height*0.2), float32(windowSize), float32(windowSize), color.RGBA{150, 200, 255, 255}, true)
	vector.DrawFilledRect(screen, float32(x+house.Width*0.65), float32(y+house.Height*0.2), float32(windowSize), float32(windowSize), color.RGBA{150, 200, 255, 255}, true)

	// Рама окон
	vector.StrokeLine(screen, float32(x+house.Width*0.2), float32(y+house.Height*0.2+windowSize/2), float32(x+house.Width*0.2+windowSize), float32(y+house.Height*0.2+windowSize/2), 1, color.RGBA{100, 100, 100, 255}, false)
	vector.StrokeLine(screen, float32(x+house.Width*0.2+windowSize/2), float32(y+house.Height*0.2), float32(x+house.Width*0.2+windowSize/2), float32(y+house.Height*0.2+windowSize), 1, color.RGBA{100, 100, 100, 255}, false)
}

// drawPlatform рисует платформу со спрайтом
func (w *PlatformerWorld) drawPlatform(screen *ebiten.Image, x float64, platform Platform) {
	tileSize := 48.0
	
	// Рисуем плитками
	for tx := 0; tx < int(platform.Width/tileSize)+1; tx++ {
		screenX := x + float64(tx)*tileSize
		
		if TerrainSprite != nil {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(screenX, platform.Y)
			// Вырезаем плитку земли с травой (верхняя левая на спрайте)
			tile := TerrainSprite.SubImage(image.Rect(0, 0, 48, 48)).(*ebiten.Image)
			screen.DrawImage(tile, opts)
		} else {
			// Резерв - векторная графика
			vector.DrawFilledRect(screen, float32(screenX), float32(platform.Y), float32(tileSize), 8, color.RGBA{100, 200, 100, 255}, true)
			vector.DrawFilledRect(screen, float32(screenX), float32(platform.Y+8), float32(tileSize), float32(tileSize-8), color.RGBA{140, 100, 60, 255}, true)
		}
	}
}

// drawGround рисует землю внизу экрана
func (w *PlatformerWorld) drawGround(screen *ebiten.Image, cameraX float64) {
	groundY := w.GroundY
	startX := cameraX
	endX := cameraX + 1280
	
	// Рисуем землю плитками
	for x := startX - 50; x < endX; x += 48 {
		screenX := x - cameraX
		if TerrainSprite != nil {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(screenX, groundY)
			// Плитка с травой
			tile := TerrainSprite.SubImage(image.Rect(0, 0, 48, 48)).(*ebiten.Image)
			screen.DrawImage(tile, opts)
		} else {
			vector.DrawFilledRect(screen, float32(screenX), float32(groundY), 48, 8, color.RGBA{100, 200, 100, 255}, true)
			vector.DrawFilledRect(screen, float32(screenX), float32(groundY+8), 48, 42, color.RGBA{140, 100, 60, 255}, true)
		}
	}
}

// drawCoin рисует монетку
func (w *PlatformerWorld) drawCoin(screen *ebiten.Image, x float64, coin *Coin) {
	// Анимация вращения
	scale := math.Abs(math.Sin(coin.AnimFrame))

	vector.DrawFilledCircle(screen, float32(x+10), float32(coin.Y+10), float32(10*scale), color.RGBA{255, 215, 0, 255}, true)
	vector.DrawFilledCircle(screen, float32(x+10), float32(coin.Y+10), float32(6*scale), color.RGBA{255, 255, 150, 255}, true)
}

// drawFlag рисует флаг
func (w *PlatformerWorld) drawFlag(screen *ebiten.Image, x float64, flag Flag) {
	// Флагшток
	vector.StrokeLine(screen, float32(x), float32(flag.Y), float32(x), float32(flag.Y+flag.Height), 3, color.RGBA{100, 100, 100, 255}, false)

	// Флаг
	vector.StrokeLine(screen, float32(x), float32(flag.Y+10), float32(x+50), float32(flag.Y+35), 3, color.RGBA{255, 50, 50, 255}, false)
	vector.StrokeLine(screen, float32(x), float32(flag.Y+10), float32(x+50), float32(flag.Y+60), 3, color.RGBA{255, 50, 50, 255}, false)
	vector.StrokeLine(screen, float32(x+50), float32(flag.Y+35), float32(x+50), float32(flag.Y+60), 3, color.RGBA{255, 50, 50, 255}, false)

	// Заполнение флага
	for i := 0; i < 25; i++ {
		width := float32(50) * float32(i) / 25
		vector.DrawFilledRect(screen, float32(x)+width, float32(flag.Y+35+float64(i)), 1, 1, color.RGBA{255, 50, 50, 255}, true)
	}

	// Шар на флагштоке
	vector.DrawFilledCircle(screen, float32(x), float32(flag.Y-5), 8, color.RGBA{200, 200, 50, 255}, true)
}
