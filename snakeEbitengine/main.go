package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"snake/internal/game"
	"snake/internal/types"
)

func main() {
	ebiten.SetWindowSize(types.ScreenW, types.ScreenH)
	ebiten.SetWindowTitle("Змейка: Подарок Судьбы")
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(game.NewGame()); err != nil {
		log.Fatal(err)
	}
}
