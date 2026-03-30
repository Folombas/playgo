// Package render - система отрисовки игры
// Go365 Day 90 - City Survivor
package render

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"city_platformer/internal/entity"
	"city_platformer/internal/level"
	"city_platformer/internal/sprite"
)

// Renderer - основной рендерер игры
type Renderer struct {
	spriteSheet *sprite.SpriteSheet
	tileSize    int
}

// NewRenderer создаёт новый рендерер
func NewRenderer(ss *sprite.SpriteSheet) *Renderer {
	return &Renderer{
		spriteSheet: ss,
		tileSize:    64,
	}
}

// DrawBackground отрисовывает фон
func (r *Renderer) DrawBackground(screen *ebiten.Image, cameraX, cameraY float64) {
	bg := r.spriteSheet.GetBackground()
	if bg == nil {
		// Заглушка - градиентное небо
		r.drawPlaceholderBackground(screen)
		return
	}

	// Параллакс эффект для фона
	parallaxX := cameraX * 0.3
	parallaxY := cameraY * 0.3

	opts := &ebiten.DrawImageOptions{}

	// Рисуем фон несколько раз для бесшовности
	screenW, screenH := screen.Size()
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			opts.GeoM.Reset()
			opts.GeoM.Translate(
				float64(x*screenW)-parallaxX,
				float64(y*screenH)-parallaxY,
			)
			screen.DrawImage(bg, opts)
		}
	}
}

// drawPlaceholderBackground рисует фон-заглушку
func (r *Renderer) drawPlaceholderBackground(screen *ebiten.Image) {
	screenW, screenH := screen.Size()

	// Градиент неба
	for y := 0; y < screenH/2; y++ {
		ratio := float32(y) / float32(screenH/2)
		r := uint8(40 + ratio*30)
		g := uint8(40 + ratio*25)
		b := uint8(60 + ratio*40)
		vector.DrawFilledRect(screen, 0, float32(y), float32(screenW), 1, color.RGBA{r, g, b, 255}, true)
	}

	// Силуэты зданий на заднем плане
	r.drawCitySilhouette(screen, screenW, screenH)
}

// drawCitySilhouette рисует силуэт города
func (r *Renderer) drawCitySilhouette(screen *ebiten.Image, screenW, screenH int) {
	buildingColor := color.RGBA{50, 50, 60, 255}

	// Генерируем здания детерминированно
	for x := 0; x < screenW; x += 60 {
		height := 100 + (x*17)%150
		vector.DrawFilledRect(
			screen,
			float32(x),
			float32(screenH/2+50-height),
			55,
			float32(height+100),
			buildingColor,
			true,
		)

		// Окна с светом
		r.drawBuildingWindows(screen, x, screenH/2+50-height, 55, height)
	}
}

// drawBuildingWindows рисует окна здания
func (r *Renderer) drawBuildingWindows(screen *ebiten.Image, x, y, w, h int) {
	windowColor := color.RGBA{150, 140, 100, 200}

	for wy := y + 10; wy < y+h-10; wy += 25 {
		for wx := x + 8; wx < x+w-8; wx += 15 {
			// Каждое окно светится с вероятностью 40%
			if (wx*wy)%10 < 4 {
				vector.DrawFilledRect(screen, float32(wx), float32(wy), 10, 15, windowColor, true)
			}
		}
	}
}

// DrawPlatform отрисовывает платформу
func (r *Renderer) DrawPlatform(screen *ebiten.Image, platform *level.Platform, cameraX, cameraY float64) {
	tileSize := float64(r.tileSize)

	// Получаем спрайт тайла
	spriteName := level.GetTileSpriteName(platform.Type)
	tileSprite := r.spriteSheet.GetTileSprite(spriteName)

	if tileSprite == nil {
		// Заглушка - цветной прямоугольник
		tileColor := color.RGBA{100, 100, 100, 255}
		vector.DrawFilledRect(
			screen,
			float32(platform.X-cameraX),
			float32(platform.Y-cameraY),
			float32(platform.Width),
			float32(platform.Height),
			tileColor,
			true,
		)
		return
	}

	// Тайлим спрайт по платформе
	opts := &ebiten.DrawImageOptions{}
	for y := platform.Y; y < platform.Y+platform.Height; y += tileSize {
		for x := platform.X; x < platform.X+platform.Width; x += tileSize {
			opts.GeoM.Reset()
			opts.GeoM.Translate(x-cameraX, y-cameraY)
			screen.DrawImage(tileSprite, opts)
		}
	}
}

// DrawPlayer отрисовывает игрока
func (r *Renderer) DrawPlayer(screen *ebiten.Image, player *entity.Player, cameraX, cameraY float64) {
	player.Draw(screen, cameraX, cameraY)
}

// DrawEnemy отрисовывает врага
func (r *Renderer) DrawEnemy(screen *ebiten.Image, enemy *entity.Enemy, cameraX, cameraY float64) {
	enemy.Draw(screen, cameraX, cameraY)

	// Полоска здоровья врага
	if enemy.Health.Current < enemy.Health.Max {
		r.drawHealthBar(screen, int(enemy.Transform.X-cameraX), int(enemy.Transform.Y-cameraY-10), 40, enemy.Health.Current, enemy.Health.Max)
	}
}

// DrawItem отрисовывает предмет
func (r *Renderer) DrawItem(screen *ebiten.Image, item *entity.Item, cameraX, cameraY float64) {
	item.Draw(screen, cameraX, cameraY)
}

// DrawProjectile отрисовывает снаряд
func (r *Renderer) DrawProjectile(screen *ebiten.Image, projectile *entity.Projectile, cameraX, cameraY float64) {
	projectile.Draw(screen, cameraX, cameraY)
}

// DrawExit отрисовывает выход
func (r *Renderer) DrawExit(screen *ebiten.Image, x, y float64, cameraX, cameraY float64) {
	screenX := x - cameraX
	screenY := y - cameraY

	// Зона выхода
	vector.DrawFilledRect(
		screen,
		float32(screenX),
		float32(screenY),
		64,
		80,
		color.RGBA{50, 200, 50, 100},
		true,
	)

	// Спрайт двери
	if doorSprite := r.spriteSheet.GetTileSprite("door_closedTop"); doorSprite != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(screenX+16, screenY)
		screen.DrawImage(doorSprite, opts)
	}

	// Стрелка/индикатор
	ebitenutil.DebugPrintAt(screen, "🚁", int(screenX)+18, int(screenY)-10)
}

// DrawHUD отрисовывает интерфейс
func (r *Renderer) DrawHUD(
	screen *ebiten.Image,
	health, maxHealth int,
	ammo, maxAmmo int,
	score int,
	levelNum int,
	levelName string,
) {
	screenW, _ := screen.Size()

	// Левая панель - здоровье
	r.drawHealthBar(screen, 10, 10, 200, health, maxHealth)

	// Правая панель - патроны и счёт
	r.drawAmmoPanel(screen, screenW-200, 10, ammo, maxAmmo)
	r.drawScorePanel(screen, screenW-200, 55, score)

	// Центр - информация об уровне
	r.drawLevelInfo(screen, screenW/2, 15, levelNum, levelName)
}

// drawHealthBar рисует полоску здоровья
func (r *Renderer) drawHealthBar(screen *ebiten.Image, x, y, width int, health, maxHealth int) {
	// Фон
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), 25, color.RGBA{40, 40, 40, 200}, true)

	// Полоска здоровья
	healthPercent := float32(health) / float32(maxHealth)
	barColor := color.RGBA{200, 50, 50, 255}
	if healthPercent > 0.5 {
		barColor = color.RGBA{50, 200, 50, 255}
	} else if healthPercent > 0.25 {
		barColor = color.RGBA{200, 200, 50, 255}
	}

	vector.DrawFilledRect(screen, float32(x)+2, float32(y)+2, (float32(width)-4)*healthPercent, 21, barColor, true)

	// Рамка
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), 25, color.RGBA{255, 255, 255, 100}, false)

	// Текст
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("❤️ %d/%d", health, maxHealth), x+10, y+4)
}

// drawAmmoPanel рисует панель патронов
func (r *Renderer) drawAmmoPanel(screen *ebiten.Image, x, y int, ammo, maxAmmo int) {
	vector.DrawFilledRect(screen, float32(x), float32(y), 180, 35, color.RGBA{40, 40, 40, 180}, true)
	vector.DrawFilledRect(screen, float32(x), float32(y), 180, 35, color.RGBA{255, 255, 255, 80}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🔫 %d/%d", ammo, maxAmmo), x+10, y+8)
}

// drawScorePanel рисует панель счёта
func (r *Renderer) drawScorePanel(screen *ebiten.Image, x, y int, score int) {
	vector.DrawFilledRect(screen, float32(x), float32(y), 180, 35, color.RGBA{40, 40, 40, 180}, true)
	vector.DrawFilledRect(screen, float32(x), float32(y), 180, 35, color.RGBA{255, 255, 255, 80}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("⭐ %d", score), x+10, y+8)
}

// drawLevelInfo рисует информацию об уровне
func (r *Renderer) drawLevelInfo(screen *ebiten.Image, x, y int, levelNum int, levelName string) {
	vector.DrawFilledRect(screen, float32(x-150), float32(y-12), 300, 24, color.RGBA{0, 0, 0, 120}, true)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("📍 Уровень %d: %s", levelNum, levelName), x-140, y)
}

// DrawMenu отрисовывает главное меню
func (r *Renderer) DrawMenu(screen *ebiten.Image) {
	screenW, screenH := screen.Size()

	// Затемнение фона
	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	// Заголовок
	title := `
╔═══════════════════════════════════════════════╗
║       🏙️ CITY SURVIVOR 🏙️                    ║
║          LAST STAND                           ║
╠═══════════════════════════════════════════════╣
║                                               ║
║           [SPACE] - Начать игру               ║
║           [ESC] - Выход                       ║
║                                               ║
║  🎮 Управление:                               ║
║     A/D или ←/→ - Бег                         ║
║     W/↑ - Прыжок                              ║
║     S/↓ - Присесть                            ║
║     J - Выстрел                               ║
║     K - Перезарядка                           ║
║                                               ║
║  🎯 Цель: Доберись до точки эвакуации!        ║
║  💀 Остерегайся мутантов и роботов!           ║
║  📦 Собирай монеты и кристаллы                ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, title, screenW/2-240, screenH/2-180)
}

// DrawPause отрисовывает паузу
func (r *Renderer) DrawPause(screen *ebiten.Image) {
	screenW, screenH := screen.Size()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	pauseText := `
╔═══════════════════════════════════════╗
║              ⏸️ ПАУЗА                  ║
╠═══════════════════════════════════════╣
║     [ESC] - Продолжить                ║
║     [SPACE] - Выйти в меню            ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, pauseText, screenW/2-180, screenH/2-60)
}

// DrawGameOver отрисовывает экран смерти
func (r *Renderer) DrawGameOver(screen *ebiten.Image, score, levelNum int) {
	screenW, screenH := screen.Size()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{80, 0, 0, 200})
	screen.DrawImage(overlay, nil)

	gameOverText := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║       💀 ВЫ ПОГИБЛИ 💀                ║
╠═══════════════════════════════════════╣
║     Финальный счёт: %6d                ║
║     Уровень: %2d                        ║
║                                       ║
║     [SPACE] - Новая попытка           ║
║     [ESC] - Выход                     ║
╚═══════════════════════════════════════╝
`, score, levelNum)

	ebitenutil.DebugPrintAt(screen, gameOverText, screenW/2-180, screenH/2-100)
}

// DrawVictory отрисовывает победу
func (r *Renderer) DrawVictory(screen *ebiten.Image, score int) {
	screenW, screenH := screen.Size()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 80, 0, 150})
	screen.DrawImage(overlay, nil)

	victoryText := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║     🚁 ЭВАКУАЦИЯ УСПЕШНА! 🚁          ║
╠═══════════════════════════════════════╣
║     Вы прошли все уровни!             ║
║     Финальный счёт: %6d                ║
║                                       ║
║     🎉 ПОЗДРАВЛЯЕМ! 🎉                ║
║                                       ║
║     [SPACE] - Играть снова            ║
║     [ESC] - Выход                     ║
╚═══════════════════════════════════════╝
`, score)

	ebitenutil.DebugPrintAt(screen, victoryText, screenW/2-180, screenH/2-120)
}

// DrawLevelComplete отрисовывает завершение уровня
func (r *Renderer) DrawLevelComplete(screen *ebiten.Image, levelNum, score int) {
	screenW, screenH := screen.Size()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 80, 0, 150})
	screen.DrawImage(overlay, nil)

	levelCompleteText := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║     ✅ УРОВЕНЬ %d ПРОЙДЕН! ✅          ║
╠═══════════════════════════════════════╣
║     Счёт: %6d                          ║
║                                       ║
║     [SPACE] - Следующий уровень       ║
║     [ESC] - Пауза                     ║
╚═══════════════════════════════════════╝
`, levelNum, score)

	ebitenutil.DebugPrintAt(screen, levelCompleteText, screenW/2-180, screenH/2-80)
}

// Particle - частица для эффектов
type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    float64
	MaxLife float64
	Color   color.Color
	Size    float64
}

// DrawParticles отрисовывает частицы
func (r *Renderer) DrawParticles(screen *ebiten.Image, particles []Particle, cameraX, cameraY float64) {
	for _, p := range particles {
		screenX := p.X - cameraX
		screenY := p.Y - cameraY
		alpha := uint8(255 * (p.Life / p.MaxLife))

		r, g, b, _ := p.Color.RGBA()
		vector.DrawFilledCircle(
			screen,
			float32(screenX),
			float32(screenY),
			float32(p.Size),
			color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), alpha},
			true,
		)
	}
}

// CreateExplosion создаёт взрыв частиц
func CreateExplosion(x, y float64, count int, c color.Color) []Particle {
	particles := make([]Particle, count)
	for i := 0; i < count; i++ {
		angle := float64(i) / float64(count) * 2 * math.Pi
		speed := 50 + float64(i%10)*10
		particles[i] = Particle{
			X: x, Y: y,
			VX: math.Cos(angle) * speed,
			VY: math.Sin(angle) * speed,
			Life:    1.0,
			MaxLife: 1.0,
			Color:   c,
			Size:    3.0 + float64(i%5)*2,
		}
	}
	return particles
}
