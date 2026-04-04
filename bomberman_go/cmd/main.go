package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/bomberman_go/internal/config"
	"github.com/playgo/bomberman_go/internal/game"
)

func main() {
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle("Bomberman Go - Go365 Challenge")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	gameInstance := game.NewGame()

	if err := ebiten.RunGame(gameInstance); err != nil {
		log.Fatal(err)
	}
}
