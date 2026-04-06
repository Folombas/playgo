package ui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// DrawText draws text with optional scaling
func DrawText(screen *ebiten.Image, str string, x, y, size int, c color.Color) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	
	if size != 13 { // basicfont.Face7x13 is 13px
		scale := float64(size) / 13.0
		op.GeoM.Scale(scale, scale)
	}
	
	// Create text image
	txtImg := text.Append(nil, str, &text.Options{
		Font: basicfont.Face7x13,
		FontSize: float64(size),
		LineSpacing: float64(size) + 2,
	})
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	
	if col, ok := c.(color.RGBA); ok {
		op2.ColorScale.SetR(float32(col.R) / 255)
		op2.ColorScale.SetG(float32(col.G) / 255)
		op2.ColorScale.SetB(float32(col.B) / 255)
		op2.ColorScale.SetA(float32(col.A) / 255)
	}
	
	screen.DrawImage(txtImg, op2)
}

// DrawMultilineText draws text with newlines
func DrawMultilineText(screen *ebiten.Image, str string, x, y int, size int, c color.Color) {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		DrawText(screen, line, x, y+i*(size+5), size, c)
	}
}
