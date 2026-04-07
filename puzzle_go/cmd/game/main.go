package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/puzzle_go/internal/game"
)

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Crystal Cascade - Puzzle GO")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
