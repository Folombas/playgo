package main

import (
	"fmt"
	"image/color"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu: g.drawMenu(screen)
	case StatePlaying, StateBossFight: g.drawPlaying(screen)
	case StateShop: g.drawShop(screen)
	case StatePaused: g.drawPaused(screen)
	case StateLevelComplete: g.drawLevelComplete(screen)
	case StateGameOver: g.drawGameOver(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	for y := 0; y < ScreenHeight; y++ { s := uint8(y % 3); screen.Fill(color.RGBA{10 + s, 10 + s, 30 + s, 255}) }
	for i := 0; i < 100; i++ { x := float32((i*73 + g.frameCount/2) % ScreenWidth); y := float32((i*91 + g.frameCount) % ScreenHeight); vector.DrawFilledCircle(screen, x, y, 1, color.RGBA{255, 255, 255, 255}, true) }
	title := "🚀 SPACE WARRIOR"
	ebitenutil.DebugPrintAt(screen, title, ScreenWidth/2-180, 150)
	ebitenutil.DebugPrintAt(screen, "Эпический космический шутер", ScreenWidth/2-120, 220)
	opts := []string{"[ENTER] Начать игру", "[S] Магазин", "[ESC] Меню"}
	for i, o := range opts { ebitenutil.DebugPrintAt(screen, o, ScreenWidth/2-100, 350+i*50) }
	ctrls := []string{"WASD - Движение", "SPACE - Стрельба", "P - Магазин", "ESC - Пауза"}
	for i, c := range ctrls { ebitenutil.DebugPrintAt(screen, c, ScreenWidth/2-100, 520+i*28) }
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Убийств: %d | Волн: %d", g.player.enemiesKilled, g.wave-1), 20, ScreenHeight-40)
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	screen.Fill(color.RGBA{5, 5, 20, 255})
	for i := 0; i < 50; i++ { x := float32((i*73 + g.frameCount/(i%5+1)) % ScreenWidth); y := float32((i*91 + g.frameCount/(i%3+1)) % ScreenHeight); vector.DrawFilledCircle(screen, x, y, 1, color.RGBA{255, 255, 255, uint8(100 + i%155)}, true) }
	g.drawItems(screen); g.drawPlayer(screen); g.drawEnemies(screen)
	if g.boss != nil { g.drawBoss(screen) }
	g.drawProjectiles(screen); g.drawParticles(screen); g.drawHUD(screen); g.drawAchievements(screen)
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	p := g.player
	if p.invincible && g.frameCount%4 < 2 { return }
	x, y := p.x, p.y
	vector.StrokeLine(screen, float32(x+p.width/2), float32(y), float32(x), float32(y+p.height), 3, color.RGBA{0, 150, 255, 255}, false)
	vector.StrokeLine(screen, float32(x), float32(y+p.height), float32(x+p.width), float32(y+p.height), 3, color.RGBA{0, 150, 255, 255}, false)
	vector.StrokeLine(screen, float32(x+p.width), float32(y+p.height), float32(x+p.width/2), float32(y), 3, color.RGBA{0, 150, 255, 255}, false)
	for i := int(y); i < int(y+p.height); i++ { w := p.width * float64(i-int(y)) / p.height; vector.DrawFilledRect(screen, float32(x+p.width/2-w/2), float32(i), float32(w), 1, color.RGBA{0, 150, 255, 200}, false) }
	vector.DrawFilledCircle(screen, float32(x+p.width/2), float32(y+20), 8, color.RGBA{100, 200, 255, 255}, false)
	vector.DrawFilledRect(screen, float32(x+15), float32(y+p.height), 10, float32(10+g.frameCount%5), color.RGBA{255, 100, 0, 255}, false)
	if p.shield > 0 { vector.StrokeCircle(screen, float32(x+p.width/2), float32(y+p.height/2), float32(p.width), 2, color.RGBA{0, 100, 255, 150}, false) }
}

func (g *Game) drawEnemies(screen *ebiten.Image) {
	for _, e := range g.enemies {
		var c color.RGBA
		switch e.enemyType { case EnemyDrone: c = color.RGBA{200, 50, 50, 255}; case EnemyFighter: c = color.RGBA{200, 100, 50, 255}; case EnemyCruiser: c = color.RGBA{150, 50, 150, 255}; case EnemyCarrier: c = color.RGBA{100, 50, 200, 255} }
		vector.DrawFilledRect(screen, float32(e.x), float32(e.y), float32(e.width), float32(e.height), c, false)
		vector.DrawFilledRect(screen, float32(e.x+10), float32(e.y+10), 20, 20, color.RGBA{255, 200, 0, 255}, false)
		vector.DrawFilledRect(screen, float32(e.x), float32(e.y-8), float32(e.width), 4, color.RGBA{50, 50, 50, 255}, false)
		vector.DrawFilledRect(screen, float32(e.x), float32(e.y-8), float32(e.width)*float32(e.health/e.maxHealth), 4, color.RGBA{0, 255, 0, 255}, false)
	}
}

func (g *Game) drawBoss(screen *ebiten.Image) {
	b := g.boss
	vector.DrawFilledRect(screen, float32(b.x), float32(b.y), float32(b.width), float32(b.height), color.RGBA{150, 0, 50, 255}, false)
	vector.DrawFilledRect(screen, float32(b.x+20), float32(b.y+20), float32(b.width-40), float32(b.height-40), color.RGBA{200, 50, 100, 255}, false)
	vector.DrawFilledCircle(screen, float32(b.x+40), float32(b.y+40), 15, color.RGBA{255, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, float32(b.x+b.width-40), float32(b.y+40), 15, color.RGBA{255, 0, 0, 255}, false)
	vector.DrawFilledRect(screen, 50, 50, float32(ScreenWidth-100), 20, color.RGBA{50, 0, 0, 255}, false)
	vector.DrawFilledRect(screen, 50, 50, float32(ScreenWidth-100)*float32(b.health/b.maxHealth), 20, color.RGBA{255, 0, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, "⚠️ БОСС", 50, 35)
}

func (g *Game) drawProjectiles(screen *ebiten.Image) {
	for _, p := range g.projectiles { vector.DrawFilledRect(screen, float32(p.x), float32(p.y), float32(p.width), float32(p.height), p.color, false) }
}

func (g *Game) drawItems(screen *ebiten.Image) {
	for _, i := range g.items {
		x, y, o := float32(i.x), float32(i.y), float32(i.animFrame)
		switch i.itemType {
		case ItemHealth: vector.DrawFilledCircle(screen, x, y+o*0.1, 12, color.RGBA{0, 255, 0, 255}, false)
		case ItemAmmo: vector.DrawFilledRect(screen, x-8, y-10, 16, 20, color.RGBA{255, 165, 0, 255}, false)
		case ItemShield: vector.DrawFilledCircle(screen, x, y, 12, color.RGBA{0, 100, 255, 200}, false)
		case ItemGem: vector.DrawFilledRect(screen, x-8, y-4, 16, 8, i.color, false)
		}
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles { a := uint8(255 * p.life / p.maxLife); vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), float32(p.size), color.RGBA{p.color.R, p.color.G, p.color.B, a}, true) }
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	p := g.player
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 40, color.RGBA{0, 0, 0, 180}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %.0f/%.0f", p.health, p.maxHealth), 20, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SHIELD: %.0f", p.shield), 200, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("AMMO: %d/%d", p.ammo, p.maxAmmo), 400, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("CREDITS: %d", p.credits), 600, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("GEMS: %d", p.gems), 800, 10)
	vector.DrawFilledRect(screen, 0, ScreenHeight-40, ScreenWidth, 40, color.RGBA{0, 0, 0, 180}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SCORE: %06d | WAVE: %d | LEVEL: %d", g.score, g.wave, p.level), 20, ScreenHeight-30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🔫 %s Lv.%d", p.weapon.name, p.weapon.level), 20, ScreenHeight-70)
}

func (g *Game) drawAchievements(screen *ebiten.Image) {
	if len(g.newAchievements) == 0 { return }
	y := 150
	for _, a := range g.newAchievements {
		vector.DrawFilledRect(screen, 20, float32(y-30), float32(ScreenWidth-40), 70, color.RGBA{0, 0, 0, 200}, false)
		vector.DrawFilledRect(screen, 20, float32(y-30), float32(ScreenWidth-40), 3, color.RGBA{255, 215, 0, 255}, false)
		ebitenutil.DebugPrintAt(screen, "🏆 "+a.name, 40, y)
		y += 80
	}
}

func (g *Game) drawShop(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})
	ebitenutil.DebugPrintAt(screen, "🛒 МАГАЗИН", ScreenWidth/2-100, 50)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Кредиты: %d", g.player.credits), 20, 100)
	y := 160
	for id, w := range weapons {
		yy := y + (id-1)*100
		bg := color.RGBA{30, 30, 60, 255}
		if id == g.selectedWeapon { bg = color.RGBA{50, 50, 100, 255} }
		vector.DrawFilledRect(screen, 50, float32(yy), float32(ScreenWidth-100), 80, bg, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s (Ур.%d/%d)", w.name, w.level, w.maxLevel), 80, yy+10)
		var price string
		if !w.unlocked { price = fmt.Sprintf("Купить: %d", w.upgradeCost*3) } else if w.level < w.maxLevel { price = fmt.Sprintf("Улучшить: %d", w.upgradeCost*w.level) } else { price = "МАКСИМУМ" }
		ebitenutil.DebugPrintAt(screen, price, 80, yy+50)
	}
	ebitenutil.DebugPrintAt(screen, "[H] Лечение +30 HP - 50 кредитов", 80, y+len(weapons)*100+20)
	ebitenutil.DebugPrintAt(screen, "← → Выбор | ENTER Купить | ESC Выход", ScreenWidth/2-150, ScreenHeight-50)
}

func (g *Game) drawPaused(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)
	ebitenutil.DebugPrintAt(screen, "ПАУЗА", ScreenWidth/2-40, ScreenHeight/2-20)
	ebitenutil.DebugPrintAt(screen, "ESC - Продолжить", ScreenWidth/2-70, ScreenHeight/2+20)
}

func (g *Game) drawLevelComplete(screen *ebiten.Image) {
	for y := 0; y < ScreenHeight; y++ { r := uint8(50 + y/10); g := uint8(30 + y/15); b := uint8(100 + y/8); screen.Fill(color.RGBA{r, g, b, 255}) }
	ebitenutil.DebugPrintAt(screen, "🎉 УРОВЕНЬ ПРОЙДЕН!", ScreenWidth/2-120, ScreenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Счёт: %06d | Волна: %d", g.score, g.wave), ScreenWidth/2-100, ScreenHeight/2)
	ebitenutil.DebugPrintAt(screen, "ENTER - Продолжить", ScreenWidth/2-80, ScreenHeight/2+50)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 0, 0, 255})
	ebitenutil.DebugPrintAt(screen, "💀 ИГРА ОКОНЧЕНА", ScreenWidth/2-100, ScreenHeight/2-60)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Счёт: %06d | Волны: %d | Убийств: %d", g.score, g.wave, g.player.enemiesKilled), ScreenWidth/2-120, ScreenHeight/2)
	ebitenutil.DebugPrintAt(screen, "ENTER - В меню", ScreenWidth/2-60, ScreenHeight/2+60)
}
