package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Star represents a background star with parallax.
type Star struct {
	X, Y       float64
	Size       float64
	Brightness float64
	TwinkleSpeed float64
	Layer      int // 0=far (slow), 1=mid, 2=near (fast)
}

// ParallaxBackground manages the animated star field.
type ParallaxBackground struct {
	Stars    []*Star
	OffsetX  float64
	OffsetY  float64
	Time     float64
}

// NewParallaxBackground creates a new animated star field.
func NewParallaxBackground(width, height int) *ParallaxBackground {
	bg := &ParallaxBackground{
		Stars: make([]*Star, 0, 150),
	}
	bg.generateStars(width, height)
	return bg
}

// generateStars creates stars at different depths for parallax effect.
func (bg *ParallaxBackground) generateStars(width, height int) {
	rng := rand.New(rand.NewSource(42)) // Fixed seed for consistency

	// Far layer - many small dim stars
	for i := 0; i < 80; i++ {
		bg.Stars = append(bg.Stars, &Star{
			X:            rng.Float64() * float64(width),
			Y:            rng.Float64() * float64(height),
			Size:         1 + rng.Float64()*1.5,
			Brightness:   0.3 + rng.Float64()*0.3,
			TwinkleSpeed: 1 + rng.Float64()*2,
			Layer:        0,
		})
	}

	// Mid layer - medium stars
	for i := 0; i < 50; i++ {
		bg.Stars = append(bg.Stars, &Star{
			X:            rng.Float64() * float64(width),
			Y:            rng.Float64() * float64(height),
			Size:         1.5 + rng.Float64()*2,
			Brightness:   0.5 + rng.Float64()*0.3,
			TwinkleSpeed: 2 + rng.Float64()*2,
			Layer:        1,
		})
	}

	// Near layer - few large bright stars
	for i := 0; i < 20; i++ {
		bg.Stars = append(bg.Stars, &Star{
			X:            rng.Float64() * float64(width),
			Y:            rng.Float64() * float64(height),
			Size:         2 + rng.Float64()*3,
			Brightness:   0.7 + rng.Float64()*0.3,
			TwinkleSpeed: 3 + rng.Float64()*2,
			Layer:        2,
		})
	}
}

// Update advances the background animation.
func (bg *ParallaxBackground) Update(dt float64) {
	bg.Time += dt

	// Slow drift for parallax feel
	bg.OffsetX += dt * 2
	bg.OffsetY += dt * 0.5

	// Wrap around
	if bg.OffsetX > float64(ScreenWidth) {
		bg.OffsetX = 0
	}
	if bg.OffsetY > float64(ScreenHeight) {
		bg.OffsetY = 0
	}
}

// Draw renders the star field with parallax offsets.
func (bg *ParallaxBackground) Draw(screen *ebiten.Image) {
	// Draw base background
	baseColor := color.RGBA{0x0A, 0x0A, 0x1A, 0xFF}
	vector.DrawFilledRect(screen, 0, 0, float32(ScreenWidth), float32(ScreenHeight), baseColor, true)

	// Draw stars with parallax
	for _, star := range bg.Stars {
		// Calculate parallax offset based on layer
		parallaxFactor := []float64{0.2, 0.5, 1.0}[star.Layer]
		starX := star.X + bg.OffsetX*parallaxFactor
		starY := star.Y + bg.OffsetY*parallaxFactor

		// Wrap around screen
		for starX < 0 {
			starX += float64(ScreenWidth)
		}
		for starX >= float64(ScreenWidth) {
			starX -= float64(ScreenWidth)
		}
		for starY < 0 {
			starY += float64(ScreenHeight)
		}
		for starY >= float64(ScreenHeight) {
			starY -= float64(ScreenHeight)
		}

		// Twinkle effect
		twinkle := 0.5 + 0.5*math.Sin(bg.Time*star.TwinkleSpeed)
		alpha := uint8(star.Brightness * twinkle * 255)

		starColor := color.RGBA{0xFF, 0xFF, 0xFF, alpha}

		// Draw star as a circle
		vector.DrawFilledCircle(screen, float32(starX), float32(starY), float32(star.Size), starColor, true)

		// Glow effect for bright stars
		if star.Layer == 2 {
			glowColor := color.RGBA{0xAA, 0xAA, 0xFF, alpha / 3}
			vector.DrawFilledCircle(screen, float32(starX), float32(starY), float32(star.Size*2), glowColor, true)
		}
	}
}
