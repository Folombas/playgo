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
	Particles   []Particle
	ScreenShake ScreenShake
}

// NewEffectSystem создаёт новую систему эффектов
func NewEffectSystem() *EffectSystem {
	return &EffectSystem{
		Particles:   []Particle{},
		ScreenShake: ScreenShake{},
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
	
	es.ScreenShake.Update()
}

// Draw отрисовывает все частицы
func (es *EffectSystem) Draw(screen *ebiten.Image) {
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
