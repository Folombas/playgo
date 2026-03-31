// Package render - отрисовка игры
// Go365 Day 91 - Cyber City Runner
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"cyber_city_runner/internal/entity"
	"cyber_city_runner/internal/level"
	"cyber_city_runner/internal/sprite"
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
func (r *Renderer) DrawBackground(screen *ebiten.Image, cameraX, cameraY float64) {
	// Киберпанк фон - тёмно-фиолетовый градиент
	screen.Fill(color.RGBA{20, 10, 40, 255})

	// Рисуем "неоновые" линии на фоне
	for i := 0; i < 20; i++ {
		x := float64(i*100) - cameraX*0.5
		y := float64(i*50) - cameraY*0.3
		ebitenutil.DrawLine(screen, x, 0, x, 720, color.RGBA{50, 20, 100, 100})
		ebitenutil.DrawLine(screen, 0, y, 1280, y, color.RGBA{50, 20, 100, 100})
	}
}

// DrawPlatform отрисовывает платформу
func (r *Renderer) DrawPlatform(screen *ebiten.Image, platform *level.Platform, cameraX, cameraY float64) {
	x := platform.X - cameraX
	y := platform.Y - cameraY

	switch platform.Type {
	case level.TileGround:
		// Земля - коричневый с зелёной травой сверху
		ebitenutil.DrawRect(screen, x, y, platform.Width, platform.Height, color.RGBA{101, 67, 33, 255})
		ebitenutil.DrawRect(screen, x, y, platform.Width, 10, color.RGBA{34, 139, 34, 255})
	case level.TileGrass:
		// Трава - зелёная платформа
		ebitenutil.DrawRect(screen, x, y, platform.Width, platform.Height, color.RGBA{34, 139, 34, 255})
		ebitenutil.DrawRect(screen, x, y, platform.Width, 5, color.RGBA{144, 238, 144, 255})
	case level.TileBrick:
		// Кирпич - серый/красноватый
		ebitenutil.DrawRect(screen, x, y, platform.Width, platform.Height, color.RGBA{139, 69, 19, 255})
		// Рисуем "кирпичики"
		for bx := x; bx < x+platform.Width; bx += 32 {
			for by := y; by < y+platform.Height; by += 16 {
				ebitenutil.DrawRect(screen, bx, by, 30, 14, color.RGBA{178, 34, 34, 255})
			}
		}
	case level.TileBox:
		// Коробка - жёлто-коричневая
		ebitenutil.DrawRect(screen, x, y, platform.Width, platform.Height, color.RGBA{210, 180, 140, 255})
		ebitenutil.DrawRect(screen, x+5, y+5, platform.Width-10, platform.Height-10, color.RGBA{139, 90, 43, 255})
	case level.TileLadder:
		// Лестница - серые перекладины
		for by := y; by < y+platform.Height; by += 20 {
			ebitenutil.DrawRect(screen, x, by, platform.Width, 4, color.RGBA{128, 128, 128, 255})
		}
		ebitenutil.DrawLine(screen, x+10, y, x+10, y+platform.Height, color.RGBA{128, 128, 128, 255})
		ebitenutil.DrawLine(screen, x+platform.Width-10, y, x+platform.Width-10, y+platform.Height, color.RGBA{128, 128, 128, 255})
	case level.TileSpike:
		// Шипы - треугольники
		for sx := x; sx < x+platform.Width; sx += 20 {
			ebitenutil.DrawLine(screen, sx, y+platform.Height, sx+10, y, color.RGBA{200, 200, 200, 255})
			ebitenutil.DrawLine(screen, sx+10, y, sx+20, y+platform.Height, color.RGBA{200, 200, 200, 255})
		}
	default:
		// По умолчанию - серый блок
		ebitenutil.DrawRect(screen, x, y, platform.Width, platform.Height, color.RGBA{100, 100, 100, 255})
	}
}

// DrawPlayer отрисовывает игрока
func (r *Renderer) DrawPlayer(screen *ebiten.Image, player *entity.Player, cameraX, cameraY float64) {
	if player.Invisible > 0 && int(player.Invisible*10)%2 == 0 {
		return // Мигание при невидимости
	}

	if player.Renderer.CurrentImg != nil {
		player.Renderer.Draw(screen, player.Transform, cameraX, cameraY)
	} else {
		// Заглушка - неоновый прямоугольник
		x := player.Transform.X - cameraX
		y := player.Transform.Y - cameraY - player.Transform.Height

		c := color.RGBA{0, 255, 255, 255} // Циан
		if player.Dashing {
			c = color.RGBA{255, 0, 255, 255} // Маджента при рывке
		}

		ebitenutil.DrawRect(screen, x, y, player.Transform.Width, player.Transform.Height, c)

		// "Глаза" для направления
		eyeX := x + player.Transform.Width/2
		if player.Transform.Facing == 1 {
			eyeX += 10
		} else {
			eyeX -= 10
		}
		ebitenutil.DrawCircle(screen, eyeX, y+15, 4, color.RGBA{255, 255, 255, 255})
	}
}

// DrawEnemy отрисовывает врага
func (r *Renderer) DrawEnemy(screen *ebiten.Image, enemy *entity.Enemy, cameraX, cameraY float64) {
	if enemy.Renderer.CurrentImg != nil {
		enemy.Renderer.Draw(screen, enemy.Transform, cameraX, cameraY)
	} else {
		// Заглушка - красный прямоугольник
		x := enemy.Transform.X - cameraX
		y := enemy.Transform.Y - cameraY - enemy.Transform.Height

		c := color.RGBA{255, 50, 50, 255}
		if enemy.Behavior == entity.EnemyFlying {
			c = color.RGBA{255, 100, 0, 255} // Оранжевый для летающих
		} else if enemy.Behavior == entity.EnemyTurret {
			c = color.RGBA{100, 100, 255, 255} // Синий для турелей
		} else if enemy.Behavior == entity.EnemyCamera {
			c = color.RGBA{255, 255, 0, 255} // Жёлтый для камер
		}

		ebitenutil.DrawRect(screen, x, y, enemy.Transform.Width, enemy.Transform.Height, c)

		// "Глаза"
		eyeX := x + enemy.Transform.Width/2
		if enemy.Transform.Facing == 1 {
			eyeX += 8
		} else {
			eyeX -= 8
		}
		ebitenutil.DrawCircle(screen, eyeX, y+12, 3, color.RGBA{255, 255, 0, 255})
	}
}

// DrawItem отрисовывает предмет
func (r *Renderer) DrawItem(screen *ebiten.Image, item *entity.Item, cameraX, cameraY float64) {
	if item.Renderer.CurrentImg != nil {
		item.Renderer.Draw(screen, item.Transform, cameraX, cameraY)
	} else {
		// Заглушка - светящийся круг
		x := item.Transform.X - cameraX + item.Transform.Width/2
		y := item.Transform.Y - cameraY + item.FloatOffset

		c := color.RGBA{255, 215, 0, 255} // Золотой
		if item.ItemType == "gemRed" {
			c = color.RGBA{255, 0, 0, 255}
		} else if item.ItemType == "gemBlue" {
			c = color.RGBA{0, 100, 255, 255}
		} else if item.ItemType == "star" {
			c = color.RGBA{255, 255, 255, 255}
		}

		ebitenutil.DrawCircle(screen, x, y, 10, c)
	}
}

// DrawProjectile отрисовывает снаряд
func (r *Renderer) DrawProjectile(screen *ebiten.Image, projectile *entity.Projectile, cameraX, cameraY float64) {
	if !projectile.Active {
		return
	}

	if projectile.Renderer.CurrentImg != nil {
		projectile.Renderer.Draw(screen, projectile.Transform, cameraX, cameraY)
	} else {
		// Заглушка - светящаяся точка
		x := projectile.Transform.X - cameraX
		y := projectile.Transform.Y - cameraY

		c := color.RGBA{255, 255, 0, 255}
		if projectile.IsEnemy {
			c = color.RGBA{255, 0, 0, 255}
		}

		ebitenutil.DrawCircle(screen, x+8, y+4, 6, c)
	}
}

// DrawExit отрисовывает выход
func (r *Renderer) DrawExit(screen *ebiten.Image, exitX, exitY, cameraX, cameraY float64) {
	x := exitX - cameraX
	y := exitY - cameraY

	// Неоновый портал
	ebitenutil.DrawRect(screen, x, y, 60, 80, color.RGBA{0, 0, 0, 100})
	ebitenutil.DrawRect(screen, x+5, y+5, 50, 70, color.RGBA{0, 255, 255, 200})
	ebitenutil.DrawRect(screen, x+10, y+10, 40, 60, color.RGBA{100, 255, 255, 255})

	// Мигающий эффект
	alpha := uint8(150 + 50*int(int(cameraX/100)%3))
	ebitenutil.DrawRect(screen, x+15, y+15, 30, 50, color.RGBA{200, 255, 255, alpha})
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
func (r *Renderer) DrawHUD(screen *ebiten.Image, health, maxHealth int, energy, maxEnergy float64, score, levelNum int, levelName string, combo int, special string) {
	// Полоска здоровья
	healthPercent := float64(health) / float64(maxHealth)
	ebitenutil.DrawRect(screen, 10, 10, 200, 20, color.RGBA{50, 50, 50, 255})
	healthColor := color.RGBA{255, 50, 50, 255}
	if healthPercent > 0.5 {
		healthColor = color.RGBA{50, 255, 50, 255}
	} else if healthPercent > 0.25 {
		healthColor = color.RGBA{255, 255, 0, 255}
	}
	ebitenutil.DrawRect(screen, 10, 10, 200*healthPercent, 20, healthColor)
	ebitenutil.DebugPrintAt(screen, "HP: "+itoa(health)+"/"+itoa(maxHealth), 15, 12)

	// Полоска энергии
	energyPercent := energy / maxEnergy
	ebitenutil.DrawRect(screen, 10, 35, 200, 15, color.RGBA{50, 50, 50, 255})
	ebitenutil.DrawRect(screen, 10, 35, 200*energyPercent, 15, color.RGBA{0, 200, 255, 255})
	ebitenutil.DebugPrintAt(screen, "NRG: "+itoa(int(energy))+"%", 15, 37)

	// Счёт
	ebitenutil.DebugPrintAt(screen, "СЧЁТ: "+itoa(score), 250, 15)

	// Уровень
	ebitenutil.DebugPrintAt(screen, "УРОВЕНЬ "+itoa(levelNum)+": "+levelName, 250, 40)

	// Комбо
	if combo > 1 {
		if combo > 20 {
			ebitenutil.DebugPrintAt(screen, "КОМБО x"+itoa(combo), 600, 15)
		} else if combo > 10 {
			ebitenutil.DebugPrintAt(screen, "КОМБО x"+itoa(combo), 600, 15)
		} else {
			ebitenutil.DebugPrintAt(screen, "КОМБО x"+itoa(combo), 600, 15)
		}
	}

	// Специальная способность
	ebitenutil.DebugPrintAt(screen, "[K] "+special, 1000, 15)
}

// DrawMenu отрисовывает главное меню
func (r *Renderer) DrawMenu(screen *ebiten.Image) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	// Затемнение
	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 0, 20, 230})
	screen.DrawImage(overlay, nil)

	// Заголовок
	title := `
╔═══════════════════════════════════════════════╗
║                                               ║
║     🌃  CYBER CITY RUNNER  🌃                ║
║                                               ║
║           Go365 Challenge - Day 91           ║
║                                               ║
╠═══════════════════════════════════════════════╣
║                                               ║
║         [SPACE] - Начать игру                ║
║         [ESC] - Выход                        ║
║                                               ║
║   Прорвись через неоновый город будущего!    ║
║   Собирай дата-чипы, избегай охранников!     ║
║                                               ║
║   Управление:                                 ║
║   A/D - Бег  |  W - Прыжок  |  Shift - Рывок ║
║   J - Огонь  |  K - Способность              ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, title, screenW/2-240, screenH/2-200)
}

// DrawPause отрисовывает паузу
func (r *Renderer) DrawPause(screen *ebiten.Image) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	pauseText := `
╔═══════════════════════════════════════╗
║           ⏸️  ПАУЗА  ⏸️               ║
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
	overlay.Fill(color.RGBA{50, 0, 0, 200})
	screen.DrawImage(overlay, nil)

	gameOverText := `
╔═══════════════════════════════════════╗
║          💀 GAME OVER 💀              ║
╠═══════════════════════════════════════╣
║                                       ║
║   Счёт: ` + itoa(score) + `                          ║
║   Уровень: ` + itoa(levelNum) + `                        ║
║                                       ║
║   [SPACE] - Попробовать снова        ║
║   [ESC] - Выход                      ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, gameOverText, screenW/2-200, screenH/2-100)
}

// DrawVictory отрисовывает победу
func (r *Renderer) DrawVictory(screen *ebiten.Image, score int) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 50, 0, 200})
	screen.DrawImage(overlay, nil)

	victoryText := `
╔═══════════════════════════════════════╗
║          🎉 ПОБЕДА! 🎉                ║
╠═══════════════════════════════════════╣
║                                       ║
║   Ты прошёл весь город!              ║
║                                       ║
║   Итоговый счёт: ` + itoa(score) + `                 ║
║                                       ║
║   [SPACE] - Играть снова             ║
║   [ESC] - Выход                      ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, victoryText, screenW/2-200, screenH/2-100)
}

// DrawLevelComplete отрисовывает завершение уровня
func (r *Renderer) DrawLevelComplete(screen *ebiten.Image, levelNum int, score int) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 0, 50, 180})
	screen.DrawImage(overlay, nil)

	levelCompleteText := `
╔═══════════════════════════════════════╗
║       ✅ УРОВЕНЬ ` + itoa(levelNum) + ` ПРОЙДЕН! ✅   ║
╠═══════════════════════════════════════╣
║                                       ║
║   Счёт: ` + itoa(score) + `                          ║
║                                       ║
║   [SPACE] - Следующий уровень        ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, levelCompleteText, screenW/2-200, screenH/2-100)
}

// itoa - простая конвертация int в string
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

	// Реверс
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	return string(digits)
}
