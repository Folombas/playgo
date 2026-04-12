package main

import (
	"log"
	"math"
	"match3/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// Мобильный формат 9:16
	screenWidth  = 450
	screenHeight = 800
	screenScale  = 1.0
	targetFPS    = 60
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("🍓 Fruit Crush Saga")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(targetFPS)

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Утилиты
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func Clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(max, val))
}

func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
