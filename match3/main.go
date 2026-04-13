package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// appGame is our main game struct implementing ebiten.Game interface.
type appGame struct {
	game *Game
}

// Update implements ebiten.Game.
func (g *appGame) Update() error {
	// Handle keyboard
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.game.RPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.game.PPressed = true
	}

	return g.game.Update()
}

// Draw implements ebiten.Game.
func (g *appGame) Draw(screen *ebiten.Image) {
	// Draw background
	if backgroundImage != nil {
		screen.DrawImage(backgroundImage, nil)
	} else {
		screen.Fill(color.RGBA{0x1A, 0x1A, 0x2E, 0xFF})
	}

	// Draw board
	drawBoard(screen, g.game)

	// Draw UI
	drawUI(screen, g.game)

	// Paused overlay
	if g.game.Paused && !g.game.GameOver {
		vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0x00, 0x00, 0x00, 0x80}, false)
		ebitenutil.DebugPrintAt(screen, "PAUSED", ScreenWidth/2-40, ScreenHeight/2)
	}
}

// Layout implements ebiten.Game.
func (g *appGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	// Initialize assets
	generateTileSprites(56) // 60px cell - 4px padding
	generateBackground(ScreenWidth, ScreenHeight)

	game := NewGame()
	ebitenGame := &appGame{game: game}

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Match-3 — Go365 Day 103")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(ebitenGame); err != nil {
		log.Fatal(err)
	}
}
