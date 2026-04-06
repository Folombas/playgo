// Puzzle GO — Match-3 Gem Crusher
// Go365 Day 100+ — Рефакторинг: модульная архитектура
package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/playgo/puzzle_go/internal/audio"
	"github.com/playgo/puzzle_go/internal/config"
	"github.com/playgo/puzzle_go/internal/game"
	"github.com/playgo/puzzle_go/internal/render"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Init subsystems
	spr := render.LoadSprites()
	snd := audio.NewManager()

	// Init game
	g := game.NewGame(spr, snd)

	ebiten.SetWindowSize(config.WinW, config.WinH)
	ebiten.SetWindowTitle("Puzzle GO - Match-3")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
