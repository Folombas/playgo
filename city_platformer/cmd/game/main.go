// Sunny Adventure: Путешествие в Облачную Страну
// Go365 Challenge - День 91 (31 марта 2026)
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"sunny_adventure/internal/game"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

func init() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("🌈 Sunny Adventure - Go365 Day 91")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
}

func main() {
	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
