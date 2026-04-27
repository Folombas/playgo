package render

import (
	"image/color"
	"math"
	mathrand "math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"snake/internal/settings"
	"snake/internal/types"
)

type DrawFunc func(str string, x, y int, clr color.Color)

func Draw(screen *ebiten.Image, g interface {
	GetState() types.GameState
	GetSnake() []types.Vec
	GetDir() types.Vec
	GetFruitX() int
	GetFruitY() int
	GetFruitType() int
	GetBombs() []types.Bomb
	GetIce() types.IceBlock
	GetIceActive() bool
	GetFrozenTimer() float64
	GetScore() int
	GetHealth() int
	GetParticles() []types.Particle
	GetShake() float64
	GetMenuPulse() float64
	GetMenuSelected() int
	GetMenuButtons() []string
	GetButtonFlash() int
	GetAppleImg() *ebiten.Image
	GetStrawberryImg() *ebiten.Image
	GetOrangeImg() *ebiten.Image
	GetBananaImg() *ebiten.Image
	GetPineappleImg() *ebiten.Image
	GetGhostFrames() []*ebiten.Image
	GetGhostFrameIdx() int
	GetGhostActive() bool
	GetGhostX() int
	GetGhostY() int
	GetGhostModeTimer() float64
	GetRoachFrames() []*ebiten.Image
	GetRoachFrameIdx() int
	GetRoachActive() bool
	GetRoachX() int
	GetRoachY() int
	GetVikingFrames() []*ebiten.Image
	GetVikingList() []types.Viking
	GetGifts() []*types.Gift
	GetGiftClosedImgs() []*ebiten.Image
	GetGiftOpenFrames() []*ebiten.Image
	GetCoins() []types.Coin
	GetCoinFrames() []*ebiten.Image
	GetCoinCount() int
	GetKeysCollected() int
	GetCarryingKey() bool
	GetKeyOnField() types.KeyOnField
	GetKeyImg() *ebiten.Image
	GetSettingsVolumeSlider() float64
	GetSettingsLanguageIndex() int
	GetSettingsDifficultyIndex() int
	GetSettingsAnimations() bool
	GetSettingsSelected() int
}, drawText DrawFunc) {

	screen.Fill(color.RGBA{12, 12, 20, 255})

	ox, oy := 0.0, 0.0
	if g.GetShake() > 0.5 {
		ox = (mathrand.Float64()*2 - 1) * g.GetShake()
		oy = (mathrand.Float64()*2 - 1) * g.GetShake()
	}

	for x := 0; x < types.GridW; x++ {
		for y := 0; y < types.GridH; y++ {
			c := color.RGBA{15, 15, 25, 255}
			if (x+y)%2 != 0 {
				c = color.RGBA{18, 18, 30, 255}
			}
			ebitenutil.DrawRect(screen, float64(x*types.TileSize)+ox, float64(y*types.TileSize)+oy, types.TileSize-1, types.TileSize-1, c)
		}
	}

	DrawBombs(screen, g.GetBombs(), g.GetMenuPulse(), ox, oy)
	DrawSnake(screen, g.GetSnake(), g.GetDir(), g.GetFrozenTimer() > 0, g.GetGhostModeTimer() > 0, ox, oy)
	DrawFruit(screen, g.GetFruitX(), g.GetFruitY(), g.GetFruitType(), g.GetAppleImg(), g.GetStrawberryImg(), g.GetOrangeImg(), g.GetBananaImg(), g.GetPineappleImg(), ox, oy)
	DrawIce(screen, g.GetIce(), g.GetIceActive(), ox, oy)
	DrawGhost(screen, g.GetGhostFrames(), g.GetGhostFrameIdx(), g.GetGhostActive(), g.GetGhostX(), g.GetGhostY(), ox, oy)
	DrawRoach(screen, g.GetRoachFrames(), g.GetRoachFrameIdx(), g.GetRoachActive(), g.GetRoachX(), g.GetRoachY(), ox, oy)
	DrawVikings(screen, g.GetVikingFrames(), g.GetVikingList(), ox, oy)
	DrawGifts(screen, g.GetGifts(), g.GetGiftClosedImgs(), g.GetGiftOpenFrames(), ox, oy)
	DrawKeys(screen, g.GetKeyOnField(), g.GetKeyImg(), ox, oy)
	DrawCoins(screen, g.GetCoins(), g.GetCoinFrames(), ox, oy)
	DrawParticles(screen, g.GetParticles(), ox, oy)

	DrawUI(screen, g, drawText)

	switch g.GetState() {
	case types.STATE_MENU:
		DrawMenu(screen, g, drawText)
	case types.STATE_PAUSED:
		DrawPaused(screen, g, drawText)
	case types.STATE_GAMEOVER:
		DrawGameOver(screen, g, drawText)
	case types.STATE_SETTINGS:
		DrawSettings(screen, g, drawText)
	}
}

func DrawBombs(screen *ebiten.Image, bombs []types.Bomb, menuPulse, ox, oy float64) {
	for _, b := range bombs {
		cx := float64(b.X*types.TileSize+types.TileSize/2) + ox
		cy := float64(b.Y*types.TileSize+types.TileSize/2) + oy
		baseRadius := float64(types.TileSize) / 2 * 1.5
		t := b.Timer
		freq := 3.0 + 9.0*(1.0-math.Min(1.0, t/5.0))
		pulse := 1.0 + 0.15*math.Sin(menuPulse*20*freq)
		radius := baseRadius * pulse
		r := uint8(20)
		if t < 2.0 {
			r = uint8(80 + int(175*(1.0-t/2.0)))
		}
		ebitenutil.DrawCircle(screen, cx, cy, radius, color.RGBA{r, 20, 25, 255})
		ebitenutil.DrawCircle(screen, cx-2, cy-2, radius-2, color.RGBA{0, 0, 0, 100})
		ebitenutil.DrawCircle(screen, cx-radius*0.3, cy-radius*0.35, radius*0.25, color.RGBA{255, 255, 255, 180})
		ebitenutil.DrawCircle(screen, cx+radius*0.2, cy+radius*0.2, radius*0.2, color.RGBA{255, 80, 80, 120})
		fuseLen := 20.0 * (t / 5.0)
		fuseStartX := cx + radius*0.7
		fuseStartY := cy - radius*1.1
		fuseEndX := fuseStartX + fuseLen*0.7
		fuseEndY := fuseStartY - fuseLen*0.5
		ebitenutil.DrawLine(screen, fuseStartX, fuseStartY, fuseEndX, fuseEndY, color.RGBA{80, 70, 50, 255})
		fireSize := 3.0 + 2*math.Sin(menuPulse*50)
		ebitenutil.DrawCircle(screen, fuseEndX, fuseEndY, fireSize, color.RGBA{255, 100, 20, 255})
		ebitenutil.DrawCircle(screen, fuseEndX, fuseEndY, fireSize*0.66, color.RGBA{255, 255, 255, 100})
	}
}

func DrawSnake(screen *ebiten.Image, snake []types.Vec, dir types.Vec, frozen, ghostModeActive bool, ox, oy float64) {
	for i, s := range snake {
		x := float64(s.X*types.TileSize) + ox
		y := float64(s.Y*types.TileSize) + oy
		var base color.RGBA
		if frozen {
			base = color.RGBA{80, 180, 255, 255}
		} else if ghostModeActive {
			base = color.RGBA{20, 220, 90, 180}
		} else {
			base = color.RGBA{20, 220, 90, 255}
		}
		if i > 0 {
			shade := uint8(100 + (i*4)%100)
			if frozen {
				base = color.RGBA{80, 180, 255, shade}
			} else if ghostModeActive {
				base = color.RGBA{20, 220, 90, uint8(180)}
			} else {
				base = color.RGBA{15, shade, 70, 255}
			}
		}
		if i == 0 {
			ebitenutil.DrawRect(screen, x-3, y-3, types.TileSize+6, types.TileSize+6, color.RGBA{0, 200, 40, 80})
			if frozen {
				ebitenutil.DrawRect(screen, x-3, y-3, types.TileSize+6, types.TileSize+6, color.RGBA{100, 200, 255, 80})
			}
			if ghostModeActive {
				ebitenutil.DrawRect(screen, x-3, y-3, types.TileSize+6, types.TileSize+6, color.RGBA{255, 255, 255, 80})
			}
		}
		ebitenutil.DrawRect(screen, x, y, types.TileSize-1, types.TileSize-1, base)
		if i == 0 {
			eyex := float64(types.TileSize)/4 - 2
			eyey := float64(types.TileSize)/4 - 2
			ebitenutil.DrawRect(screen, x+eyex, y+eyey, 4, 4, color.White)
			ebitenutil.DrawRect(screen, x+float64(types.TileSize)-eyex-6, y+eyey, 4, 4, color.White)
			ebitenutil.DrawRect(screen, x+eyex+1, y+eyey+1, 2, 2, color.Black)
			ebitenutil.DrawRect(screen, x+float64(types.TileSize)-eyex-5, y+eyey+1, 2, 2, color.Black)
			var tx, ty, w, h float64
			switch dir {
			case types.Vec{1, 0}:
				tx, ty, w, h = types.TileSize-4, types.TileSize/2-2, 6, 4
			case types.Vec{-1, 0}:
				tx, ty, w, h = -2, types.TileSize/2-2, 6, 4
			case types.Vec{0, 1}:
				tx, ty, w, h = types.TileSize/2-2, types.TileSize-4, 4, 6
			case types.Vec{0, -1}:
				tx, ty, w, h = types.TileSize/2-2, -2, 4, 6
			}
			ebitenutil.DrawRect(screen, x+tx+ox, y+ty+oy, w, h, color.RGBA{255, 70, 100, 255})
		}
	}
}

func DrawFruit(screen *ebiten.Image, fruitX, fruitY, fruitType int, apple, strawberry, orange, banana, pineapple *ebiten.Image, ox, oy float64) {
	cx := float64(fruitX*types.TileSize+types.TileSize/2) + ox
	cy := float64(fruitY*types.TileSize+types.TileSize/2) + oy
	var img *ebiten.Image
	switch fruitType {
	case types.FruitApple:
		img = apple
	case types.FruitStrawberry:
		img = strawberry
	case types.FruitOrange:
		img = orange
	case types.FruitBanana:
		img = banana
	case types.FruitPineapple:
		img = pineapple
	}
	if img != nil {
		op := &ebiten.DrawImageOptions{}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		scale := 1.5
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
		op.GeoM.Translate(ox, oy)
		screen.DrawImage(img, op)
	} else {
		ebitenutil.DrawRect(screen, cx-float64(types.TileSize)/2+ox, cy-float64(types.TileSize)/2+oy, float64(types.TileSize), float64(types.TileSize), color.RGBA{200, 100, 50, 255})
	}
}

func DrawIce(screen *ebiten.Image, ice types.IceBlock, iceActive bool, ox, oy float64) {
	if iceActive {
		cx := float64(ice.X*types.TileSize+types.TileSize/2) + ox
		cy := float64(ice.Y*types.TileSize+types.TileSize/2) + oy
		radius := float64(types.TileSize) / 2 * 1.2
		ebitenutil.DrawCircle(screen, cx, cy, radius, color.RGBA{150, 220, 255, 255})
		ebitenutil.DrawCircle(screen, cx-2, cy-2, radius-3, color.RGBA{100, 200, 240, 200})
		ebitenutil.DrawCircle(screen, cx+radius*0.3, cy-radius*0.3, radius*0.25, color.RGBA{255, 255, 255, 200})
	}
}

func DrawGhost(screen *ebiten.Image, frames []*ebiten.Image, frameIdx int, active bool, gx, gy int, ox, oy float64) {
	if active && len(frames) > 0 {
		cx := float64(gx*types.TileSize+types.TileSize/2) + ox
		cy := float64(gy*types.TileSize+types.TileSize/2) + oy
		frame := frames[frameIdx]
		if frame != nil {
			op := &ebiten.DrawImageOptions{}
			w, h := frame.Bounds().Dx(), frame.Bounds().Dy()
			scale := float64(types.TileSize) / float64(w)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
			op.GeoM.Translate(ox, oy)
			screen.DrawImage(frame, op)
		}
	}
}

func DrawRoach(screen *ebiten.Image, frames []*ebiten.Image, frameIdx int, active bool, rx, ry int, ox, oy float64) {
	if active && len(frames) > 0 && settings.Current.BackgroundAnimation {
		cx := float64(rx*types.TileSize+types.TileSize/2) + ox
		cy := float64(ry*types.TileSize+types.TileSize/2) + oy
		frame := frames[frameIdx]
		if frame != nil {
			op := &ebiten.DrawImageOptions{}
			w, h := frame.Bounds().Dx(), frame.Bounds().Dy()
			scale := float64(types.TileSize) / float64(w)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
			op.GeoM.Translate(ox, oy)
			screen.DrawImage(frame, op)
		}
	}
}

func DrawVikings(screen *ebiten.Image, frames []*ebiten.Image, vikings []types.Viking, ox, oy float64) {
	if len(frames) > 0 && settings.Current.BackgroundAnimation {
		for _, v := range vikings {
			cx := float64(v.X*types.TileSize+types.TileSize/2) + ox
			cy := float64(v.Y*types.TileSize+types.TileSize/2) + oy
			frame := frames[v.Frame%len(frames)]
			if frame != nil {
				op := &ebiten.DrawImageOptions{}
				w, h := frame.Bounds().Dx(), frame.Bounds().Dy()
				scale := float64(types.TileSize) / float64(w)
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
				op.GeoM.Translate(ox, oy)
				screen.DrawImage(frame, op)
			}
		}
	}
}

func DrawGifts(screen *ebiten.Image, gifts []*types.Gift, closedImgs, openFrames []*ebiten.Image, ox, oy float64) {
	for _, gift := range gifts {
		cx := float64(gift.X*types.TileSize+types.TileSize/2) + ox
		cy := float64(gift.Y*types.TileSize+types.TileSize/2) + oy
		var img *ebiten.Image
		if gift.Opened {
			if len(openFrames) > 0 {
				img = openFrames[0]
				if img != nil {
					op := &ebiten.DrawImageOptions{}
					w, h := img.Bounds().Dx(), img.Bounds().Dy()
					scale := float64(types.TileSize) / float64(w)
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
					op.GeoM.Translate(ox, oy)
					if gift.Life < 2.0 && gift.Life > 0 {
						alpha := gift.Life / 2.0
						if alpha < 0 {
							alpha = 0
						}
						op.ColorM.Scale(1, 1, 1, alpha)
					}
					screen.DrawImage(img, op)
				}
			}
		} else {
			if len(closedImgs) > 0 {
				if gift.Color >= 0 && gift.Color < len(closedImgs) {
					img = closedImgs[gift.Color]
				} else {
					img = closedImgs[0]
				}
				if img != nil {
					op := &ebiten.DrawImageOptions{}
					w, h := img.Bounds().Dx(), img.Bounds().Dy()
					scale := float64(types.TileSize) / float64(w)
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
					op.GeoM.Translate(ox, oy)
					screen.DrawImage(img, op)
				}
			}
		}
	}
}

func DrawKeys(screen *ebiten.Image, keyOnField types.KeyOnField, keyImg *ebiten.Image, ox, oy float64) {
	if keyOnField.Active && keyImg != nil {
		cx := float64(keyOnField.X*types.TileSize+types.TileSize/2) + ox
		cy := float64(keyOnField.Y*types.TileSize+types.TileSize/2) + oy
		op := &ebiten.DrawImageOptions{}
		w, h := keyImg.Bounds().Dx(), keyImg.Bounds().Dy()
		scale := float64(types.TileSize) / float64(w)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
		op.GeoM.Translate(ox, oy)
		if keyOnField.Life < 2.0 && keyOnField.Life > 0 {
			alpha := keyOnField.Life / 2.0
			if alpha < 0 {
				alpha = 0
			}
			op.ColorM.Scale(1, 1, 1, alpha)
		}
		screen.DrawImage(keyImg, op)
	}
}

func DrawCoins(screen *ebiten.Image, coins []types.Coin, coinFrames []*ebiten.Image, ox, oy float64) {
	for _, c := range coins {
		if len(coinFrames) == 0 {
			continue
		}
		frameIdx := c.Frame % len(coinFrames)
		img := coinFrames[frameIdx]
		if img == nil {
			continue
		}
		cx := float64(c.X*types.TileSize+types.TileSize/2) + ox
		cy := float64(c.Y*types.TileSize+types.TileSize/2) + oy
		op := &ebiten.DrawImageOptions{}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		scale := float64(types.TileSize) / float64(w)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx-float64(w)*scale/2, cy-float64(h)*scale/2)
		op.GeoM.Translate(ox, oy)
		screen.DrawImage(img, op)
	}
}

func DrawParticles(screen *ebiten.Image, particles []types.Particle, ox, oy float64) {
	for _, p := range particles {
		c := p.Color
		if p.Glow {
			ebitenutil.DrawRect(screen, p.X-p.Size*1.5+ox, p.Y-p.Size*1.5+oy, p.Size*3, p.Size*3, color.RGBA{c.R, c.G, c.B, uint8(float64(c.A) * 0.4 * p.Life)})
		}
		ebitenutil.DrawRect(screen, p.X-p.Size+ox, p.Y-p.Size+oy, p.Size*2, p.Size*2, c)
	}
}

func DrawUI(screen *ebiten.Image, g interface {
	GetScore() int
	GetHealth() int
	GetKeysCollected() int
	GetCarryingKey() bool
	GetCoinCount() int
	GetFrozenTimer() float64
	GetGhostModeTimer() float64
}, drawText DrawFunc) {
	if settings.Current.Language == "ru" {
		drawText("Счёт: "+strconv.Itoa(g.GetScore()), 10, 25, color.White)
	} else {
		drawText("Score: "+strconv.Itoa(g.GetScore()), 10, 25, color.White)
	}

	barX := float64(types.ScreenW - 20)
	barW := 150.0
	barH := 14.0
	healthPct := float64(g.GetHealth()) / float64(types.MaxHealth)
	ebitenutil.DrawRect(screen, barX-barW, 10, barW, barH, color.RGBA{30, 30, 40, 200})
	ebitenutil.DrawRect(screen, barX-barW, 10, barW*healthPct, barH, color.RGBA{50, 255, 80, 255})
	if settings.Current.Language == "ru" {
		drawText("ЗДОРОВЬЕ", int(barX-barW+40), 25, color.White)
	} else {
		drawText("HEALTH", int(barX-barW+40), 25, color.White)
	}

	if settings.Current.Language == "ru" {
		drawText("Ключи: "+strconv.Itoa(g.GetKeysCollected()), 10, 55, color.RGBA{255, 215, 0, 255})
		if g.GetCarryingKey() {
			drawText("Ключ активирован!", 10, 80, color.RGBA{255, 200, 100, 255})
		}
		drawText("Монеты: "+strconv.Itoa(g.GetCoinCount()), 10, 105, color.RGBA{255, 215, 0, 255})
	} else {
		drawText("Keys: "+strconv.Itoa(g.GetKeysCollected()), 10, 55, color.RGBA{255, 215, 0, 255})
		if g.GetCarryingKey() {
			drawText("Key activated!", 10, 80, color.RGBA{255, 200, 100, 255})
		}
		drawText("Coins: "+strconv.Itoa(g.GetCoinCount()), 10, 105, color.RGBA{255, 215, 0, 255})
	}

	if settings.Current.Language == "ru" {
		drawText("ESC - меню", types.ScreenW-100, types.ScreenH-20, color.White)
		drawText("P - пауза", types.ScreenW-100, types.ScreenH-40, color.White)
		drawText("K - взять ключ", types.ScreenW-100, types.ScreenH-60, color.White)
	} else {
		drawText("ESC - menu", types.ScreenW-100, types.ScreenH-20, color.White)
		drawText("P - pause", types.ScreenW-100, types.ScreenH-40, color.White)
		drawText("K - get key", types.ScreenW-100, types.ScreenH-60, color.White)
	}

	if g.GetFrozenTimer() > 0 {
		if settings.Current.Language == "ru" {
			drawText("ЗАМОРОЗКА", types.ScreenW/2-60, types.ScreenH-30, color.RGBA{100, 200, 255, 255})
		} else {
			drawText("FROZEN", types.ScreenW/2-60, types.ScreenH-30, color.RGBA{100, 200, 255, 255})
		}
	}
	if g.GetGhostModeTimer() > 0 {
		if settings.Current.Language == "ru" {
			drawText("ПРИЗРАЧНЫЙ РЕЖИМ", types.ScreenW/2-100, types.ScreenH-60, color.RGBA{200, 200, 255, 255})
		} else {
			drawText("GHOST MODE", types.ScreenW/2-100, types.ScreenH-60, color.RGBA{200, 200, 255, 255})
		}
	}
}

func DrawMenu(screen *ebiten.Image, g interface {
	GetMenuButtons() []string
	GetMenuSelected() int
	GetButtonFlash() int
}, drawText DrawFunc) {
	ebitenutil.DrawRect(screen, 0, 0, types.ScreenW, types.ScreenH, color.RGBA{0, 0, 0, 255})
	drawText("S N A K E   R E V I V E D", types.ScreenW/2-180, 150, color.RGBA{255, 200, 100, 255})
	startY := 280
	step := 50

	for i, btn := range g.GetMenuButtons() {
		y := startY + i*step
		if i == g.GetMenuSelected() {
			bg := color.RGBA{100, 100, 150, 255}
			if g.GetButtonFlash() > 0 {
				bg = color.RGBA{200, 200, 255, 255}
			}
			ebitenutil.DrawRect(screen, types.ScreenW/2-150, float64(y)-15, 300, 35, bg)
			drawText(btn, types.ScreenW/2-len(btn)*3, y, color.RGBA{255, 0, 255, 255})
		} else {
			drawText(btn, types.ScreenW/2-len(btn)*3, y, color.White)
		}
	}
	if settings.Current.Language == "ru" {
		drawText("Стрелки вверх/вниз, Enter - выбор", types.ScreenW/2-220, types.ScreenH-70, color.RGBA{200, 200, 200, 255})
	} else {
		drawText("Arrow keys up/down, Enter - select", types.ScreenW/2-240, types.ScreenH-70, color.RGBA{200, 200, 200, 255})
	}
}

func DrawPaused(screen *ebiten.Image, g interface{}, drawText DrawFunc) {
	ebitenutil.DrawRect(screen, 0, 0, types.ScreenW, types.ScreenH, color.RGBA{0, 0, 0, 200})
	if settings.Current.Language == "ru" {
		drawText("ПАУЗА", types.ScreenW/2-40, types.ScreenH/2, color.RGBA{255, 255, 150, 255})
		drawText("Нажмите P для продолжения", types.ScreenW/2-150, types.ScreenH/2+40, color.White)
	} else {
		drawText("PAUSED", types.ScreenW/2-40, types.ScreenH/2, color.RGBA{255, 255, 150, 255})
		drawText("Press P to resume", types.ScreenW/2-150, types.ScreenH/2+40, color.White)
	}
}

func DrawGameOver(screen *ebiten.Image, g interface {
	GetScore() int
}, drawText DrawFunc) {
	ebitenutil.DrawRect(screen, 100, 80, types.ScreenW-200, types.ScreenH-180, color.RGBA{40, 0, 0, 255})
	if settings.Current.Language == "ru" {
		drawText("ИГРА ОКОНЧЕНА", types.ScreenW/2-80, types.ScreenH/2-40, color.RGBA{255, 100, 100, 255})
		drawText("Счёт: "+strconv.Itoa(g.GetScore()), types.ScreenW/2-60, types.ScreenH/2, color.White)
		drawText("Нажмите любую клавишу для меню", types.ScreenW/2-150, types.ScreenH/2+40, color.White)
	} else {
		drawText("GAME OVER", types.ScreenW/2-80, types.ScreenH/2-40, color.RGBA{255, 100, 100, 255})
		drawText("Score: "+strconv.Itoa(g.GetScore()), types.ScreenW/2-60, types.ScreenH/2, color.White)
		drawText("Press any key for menu", types.ScreenW/2-150, types.ScreenH/2+40, color.White)
	}
}

func DrawSettings(screen *ebiten.Image, g interface {
	GetSettingsVolumeSlider() float64
	GetSettingsLanguageIndex() int
	GetSettingsDifficultyIndex() int
	GetSettingsAnimations() bool
	GetSettingsSelected() int
}, drawText DrawFunc) {
	ebitenutil.DrawRect(screen, 0, 0, types.ScreenW, types.ScreenH, color.RGBA{0, 0, 0, 220})
	if settings.Current.Language == "ru" {
		drawText("НАСТРОЙКИ", types.ScreenW/2-100, 100, color.RGBA{255, 255, 150, 255})
	} else {
		drawText("SETTINGS", types.ScreenW/2-80, 100, color.RGBA{255, 255, 150, 255})
	}

	y := 180
	stepY := 60
	sliderX := 500
	sliderW := 300

	if g.GetSettingsSelected() == 0 {
		ebitenutil.DrawRect(screen, 280, float64(y)-10, 350, 30, color.RGBA{80, 80, 120, 255})
	}
	if settings.Current.Language == "ru" {
		drawText("Громкость:", 300, y, color.White)
	} else {
		drawText("Volume:", 300, y, color.White)
	}
	ebitenutil.DrawRect(screen, float64(sliderX), float64(y)-5, float64(sliderW), 10, color.RGBA{100, 100, 100, 255})
	handleX := sliderX + int(g.GetSettingsVolumeSlider()*float64(sliderW))
	ebitenutil.DrawRect(screen, float64(handleX-8), float64(y)-15, 16, 24, color.RGBA{255, 200, 100, 255})
	drawText(strconv.Itoa(int(g.GetSettingsVolumeSlider()*100))+"%", sliderX+sliderW+20, y+5, color.White)
	y += stepY

	if g.GetSettingsSelected() == 1 {
		ebitenutil.DrawRect(screen, 280, float64(y)-10, 350, 30, color.RGBA{80, 80, 120, 255})
	}
	if settings.Current.Language == "ru" {
		drawText("Язык:", 300, y, color.White)
		if g.GetSettingsLanguageIndex() == 0 {
			drawText("Русский", 500, y, color.RGBA{255, 255, 0, 255})
		} else {
			drawText("English", 500, y, color.RGBA{200, 200, 200, 255})
		}
	} else {
		drawText("Language:", 300, y, color.White)
		if g.GetSettingsLanguageIndex() == 0 {
			drawText("Russian", 500, y, color.RGBA{200, 200, 200, 255})
		} else {
			drawText("English", 500, y, color.RGBA{255, 255, 0, 255})
		}
	}
	y += stepY

	if g.GetSettingsSelected() == 2 {
		ebitenutil.DrawRect(screen, 280, float64(y)-10, 350, 30, color.RGBA{80, 80, 120, 255})
	}
	if settings.Current.Language == "ru" {
		drawText("Сложность:", 300, y, color.White)
		switch g.GetSettingsDifficultyIndex() {
		case 0:
			drawText("Лёгкая", 500, y, color.RGBA{100, 200, 100, 255})
		case 1:
			drawText("Средняя", 500, y, color.RGBA{255, 255, 0, 255})
		case 2:
			drawText("Сложная", 500, y, color.RGBA{255, 100, 100, 255})
		}
	} else {
		drawText("Difficulty:", 300, y, color.White)
		switch g.GetSettingsDifficultyIndex() {
		case 0:
			drawText("Easy", 500, y, color.RGBA{100, 200, 100, 255})
		case 1:
			drawText("Normal", 500, y, color.RGBA{255, 255, 0, 255})
		case 2:
			drawText("Hard", 500, y, color.RGBA{255, 100, 100, 255})
		}
	}
	y += stepY

	if g.GetSettingsSelected() == 3 {
		ebitenutil.DrawRect(screen, 280, float64(y)-10, 350, 30, color.RGBA{80, 80, 120, 255})
	}
	if settings.Current.Language == "ru" {
		drawText("Фоновые анимации:", 300, y, color.White)
		if g.GetSettingsAnimations() {
			drawText("Вкл", 600, y, color.RGBA{100, 255, 100, 255})
		} else {
			drawText("Выкл", 600, y, color.RGBA{255, 100, 100, 255})
		}
	} else {
		drawText("BG animations:", 300, y, color.White)
		if g.GetSettingsAnimations() {
			drawText("ON", 600, y, color.RGBA{100, 255, 100, 255})
		} else {
			drawText("OFF", 600, y, color.RGBA{255, 100, 100, 255})
		}
	}

	drawText("ESC - назад", types.ScreenW/2-80, types.ScreenH-50, color.RGBA{200, 200, 200, 255})
}