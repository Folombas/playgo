package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// HUD manages the game's heads-up display
type HUD struct {
	score     int
	combo     int
	timer     float64
	level     int
	maxCombo  int
	particles []Particle
}

// Particle represents a visual effect particle
type Particle struct {
	X, Y     float64
	VX, VY   float64
	Life     float64
	MaxLife  float64
	Color    color.RGBA
	Size     float64
}

// NewHUD creates a new HUD instance
func NewHUD() *HUD {
	return &HUD{
		particles: make([]Particle, 0),
	}
}

// Update updates the HUD state
func (h *HUD) Update(score, combo int, timer float64, maxCombo int) {
	h.score = score
	h.combo = combo
	h.timer = timer
	h.maxCombo = maxCombo
}

// AddScoreParticle adds a floating score particle effect
func (h *HUD) AddScoreParticle(x, y float64, points int, combo int) {
	p := Particle{
		X:       x,
		Y:       y,
		VY:      -2,
		Life:    1.0,
		MaxLife: 1.0,
		Size:    20,
	}

	if combo >= 3 {
		p.Color = color.RGBA{255, 215, 0, 255} // Gold for big combos
	} else {
		p.Color = color.RGBA{255, 255, 255, 255}
	}

	h.particles = append(h.particles, p)
	_ = points
}

// Update updates all particles
func (h *HUD) UpdateParticles() {
	for i := len(h.particles) - 1; i >= 0; i-- {
		p := &h.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.Life -= 1.0 / 60.0

		if p.Life <= 0 {
			h.particles = append(h.particles[:i], h.particles[i+1:]...)
		}
	}
}

// Draw draws the HUD on screen
func (h *HUD) Draw(screen *ebiten.Image) {
	h.drawScore(screen)
	h.drawTimer(screen)
	h.drawCombo(screen)
	h.drawParticles(screen)
}

func (h *HUD) drawScore(screen *ebiten.Image) {
	// Score background
	vector.DrawFilledRect(screen, 20, 20, 200, 50, color.RGBA{0, 0, 0, 150}, false)
	
	// Score text
	scoreText := fmt.Sprintf("Score: %d", h.score)
	drawText(screen, scoreText, 30, 35, 24, color.RGBA{255, 215, 0, 255})
}

func (h *HUD) drawTimer(screen *ebiten.Image) {
	// Timer background
	timerWidth := 150
	vector.DrawFilledRect(screen, float32(screenWidth-timerWidth-20), 20, float32(timerWidth), 50, color.RGBA{0, 0, 0, 150}, false)

	// Timer text
	timerColor := color.RGBA{255, 255, 255, 255}
	if h.timer < 10 {
		timerColor = color.RGBA{255, 50, 50, 255} // Red when low
	}
	timerText := fmt.Sprintf("Time: %.1f", h.timer)
	drawText(screen, timerText, float64(screenWidth-timerWidth), 35, 24, timerColor)

	// Timer bar
	barWidth := float32(timerWidth - 20)
	progress := h.timer / gameDuration
	vector.DrawFilledRect(screen,
		float32(screenWidth-timerWidth-10), float32(75),
		barWidth, 8,
		color.RGBA{100, 100, 100, 255}, false)
	
	barColor := color.RGBA{100, 255, 100, 255}
	if progress < 0.3 {
		barColor = color.RGBA{255, 100, 100, 255}
	} else if progress < 0.6 {
		barColor = color.RGBA{255, 215, 0, 255}
	}
	
	vector.DrawFilledRect(screen,
		float32(screenWidth-timerWidth-10), float32(75),
		barWidth*float32(progress), 8,
		barColor, false)
}

func (h *HUD) drawCombo(screen *ebiten.Image) {
	if h.combo > 1 {
		// Combo background
		comboWidth := 200
		vector.DrawFilledRect(screen, float32(screenWidth/2-comboWidth/2), 90, float32(comboWidth), 40, color.RGBA{0, 0, 0, 150}, false)
		
		// Combo text with pulsing effect
		pulse := math.Sin(float64(ebiten.ActualFPS())*0.1)*0.2 + 0.8
		comboText := fmt.Sprintf("COMBO x%d!", h.combo)
		comboColor := color.RGBA{255, 165, 0, uint8(255 * pulse)}
		drawText(screen, comboText, float64(screenWidth/2)-60, 100, 28, comboColor)
	}
}

func (h *HUD) drawParticles(screen *ebiten.Image) {
	for _, p := range h.particles {
		alpha := uint8(float64(p.Color.A) * (p.Life / p.MaxLife))
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		
		size := p.Size * (p.Life / p.MaxLife)
		vector.DrawFilledCircle(screen,
			float32(p.X), float32(p.Y),
			float32(size),
			c, false)
	}
}

// drawText helper функция для отрисовки текста
func drawText(screen *ebiten.Image, text string, x, y float64, size float64, color color.Color) {
	// Используем ebitenutil.DebugPrintAt как простой вариант
	// Для production лучше использовать text/v2 с настоящими шрифтами
	origColor := color
	_ = origColor
	_ = size
	
	// Временное решение через DebugPrintAt
	ebitenutil.DebugPrintAt(screen, text, int(x), int(y))
}
