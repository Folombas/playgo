// City Platformer - 2D платформер в постапокалиптическом городе
// Go365 Challenge - День 88 (29 марта 2026)
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/city_platformer/pkg/game"
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
