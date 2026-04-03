package hud

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// DrawHUD draws the game HUD.
func DrawHUD(screen *ebiten.Image, hp, maxHP, score, coins int) {
	// HP
	for i := 0; i < maxHP; i++ {
		c := color.RGBA{255, 50, 50, 255}
		if i >= hp {
			c = color.RGBA{80, 80, 80, 255}
		}
		text.Draw(screen, "♥", basicfont.Face7x13, 10+i*18, 25, c)
	}

	// Score
	text.Draw(screen, fmt.Sprintf("Score: %d", score), basicfont.Face7x13, 10, 50, color.White)

	// Coins
	text.Draw(screen, fmt.Sprintf("Coins: %d", coins), basicfont.Face7x13, 10, 70, color.RGBA{255, 215, 0, 255})
}

// DrawMenu draws the title screen.
func DrawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})
	drawCentered(screen, "CITY RUNNER", 200, color.RGBA{0, 200, 255, 255})
	drawCentered(screen, "Press ENTER to Start", 320, color.White)
	drawCentered(screen, "SPACE / W / UP = Jump (Double Jump!)", 400, color.RGBA{180, 180, 180, 255})
	drawCentered(screen, "ESC = Quit", 430, color.RGBA{180, 180, 180, 255})
}

// DrawGameOver draws game over screen.
func DrawGameOver(screen *ebiten.Image, score int) {
	screen.Fill(color.RGBA{0, 0, 0, 200})
	drawCentered(screen, "GAME OVER", 280, color.RGBA{255, 80, 80, 255})
	drawCentered(screen, fmt.Sprintf("Score: %d", score), 340, color.White)
	drawCentered(screen, "Press ENTER to Restart", 400, color.RGBA{100, 200, 255, 255})
}

func drawCentered(screen *ebiten.Image, msg string, y int, c color.Color) {
	w := len(msg) * 7
	x := (1280 - w) / 2
	text.Draw(screen, msg, basicfont.Face7x13, x, y, c)
}
