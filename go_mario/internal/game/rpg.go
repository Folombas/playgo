package main

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// addExperience добавляет опыт игроку
func (g *Game) addExperience(amount int) {
	g.player.stats.experience += amount
	
	// Level up check
	for g.player.stats.experience >= g.player.stats.maxExp {
		g.player.stats.experience -= g.player.stats.maxExp
		g.player.stats.level++
		g.player.stats.maxExp = int(float64(g.player.stats.maxExp) * 1.5)
		g.player.stats.statPoints += 3
		g.player.maxHealth += 10
		g.player.currentHealth = g.player.maxHealth
		
		// Level up effect
		g.audio.PlayPowerup()
		g.spawnSparkParticles(float32(g.player.x)+g.player.width/2, float32(g.player.y), 30, color.RGBA{255, 215, 0, 255})
		g.spawnFloatingText(float32(g.player.x), float32(g.player.y), "LEVEL UP!", color.RGBA{255, 215, 0, 255})
	}
}

// openChest открывает сундук
func (g *Game) openChest(chestIndex int) {
	if chestIndex < 0 || chestIndex >= len(g.world.chests) {
		return
	}
	
	chest := &g.world.chests[chestIndex]
	if chest.opened {
		return
	}
	
	// Check collision with player
	playerRect := struct{ x, y, w, h float32 }{
		x: float32(g.player.x),
		y: float32(g.player.y),
		w: g.player.width,
		h: g.player.height,
	}
	
	if checkRectCollision(playerRect.x, playerRect.y, playerRect.w, playerRect.h,
		chest.x, chest.y, chest.width, chest.height) {
		
		chest.opened = true
		
		// Give loot to player
		for _, item := range chest.loot {
			g.inventory.AddItem(BlockType(item.id), 1)
			g.spawnFloatingText(chest.x, chest.y-20, item.name, getRarityColor(item.rarity))
		}
		
		// Play sound based on chest type
		switch chest.chestType {
		case WoodenChest:
			g.audio.PlayCollect()
		case IronChest:
			g.audio.PlayCoin()
		case GoldChest:
			g.audio.PlayPowerup()
		case DiamondChest:
			g.audio.PlayStar()
		case AncientChest:
			g.audio.PlayExtraLife()
		}
		
		// Visual effects
		g.spawnSparkParticles(chest.x+chest.width/2, chest.y+chest.height/2, 20, color.RGBA{255, 215, 0, 255})
		
		// Add experience
		expReward := (int(chest.chestType) + 1) * 50
		g.addExperience(expReward)
	}
}

// getRarityColor возвращает цвет редкости
func getRarityColor(rarity ItemRarity) color.RGBA {
	switch rarity {
	case Common:
		return color.RGBA{255, 255, 255, 255}
	case Uncommon:
		return color.RGBA{30, 255, 30, 255}
	case Rare:
		return color.RGBA{30, 30, 255, 255}
	case Epic:
		return color.RGBA{180, 30, 255, 255}
	case Legendary:
		return color.RGBA{255, 180, 30, 255}
	default:
		return color.RGBA{255, 255, 255, 255}
	}
}

// getChestColor возвращает цвет сундука
func getChestColor(chestType ChestType) color.RGBA {
	switch chestType {
	case WoodenChest:
		return color.RGBA{139, 69, 19, 255}
	case IronChest:
		return color.RGBA{128, 128, 128, 255}
	case GoldChest:
		return color.RGBA{255, 215, 0, 255}
	case DiamondChest:
		return color.RGBA{0, 255, 255, 255}
	case AncientChest:
		return color.RGBA{180, 30, 255, 255}
	default:
		return color.RGBA{139, 69, 19, 255}
	}
}

// drawChests отрисовывает сундуки
func (g *Game) drawChests(screen *ebiten.Image) {
	for _, chest := range g.world.chests {
		// Only draw if on screen
		drawX := chest.x - float32(g.camera.x)
		drawY := chest.y - float32(g.camera.y)
		
		if drawX < -50 || drawX > screenWidth+50 || drawY < -50 || drawY > screenHeight+50 {
			continue
		}
		
		chestColor := getChestColor(chest.chestType)
		
		if chest.opened {
			// Opened chest (darker, lid open)
			vector.DrawFilledRect(screen, drawX, drawY+10, chest.width, chest.height-10, chestColor, false)
			// Open lid
			vector.DrawFilledRect(screen, drawX-5, drawY-5, chest.width+10, 10, chestColor, false)
		} else {
			// Closed chest
			vector.DrawFilledRect(screen, drawX, drawY, chest.width, chest.height, chestColor, false)
			// Lid
			vector.DrawFilledRect(screen, drawX-2, drawY-5, chest.width+4, 10, chestColor, false)
			// Lock
			vector.DrawFilledCircle(screen, drawX+chest.width/2, drawY+chest.height/2, 4, color.RGBA{255, 215, 0, 255}, false)
			
			// Glow for rare chests
			if chest.chestType >= GoldChest {
				glowColor := getChestColor(chest.chestType)
				glowColor.A = 50
				vector.DrawFilledCircle(screen, drawX+chest.width/2, drawY+chest.height/2, chest.width, glowColor, false)
			}
		}
	}
}

// drawBiomeIndicator отрисовывает индикатор текущего биома
func (g *Game) drawBiomeIndicator(screen *ebiten.Image) {
	if g.world == nil || len(g.world.biomes) == 0 {
		return
	}
	
	// Find current biome
	playerBlockX := int((g.player.x + g.camera.x) / blockSize)
	currentBiome := ""
	
	for _, biome := range g.world.biomes {
		if playerBlockX >= biome.xStart && playerBlockX < biome.xEnd {
			currentBiome = biome.name
			break
		}
	}
	
	if currentBiome == "" {
		return
	}
	
	// Draw biome name at top center
	textX := screenWidth/2 - len(currentBiome)*6
	vector.DrawFilledRect(screen, float32(textX-10), 55, float32(len(currentBiome)*12+20), 25, color.RGBA{0, 0, 0, 150}, false)
	ebitenutil.DebugPrintAt(screen, currentBiome, textX, 58)
}

// drawPlayerStats отрисовывает характеристики игрока
func (g *Game) drawPlayerStats(screen *ebiten.Image) {
	// Draw health bar
	barX := 10
	barY := screenHeight - 100
	barW := 200
	barH := 20
	
	// Background
	vector.DrawFilledRect(screen, float32(barX), float32(barY), float32(barW), float32(barH), color.RGBA{50, 50, 50, 255}, false)
	
	// Health bar
	healthPercent := float32(g.player.currentHealth) / float32(g.player.maxHealth)
	healthW := int(float32(barW-4) * healthPercent)
	vector.DrawFilledRect(screen, float32(barX+2), float32(barY+2), float32(healthW), float32(barH-4), color.RGBA{255, 50, 50, 255}, false)
	
	// Health text
	healthText := fmt.Sprintf("HP: %d/%d", g.player.currentHealth, g.player.maxHealth)
	ebitenutil.DebugPrintAt(screen, healthText, barX+10, barY+4)
	
	// Experience bar
	expY := barY + 25
	vector.DrawFilledRect(screen, float32(barX), float32(expY), float32(barW), float32(barH), color.RGBA{50, 50, 50, 255}, false)
	
	expPercent := float32(g.player.stats.experience) / float32(g.player.stats.maxExp)
	expW := int(float32(barW-4) * expPercent)
	vector.DrawFilledRect(screen, float32(barX+2), float32(expY+2), float32(expW), float32(barH-4), color.RGBA{50, 100, 255, 255}, false)
	
	// Level text
	levelText := fmt.Sprintf("Lv.%d EXP: %d/%d", g.player.stats.level, g.player.stats.experience, g.player.stats.maxExp)
	ebitenutil.DebugPrintAt(screen, levelText, barX+10, expY+4)
	
	// Stats text
	statsY := barY + 55
	stats := []string{
		fmt.Sprintf("STR: %d | DEF: %d | VIT: %d", g.player.stats.strength, g.player.stats.defense, g.player.stats.vitality),
		fmt.Sprintf("AGI: %d | LCK: %d | Pts: %d", g.player.stats.agility, g.player.stats.luck, g.player.stats.statPoints),
	}
	
	vector.DrawFilledRect(screen, float32(barX), float32(statsY-5), float32(barW), 50, color.RGBA{0, 0, 0, 180}, false)
	for i, line := range stats {
		ebitenutil.DebugPrintAt(screen, line, barX+10, statsY+i*20)
	}
}

// handleStatUpgrade обрабатывает прокачку характеристик
func (g *Game) handleStatUpgrade() {
	if g.player.stats.statPoints <= 0 {
		return
	}
	
	// Press U to open stat screen (simplified: direct upgrade with keys)
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.player.stats.strength++
		g.player.stats.statPoints--
		g.audio.PlayCollect()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.player.stats.defense++
		g.player.stats.statPoints--
		g.audio.PlayCollect()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		g.player.stats.vitality++
		g.player.maxHealth += 5
		g.player.currentHealth += 5
		g.player.stats.statPoints--
		g.audio.PlayCollect()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.player.stats.agility++
		g.player.stats.statPoints--
		g.audio.PlayCollect()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.player.stats.luck++
		g.player.stats.statPoints--
		g.audio.PlayCollect()
	}
}

// drawStatUpgradeHint отрисовывает подсказку по прокачке
func (g *Game) drawStatUpgradeHint(screen *ebiten.Image) {
	if g.player.stats.statPoints <= 0 {
		return
	}
	
	hint := fmt.Sprintf("Stat Points: %d | S:STR D:DEF V:VIT A:AGI L:LCK", g.player.stats.statPoints)
	hintX := screenWidth/2 - len(hint)*6
	vector.DrawFilledRect(screen, float32(hintX-10), 80, float32(len(hint)*12+20), 25, color.RGBA{255, 100, 0, 200}, false)
	ebitenutil.DebugPrintAt(screen, hint, hintX, 83)
}

// calculateDamage рассчитывает урон с учётом статов
func (g *Game) calculateDamage(baseDamage int) int {
	critChance := float32(g.player.stats.luck) * 0.5
	if rand.Float32()*100 < critChance {
		return baseDamage * 2 // Critical hit!
	}
	return baseDamage + g.player.stats.strength
}

// calculateDefense рассчитывает защиту
func (g *Game) calculateDefense(baseDefense int) int {
	return baseDefense + g.player.stats.defense
}

// usePotion использует зелье
func (g *Game) usePotion(item Item) {
	if item.itemType != Potion {
		return
	}
	
	if g.player.currentHealth >= g.player.maxHealth {
		return
	}
	
	g.player.currentHealth += item.health
	if g.player.currentHealth > g.player.maxHealth {
		g.player.currentHealth = g.player.maxHealth
	}
	
	g.audio.PlayCollect()
	g.spawnFloatingText(float32(g.player.x), float32(g.player.y), fmt.Sprintf("+%d HP", item.health), color.RGBA{255, 100, 100, 255})
}
