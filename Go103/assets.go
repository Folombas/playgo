package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// tileColors defines the 6 gem colors used for procedural sprite generation.
var tileColors = []color.RGBA{
	{0xE7, 0x4C, 0x3C, 0xFF}, // Red
	{0x34, 0x98, 0xDB, 0xFF}, // Blue
	{0x2E, 0xCC, 0x71, 0xFF}, // Green
	{0xF1, 0xC4, 0x0F, 0xFF}, // Yellow
	{0x9B, 0x59, 0xB6, 0xFF}, // Purple
	{0xE6, 0x7E, 0x22, 0xFF}, // Orange
}

// tileImages holds the pre-rendered tile sprites.
var tileImages []*ebiten.Image

// backgroundImage holds the game background.
var backgroundImage *ebiten.Image

// generateTileSprites creates colored circle sprites with a subtle glow.
func generateTileSprites(size int) {
	tileImages = make([]*ebiten.Image, len(tileColors))

	for i, c := range tileColors {
		img := ebiten.NewImage(size, size)

		// Draw a subtle shadow
		shadowColor := color.RGBA{0x00, 0x00, 0x00, 0x30}
		vector.DrawFilledCircle(img, float32(size/2)+2, float32(size/2)+2, float32(size/2-2), shadowColor, true)

		// Draw main gem
		vector.DrawFilledCircle(img, float32(size/2), float32(size/2), float32(size/2-2), c, true)

		// Draw highlight (glossy effect)
		highlightColor := color.RGBA{0xFF, 0xFF, 0xFF, 0x50}
		vector.DrawFilledCircle(img, float32(size/3), float32(size/3), float32(size/6), highlightColor, true)

		// Draw border ring
		borderColor := color.RGBA{0xFF, 0xFF, 0xFF, 0x80}
		vector.StrokeCircle(img, float32(size/2), float32(size/2), float32(size/2-1), 2, borderColor, true)

		tileImages[i] = img
	}
}

// generateBackground creates a dark gradient-like background.
func generateBackground(w, h int) {
	backgroundImage = ebiten.NewImage(w, h)
	bgColor := color.RGBA{0x1A, 0x1A, 0x2E, 0xFF}
	vector.DrawFilledRect(backgroundImage, 0, 0, float32(w), float32(h), bgColor, true)

	// Subtle grid pattern
	gridColor := color.RGBA{0x2A, 0x2A, 0x3E, 0xFF}
	for x := 0; x < w; x += 40 {
		vector.StrokeLine(backgroundImage, float32(x), 0, float32(x), float32(h), 1, gridColor, true)
	}
	for y := 0; y < h; y += 40 {
		vector.StrokeLine(backgroundImage, 0, float32(y), float32(w), float32(y), 1, gridColor, true)
	}
}
