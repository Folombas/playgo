package hud

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// HUD displays game information on screen.
type HUD struct {
	heartFull  *ebiten.Image
	heartEmpty *ebiten.Image
	coinSprite *ebiten.Image
}

// NewHUD creates a new HUD.
func NewHUD(heartFull, heartEmpty, coinSprite *ebiten.Image) *HUD {
	return &HUD{
		heartFull:  heartFull,
		heartEmpty: heartEmpty,
		coinSprite: coinSprite,
	}
}

// Draw renders the HUD on the screen.
func (h *HUD) Draw(screen *ebiten.Image, hp, maxHP, score, coins int) {
	// Draw hearts
	for i := 0; i < maxHP; i++ {
		x := 20 + i*28
		y := 20
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		if i < hp {
			if h.heartFull != nil {
				screen.DrawImage(h.heartFull, op)
			}
		} else {
			if h.heartEmpty != nil {
				screen.DrawImage(h.heartEmpty, op)
			}
		}
	}

	// Draw coins
	if h.coinSprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(20, 52)
		screen.DrawImage(h.coinSprite, op)
	}

	// Coin count
	coinText := fmt.Sprintf("x %d", coins)
	text.Draw(screen, coinText, basicfont.Face7x13, 40, 66, color.White)

	// Score
	scoreText := fmt.Sprintf("Score: %d", score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 20, 90, color.White)
}

// DrawCenteredMessage draws a centered message on screen.
func DrawCenteredMessage(screen *ebiten.Image, msg string, y int, c color.Color) {
	textWidth := len(msg) * 7
	x := (1280 - textWidth) / 2
	text.Draw(screen, msg, basicfont.Face7x13, x, y, c)
}

// DrawMenu draws the main menu screen.
func DrawMenu(screen *ebiten.Image) {
	// Background overlay
	screen.Fill(color.RGBA{0, 0, 0, 180})

	DrawCenteredMessage(screen, "CITY PLATFORMER", 200, color.White)
	DrawCenteredMessage(screen, "Press ENTER to Start", 300, color.RGBA{100, 200, 255, 255})
	DrawCenteredMessage(screen, "A/D or Arrows: Move", 380, color.RGBA{200, 200, 200, 255})
	DrawCenteredMessage(screen, "W/Up/Space: Jump (Double Jump!)", 410, color.RGBA{200, 200, 200, 255})
	DrawCenteredMessage(screen, "ESC: Pause", 440, color.RGBA{200, 200, 200, 255})
}

// DrawPaused draws the pause overlay.
func DrawPaused(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 128})
	DrawCenteredMessage(screen, "PAUSED", 340, color.White)
	DrawCenteredMessage(screen, "Press ESC to Resume", 380, color.RGBA{200, 200, 200, 255})
}

// DrawGameOver draws the game over screen.
func DrawGameOver(screen *ebiten.Image, score int) {
	screen.Fill(color.RGBA{0, 0, 0, 200})
	DrawCenteredMessage(screen, "GAME OVER", 280, color.RGBA{255, 80, 80, 255})
	DrawCenteredMessage(screen, fmt.Sprintf("Final Score: %d", score), 340, color.White)
	DrawCenteredMessage(screen, "Press R to Restart", 400, color.RGBA{100, 200, 255, 255})
}
