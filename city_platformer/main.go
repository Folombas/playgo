// City Platformer - 2D платформер в городе с PMC контрактником
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/city_platformer/internal/game"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("🏙️ City Platformer - PMC Contractor")

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
