package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// DrawText рисует текст на экране
func DrawText(screen *ebiten.Image, str string, x, y int, c color.Color) {
	text.Draw(screen, str, basicfont.Face7x13, x, y+12, c)
}
