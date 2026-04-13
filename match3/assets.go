package main

import (
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	_ "image/png"
)

// tileColors defines the 6 food tile colors (for particles/effects).
var tileColors = []color.RGBA{
	{0xE7, 0x4C, 0x3C, 0xFF}, // Red (apple)
	{0x8B, 0x45, 0x13, 0xFF}, // Brown (burger)
	{0xFF, 0x8C, 0x00, 0xFF}, // Orange (carrot)
	{0xFF, 0x69, 0xB4, 0xFF}, // Pink (cupcake)
	{0x80, 0x00, 0x80, 0xFF}, // Purple (grapes)
	{0xFF, 0xD7, 0x00, 0xFF}, // Yellow (pizza)
}

// tileImages holds the loaded food sprites.
var tileImages []*ebiten.Image

// backgroundImage holds the game background.
var backgroundImage *ebiten.Image

// foodSpriteFiles lists the PNG files to load for each tile type (0..5).
var foodSpriteFiles = []string{
	"apple.png",    // 0 - Red
	"burger.png",   // 1 - Brown
	"carrot.png",   // 2 - Orange
	"cupcake.png",  // 3 - Pink
	"grapes.png",   // 4 - Purple
	"pizza.png",    // 5 - Yellow
}

// loadTileSprites loads food sprites from the sprites/ directory.
func loadTileSprites(size int) {
	tileImages = make([]*ebiten.Image, len(foodSpriteFiles))

	for i, filename := range foodSpriteFiles {
		path := filepath.Join("sprites", filename)
		f, err := os.Open(path)
		if err != nil {
			log.Fatalf("Failed to load sprite %s: %v", path, err)
		}

		imgData, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			log.Fatalf("Failed to decode sprite %s: %v", path, err)
		}

		img := ebiten.NewImageFromImage(imgData)

		// Scale sprite to fit cell size
		scaled := ebiten.NewImage(size, size)
		op := &ebiten.DrawImageOptions{}
		w, h := img.Size()
		scale := float64(size) / float64(w)
		if float64(size)/float64(h) < scale {
			scale = float64(size) / float64(h)
		}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate((float64(size)-float64(w)*scale)/2, (float64(size)-float64(h)*scale)/2)
		scaled.DrawImage(img, op)

		tileImages[i] = scaled
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
