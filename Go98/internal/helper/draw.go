package helper

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// DrawRect draws a colored rectangle on the screen
func DrawRect(screen *ebiten.Image, x, y, w, h float64, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	
	// Create a 1x1 image and scale it
	img := ebiten.NewImage(1, 1)
	
	if col, ok := c.(color.RGBA); ok {
		img.Fill(color.RGBA{col.R, col.G, col.B, col.A})
	} else {
		img.Fill(c)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(w, h)
	screen.DrawImage(img, op)
}
