package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// EnemyType определяет тип врага
type EnemyType int

const (
	EnemySlime EnemyType = iota
	EnemyBat
	EnemySkeleton
	EnemyGhost
	EnemyBoss
)

// EnemyConfig конфигурация типа врага
type EnemyConfig struct {
	HP      float64
	Damage  float64
	Speed   float64
	Width   float64
	Height  float64
	XPValue float64
	Color   color.Color
}

// EnemyConfigs содержит конфигурации всех типов врагов
var EnemyConfigs = map[EnemyType]EnemyConfig{
	EnemySlime: {
		HP:      30,
		Damage:  5,
		Speed:   1.5,
		Width:   28,
		Height:  28,
		XPValue: 5,
		Color:   color.RGBA{50, 200, 50, 255},
	},
	EnemyBat: {
		HP:      20,
		Damage:  3,
		Speed:   2.5,
		Width:   24,
		Height:  24,
		XPValue: 3,
		Color:   color.RGBA{150, 50, 150, 255},
	},
	EnemySkeleton: {
		HP:      50,
		Damage:  8,
		Speed:   1.2,
		Width:   30,
		Height:  30,
		XPValue: 8,
		Color:   color.RGBA{200, 200, 200, 255},
	},
	EnemyGhost: {
		HP:      40,
		Damage:  6,
		Speed:   1.8,
		Width:   32,
		Height:  32,
		XPValue: 6,
		Color:   color.RGBA{100, 100, 255, 180},
	},
	EnemyBoss: {
		HP:      500,
		Damage:  20,
		Speed:   0.8,
		Width:   80,
		Height:  80,
		XPValue: 100,
		Color:   color.RGBA{255, 50, 50, 255},
	},
}

// Enemy представляет врага
type Enemy struct {
	*Entity
	Type        EnemyType
	XPValue     float64
	AttackTimer float64
	AttackSpeed float64
}

// NewEnemy создаёт нового врага
func NewEnemy(x, y float64, enemyType EnemyType) *Enemy {
	cfg := EnemyConfigs[enemyType]
	e := &Enemy{
		Entity:      NewEntity(x, y, cfg.Width, cfg.Height, cfg.Speed, cfg.HP, cfg.Damage, cfg.Color),
		Type:        enemyType,
		XPValue:     cfg.XPValue,
		AttackTimer: 0,
		AttackSpeed: 1.0,
	}
	e.Speed = cfg.Speed
	return e
}

// Update обновляет врага
func (e *Enemy) Update(player *Player) {
	if !e.IsActive {
		return
	}

	// AI преследования игрока
	dx := player.X - e.X
	dy := player.Y - e.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > 0 {
		e.X += (dx / distance) * e.Speed
		e.Y += (dy / distance) * e.Speed
	}

	// Таймер атаки
	if e.AttackTimer > 0 {
		e.AttackTimer -= 1.0 / 60.0
	}
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image) {
	if !e.IsActive {
		return
	}

	// Тело врага
	vector.DrawFilledRect(
		screen,
		float32(e.X-e.Width/2),
		float32(e.Y-e.Height/2),
		float32(e.Width),
		float32(e.Height),
		e.Color,
		true,
	)

	// Глаза (кроме босса)
	if e.Type != EnemyBoss {
		eyeSize := float32(4)
		vector.DrawFilledRect(screen, float32(e.X-6), float32(e.Y-3), eyeSize, eyeSize, color.White, true)
		vector.DrawFilledRect(screen, float32(e.X+2), float32(e.Y-3), eyeSize, eyeSize, color.White, true)
	}

	// Полоска здоровья для босса
	if e.Type == EnemyBoss {
		e.drawBossHealthBar(screen)
	}
}

// drawBossHealthBar рисует полоску здоровья босса
func (e *Enemy) drawBossHealthBar(screen *ebiten.Image) {
	barWidth := float32(e.Width + 20)
	barHeight := float32(6)
	x := float32(e.X) - barWidth/2
	y := float32(e.Y - e.Height/2 - 12)

	// Фон
	vector.DrawFilledRect(screen, x, y, barWidth, barHeight, color.RGBA{50, 50, 50, 255}, true)

	// Здоровье
	hpPercent := e.HP / e.MaxHP
	hpWidth := barWidth * float32(hpPercent)
	vector.DrawFilledRect(screen, x, y, hpWidth, barHeight, color.RGBA{255, 50, 50, 255}, true)
}

// CanAttack проверяет, может ли враг атаковать
func (e *Enemy) CanAttack() bool {
	return e.AttackTimer <= 0
}

// ResetAttack сбрасывает таймер атаки
func (e *Enemy) ResetAttack() {
	e.AttackTimer = 1.0 / e.AttackSpeed
}

// GetEnemyTypeForWave возвращает тип врага для волны
func GetEnemyTypeForWave(wave int) EnemyType {
	switch {
	case wave < 3:
		return EnemySlime
	case wave < 5:
		return EnemyBat
	case wave < 8:
		return EnemySkeleton
	case wave < 10:
		return EnemyGhost
	default:
		// Босс каждые 10 волн
		if wave%10 == 0 {
			return EnemyBoss
		}
		return EnemyGhost
	}
}
