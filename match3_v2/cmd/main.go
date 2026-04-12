package main

import (
	"image/color"
	"log"
	"match3_v2/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 800
	targetFPS    = 60
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("💎 Gem Crush - Match 3")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetTPS(targetFPS)

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// drawRoundedRect рисует скруглённый прямоугольник
func drawRoundedRect(screen *ebiten.Image, x, y, w, h float64, cr float64, c color.Color) {
	// Основная часть
	vector.DrawFilledRect(screen, float32(x+cr), float32(y), float32(w-cr*2), float32(h), c, false)
	vector.DrawFilledRect(screen, float32(x), float32(y+cr), float32(w), float32(h-cr*2), c, false)

	// Углы
	vector.DrawFilledCircle(screen, float32(x+cr), float32(y+cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+w-cr), float32(y+cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+cr), float32(y+h-cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+w-cr), float32(y+h-cr), float32(cr), c, false)
}
