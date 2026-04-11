package logic

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// Particle представляет частицу для эффектов
type Particle struct {
	X, Y     float64
	VX, VY   float64
	Life     float64
	MaxLife  float64
	Color    color.Color
	Size     float64
	Rotation float64
}

// EffectSystem управляет визуальными эффектами
type EffectSystem struct {
	Particles []*Particle
}

// NewEffectSystem создаёт новую систему эффектов
func NewEffectSystem() *EffectSystem {
	return &EffectSystem{
		Particles: make([]*Particle, 0),
	}
}

// SpawnMatchEffect создаёт эффект при уничтожении матча
func (es *EffectSystem) SpawnMatchEffect(x, y int, gemColor color.Color, count int) {
	// Создать взрыв частиц
	particleCount := count * 5
	for i := 0; i < particleCount; i++ {
		angle := (math.Pi * 2 * float64(i)) / float64(particleCount)
		speed := 2 + rand.Float64()*3

		p := &Particle{
			X:    float64(x) + 30,
			Y:    float64(y) + 30,
			VX:   math.Cos(angle) * speed,
			VY:   math.Sin(angle) * speed,
			Life: 1.0,
			MaxLife: 1.0,
			Color: gemColor,
			Size:  3 + rand.Float64()*4,
		}
		es.Particles = append(es.Particles, p)
	}

	// Добавить искры
	for i := 0; i < 10; i++ {
		p := &Particle{
			X:    float64(x) + 30 + (rand.Float64()-0.5)*40,
			Y:    float64(y) + 30 + (rand.Float64()-0.5)*40,
			VX:   (rand.Float64() - 0.5) * 4,
			VY:   -2 - rand.Float64()*3,
			Life: 0.8,
			MaxLife: 0.8,
			Color: color.RGBA{255, 255, 255, 255},
			Size:  2,
		}
		es.Particles = append(es.Particles, p)
	}
}

// Update обновляет все частицы
func (es *EffectSystem) Update() {
	alive := make([]*Particle, 0)
	for _, p := range es.Particles {
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.1 // Гравитация частиц
		p.Life -= 0.02
		p.Rotation += 0.1

		if p.Life > 0 {
			alive = append(alive, p)
		}
	}
	es.Particles = alive
}

// Draw отрисовывает все частицы
func (es *EffectSystem) Draw(screen *ebiten.Image) {
	for _, p := range es.Particles {
		alpha := uint8((p.Life / p.MaxLife) * 255)
		c := p.Color
		r, g, b, _ := c.RGBA()
		
		particleColor := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: alpha,
		}

		img := ebiten.NewImage(int(p.Size), int(p.Size))
		img.Fill(particleColor)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-p.Size/2, p.Y-p.Size/2)
		op.GeoM.Rotate(p.Rotation)
		
		screen.DrawImage(img, op)
	}
}

// ComboEffect создаёт эффект комбо (большой взрыв)
func (es *EffectSystem) SpawnComboEffect(x, y int, comboLevel int) {
	// Большие кольца
	for i := 0; i < 360; i += 10 {
		angle := float64(i) * math.Pi / 180
		p := &Particle{
			X:    float64(x) + 30,
			Y:    float64(y) + 30,
			VX:   math.Cos(angle) * float64(comboLevel) * 1.5,
			VY:   math.Sin(angle) * float64(comboLevel) * 1.5,
			Life: 1.5,
			MaxLife: 1.5,
			Color: color.RGBA{255, 215, 0, 255}, // Золотой
			Size:  4,
		}
		es.Particles = append(es.Particles, p)
	}
	
	// Добавить дополнительные частицы для высоких комбо
	if comboLevel >= 3 {
		for i := 0; i < 20*comboLevel; i++ {
			angle := rng.Float64() * math.Pi * 2
			speed := 1 + rng.Float64()*float64(comboLevel)
			
			p := &Particle{
				X:    float64(x) + 30,
				Y:    float64(y) + 30,
				VX:   math.Cos(angle) * speed,
				VY:   math.Sin(angle) * speed,
				Life: 2.0,
				MaxLife: 2.0,
				Color: color.RGBA{
					R: uint8(200 + rng.Intn(55)),
					G: uint8(rng.Intn(100)),
					B: uint8(rng.Intn(100)),
					A: 255,
				},
				Size: 3 + rng.Float64()*5,
			}
			es.Particles = append(es.Particles, p)
		}
	}
}

// SpawnSparkEffect создаёт эффект искр
func (es *EffectSystem) SpawnSparkEffect(x, y int, count int) {
	for i := 0; i < count; i++ {
		angle := rng.Float64() * math.Pi * 2
		speed := 2 + rng.Float64()*4
		
		p := &Particle{
			X:    float64(x),
			Y:    float64(y),
			VX:   math.Cos(angle) * speed,
			VY:   math.Sin(angle) * speed,
			Life: 0.5 + rng.Float64()*0.5,
			MaxLife: 1.0,
			Color: color.RGBA{255, 255, 200, 255},
			Size:  1 + rng.Float64()*2,
		}
		es.Particles = append(es.Particles, p)
	}
}

// SpawnCelebrationEffect создаёт эффект празднования (фейерверк)
func (es *EffectSystem) SpawnCelebrationEffect(x, y int) {
	// Несколько волн частиц
	for wave := 0; wave < 3; wave++ {
		for i := 0; i < 30; i++ {
			angle := float64(i) * math.Pi * 2 / 30
			speed := 3 + rng.Float64()*3
			
			// Разные цвета для каждой волны
			var c color.Color
			switch wave {
			case 0:
				c = color.RGBA{255, 215, 0, 255} // Золотой
			case 1:
				c = color.RGBA{255, 100, 100, 255} // Красный
			case 2:
				c = color.RGBA{100, 255, 100, 255} // Зелёный
			}
			
			p := &Particle{
				X:    float64(x),
				Y:    float64(y),
				VX:   math.Cos(angle) * speed,
				VY:   math.Sin(angle) * speed - 2,
				Life: 2.0 + float64(wave)*0.5,
				MaxLife: 3.0,
				Color: c,
				Size:  3 + rng.Float64()*4,
			}
			es.Particles = append(es.Particles, p)
		}
	}
}

// ScorePopup представляет всплывающий текст счёта
type ScorePopup struct {
	X, Y     float64
	Score    int
	Life     float64
	MaxLife  float64
}

// ScorePopupSystem управляет всплывающими очками
type ScorePopupSystem struct {
	Popups []*ScorePopup
}

// NewScorePopupSystem создаёт систему всплывающих очков
func NewScorePopupSystem() *ScorePopupSystem {
	return &ScorePopupSystem{
		Popups: make([]*ScorePopup, 0),
	}
}

// AddScorePopup добавляет всплывающий счёт
func (sps *ScorePopupSystem) AddScorePopup(x, y int, score int) {
	sps.Popups = append(sps.Popups, &ScorePopup{
		X:       float64(x) + 30,
		Y:       float64(y) + 30,
		Score:   score,
		Life:    1.5,
		MaxLife: 1.5,
	})
}

// Update обновляет всплывающие очки
func (sps *ScorePopupSystem) Update() {
	alive := make([]*ScorePopup, 0)
	for _, p := range sps.Popups {
		p.Y -= 1 // Всплывает вверх
		p.Life -= 0.02

		if p.Life > 0 {
			alive = append(alive, p)
		}
	}
	sps.Popups = alive
}

// IsEmpty проверяет, есть ли активные эффекты
func (es *EffectSystem) IsEmpty() bool {
	return len(es.Particles) == 0
}

// IsEmpty проверяет, есть ли активные попапы
func (sps *ScorePopupSystem) IsEmpty() bool {
	return len(sps.Popups) == 0
}
