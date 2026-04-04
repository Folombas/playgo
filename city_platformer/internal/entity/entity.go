package entity

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Player - игровой персонаж
type Player struct {
	X, Y      float64
	VX, VY    float64
	Width     float64
	Height    float64
	OnGround  bool
	JumpCount int
	Speed     float64
	JumpForce float64
	Gravity   float64
}

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64) *Player {
	return &Player{
		X:         x,
		Y:         y,
		Width:     32,
		Height:    48,
		Speed:     5,
		JumpForce: -12,
		Gravity:   0.6,
	}
}

// Update обновляет состояние игрока
func (p *Player) Update() {
	// Гравитация
	p.VY += p.Gravity

	// Ограничение скорости падения
	if p.VY > 15 {
		p.VY = 15
	}

	// Движение по X
	p.X += p.VX

	// Движение по Y
	p.Y += p.VY

	// Трение
	p.VX *= 0.8

	// Сброс счётчика прыжков на земле
	if p.OnGround {
		p.JumpCount = 0
	}
}

// Jump выполняет прыжок
func (p *Player) Jump() {
	if p.JumpCount < 2 { // Двойной прыжок
		p.VY = p.JumpForce
		p.JumpCount++
		p.OnGround = false
	}
}

// MoveLeft двигает влево
func (p *Player) MoveLeft() {
	p.VX = -p.Speed
}

// MoveRight двигает вправо
func (p *Player) MoveRight() {
	p.VX = p.Speed
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX float64) {
	// Тело (синее)
	vector.DrawFilledRect(
		screen,
		float32(p.X-cameraX),
		float32(p.Y),
		float32(p.Width),
		float32(p.Height),
		color.RGBA{30, 144, 255, 255},
		false,
	)

	// Голова
	vector.DrawFilledRect(
		screen,
		float32(p.X-cameraX)+8,
		float32(p.Y)-8,
		float32(p.Width-16),
		float32(p.Height-40),
		color.RGBA{255, 200, 150, 255},
		false,
	)

	// Глаза
	vector.DrawFilledRect(
		screen,
		float32(p.X-cameraX)+18,
		float32(p.Y)-4,
		float32(4),
		float32(4),
		color.RGBA{0, 0, 0, 255},
		false,
	)

	// Повязка (бандана)
	vector.DrawFilledRect(
		screen,
		float32(p.X-cameraX)+6,
		float32(p.Y)-6,
		float32(p.Width-12),
		float32(6),
		color.RGBA{255, 50, 50, 255},
		false,
	)
}

// Enemy - враг
type Enemy struct {
	X, Y        float64
	VX, VY      float64
	Width       float64
	Height      float64
	Alive       bool
	Speed       float64
	PatrolRange float64
	StartX      float64
	Type        string
}

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType string) *Enemy {
	width, height := 32.0, 32.0
	speed := 1.0

	switch enemyType {
	case "slime":
		width, height = 24, 16
		speed = 0.5
	case "snake":
		width, height = 32, 16
		speed = 1.5
	case "spider":
		width, height = 20, 20
		speed = 0.8
	}

	return &Enemy{
		X:           x,
		Y:           y,
		Width:       width,
		Height:      height,
		Alive:       true,
		Speed:       speed,
		PatrolRange: 100,
		StartX:      x,
		Type:        enemyType,
	}
}

// Update обновляет врага
func (e *Enemy) Update() {
	if !e.Alive {
		return
	}

	// Патрулирование
	e.X += e.VX

	// Разворот на границах патруля
	if e.X < e.StartX-e.PatrolRange || e.X > e.StartX+e.PatrolRange {
		e.VX = -e.VX
	}

	// Начальное движение
	if e.VX == 0 {
		e.VX = e.Speed
	}
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX float64) {
	if !e.Alive {
		return
	}

	switch e.Type {
	case "slime":
		// Слизень (зелёный)
		vector.DrawFilledRect(
			screen,
			float32(e.X-cameraX),
			float32(e.Y),
			float32(e.Width),
			float32(e.Height),
			color.RGBA{100, 255, 100, 255},
			false,
		)
		// Глаза
		vector.DrawFilledRect(
			screen,
			float32(e.X-cameraX)+4,
			float32(e.Y)+2,
			float32(4),
			float32(4),
			color.RGBA{0, 0, 0, 255},
			false,
		)
		vector.DrawFilledRect(
			screen,
			float32(e.X-cameraX)+float32(e.Width)-8,
			float32(e.Y)+2,
			float32(4),
			float32(4),
			color.RGBA{0, 0, 0, 255},
			false,
		)

	case "snake":
		// Змейка (коричневая)
		vector.DrawFilledRect(
			screen,
			float32(e.X-cameraX),
			float32(e.Y),
			float32(e.Width),
			float32(e.Height),
			color.RGBA{139, 69, 19, 255},
			false,
		)
		// Язык
		vector.DrawFilledRect(
			screen,
			float32(e.X-cameraX)+float32(e.Width),
			float32(e.Y)+4,
			float32(8),
			float32(2),
			color.RGBA{255, 0, 0, 255},
			false,
		)

	case "spider":
		// Паучок (чёрный)
		vector.DrawFilledRect(
			screen,
			float32(e.X-cameraX),
			float32(e.Y),
			float32(e.Width),
			float32(e.Height),
			color.RGBA{30, 30, 30, 255},
			false,
		)
		// Ноги
		vector.DrawFilledRect(
			screen,
			float32(e.X-cameraX)-4,
			float32(e.Y)+8,
			float32(4),
			float32(8),
			color.RGBA{30, 30, 30, 255},
			false,
		)
		vector.DrawFilledRect(
			screen,
			float32(e.X-cameraX)+float32(e.Width),
			float32(e.Y)+8,
			float32(4),
			float32(8),
			color.RGBA{30, 30, 30, 255},
			false,
		)
	}
}

// Particle - частица для эффектов
type Particle struct {
	X, Y      float64
	VX, VY    float64
	Life      int
	MaxLife   int
	Color     color.Color
	Gravity   float64
}

// NewParticle создаёт новую частицу
func NewParticle(x, y float64, c color.Color) *Particle {
	return &Particle{
		X:       x,
		Y:       y,
		VX:      (float64(int(randomRange(-3, 4)))) * 0.5,
		VY:      (float64(int(randomRange(-5, 2)))) * 0.5,
		Life:    30,
		MaxLife: 30,
		Color:   c,
		Gravity: 0.2,
	}
}

// Update обновляет частицу
func (p *Particle) Update() {
	p.VY += p.Gravity
	p.X += p.VX
	p.Y += p.VY
	p.Life--
}

// Draw отрисовывает частицу
func (p *Particle) Draw(screen *ebiten.Image, cameraX float64) {
	alpha := uint8(float32(p.Life) / float32(p.MaxLife) * 255)
	r, g, b, _ := p.Color.RGBA()
	vector.DrawFilledRect(
		screen,
		float32(p.X-cameraX),
		float32(p.Y),
		float32(4),
		float32(4),
		color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), alpha},
		false,
	)
}

// Вспомогательная функция для случайных чисел
func randomRange(min, max int) int {
	// Простой псевдослучайный генератор
	return min + (int(uint32(min*max)%uint32(max-min+1)))
}
