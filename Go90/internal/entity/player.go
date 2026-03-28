package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Player представляет игрока
type Player struct {
	*Entity
	Level        int
	XP           float64
	XPToLevel    float64
	Damage       float64
	AttackSpeed  float64 // атак в секунду
	AttackTimer  float64
	MoveSpeed    float64
	Range        float64
	PickupRange  float64
	Luck         float64
	Area         float64
}

// NewPlayer создаёт нового игрока
func NewPlayer(x, y float64) *Player {
	return &Player{
		Entity:      NewEntity(x, y, 32, 32, 0, 100, 10, color.RGBA{0, 150, 255, 255}),
		Level:       1,
		XP:          0,
		XPToLevel:   20,
		Damage:      15,
		AttackSpeed: 1.0,
		AttackTimer: 0,
		MoveSpeed:   4.0,
		Range:       200,
		PickupRange: 100,
		Luck:        0,
		Area:        1.0,
	}
}

// Update обновляет состояние игрока
func (p *Player) Update() {
	// Таймер атаки
	if p.AttackTimer > 0 {
		p.AttackTimer -= 1.0 / 60.0 // 60 FPS
	}
}

// Draw отрисовывает игрока с деталями
func (p *Player) Draw(screen *ebiten.Image) {
	// Тело игрока (синий квадрат)
	vector.DrawFilledRect(
		screen,
		float32(p.X-p.Width/2),
		float32(p.Y-p.Height/2),
		float32(p.Width),
		float32(p.Height),
		color.RGBA{0, 150, 255, 255},
		true,
	)

	// Глаза
	eyeSize := float32(6)
	vector.DrawFilledRect(screen, float32(p.X-8), float32(p.Y-4), eyeSize, eyeSize, color.White, true)
	vector.DrawFilledRect(screen, float32(p.X+2), float32(p.Y-4), eyeSize, eyeSize, color.White, true)

	// Радиус атаки (полупрозрачный круг)
	p.drawRangeCircle(screen)
}

// drawRangeCircle рисует круг радиуса атаки
func (p *Player) drawRangeCircle(screen *ebiten.Image) {
	// Упрощённая визуализация - 4 точки по кругу
	r := float32(p.Range)
	cx, cy := float32(p.X), float32(p.Y)

	points := []struct{ x, y float32 }{
		{cx + r, cy},
		{cx - r, cy},
		{cx, cy + r},
		{cx, cy - r},
	}

	for _, pt := range points {
		vector.DrawFilledRect(screen, pt.x-2, pt.y-2, 4, 4, color.RGBA{255, 100, 100, 100}, true)
	}
}

// CanAttack проверяет, может ли игрок атаковать
func (p *Player) CanAttack() bool {
	return p.AttackTimer <= 0
}

// ResetAttack сбрасывает таймер атаки
func (p *Player) ResetAttack() {
	p.AttackTimer = 1.0 / p.AttackSpeed
}

// AddXP добавляет опыт и повышает уровень
func (p *Player) AddXP(amount float64) int {
	p.XP += amount
	levelsGained := 0

	for p.XP >= p.XPToLevel {
		p.XP -= p.XPToLevel
		p.Level++
		p.XPToLevel = float64(p.Level) * 20
		levelsGained++

		// Увеличение статов при повышении уровня
		p.Damage += 2
		p.MaxHP += 10
		p.HP = p.MaxHP
	}

	return levelsGained
}

// AngleTo вычисляет угол до цели
func (p *Player) AngleTo(targetX, targetY float64) float64 {
	dx := targetX - p.X
	dy := targetY - p.Y
	return math.Atan2(dy, dx)
}

// DistanceTo вычисляет расстояние до точки
func (p *Player) DistanceTo(x, y float64) float64 {
	dx := x - p.X
	dy := y - p.Y
	return math.Sqrt(dx*dx + dy*dy)
}
