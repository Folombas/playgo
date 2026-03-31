// Package render - отрисовка для Sunny Adventure
// Go365 Day 91 - Доброе сказочное приключение
package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"sunny_adventure/internal/entity"
	"sunny_adventure/internal/level"
	"sunny_adventure/internal/sprite"
)

// Particle - частица
type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    float64
	MaxLife float64
	Color   color.Color
	Size    float64
}

// Renderer - рендерер игры
type Renderer struct {
	spriteSheet *sprite.SpriteSheet
}

// NewRenderer создаёт рендерер
func NewRenderer(ss *sprite.SpriteSheet) *Renderer {
	return &Renderer{
		spriteSheet: ss,
	}
}

// DrawBackground отрисовывает фон
func (r *Renderer) DrawBackground(screen *ebiten.Image, cameraX, cameraY float64, levelNum int) {
	// Градиентное небо: от голубого к розоватому
	for y := 0; y < 720; y++ {
		r := uint8(135 + int(float64(y)/720*50))
		g := uint8(206 - int(float64(y)/720*50))
		b := uint8(235 - int(float64(y)/720*30))
		ebitenutil.DrawRect(screen, 0, float64(y), 1280, 1, color.RGBA{r, g, b, 255})
	}

	// Рисуем солнышко на фоне
	sunX := 100 - cameraX*0.2
	sunY := 80 - cameraY*0.1
	r.drawSunOnBackground(screen, sunX, sunY)

	// Рисуем облака на фоне
	for i := 0; i < 5; i++ {
		cloudX := float64(i*280+100) - cameraX*0.3
		cloudY := float64(50+i*30) - cameraY*0.15
		r.drawCloudOnBackground(screen, cloudX, cloudY, (i%3)+1)
	}

	// Дальние горы/холмы (параллакс)
	r.drawHills(screen, cameraX*0.4)
}

// drawSunOnBackground рисует солнышко на фоне
func (r *Renderer) drawSunOnBackground(screen *ebiten.Image, x, y float64) {
	// Жёлтое светящееся солнышко
	ebitenutil.DrawCircle(screen, x, y, 40, color.RGBA{255, 255, 100, 255})
	ebitenutil.DrawCircle(screen, x, y, 35, color.RGBA{255, 255, 0, 255})
	ebitenutil.DrawCircle(screen, x, y, 25, color.RGBA{255, 200, 0, 255})

	// Лучики солнышка
	for i := 0; i < 8; i++ {
		angle := float64(i)*math.Pi/4 + float64(int(x*100)%360)*math.Pi/180
		rayX := x + math.Cos(angle)*50
		rayY := y + math.Sin(angle)*50
		ebitenutil.DrawCircle(screen, rayX, rayY, 8, color.RGBA{255, 255, 50, 200})
	}
}

// drawCloudOnBackground рисует облачко на фоне
func (r *Renderer) drawCloudOnBackground(screen *ebiten.Image, x, y float64, cloudNum int) {
	// Белые пушистые облака
	c := color.RGBA{255, 255, 255, 200}
	size := 20.0 + float64(cloudNum)*8

	ebitenutil.DrawCircle(screen, x, y, size, c)
	ebitenutil.DrawCircle(screen, x+size*0.8, y+5, size*0.7, c)
	ebitenutil.DrawCircle(screen, x-size*0.8, y+5, size*0.7, c)
	ebitenutil.DrawCircle(screen, x, y+size*0.5, size*0.9, c)
}

// drawHills рисует холмы на заднем плане
func (r *Renderer) drawHills(screen *ebiten.Image, offsetX float64) {
	// Зелёные холмы
	for i := 0; i < 6; i++ {
		hillX := float64(i*300) - offsetX
		hillY := 500.0 + float64(i%3)*30

		// Рисуем холм как серию кругов
		for j := 0; j < 5; j++ {
			x := hillX + float64(j)*60
			y := hillY - math.Abs(float64(j-2))*20
			radius := 80.0 - math.Abs(float64(j-2))*10
			ebitenutil.DrawCircle(screen, x, y, radius, color.RGBA{34, 139, 34, 200})
		}
	}
}

// DrawPlatform отрисовывает платформу
func (r *Renderer) DrawPlatform(screen *ebiten.Image, platform *level.Platform, cameraX, cameraY float64) {
	x := platform.X - cameraX
	y := platform.Y - cameraY
	ts := platform.Height // Размер тайла

	// Рисуем платформу из спрайтов тайлами
	switch platform.Type {
	case level.TileGround, level.TileGrass:
		// Земля с травой - рисуем тайлами
		for tx := x; tx < x+platform.Width; tx += ts {
			// Верхний ряд - трава
			if grassSprite := r.spriteSheet.GetTileSprite("grassHalf"); grassSprite != nil {
				screen.DrawImage(grassSprite, &ebiten.DrawImageOptions{
					GeoM: func() ebiten.GeoM {
						var g ebiten.GeoM
						g.Translate(tx, y)
						return g
					}(),
				})
			}
			// Нижние ряды - земля
			for ty := y + ts; ty < y+platform.Height; ty += ts {
				if dirtSprite := r.spriteSheet.GetTileSprite("dirt"); dirtSprite != nil {
					screen.DrawImage(dirtSprite, &ebiten.DrawImageOptions{
						GeoM: func() ebiten.GeoM {
							var g ebiten.GeoM
							g.Translate(tx, ty)
							return g
						}(),
					})
				}
			}
		}

	case level.TileBrick, level.TileCastle:
		// Замок/кирпич - рисуем тайлами
		for tx := x; tx < x+platform.Width; tx += ts {
			for ty := y; ty < y+platform.Height; ty += ts {
				if brickSprite := r.spriteSheet.GetTileSprite("brickWall"); brickSprite != nil {
					screen.DrawImage(brickSprite, &ebiten.DrawImageOptions{
						GeoM: func() ebiten.GeoM {
							var g ebiten.GeoM
							g.Translate(tx, ty)
							return g
						}(),
					})
				}
			}
		}

	case level.TileBox:
		// Коробки - рисуем тайлами
		for tx := x; tx < x+platform.Width; tx += ts {
			for ty := y; ty < y+platform.Height; ty += ts {
				if boxSprite := r.spriteSheet.GetTileSprite("box"); boxSprite != nil {
					screen.DrawImage(boxSprite, &ebiten.DrawImageOptions{
						GeoM: func() ebiten.GeoM {
							var g ebiten.GeoM
							g.Translate(tx, ty)
							return g
						}(),
					})
				}
			}
		}

	case level.TileCandy:
		// Конфетная платформа
		for tx := x; tx < x+platform.Width; tx += ts {
			for ty := y; ty < y+platform.Height; ty += ts {
				if candySprite := r.spriteSheet.GetTileSprite("candy"); candySprite != nil {
					screen.DrawImage(candySprite, &ebiten.DrawImageOptions{
						GeoM: func() ebiten.GeoM {
							var g ebiten.GeoM
							g.Translate(tx, ty)
							return g
						}(),
					})
				}
			}
		}

	case level.TileIce:
		// Ледяная платформа
		for tx := x; tx < x+platform.Width; tx += ts {
			for ty := y; ty < y+platform.Height; ty += ts {
				if iceSprite := r.spriteSheet.GetTileSprite("iceHalf"); iceSprite != nil {
					screen.DrawImage(iceSprite, &ebiten.DrawImageOptions{
						GeoM: func() ebiten.GeoM {
							var g ebiten.GeoM
							g.Translate(tx, ty)
							return g
						}(),
					})
				}
			}
		}

	case level.TileLadder:
		// Лестница
		for ty := y; ty < y+platform.Height; ty += ts {
			if ladderSprite := r.spriteSheet.GetTileSprite("ladder_mid"); ladderSprite != nil {
				screen.DrawImage(ladderSprite, &ebiten.DrawImageOptions{
					GeoM: func() ebiten.GeoM {
						var g ebiten.GeoM
						g.Translate(x, ty)
						return g
					}(),
				})
			}
		}
	}
}

// DrawPlayer отрисовывает игрока
func (r *Renderer) DrawPlayer(screen *ebiten.Image, player *entity.Player, cameraX, cameraY float64) {
	if player.Health.Invincible > 0 && int(player.Health.Invincible*10)%2 == 0 {
		return // Мигание при неуязвимости
	}

	// Отрисовка спрайта через SpriteRenderer
	if player.Renderer != nil && player.Renderer.CurrentImg != nil {
		player.Renderer.Draw(screen, player.Transform, cameraX, cameraY)
	}
}

// DrawFriend отрисовывает друга
func (r *Renderer) DrawFriend(screen *ebiten.Image, friend *entity.Friend, cameraX, cameraY float64) {
	if friend.Collected {
		return
	}

	if friend.Renderer.CurrentImg != nil {
		friend.Renderer.Draw(screen, friend.Transform, cameraX, cameraY)
	} else {
		// Заглушка - милый круглый друг
		x := friend.Transform.X - cameraX + friend.Transform.Width/2
		y := friend.Transform.Y - cameraY + friend.Transform.Height/2

		c := color.RGBA{255, 150, 150, 255}
		switch friend.FriendType {
		case entity.FriendBee:
			c = color.RGBA{255, 255, 0, 255} // Жёлтая пчёлка
		case entity.FriendLadybug:
			c = color.RGBA{255, 50, 50, 255} // Красная божья коровка
		case entity.FriendFrog:
			c = color.RGBA{50, 255, 50, 255} // Зелёный лягушонок
		case entity.FriendSnail:
			c = color.RGBA{139, 90, 43, 255} // Коричневая улитка
		case entity.FriendGhost:
			c = color.RGBA{200, 200, 255, 200} // Голубой призрачок
		}

		ebitenutil.DrawCircle(screen, x, y, 12, c)
		// Глазки
		ebitenutil.DrawCircle(screen, x-3, y-2, 2, color.RGBA{0, 0, 0, 255})
		ebitenutil.DrawCircle(screen, x+3, y-2, 2, color.RGBA{0, 0, 0, 255})
	}
}

// DrawEnemy отрисовывает врага
func (r *Renderer) DrawEnemy(screen *ebiten.Image, enemy *entity.Enemy, cameraX, cameraY float64) {
	if enemy.Renderer.CurrentImg != nil {
		enemy.Renderer.Draw(screen, enemy.Transform, cameraX, cameraY)
	} else {
		x := enemy.Transform.X - cameraX
		y := enemy.Transform.Y - cameraY - enemy.Transform.Height

		c := color.RGBA{150, 150, 150, 255}
		if enemy.Converted {
			// Превращённый враг - розовый и добрый!
			c = color.RGBA{255, 150, 200, 255}
		} else {
			switch enemy.EnemyType {
			case entity.EnemyWind:
				c = color.RGBA{200, 200, 255, 255} // Голубой ветерок
			case entity.EnemyStorm:
				c = color.RGBA{100, 100, 150, 255} // Тёмная тучка
			case entity.EnemyBat:
				c = color.RGBA{100, 50, 100, 255} // Фиолетовая мышь
			case entity.EnemySpider:
				c = color.RGBA{50, 50, 50, 255} // Чёрный паук
			case entity.EnemySnake:
				c = color.RGBA{50, 150, 50, 255} // Зелёная змейка
			}
		}

		ebitenutil.DrawRect(screen, x, y, enemy.Transform.Width, enemy.Transform.Height, c)

		// Глаза
		eyeX := x + enemy.Transform.Width/2
		if enemy.Transform.Facing == 1 {
			eyeX += 8
		} else {
			eyeX -= 8
		}
		ebitenutil.DrawCircle(screen, eyeX, y+12, 4, color.RGBA{255, 255, 255, 255})
	}
}

// DrawItem отрисовывает предмет
func (r *Renderer) DrawItem(screen *ebiten.Image, item *entity.Item, cameraX, cameraY float64) {
	if item.Collected {
		return
	}

	if item.Renderer.CurrentImg != nil {
		item.Renderer.Draw(screen, item.Transform, cameraX, cameraY)
	} else {
		x := item.Transform.X - cameraX + item.Transform.Width/2
		y := item.Transform.Y - cameraY + item.FloatOffset

		c := color.RGBA{255, 215, 0, 255}
		if item.ItemType == "gemRed" {
			c = color.RGBA{255, 50, 50, 255}
		} else if item.ItemType == "gemBlue" {
			c = color.RGBA{50, 100, 255, 255}
		} else if item.ItemType == "star" {
			c = color.RGBA{255, 255, 255, 255}
		}

		ebitenutil.DrawCircle(screen, x, y, 10, c)
	}
}

// DrawCloud отрисовывает облачко
func (r *Renderer) DrawCloud(screen *ebiten.Image, cloud *entity.Cloud, cameraX, cameraY float64) {
	if cloud.Collected {
		return
	}

	if cloud.Renderer.CurrentImg != nil {
		cloud.Renderer.Draw(screen, cloud.Transform, cameraX, cameraY)
	} else {
		x := cloud.Transform.X - cameraX + cloud.Transform.Width/2
		y := cloud.Transform.Y - cameraY + cloud.FloatY + cloud.Transform.Height/2

		// Пушистое белое облачко
		c := color.RGBA{255, 255, 255, 230}
		size := 15.0 + float64(cloud.CloudNum)*5

		ebitenutil.DrawCircle(screen, x, y, size, c)
		ebitenutil.DrawCircle(screen, x+size*0.7, y+3, size*0.6, c)
		ebitenutil.DrawCircle(screen, x-size*0.7, y+3, size*0.6, c)
	}
}

// DrawProjectile отрисовывает морковку-снаряд
func (r *Renderer) DrawProjectile(screen *ebiten.Image, projectile *entity.Projectile, cameraX, cameraY float64) {
	if !projectile.Active {
		return
	}

	x := projectile.Transform.X - cameraX
	y := projectile.Transform.Y - cameraY

	// Оранжевая морковка
	ebitenutil.DrawCircle(screen, x+12, y+4, 7, color.RGBA{255, 140, 0, 255})
	ebitenutil.DrawCircle(screen, x+12, y+6, 5, color.RGBA{255, 165, 0, 255})

	// Зелёный хвостик
	ebitenutil.DrawRect(screen, x+4, y+2, 6, 4, color.RGBA{50, 205, 50, 255})

	// Траектория
	for i := 0; i < 3; i++ {
		trailX := x - float64(i)*6
		alpha := uint8(200 - i*50)
		ebitenutil.DrawCircle(screen, trailX+10, y+4, 4-float64(i), color.RGBA{255, 200, 0, alpha})
	}
}

// DrawExit отрисовывает выход (флаг)
func (r *Renderer) DrawExit(screen *ebiten.Image, exitX, exitY, cameraX, cameraY float64) {
	x := exitX - cameraX
	y := exitY - cameraY

	// Флагшток
	ebitenutil.DrawRect(screen, x+5, y, 4, 80, color.RGBA{139, 90, 43, 255})

	// Зелёный флаг
	ebitenutil.DrawRect(screen, x+9, y+5, 40, 25, color.RGBA{50, 205, 50, 255})
	ebitenutil.DrawRect(screen, x+12, y+8, 15, 15, color.RGBA{255, 255, 255, 255}) // Звёздочка

	// Мигающий эффект
	alpha := uint8(150 + 50*int(math.Sin(float64(int(exitY)*100))*0.5+0.5))
	ebitenutil.DrawCircle(screen, x+7, y+3, 5, color.RGBA{255, 215, 0, alpha})
}

// DrawParticles отрисовывает частицы
func (r *Renderer) DrawParticles(screen *ebiten.Image, particles []Particle, cameraX, cameraY float64) {
	for _, p := range particles {
		x := p.X - cameraX
		y := p.Y - cameraY
		alpha := uint8(p.Life * 255)

		switch c := p.Color.(type) {
		case color.RGBA:
			ebitenutil.DrawCircle(screen, x, y, p.Size, color.RGBA{c.R, c.G, c.B, alpha})
		}
	}
}

// DrawHUD отрисовывает интерфейс
func (r *Renderer) DrawHUD(screen *ebiten.Image, health, maxHealth int, light, maxLight float64, score, levelNum int, levelName string, friendCount, cloudCount int) {
	// Полоска здоровья - розовая
	healthPercent := float64(health) / float64(maxHealth)
	ebitenutil.DrawRect(screen, 10, 10, 200, 20, color.RGBA{50, 50, 50, 255})
	healthColor := color.RGBA{255, 100, 150, 255}
	if healthPercent > 0.5 {
		healthColor = color.RGBA{100, 255, 100, 255}
	}
	ebitenutil.DrawRect(screen, 10, 10, 200*healthPercent, 20, healthColor)
	ebitenutil.DebugPrintAt(screen, "❤ "+itoa(health)+"/"+itoa(maxHealth), 15, 12)

	// Полоска света - жёлтая
	lightPercent := light / maxLight
	ebitenutil.DrawRect(screen, 10, 35, 200, 15, color.RGBA{50, 50, 50, 255})
	ebitenutil.DrawRect(screen, 10, 35, 200*lightPercent, 15, color.RGBA{255, 255, 0, 255})
	ebitenutil.DebugPrintAt(screen, "☀ "+itoa(int(light))+"%", 15, 37)

	// Счёт
	ebitenutil.DebugPrintAt(screen, "★ "+itoa(score), 250, 15)

	// Друзья
	ebitenutil.DebugPrintAt(screen, "🐾 Друзья: "+itoa(friendCount), 250, 40)

	// Облачка
	ebitenutil.DebugPrintAt(screen, "☁ "+itoa(cloudCount)+" облачков", 450, 15)

	// Уровень
	ebitenutil.DebugPrintAt(screen, "Уровень "+itoa(levelNum)+": "+levelName, 450, 40)
}

// DrawMenu отрисовывает главное меню
func (r *Renderer) DrawMenu(screen *ebiten.Image) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	// Градиентный фон
	for y := 0; y < screenH; y++ {
		r := uint8(135 + int(float64(y)/float64(screenH)*50))
		g := uint8(206 - int(float64(y)/float64(screenH)*50))
		b := uint8(235 - int(float64(y)/float64(screenH)*30))
		ebitenutil.DrawRect(screen, 0, float64(y), float64(screenW), 1, color.RGBA{r, g, b, 255})
	}

	// Облачка
	r.drawCloudOnBackground(screen, 200, 100, 2)
	r.drawCloudOnBackground(screen, 900, 120, 3)
	r.drawCloudOnBackground(screen, 500, 80, 1)

	// Морковки вокруг
	r.drawCarrot(screen, 300, 250)
	r.drawCarrot(screen, 800, 230)
	r.drawCarrot(screen, 400, 300)

	// Заголовок
	title := `
╔═══════════════════════════════════════════════╗
║                                               ║
║     🌈  SUNNY ADVENTURE  🏃                   ║
║                                               ║
║     Приключения Героя в Облачной Стране      ║
║                                               ║
╠═══════════════════════════════════════════════╣
║                                               ║
║         [SPACE] - Начать приключение        ║
║         [ESC] - Выход                        ║
║                                               ║
║   Собери всех друзей и верни облачка!       ║
║   Стреляй морковками!                        ║
║                                               ║
║   Управление:                                 ║
║   A/D - Бег  |  W - Прыжок  |  J - Морковка  ║
║   K - Обнимашки с друзьями                   ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, title, screenW/2-240, screenH/2-150)
}

// drawCarrot рисует морковку
func (r *Renderer) drawCarrot(screen *ebiten.Image, x, y float64) {
	// Оранжевая морковка
	ebitenutil.DrawCircle(screen, x, y, 8, color.RGBA{255, 140, 0, 255})
	ebitenutil.DrawCircle(screen, x, y+5, 6, color.RGBA{255, 165, 0, 255})
	// Зелёный хвостик
	ebitenutil.DrawRect(screen, x-3, y-8, 2, 8, color.RGBA{50, 205, 50, 255})
	ebitenutil.DrawRect(screen, x+1, y-8, 2, 8, color.RGBA{50, 205, 50, 255})
}

// DrawPause отрисовывает паузу
func (r *Renderer) DrawPause(screen *ebiten.Image) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{255, 255, 200, 200})
	screen.DrawImage(overlay, nil)

	pauseText := `
╔═══════════════════════════════════════╗
║       🌸  ПАУЗА  🌸                   ║
╠═══════════════════════════════════════╣
║                                       ║
║   [ESC] - Продолжить                 ║
║   [SPACE] - Рестарт                  ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, pauseText, screenW/2-200, screenH/2-100)
}

// DrawGameOver отрисовывает Game Over
func (r *Renderer) DrawGameOver(screen *ebiten.Image, score, levelNum int) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{200, 150, 150, 200})
	screen.DrawImage(overlay, nil)

	gameOverText := `
╔═══════════════════════════════════════╗
║      😢  GAME OVER  😢                ║
╠═══════════════════════════════════════╣
║                                       ║
║   Счёт: ` + itoa(score) + `                          ║
║   Уровень: ` + itoa(levelNum) + `                        ║
║                                       ║
║   Не сдавайся! Попробуй снова!       ║
║                                       ║
║   [SPACE] - Попробовать снова        ║
║   [ESC] - Выход                      ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, gameOverText, screenW/2-200, screenH/2-100)
}

// DrawVictory отрисовывает победу
func (r *Renderer) DrawVictory(screen *ebiten.Image, score, friendCount int) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{255, 200, 150, 200})
	screen.DrawImage(overlay, nil)

	victoryText := `
╔═══════════════════════════════════════╗
║      🎉 ПОБЕДА! 🎉                    ║
╠═══════════════════════════════════════╣
║                                       ║
║   Ты спас всех друзей!               ║
║   Друзей собрано: ` + itoa(friendCount) + `              ║
║                                       ║
║   Итоговый счёт: ` + itoa(score) + `                 ║
║                                       ║
║   Спасибо за игру! ❤                 ║
║                                       ║
║   [SPACE] - Играть снова             ║
║   [ESC] - Выход                      ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, victoryText, screenW/2-200, screenH/2-120)
}

// DrawLevelComplete отрисовывает завершение уровня
func (r *Renderer) DrawLevelComplete(screen *ebiten.Image, levelNum int, score, cloudCount int) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{150, 255, 200, 180})
	screen.DrawImage(overlay, nil)

	levelCompleteText := `
╔═══════════════════════════════════════╗
║   ✅ Уровень ` + itoa(levelNum) + ` пройден! ✅     ║
╠═══════════════════════════════════════╣
║                                       ║
║   Облачков собрано: ` + itoa(cloudCount) + `            ║
║   Счёт: ` + itoa(score) + `                          ║
║                                       ║
║   [SPACE] - Следующий уровень        ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, levelCompleteText, screenW/2-200, screenH/2-100)
}

// DrawArc рисует дугу
func (r *Renderer) DrawArc(screen *ebiten.Image, cx, cy, radius, start, end float64, width int, c color.Color) {
	steps := int(radius * 2)
	for i := 0; i < steps; i++ {
		angle := start + (end-start)*float64(i)/float64(steps)
		x := cx + math.Cos(angle)*radius
		y := cy + math.Sin(angle)*radius
		ebitenutil.DrawCircle(screen, x, y, float64(width)/2, c)
	}
}

// itoa - конвертация int в string
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := false
	if n < 0 {
		negative = true
		n = -n
	}

	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	if negative {
		digits = append(digits, '-')
	}

	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	return string(digits)
}
