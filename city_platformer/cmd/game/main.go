// City Platformer: Neon Runner - Cyber Escape
// Go365 Challenge - День 93 (2 апреля 2026)
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
	ebiten.SetWindowTitle("Neon Runner: Cyber Escape - Go365 Day 93")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
}

func main() {
	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
