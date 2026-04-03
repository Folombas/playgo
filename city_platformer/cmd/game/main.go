package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"city_platformer/internal/engine"
)

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("City Runner - Go365 Day 94")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := engine.NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
