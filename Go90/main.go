// Package main - точка входа игры Dungeon Survivor
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/go90/internal/game"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("🎮 Dungeon Survivor - Go90 Roguelike")

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
