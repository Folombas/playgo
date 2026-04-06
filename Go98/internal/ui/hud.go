package ui

import (
	"fmt"
	"image/color"

	"dungeon_crawler/internal/config"
	"dungeon_crawler/internal/entity"
	"dungeon_crawler/internal/helper"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// HUD draws player stats
func DrawHUD(screen *ebiten.Image, player *entity.Player) {
	// Background bar
	helper.DrawRect(screen, 10, 10, 204, 34, color.RGBA{0, 0, 0, 179})

	// HP Bar background
	helper.DrawRect(screen, 15, 15, 194, 24, color.RGBA{51, 51, 51, 255})

	// HP Bar fill
	hpRatio := float64(player.HP) / float64(player.MaxHP)
	if hpRatio < 0 {
		hpRatio = 0
	}

	var hpColor color.Color
	if hpRatio > 0.6 {
		hpColor = color.RGBA{0, 204, 0, 255}
	} else if hpRatio > 0.3 {
		hpColor = color.RGBA{255, 204, 0, 255}
	} else {
		hpColor = color.RGBA{230, 0, 0, 255}
	}
	helper.DrawRect(screen, 15, 15, 194*hpRatio, 24, hpColor)

	// HP Text
	hpText := fmt.Sprintf("HP: %d/%d", player.HP, player.MaxHP)
	text.Draw(screen, hpText, basicfont.Face7x13, 20, 32, color.White)

	// Floor
	floorText := fmt.Sprintf("Floor: %d/%d", player.Floor, config.MaxFloors)
	text.Draw(screen, floorText, basicfont.Face7x13, 220, 27, color.White)

	// Score
	scoreText := fmt.Sprintf("Score: %d", player.Score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 330, 27, color.RGBA{255, 215, 0, 255})

	// Items background
	helper.DrawRect(screen, 10, 50, 150, 70, color.RGBA{0, 0, 0, 179})

	// Coins
	coinText := fmt.Sprintf("Coins: %d", player.Coins)
	text.Draw(screen, coinText, basicfont.Face7x13, 15, 68, color.RGBA{255, 215, 0, 255})

	// Gems
	gemText := fmt.Sprintf("Gems: %d", player.Gems)
	text.Draw(screen, gemText, basicfont.Face7x13, 15, 83, color.RGBA{0, 200, 255, 255})

	// Keys
	keyText := fmt.Sprintf("Keys: %d", player.Keys)
	text.Draw(screen, keyText, basicfont.Face7x13, 15, 98, color.RGBA{255, 255, 0, 255})

	// Potions
	potionText := fmt.Sprintf("Potions: %d", player.Potions)
	text.Draw(screen, potionText, basicfont.Face7x13, 15, 113, color.RGBA{255, 100, 200, 255})

	// Controls hint
	helper.DrawRect(screen, config.ScreenWidth-220, config.ScreenHeight-35, 210, 25, color.RGBA{0, 0, 0, 153})

	controlsText := "WASD:Move J:Attack K:Item ESC:Pause"
	text.Draw(screen, controlsText, basicfont.Face7x13, config.ScreenWidth-215, config.ScreenHeight-18, color.RGBA{200, 200, 200, 255})
}

// DrawMenu draws the main menu screen
func DrawMenu(screen *ebiten.Image, selected int) {
	// Dark overlay
	helper.DrawRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{0, 0, 0, 217})

	// Title
	titleText := "DUNGEON CRAWLER"
	text.Draw(screen, titleText, basicfont.Face7x13, config.ScreenWidth/2-80, 150, color.RGBA{255, 215, 0, 255})

	subtitleText := "Go365 Day 98 - Ebitengine"
	text.Draw(screen, subtitleText, basicfont.Face7x13, config.ScreenWidth/2-75, 170, color.RGBA{200, 200, 200, 255})

	// Menu items
	menuItems := []string{"START GAME", "HOW TO PLAY", "EXIT"}
	for i, item := range menuItems {
		menuY := 250 + i*40
		menuX := config.ScreenWidth/2 - 60

		if i == selected {
			// Highlight
			helper.DrawRect(screen, float64(menuX-10), float64(menuY-15), 150, 25, color.RGBA{76, 76, 128, 204})

			text.Draw(screen, "> "+item, basicfont.Face7x13, menuX, menuY, color.RGBA{255, 255, 100, 255})
		} else {
			text.Draw(screen, "  "+item, basicfont.Face7x13, menuX, menuY, color.RGBA{200, 200, 200, 255})
		}
	}

	// Bottom info
	text.Draw(screen, "W/S or Up/Down: Navigate", basicfont.Face7x13, config.ScreenWidth/2-75, 400, color.RGBA{150, 150, 150, 255})
	text.Draw(screen, "ENTER or J: Select", basicfont.Face7x13, config.ScreenWidth/2-65, 415, color.RGBA{150, 150, 150, 255})
	text.Draw(screen, "A deep dungeon adventure", basicfont.Face7x13, config.ScreenWidth/2-75, 480, color.RGBA{100, 100, 100, 255})
}

// DrawPause draws pause screen
func DrawPause(screen *ebiten.Image) {
	// Dark overlay
	helper.DrawRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{0, 0, 0, 179})

	// Pause text
	text.Draw(screen, "PAUSED", basicfont.Face7x13, config.ScreenWidth/2-35, config.ScreenHeight/2-20, color.White)
	text.Draw(screen, "Press ESC to resume", basicfont.Face7x13, config.ScreenWidth/2-70, config.ScreenHeight/2+10, color.RGBA{200, 200, 200, 255})
}

// DrawGameOver draws game over screen
func DrawGameOver(screen *ebiten.Image, score int, floor int) {
	// Dark overlay
	helper.DrawRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{76, 0, 0, 217})

	// Game Over text
	text.Draw(screen, "GAME OVER", basicfont.Face7x13, config.ScreenWidth/2-50, config.ScreenHeight/2-40, color.RGBA{255, 50, 50, 255})

	// Stats
	scoreText := fmt.Sprintf("Final Score: %d", score)
	text.Draw(screen, scoreText, basicfont.Face7x13, config.ScreenWidth/2-60, config.ScreenHeight/2, color.RGBA{255, 215, 0, 255})

	floorText := fmt.Sprintf("Reached Floor: %d", floor)
	text.Draw(screen, floorText, basicfont.Face7x13, config.ScreenWidth/2-65, config.ScreenHeight/2+20, color.White)

	text.Draw(screen, "Press ENTER to restart", basicfont.Face7x13, config.ScreenWidth/2-75, config.ScreenHeight/2+60, color.RGBA{200, 200, 200, 255})
}

// DrawVictory draws victory screen
func DrawVictory(screen *ebiten.Image, score int) {
	// Gold overlay
	helper.DrawRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{102, 76, 0, 217})

	// Victory text
	text.Draw(screen, "VICTORY!", basicfont.Face7x13, config.ScreenWidth/2-45, config.ScreenHeight/2-40, color.RGBA{255, 215, 0, 255})

	subtitleText := "You conquered all 10 floors!"
	text.Draw(screen, subtitleText, basicfont.Face7x13, config.ScreenWidth/2-85, config.ScreenHeight/2, color.White)

	scoreText := fmt.Sprintf("Final Score: %d", score)
	text.Draw(screen, scoreText, basicfont.Face7x13, config.ScreenWidth/2-60, config.ScreenHeight/2+30, color.RGBA{255, 215, 0, 255})

	text.Draw(screen, "Press ENTER to play again", basicfont.Face7x13, config.ScreenWidth/2-80, config.ScreenHeight/2+70, color.RGBA{200, 200, 200, 255})
}

// DrawNextFloor draws floor transition
func DrawNextFloor(screen *ebiten.Image, floor int) {
	// Dark overlay
	helper.DrawRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{0, 0, 76, 217})

	floorText := fmt.Sprintf("Floor %d", floor)
	text.Draw(screen, floorText, basicfont.Face7x13, config.ScreenWidth/2-40, config.ScreenHeight/2-10, color.RGBA{100, 200, 255, 255})

	subtitleText := "Descending deeper..."
	text.Draw(screen, subtitleText, basicfont.Face7x13, config.ScreenWidth/2-75, config.ScreenHeight/2+15, color.RGBA{200, 200, 200, 255})
}
