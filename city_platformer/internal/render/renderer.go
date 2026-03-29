// Package render - система отрисовки City Platformer
// Go365 Day 91
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

// Renderer - система отрисовки
type Renderer struct {
	spriteSheet *sprite.SpriteSheet
}

// NewRenderer - создание рендерера
func NewRenderer(ss *sprite.SpriteSheet) *Renderer {
	return &Renderer{
		spriteSheet: ss,
	}
}

// DrawPlayer - отрисовка игрока
func (r *Renderer) DrawPlayer(screen *ebiten.Image, player *entity.Player, cameraX, cameraY float64) {
	player.Draw(screen, cameraX, cameraY)
}

// DrawEnemy - отрисовка врага
func (r *Renderer) DrawEnemy(screen *ebiten.Image, enemy *entity.Enemy, cameraX, cameraY float64) {
	enemy.Draw(screen, cameraX, cameraY)
}

// DrawProjectile - отрисовка снаряда
func (r *Renderer) DrawProjectile(screen *ebiten.Image, projectile *entity.Projectile, cameraX, cameraY float64) {
	projectile.Draw(screen, cameraX, cameraY)
}

// DrawItem - отрисовка предмета
func (r *Renderer) DrawItem(screen *ebiten.Image, item *level.LevelItem, cameraX, cameraY float64) {
	if item.Collected {
		return
	}

	screenX := item.X - cameraX
	screenY := item.Y - cameraY

	// Анимация парения
	offsetY := 0.0

	spriteImg := r.spriteSheet.GetItemSprite(item.Type)
	if spriteImg != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(screenX, screenY+offsetY)
		screen.DrawImage(spriteImg, opts)
	}
}

// DrawPlatform - отрисовка платформы
func (r *Renderer) DrawPlatform(screen *ebiten.Image, platform *level.Platform, cameraX, cameraY float64) {
	tile := r.spriteSheet.GetTile(platform.Type)

	if tile != nil {
		tileW := float64(tile.Bounds().Dx())
		tileH := float64(tile.Bounds().Dy())

		startX := (platform.X - cameraX)
		startY := (platform.Y - cameraY)

		for y := startY; y < startY+platform.Height; y += tileH {
			for x := startX; x < startX+platform.Width; x += tileW {
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(x, y)
				screen.DrawImage(tile, opts)
			}
		}
	}
}

// DrawBackground - отрисовка фона (постапокалиптический город)
func (r *Renderer) DrawBackground(screen *ebiten.Image, cameraX, cameraY float64) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	// Небо - тёмное, загрязнённое
	for y := 0; y < screenH; y++ {
		ratio := float64(y) / float64(screenH)
		r := uint8(80 - ratio*20)
		g := uint8(70 - ratio*15)
		b := uint8(90 - ratio*20)
		vector.DrawFilledRect(screen, 0, float32(y), float32(screenW), 1, color.RGBA{r, g, b, 255}, true)
	}

	// Дальние здания (параллакс слой 1)
	r.drawBuildingsLayer(screen, cameraX*0.2, 200, color.RGBA{50, 50, 60, 255}, 150, 300)

	// Ближние здания (параллакс слой 2)
	r.drawBuildingsLayer(screen, cameraX*0.5, 400, color.RGBA{70, 70, 80, 255}, 100, 200)

	// Разрушенные здания на переднем плане
	r.drawRuinedBuildings(screen, cameraX)
}

// drawBuildingsLayer - отрисовка слоя зданий
func (r *Renderer) drawBuildingsLayer(screen *ebiten.Image, offsetX, baseY float64, c color.RGBA, minWidth, maxWidth float64) {
	screenW := screen.Bounds().Dx()

	for x := -100.0; x < float64(screenW)+100; x += 80 {
		buildingX := x - math.Mod(offsetX, 80)
		buildingW := minWidth + math.Mod(x*17, maxWidth-minWidth)
		buildingH := 100.0 + math.Mod(x*23, 200)

		vector.DrawFilledRect(
			screen,
			float32(buildingX),
			float32(baseY-buildingH),
			float32(buildingW),
			float32(buildingH),
			c,
			true,
		)

		// Окна
		r.drawWindows(screen, buildingX, baseY-buildingH+10, buildingW, buildingH-20, c)
	}
}

// drawWindows - отрисовка окон
func (r *Renderer) drawWindows(screen *ebiten.Image, x, y, w, h float64, baseColor color.RGBA) {
	windowColor := color.RGBA{100, 100, 80, 200}
	
	for wy := y + 10; wy < y+h-10; wy += 30 {
		for wx := x + 10; wx < x+w-10; wx += 25 {
			// Каждое окно светится с вероятностью 30%
			if int(wx*wy)%10 < 3 {
				vector.DrawFilledRect(screen, float32(wx), float32(wy), 15, 20, windowColor, true)
			}
		}
	}
}

// drawRuinedBuildings - отрисовка разрушенных зданий
func (r *Renderer) drawRuinedBuildings(screen *ebiten.Image, cameraX float64) {
	screenH := screen.Bounds().Dy()

	// Разрушенные силуэты
	for i := 0; i < 5; i++ {
		buildingX := float64(i*300) - (cameraX * 0.8)
		buildingW := 120.0
		buildingH := 250.0 + math.Mod(float64(i*37), 100)

		// Основная форма здания
		vector.DrawFilledRect(
			screen,
			float32(buildingX),
			float32(float64(screenH)-150-buildingH),
			float32(buildingW),
			float32(buildingH),
			color.RGBA{40, 40, 50, 255},
			true,
		)

		// Разрушенная крыша (зубцы)
		for j := 0; j < 5; j++ {
			if (i+j)%2 == 0 {
				vector.DrawFilledRect(
					screen,
					float32(buildingX+float64(j)*25),
					float32(float64(screenH)-150-buildingH-20),
					20,
					20,
					color.RGBA{40, 40, 50, 255},
					true,
				)
			}
		}
	}
}

// DrawHUD - отрисовка интерфейса
func (r *Renderer) DrawHUD(screen *ebiten.Image, health, maxHealth, ammo, maxAmmo, score, level int, levelName string) {
	screenW := screen.Bounds().Dx()

	// Левая панель - здоровье
	r.drawHealthBar(screen, 10, 10, health, maxHealth)

	// Правая панель - патроны и счёт
	r.drawAmmoPanel(screen, screenW-200, 10, ammo, maxAmmo)
	r.drawScorePanel(screen, screenW-200, 60, score)

	// Центр - название уровня и номер
	r.drawLevelInfo(screen, screenW/2, 20, level, levelName)
}

// drawHealthBar - полоска здоровья
func (r *Renderer) drawHealthBar(screen *ebiten.Image, x, y int, health, maxHealth int) {
	barWidth := 200
	barHeight := 20

	// Фон
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(barWidth), float32(barHeight), color.RGBA{50, 50, 50, 255}, true)

	// Здоровье
	healthPercent := float32(health) / float32(maxHealth)
	barColor := color.RGBA{200, 50, 50, 255}
	if healthPercent > 0.5 {
		barColor = color.RGBA{50, 200, 50, 255}
	} else if healthPercent > 0.25 {
		barColor = color.RGBA{200, 200, 50, 255}
	}

	vector.DrawFilledRect(screen, float32(x)+2, float32(y)+2, (float32(barWidth)-4)*healthPercent, float32(barHeight)-4, barColor, true)

	// Рамка
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(barWidth), float32(barHeight), color.RGBA{255, 255, 255, 255}, false)

	// Текст
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("❤️ %d/%d", health, maxHealth), x+10, y+2)
}

// drawAmmoPanel - панель патронов
func (r *Renderer) drawAmmoPanel(screen *ebiten.Image, x, y int, ammo, maxAmmo int) {
	// Фон
	vector.DrawFilledRect(screen, float32(x), float32(y), 180, 40, color.RGBA{50, 50, 50, 200}, true)
	vector.DrawFilledRect(screen, float32(x), float32(y), 180, 40, color.RGBA{255, 255, 255, 100}, false)

	// Текст
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🔫 Патроны: %d/%d", ammo, maxAmmo), x+10, y+12)
}

// drawScorePanel - панель счёта
func (r *Renderer) drawScorePanel(screen *ebiten.Image, x, y int, score int) {
	// Фон
	vector.DrawFilledRect(screen, float32(x), float32(y), 180, 40, color.RGBA{50, 50, 50, 200}, true)
	vector.DrawFilledRect(screen, float32(x), float32(y), 180, 40, color.RGBA{255, 255, 255, 100}, false)

	// Текст
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("⭐ Счёт: %d", score), x+10, y+12)
}

// drawLevelInfo - информация об уровне
func (r *Renderer) drawLevelInfo(screen *ebiten.Image, x, y int, levelNum int, levelName string) {
	// Фон
	vector.DrawFilledRect(screen, float32(x-150), float32(y-15), 300, 30, color.RGBA{0, 0, 0, 150}, true)

	// Текст
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("📍 Уровень %d: %s", levelNum, levelName), x-140, y)
}

// DrawExit - отрисовка выхода
func (r *Renderer) DrawExit(screen *ebiten.Image, x, y, cameraX, cameraY float64) {
	screenX := x - cameraX
	screenY := y - cameraY

	// Зелёная зона эвакуации
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), 60, 80, color.RGBA{50, 200, 50, 150}, true)
	vector.DrawFilledRect(screen, float32(screenX)+20, float32(screenY)+10, 20, 60, color.RGBA{100, 255, 100, 255}, true)

	// Стрелка вверх
	ebitenutil.DebugPrintAt(screen, "🚁", int(screenX)+15, int(screenY)-10)
}

// DrawMessage - отрисовка сообщения
func (r *Renderer) DrawMessage(screen *ebiten.Image, message string, x, y int) {
	// Фон
	vector.DrawFilledRect(screen, float32(x-10), float32(y-10), float32(len(message)*12+20), 40, color.RGBA{0, 0, 0, 200}, true)

	// Текст
	ebitenutil.DebugPrintAt(screen, message, x, y)
}

// DrawParticles - отрисовка частиц
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

// Particle - частица
type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    float64
	MaxLife float64
	Color   color.Color
	Size    float64
}

// CreateExplosion - создание взрыва частиц
func CreateExplosion(x, y float64, count int, c color.Color) []Particle {
	particles := make([]Particle, count)
	for i := 0; i < count; i++ {
		particles[i] = Particle{
			X: x, Y: y,
			VX: (float64(i%10) - 5) * 2,
			VY: (float64(i/10) - 2) * 2,
			Life:    1.0,
			MaxLife: 1.0,
			Color:   c,
			Size:    3.0 + float64(i%5),
		}
	}
	return particles
}
