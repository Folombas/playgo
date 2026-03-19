// Package effects содержит систему визуальных эффектов
package effects

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Particle представляет частицу для визуальных эффектов
type Particle struct {
	X, Y    float32
	VX, VY  float32
	Life    int
	MaxLife int
	Color   color.RGBA
	Size    float32
	Gravity float32
}

// BloodParticle представляет частицу крови
type BloodParticle struct {
	X, Y    float32
	VX, VY  float32
	Life    int
	MaxLife int
	Size    float32
	Gravity float32
	Color   color.RGBA
}

// BloodStain представляет пятно крови на земле
type BloodStain struct {
	X, Y    float32
	Size    float32
	Color   color.RGBA
	Life    int
	MaxLife int
}

// ScreenShake управляет тряской экрана
type ScreenShake struct {
	Intensity float32
	Duration  int
	Timer     int
	Angle     float64
}

// Update обновляет тряску экрана
func (ss *ScreenShake) Update() {
	if ss.Timer > 0 {
		ss.Timer--
		ss.Angle += math.Pi / 8
		if ss.Timer <= 0 {
			ss.Intensity = 0
		}
	}
}

// IsActive возвращает true, если тряска активна
func (ss *ScreenShake) IsActive() bool {
	return ss.Timer > 0
}

// GetOffset возвращает смещение экрана
func (ss *ScreenShake) GetOffset() (float32, float32) {
	if !ss.IsActive() {
		return 0, 0
	}
	offset := ss.Intensity * float32(ss.Timer) / float32(ss.Duration)
	dx := offset * float32(math.Sin(ss.Angle))
	dy := offset * float32(math.Cos(ss.Angle))
	return dx, dy
}

// Trigger запускает тряску экрана
func (ss *ScreenShake) Trigger(intensity float32, duration int) {
	ss.Intensity = intensity
	ss.Duration = duration
	ss.Timer = duration
	ss.Angle = rand.Float64() * math.Pi * 2
}

// EffectSystem управляет всеми эффектами
type EffectSystem struct {
	Particles    []Particle
	BloodParticles []BloodParticle
	BloodStains  []BloodStain
	ScreenShake  ScreenShake
}

// NewEffectSystem создаёт новую систему эффектов
func NewEffectSystem() *EffectSystem {
	return &EffectSystem{
		Particles:    []Particle{},
		BloodParticles: []BloodParticle{},
		BloodStains:  []BloodStain{},
		ScreenShake:  ScreenShake{},
	}
}

// SpawnParticles создаёт частицы в указанной позиции
func (es *EffectSystem) SpawnParticles(x, y float32, count int, baseColor color.RGBA, spread float32) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := rand.Float64() * float64(spread)
		particle := Particle{
			X: x,
			Y: y,
			VX: float32(math.Cos(angle) * speed),
			VY: float32(math.Sin(angle) * speed),
			Life:    20 + rand.Intn(10),
			MaxLife: 30,
			Color:   baseColor,
			Size:    2 + rand.Float32()*3,
			Gravity: 0.1,
		}
		es.Particles = append(es.Particles, particle)
	}
}

// Update обновляет все частицы
func (es *EffectSystem) Update() {
	// Обновление обычных частиц
	for i := len(es.Particles) - 1; i >= 0; i-- {
		p := &es.Particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += p.Gravity
		p.Life--

		if p.Life <= 0 {
			es.Particles = append(es.Particles[:i], es.Particles[i+1:]...)
		}
	}

	// Обновление частиц крови
	for i := len(es.BloodParticles) - 1; i >= 0; i-- {
		bp := &es.BloodParticles[i]
		bp.X += bp.VX
		bp.Y += bp.VY
		bp.VY += bp.Gravity
		bp.Life--

		if bp.Life <= 0 {
			// Создаём пятно крови
			es.BloodStains = append(es.BloodStains, BloodStain{
				X: bp.X,
				Y: bp.Y,
				Size: bp.Size * 0.8,
				Color: color.RGBA{bp.Color.R, bp.Color.G, bp.Color.B, 180},
				Life: 300, // Пятно держится 5 секунд
				MaxLife: 300,
			})
			es.BloodParticles = append(es.BloodParticles[:i], es.BloodParticles[i+1:]...)
		}
	}

	// Обновление пятен крови
	for i := len(es.BloodStains) - 1; i >= 0; i-- {
		bs := &es.BloodStains[i]
		bs.Life--
		if bs.Life <= 0 {
			es.BloodStains = append(es.BloodStains[:i], es.BloodStains[i+1:]...)
		}
	}

	es.ScreenShake.Update()
}

// SpawnBlood создаёт брызги крови
func (es *EffectSystem) SpawnBlood(x, y float32, count int, spread float32) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := rand.Float64() * float64(spread)
		blood := BloodParticle{
			X: x,
			Y: y,
			VX: float32(math.Cos(angle) * speed),
			VY: float32(math.Sin(angle) * speed) - 2, // Начальный импульс вверх
			Life:    15 + rand.Intn(10),
			MaxLife: 25,
			Size:    2 + rand.Float32()*3,
			Gravity: 0.2,
			Color:   color.RGBA{180, 0, 0, 255}, // Тёмно-красный
		}
		es.BloodParticles = append(es.BloodParticles, blood)
	}
}

// Draw отрисовывает все частицы
func (es *EffectSystem) Draw(screen *ebiten.Image) {
	// Отрисовка пятен крови
	for _, bs := range es.BloodStains {
		alpha := uint8(float32(bs.Color.A) * float32(bs.Life) / float32(bs.MaxLife))
		c := color.RGBA{bs.Color.R, bs.Color.G, bs.Color.B, alpha}
		vector.DrawFilledCircle(screen, bs.X, bs.Y, bs.Size, c, false)
	}

	// Отрисовка частиц крови
	for _, bp := range es.BloodParticles {
		alpha := uint8(float32(bp.Color.A) * float32(bp.Life) / float32(bp.MaxLife))
		c := color.RGBA{bp.Color.R, bp.Color.G, bp.Color.B, alpha}
		vector.DrawFilledCircle(screen, bp.X, bp.Y, bp.Size, c, false)
	}

	// Отрисовка обычных частиц
	for _, p := range es.Particles {
		alpha := uint8(float32(p.Color.A) * float32(p.Life) / float32(p.MaxLife))
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		vector.DrawFilledCircle(screen, p.X, p.Y, p.Size, c, false)
	}
}

// TriggerShake запускает тряску экрана
func (es *EffectSystem) TriggerShake(intensity float32, duration int) {
	es.ScreenShake.Trigger(intensity, duration)
}

// CreateGradientBackground создаёт градиентный фон
func CreateGradientBackground(screenWidth, screenHeight int) *ebiten.Image {
	gradient := ebiten.NewImage(screenWidth, screenHeight)

	for y := 0; y < screenHeight; y++ {
		ratio := float32(y) / float32(screenHeight)
		r := uint8(float32(10) * (1 - ratio))
		g := uint8(float32(10) * (1 - ratio))
		b := uint8(float32(30) * (1 - ratio))

		for x := 0; x < screenWidth; x++ {
			gradient.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return gradient
}

// PulseScale возвращает коэффициент пульсации
func PulseScale(phase float64, amplitude float64) float32 {
	return float32(1.0 + amplitude*math.Sin(phase))
}

// FoodPulseScale возвращает коэффициент пульсации для еды
func FoodPulseScale(timer int) float32 {
	if timer > 0 {
		return float32(1.0 + 0.3*math.Sin(float64(timer)*math.Pi/10))
	}
	return 1.0
}
