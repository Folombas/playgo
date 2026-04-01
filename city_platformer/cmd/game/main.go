// Cyber City Runner - Киберпанк-платформер с паркуром и хакерством
// Go365 Challenge - День 92 (1 апреля 2026)
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"cyber_city/internal/game"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

func init() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("🌃 Cyber City Runner - Go365 Day 92")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
}

func main() {
	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
