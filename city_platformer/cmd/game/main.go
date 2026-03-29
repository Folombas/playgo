// City Platformer - 2D платформер в постапокалиптическом городе
// Go365 Challenge - День 91 (30 марта 2026)
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
	ebiten.SetWindowTitle("🏙️ City Platformer - Last Survivor")
}

func main() {
	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
