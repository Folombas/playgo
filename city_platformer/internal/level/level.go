package level

import (
	"image/color"

	"city_platformer/internal/entity"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Platform - платформа
type Platform struct {
	X, Y     float64
	Width    float64
	Height   float64
	Type     string
}

// Item - предмет
type Item struct {
	X, Y     float64
	Width    float64
	Height   float64
	Type     string
	Collected bool
}

// Level - игровой уровень
type Level struct {
	Width     int
	Height    int
	Platforms []*Platform
	Enemies   []*entity.Enemy
	Items     []*Item
	Door      *Door
}

// Door - выход с уровня
type Door struct {
	X, Y   float64
	Width  float64
	Height float64
	Open   bool
}

// LoadLevel загружает уровень по номеру
func LoadLevel(num int) (*Level, error) {
	switch num {
	case 1:
		return createLevel1(), nil
	case 2:
		return createLevel2(), nil
	case 3:
		return createLevel3(), nil
	default:
		return createLevel1(), nil
	}
}

// CreateDefaultLevel создаёт уровень по умолчанию
func CreateDefaultLevel() *Level {
	return createLevel1()
}

func createLevel1() *Level {
	level := &Level{
		Width:  2000,
		Height: 600,
	}

	// Земля
	level.Platforms = append(level.Platforms, &Platform{
		X: 0, Y: 500, Width: 2000, Height: 100, Type: "ground",
	})

	// Платформы
	level.Platforms = append(level.Platforms,
		&Platform{X: 200, Y: 400, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 400, Y: 350, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 600, Y: 300, Width: 150, Height: 20, Type: "platform"},
		&Platform{X: 850, Y: 350, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 1050, Y: 400, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 1300, Y: 350, Width: 150, Height: 20, Type: "platform"},
		&Platform{X: 1550, Y: 300, Width: 100, Height: 20, Type: "platform"},
	)

	// Препятствия (ямы)
	// Удаляем часть земли для создания ям
	level.Platforms = append(level.Platforms,
		&Platform{X: 750, Y: 500, Width: 50, Height: 100, Type: "ground"},
		&Platform{X: 900, Y: 500, Width: 50, Height: 100, Type: "ground"},
	)

	// Враги
	level.Enemies = append(level.Enemies,
		entity.NewEnemy(500, 484, "slime"),
		entity.NewEnemy(900, 484, "snake"),
		entity.NewEnemy(1400, 484, "slime"),
	)

	// Монеты
	for i := 0; i < 20; i++ {
		level.Items = append(level.Items, &Item{
			X: float64(300 + i*80), Y: 450, Width: 16, Height: 16, Type: "coin",
		})
	}

	// Звёзды
	level.Items = append(level.Items,
		&Item{X: 650, Y: 260, Width: 20, Height: 20, Type: "star"},
		&Item{X: 1350, Y: 310, Width: 20, Height: 20, Type: "star"},
		&Item{X: 1600, Y: 260, Width: 20, Height: 20, Type: "star"},
	)

	// Дверь выхода
	level.Door = &Door{
		X: 1900, Y: 420, Width: 60, Height: 80, Open: false,
	}

	return level
}

func createLevel2() *Level {
	level := &Level{
		Width:  2500,
		Height: 600,
	}

	// Земля с разрывами
	level.Platforms = append(level.Platforms,
		&Platform{X: 0, Y: 500, Width: 600, Height: 100, Type: "ground"},
		&Platform{X: 700, Y: 500, Width: 500, Height: 100, Type: "ground"},
		&Platform{X: 1300, Y: 500, Width: 1200, Height: 100, Type: "ground"},
	)

	// Платформы
	level.Platforms = append(level.Platforms,
		&Platform{X: 150, Y: 400, Width: 120, Height: 20, Type: "platform"},
		&Platform{X: 350, Y: 320, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 550, Y: 380, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 750, Y: 400, Width: 120, Height: 20, Type: "platform"},
		&Platform{X: 950, Y: 320, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 1150, Y: 380, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 1400, Y: 400, Width: 150, Height: 20, Type: "platform"},
		&Platform{X: 1650, Y: 350, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 1850, Y: 300, Width: 120, Height: 20, Type: "platform"},
		&Platform{X: 2050, Y: 350, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 2250, Y: 400, Width: 100, Height: 20, Type: "platform"},
	)

	// Враги
	level.Enemies = append(level.Enemies,
		entity.NewEnemy(300, 484, "slime"),
		entity.NewEnemy(800, 484, "snake"),
		entity.NewEnemy(1000, 484, "spider"),
		entity.NewEnemy(1500, 484, "slime"),
		entity.NewEnemy(1800, 484, "snake"),
		entity.NewEnemy(2100, 484, "spider"),
	)

	// Монеты
	for i := 0; i < 30; i++ {
		level.Items = append(level.Items, &Item{
			X: float64(200 + i*70), Y: 450, Width: 16, Height: 16, Type: "coin",
		})
	}

	// Звёзды
	level.Items = append(level.Items,
		&Item{X: 400, Y: 280, Width: 20, Height: 20, Type: "star"},
		&Item{X: 1000, Y: 280, Width: 20, Height: 20, Type: "star"},
		&Item{X: 1450, Y: 360, Width: 20, Height: 20, Type: "star"},
		&Item{X: 1900, Y: 260, Width: 20, Height: 20, Type: "star"},
		&Item{X: 2300, Y: 360, Width: 20, Height: 20, Type: "star"},
	)

	// Дверь
	level.Door = &Door{
		X: 2400, Y: 420, Width: 60, Height: 80, Open: false,
	}

	return level
}

func createLevel3() *Level {
	level := &Level{
		Width:  3000,
		Height: 600,
	}

	// Земля с множественными разрывами
	level.Platforms = append(level.Platforms,
		&Platform{X: 0, Y: 500, Width: 400, Height: 100, Type: "ground"},
		&Platform{X: 500, Y: 500, Width: 300, Height: 100, Type: "ground"},
		&Platform{X: 900, Y: 500, Width: 400, Height: 100, Type: "ground"},
		&Platform{X: 1400, Y: 500, Width: 300, Height: 100, Type: "ground"},
		&Platform{X: 1800, Y: 500, Width: 500, Height: 100, Type: "ground"},
		&Platform{X: 2400, Y: 500, Width: 600, Height: 100, Type: "ground"},
	)

	// Платформы
	level.Platforms = append(level.Platforms,
		&Platform{X: 100, Y: 400, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 300, Y: 320, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 550, Y: 380, Width: 80, Height: 20, Type: "platform"},
		&Platform{X: 750, Y: 300, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 950, Y: 350, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 1150, Y: 280, Width: 120, Height: 20, Type: "platform"},
		&Platform{X: 1450, Y: 350, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 1650, Y: 280, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 1900, Y: 350, Width: 150, Height: 20, Type: "platform"},
		&Platform{X: 2150, Y: 280, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 2350, Y: 350, Width: 100, Height: 20, Type: "platform"},
		&Platform{X: 2550, Y: 400, Width: 120, Height: 20, Type: "platform"},
		&Platform{X: 2750, Y: 350, Width: 100, Height: 20, Type: "platform"},
	)

	// Враги
	level.Enemies = append(level.Enemies,
		entity.NewEnemy(200, 484, "slime"),
		entity.NewEnemy(600, 484, "snake"),
		entity.NewEnemy(1000, 484, "spider"),
		entity.NewEnemy(1200, 484, "slime"),
		entity.NewEnemy(1500, 484, "snake"),
		entity.NewEnemy(2000, 484, "spider"),
		entity.NewEnemy(2200, 484, "slime"),
		entity.NewEnemy(2600, 484, "snake"),
	)

	// Монеты
	for i := 0; i < 40; i++ {
		level.Items = append(level.Items, &Item{
			X: float64(150 + i*60), Y: 450, Width: 16, Height: 16, Type: "coin",
		})
	}

	// Звёзды
	level.Items = append(level.Items,
		&Item{X: 350, Y: 280, Width: 20, Height: 20, Type: "star"},
		&Item{X: 800, Y: 260, Width: 20, Height: 20, Type: "star"},
		&Item{X: 1200, Y: 240, Width: 20, Height: 20, Type: "star"},
		&Item{X: 1700, Y: 240, Width: 20, Height: 20, Type: "star"},
		&Item{X: 2200, Y: 240, Width: 20, Height: 20, Type: "star"},
		&Item{X: 2800, Y: 310, Width: 20, Height: 20, Type: "star"},
	)

	// Дверь
	level.Door = &Door{
		X: 2900, Y: 420, Width: 60, Height: 80, Open: false,
	}

	return level
}

// Update обновляет уровень
func (l *Level) Update() {
	for _, enemy := range l.Enemies {
		enemy.Update()
	}
}

// Draw отрисовывает уровень
func (l *Level) Draw(screen *ebiten.Image, cameraX float64) {
	// Отрисовка платформ
	for _, platform := range l.Platforms {
		l.drawPlatform(screen, platform, cameraX)
	}

	// Отрисовка предметов
	for _, item := range l.Items {
		l.drawItem(screen, item, cameraX)
	}

	// Отрисовка врагов
	for _, enemy := range l.Enemies {
		enemy.Draw(screen, cameraX)
	}

	// Отрисовка двери
	l.drawDoor(screen, l.Door, cameraX)
}

func (l *Level) drawPlatform(screen *ebiten.Image, p *Platform, cameraX float64) {
	// Не рисуем, если за пределами экрана
	if p.X+cameraX < -p.Width || p.X+cameraX > float64(screen.Bounds().Dx()) {
		return
	}

	switch p.Type {
	case "ground":
		// Земля с травой
		vector.DrawFilledRect(
			screen,
			float32(p.X-cameraX),
			float32(p.Y),
			float32(p.Width),
			float32(p.Height),
			color.RGBA{139, 69, 19, 255},
			false,
		)
		// Трава сверху
		vector.DrawFilledRect(
			screen,
			float32(p.X-cameraX),
			float32(p.Y),
			float32(p.Width),
			float32(10),
			color.RGBA{34, 139, 34, 255},
			false,
		)

	case "platform":
		// Платформа (кирпичная)
		vector.DrawFilledRect(
			screen,
			float32(p.X-cameraX),
			float32(p.Y),
			float32(p.Width),
			float32(p.Height),
			color.RGBA{178, 34, 34, 255},
			false,
		)
		// Верх платформы
		vector.DrawFilledRect(
			screen,
			float32(p.X-cameraX),
			float32(p.Y),
			float32(p.Width),
			float32(4),
			color.RGBA{205, 92, 92, 255},
			false,
		)
	}
}

func (l *Level) drawItem(screen *ebiten.Image, item *Item, cameraX float64) {
	switch item.Type {
	case "coin":
		// Монетка (жёлтый круг)
		vector.DrawFilledRect(
			screen,
			float32(item.X-cameraX),
			float32(item.Y),
			float32(item.Width),
			float32(item.Height),
			color.RGBA{255, 215, 0, 255},
			false,
		)
		// Блеск
		vector.DrawFilledRect(
			screen,
			float32(item.X-cameraX)+4,
			float32(item.Y)+2,
			float32(4),
			float32(4),
			color.RGBA{255, 255, 200, 255},
			false,
		)

	case "star":
		// Звезда (жёлтая)
		vector.DrawFilledRect(
			screen,
			float32(item.X-cameraX),
			float32(item.Y),
			float32(item.Width),
			float32(item.Height),
			color.RGBA{255, 255, 0, 255},
			false,
		)

	case "heart":
		// Сердце (красное)
		vector.DrawFilledRect(
			screen,
			float32(item.X-cameraX),
			float32(item.Y),
			float32(item.Width),
			float32(item.Height),
			color.RGBA{255, 50, 50, 255},
			false,
		)
	}
}

func (l *Level) drawDoor(screen *ebiten.Image, door *Door, cameraX float64) {
	// Дверь (тёмно-коричневая)
	vector.DrawFilledRect(
		screen,
		float32(door.X-cameraX),
		float32(door.Y),
		float32(door.Width),
		float32(door.Height),
		color.RGBA{101, 67, 33, 255},
		false,
	)

	// Дверная ручка
	vector.DrawFilledRect(
		screen,
		float32(door.X-cameraX)+float32(door.Width)-15,
		float32(door.Y)+float32(door.Height)/2,
		float32(8),
		float32(8),
		color.RGBA{255, 215, 0, 255},
		false,
	)

	// Свет от открытой двери
	if door.Open {
		vector.DrawFilledRect(
			screen,
			float32(door.X-cameraX)+10,
			float32(door.Y)+10,
			float32(door.Width-20),
			float32(door.Height-20),
			color.RGBA{255, 255, 200, 100},
			false,
		)
	}
}
