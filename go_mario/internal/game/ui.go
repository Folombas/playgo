package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawTextWithShadow - отрисовка текста с тенью
func drawTextWithShadow(screen *ebiten.Image, text string, x, y int, c color.RGBA) {
	ebitenutil.DebugPrintAt(screen, text, x+1, y+1)
	ebitenutil.DebugPrintAt(screen, text, x, y)
}

// DrawUI отрисовывает интерфейс
func DrawUI(screen *ebiten.Image, player *Player, frameCount int, totalCoins, collectedCoins int) {
	// Gradient background
	for y := 0; y < 45; y++ {
		alpha := uint8(150 - y*2)
		vector.StrokeLine(screen, 0, float32(y), screenWidth, float32(y), 1, color.RGBA{0, 0, 0, alpha}, false)
	}

	// Score
	scoreText := fmt.Sprintf("🍎 %d", player.score)
	drawTextWithShadow(screen, scoreText, 20, 15, color.RGBA{255, 255, 255, 255})

	// Coins
	coinText := fmt.Sprintf("🪙 %d", player.coins)
	drawTextWithShadow(screen, coinText, 140, 15, color.RGBA{255, 215, 0, 255})

	// Lives
	if player.lives > 0 {
		for i := 0; i < player.lives && i < 5; i++ {
			heartX := 260 + i*25
			vector.DrawFilledCircle(screen, float32(heartX+8), 22, 8, color.RGBA{255, 50, 50, 200}, false)
			vector.DrawFilledCircle(screen, float32(heartX+8), 22, 5, color.RGBA{255, 100, 100, 255}, false)
		}
	}

	// Time of day
	vector.DrawFilledCircle(screen, 420, 22, 12, color.RGBA{255, 255, 100, 255}, false)
	vector.DrawFilledCircle(screen, 420, 22, 8, color.RGBA{255, 255, 200, 255}, false)

	// Invincibility
	if player.invincible > 0 {
		glowIntensity := uint8(100 + int(float64(frameCount)*0.3)*50%100)
		vector.DrawFilledCircle(screen, 500, 22, 15, color.RGBA{255, 255, 200, glowIntensity}, false)
	}

	// Coin progress bar
	if totalCoins > 0 {
		progressX := screenWidth - 120
		progressW := 100
		progressH := 8

		vector.DrawFilledRect(screen, float32(progressX), 16, float32(progressW), float32(progressH), color.RGBA{50, 50, 50, 200}, false)

		if collectedCoins > 0 {
			filledW := int(float32(progressW) * float32(collectedCoins) / float32(totalCoins))
			vector.DrawFilledRect(screen, float32(progressX), 16, float32(filledW), float32(progressH), color.RGBA{255, 215, 0, 255}, false)
		}

		vector.StrokeRect(screen, float32(progressX), 16, float32(progressW), float32(progressH), 1, color.RGBA{150, 150, 150, 255}, false)
	}
}

// DrawInventory отрисовывает инвентарь
func DrawInventory(screen *ebiten.Image, inventory *Inventory, selected int, state GameState) {
	barY := screenHeight - 55
	barHeight := 45

	// Background
	for y := 0; y < barHeight; y++ {
		alpha := uint8(180 - y*2)
		vector.StrokeLine(screen, 10, float32(barY+y), float32(inventorySize*44+16), float32(barY+y), 1, color.RGBA{20, 20, 30, alpha}, false)
	}

	// Slots
	for i := 0; i < inventorySize; i++ {
		slotX := 18 + i*44
		slotY := barY + 4

		if i == selected {
			vector.DrawFilledRect(screen, float32(slotX-2), float32(slotY-2), 44, 44, color.RGBA{255, 215, 0, 100}, false)
			vector.StrokeRect(screen, float32(slotX), float32(slotY), 40, 40, 2, color.RGBA{255, 215, 0, 255}, false)
		} else {
			vector.DrawFilledRect(screen, float32(slotX), float32(slotY), 40, 40, color.RGBA{60, 60, 70, 200}, false)
			vector.StrokeRect(screen, float32(slotX), float32(slotY), 40, 40, 1, color.RGBA{100, 100, 110, 255}, false)
		}

		slot := inventory.slots[i]
		if slot.count > 0 {
			itemColor := getBlockColor(slot.item)
			vector.DrawFilledRect(screen, float32(slotX+6), float32(slotY+6), 28, 28, itemColor, false)
			vector.StrokeLine(screen, float32(slotX+6), float32(slotY+6), float32(slotX+34), float32(slotY+6), 2, color.RGBA{255, 255, 255, 150}, false)
			vector.StrokeLine(screen, float32(slotX+6), float32(slotY+6), float32(slotX+6), float32(slotY+34), 2, color.RGBA{255, 255, 255, 150}, false)
			vector.StrokeLine(screen, float32(slotX+6), float32(slotY+34), float32(slotX+34), float32(slotY+34), 2, color.RGBA{0, 0, 0, 150}, false)
			vector.StrokeLine(screen, float32(slotX+34), float32(slotY+6), float32(slotX+34), float32(slotY+34), 2, color.RGBA{0, 0, 0, 150}, false)

			countText := fmt.Sprintf("%d", slot.count)
			drawTextWithShadow(screen, countText, slotX+28, slotY+28, color.RGBA{255, 255, 255, 255})
		}

		slotNum := fmt.Sprintf("%d", i+1)
		drawTextWithShadow(screen, slotNum, slotX+2, slotY+2, color.RGBA{150, 150, 150, 255})
	}

	if state == Crafting {
		drawTextWithShadow(screen, "1-5: Craft | ESC: Exit", 20, barY-20, color.RGBA{255, 215, 0, 255})
	}
}

// DrawTutorial отрисовывает подсказки туториала
func DrawTutorial(screen *ebiten.Image, tutorial *Tutorial) {
	if tutorial == nil || !tutorial.visible {
		return
	}

	if tutorial.currentStep < len(tutorial.steps) {
		step := tutorial.steps[tutorial.currentStep]
		if !step.completed {
			hintText := fmt.Sprintf("💡 %s", step.title)
			hintX := screenWidth/2 - len(hintText)*6
			hintY := screenHeight - 85

			vector.DrawFilledRect(screen, float32(hintX-5), float32(hintY-5), float32(len(hintText)*12+10), 20, color.RGBA{0, 0, 0, 100}, false)
			drawTextWithShadow(screen, hintText, hintX, hintY, color.RGBA{255, 255, 100, 255})
		}
	}
}

// DrawQuests отрисовывает прогресс квестов
func DrawQuests(screen *ebiten.Image, quests []Quest) {
	completed := 0
	total := len(quests)

	for _, quest := range quests {
		if quest.completed {
			completed++
		}
	}

	if total > 0 {
		questText := fmt.Sprintf("📜 %d/%d", completed, total)
		drawTextWithShadow(screen, questText, screenWidth-80, 15, color.RGBA{255, 100, 100, 255})
	}
}

// DrawPaused отрисовывает меню паузы
func DrawPaused(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 180}, false)

	title := "⏸️ PAUSED"
	titleX := screenWidth/2 - len(title)*8
	titleY := screenHeight/2 - 60
	drawTextWithShadow(screen, title, titleX, titleY, color.RGBA{255, 255, 255, 255})

	options := []string{"P / ESC - Resume", "S - Settings", "Enter - Menu"}
	for i, opt := range options {
		optX := screenWidth/2 - len(opt)*6
		optY := titleY + 50 + i*30
		drawTextWithShadow(screen, opt, optX, optY, color.RGBA{200, 200, 200, 255})
	}
}

// DrawSettings отрисовывает настройки
func DrawSettings(screen *ebiten.Image, audioEnabled, musicEnabled bool) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 200}, false)

	title := "⚙️ SETTINGS"
	titleX := screenWidth/2 - len(title)*8
	titleY := screenHeight/2 - 100
	drawTextWithShadow(screen, title, titleX, titleY, color.RGBA{255, 255, 255, 255})

	soundStatus := "ON"
	soundColor := color.RGBA{0, 255, 0, 255}
	if !audioEnabled {
		soundStatus = "OFF"
		soundColor = color.RGBA{255, 100, 100, 255}
	}
	soundText := fmt.Sprintf("Sound: %s (M)", soundStatus)
	drawTextWithShadow(screen, soundText, screenWidth/2-len(soundText)*6, titleY+60, soundColor)

	musicStatus := "ON"
	musicColor := color.RGBA{0, 255, 0, 255}
	if !musicEnabled {
		musicStatus = "OFF"
		musicColor = color.RGBA{255, 100, 100, 255}
	}
	musicText := fmt.Sprintf("Music: %s (N)", musicStatus)
	drawTextWithShadow(screen, musicText, screenWidth/2-len(musicText)*6, titleY+100, musicColor)

	drawTextWithShadow(screen, "ESC - Back", screenWidth/2-40, titleY+160, color.RGBA{200, 200, 200, 255})
}

// DrawAchievementAlbum отрисовывает альбом достижений
func DrawAchievementAlbum(screen *ebiten.Image, album *AchievementAlbum) {
	if album == nil || !album.showAlbum {
		return
	}

	panelX := 50
	panelY := 50
	panelW := screenWidth - 100
	panelH := screenHeight - 100

	vector.DrawFilledRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), color.RGBA{0, 0, 0, 200}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), 2, color.RGBA{255, 215, 0, 255}, false)

	title := "🏆 ACHIEVEMENTS"
	drawTextWithShadow(screen, title, panelX+20, panelY+20, color.RGBA{255, 215, 0, 255})

	subtitle := fmt.Sprintf("Unlocked: %d/%d", album.totalUnlocked, len(album.achievements))
	drawTextWithShadow(screen, subtitle, panelX+20, panelY+45, color.RGBA{255, 255, 255, 255})

	for i, ach := range album.achievements {
		y := panelY + 80 + i*35
		marker := "⬜"
		color := color.RGBA{150, 150, 150, 255}

		if ach.completed {
			marker = "✅"
			color = GetMedalColor(ach.medalTier)
		}

		text := fmt.Sprintf("%s %s [%s]", marker, ach.title, GetTierName(ach.medalTier))
		drawTextWithShadow(screen, text, panelX+20, y, color)
	}
}
