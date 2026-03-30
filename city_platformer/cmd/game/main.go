// City Survivor - постапокалиптический платформер на Go + Ebitengine
// Go365 Challenge - День 90 (30 марта 2026)
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"city_platformer/internal/game"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

func init() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("🏙️ City Survivor - Go365 Day 90")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
}

func main() {
	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
