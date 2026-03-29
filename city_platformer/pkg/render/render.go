// Package render - система отрисовки Food Platformer
// Go365 Day 88
package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/playgo/city_platformer/pkg/entity"
	"github.com/playgo/city_platformer/pkg/level"
	"github.com/playgo/city_platformer/pkg/sprite"
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

// DrawPlayer - отрисовка игрока (повара)
func (r *Renderer) DrawPlayer(screen *ebiten.Image, p entity.Player, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	// Мигание если неуязвим
	if p.Invincible > 0 && (p.Invincible%6 < 3) {
		return
	}
	
	// Получаем спрайт игрока
	state := "stand"
	if p.IsCrouching {
		state = "crouch"
	} else if !p.OnGround {
		state = "jump"
	} else if p.IsMoving {
		state = "walk"
	}
	
	frame := int(p.AnimFrame)
	playerSprite := r.spriteSheet.GetPlayerFrame(state, frame)
	
	if playerSprite != nil {
		opts := &ebiten.DrawImageOptions{}
		
		if p.Facing == -1 {
			opts.GeoM.Scale(-1, 1)
			opts.GeoM.Translate(float64(p.Width), 0)
		}
		
		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(playerSprite, opts)
	} else {
		p.Draw(screen, cameraX, cameraY)
	}
}

// DrawFood - отрисовка еды
func (r *Renderer) DrawFood(screen *ebiten.Image, f *entity.Food, cameraX, cameraY float64) {
	if f.Collected {
		return
	}
	
	// Анимация парения
	offsetY := 0.0
	
	// Получаем спрайт еды или рисуем заглушкой
	foodSprite := r.spriteSheet.GetFood(sprite.FoodType(f.Type), int(f.AnimFrame)%10)
	
	if foodSprite != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(f.X-cameraX, f.Y-cameraY+offsetY)
		screen.DrawImage(foodSprite, opts)
	} else {
		f.Draw(screen, cameraX, cameraY)
	}
}

// DrawEnemy - отрисовка врага
func (r *Renderer) DrawEnemy(screen *ebiten.Image, e entity.Enemy, cameraX, cameraY float64) {
	frame := int(e.AnimFrame)
	enemySprite := r.spriteSheet.GetEnemy(e.Type, frame)
	
	if enemySprite != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(e.X-cameraX, e.Y-cameraY)
		screen.DrawImage(enemySprite, opts)
	} else {
		e.Draw(screen, cameraX, cameraY)
	}
}

// DrawBoss - отрисовка босса
func (r *Renderer) DrawBoss(screen *ebiten.Image, b entity.Boss, cameraX, cameraY float64) {
	// Рисуем босса
	b.Draw(screen, cameraX, cameraY)
}

// DrawPlatform - отрисовка платформы
func (r *Renderer) DrawPlatform(screen *ebiten.Image, p level.Platform, cameraX, cameraY float64) {
	tile := r.spriteSheet.GetTile(p.Type)
	
	if tile != nil {
		tileW := float64(tile.Bounds().Dx())
		tileH := float64(tile.Bounds().Dy())
		
		for ty := p.Y - cameraY; ty < p.Y+p.Height-cameraY; ty += tileH {
			for tx := p.X - cameraX; tx < p.X+p.Width-cameraX; tx += tileW {
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(tx, ty)
				screen.DrawImage(tile, opts)
			}
		}
	} else {
		// Резервная отрисовка
		var c color.RGBA
		switch p.Type {
		case "counter":
			c = color.RGBA{180, 140, 100, 255}
		case "floor":
			c = color.RGBA{200, 200, 200, 255}
		case "shelf":
			c = color.RGBA{160, 120, 80, 255}
		default:
			c = color.RGBA{150, 150, 150, 255}
		}
		vector.DrawFilledRect(screen, float32(p.X-cameraX), float32(p.Y-cameraY), float32(p.Width), float32(p.Height), c, true)
	}
}

// DrawCollectible - отрисовка предмета
func (r *Renderer) DrawCollectible(screen *ebiten.Image, c level.Collectible, cameraX, cameraY float64) {
	if c.Collected {
		return
	}
	
	// Рисуем как еду
	collectible := entity.NewFood(c.X, c.Y, c.TypeInt, c.Value)
	collectible.Draw(screen, cameraX, cameraY)
}

// DrawProjectile - отрисовка снаряда
func (r *Renderer) DrawProjectile(screen *ebiten.Image, p entity.Projectile, cameraX, cameraY float64) {
	if !p.Active {
		return
	}
	
	p.Draw(screen, cameraX, cameraY)
}

// DrawBackground - отрисовка фона (кухня)
func (r *Renderer) DrawBackground(screen *ebiten.Image, cameraX, cameraY float64) {
	// Стены кухни - светло-жёлтые
	for y := 0; y < screen.Bounds().Dy(); y++ {
		vector.DrawFilledRect(screen, 0, float32(y), float32(screen.Bounds().Dx()), 1, color.RGBA{255, 240, 200, 255}, true)
	}
	
	// Плитка на стене
	tileSize := 40.0
	startX := -int(cameraX*0.1) % int(tileSize)
	startY := int(cameraY*0.1) % int(tileSize)
	
	for y := startY; y < screen.Bounds().Dy(); y += int(tileSize) {
		for x := startX; x < screen.Bounds().Dx(); x += int(tileSize) {
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(tileSize-2), float32(tileSize-2), color.RGBA{255, 255, 255, 100}, true)
		}
	}
	
	// Кухонные полки на фоне
	for i := 0; i < 5; i++ {
		x := float64(i*300) - cameraX*0.3
		r.drawKitchenShelf(screen, x, 200)
	}
}

// drawKitchenShelf - отрисовка кухонной полки
func (r *Renderer) drawKitchenShelf(screen *ebiten.Image, x, y float64) {
	// Полка
	vector.DrawFilledRect(screen, float32(x), float32(y), 200, 10, color.RGBA{160, 120, 80, 255}, true)
	
	// Банки на полке
	for i := 0; i < 4; i++ {
		jarX := x + 20 + float64(i*50)
		vector.DrawFilledRect(screen, float32(jarX), float32(y)-25, 20, 25, color.RGBA{200, 180, 150, 200}, true)
	}
}

// DrawHUD - отрисовка интерфейса
func (r *Renderer) DrawHUD(screen *ebiten.Image, score, lives, level, health, maxHealth int) {
	// Фон HUD слева
	vector.DrawFilledRect(screen, 10, 10, 280, 90, color.RGBA{0, 0, 0, 150}, true)
	vector.DrawFilledRect(screen, 10, 10, 280, 90, color.RGBA{255, 255, 255, 80}, false)
	
	// Счёт
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🍳 Счёт: %d", score), 20, 20)
	
	// Здоровье
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("❤️ Здоровье: %d/%d", health, maxHealth), 20, 45)
	
	// Уровень
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("📍 Уровень: %d", level), 20, 70)
	
	// Полоска здоровья
	barWidth := 200
	barHeight := 15
	x := float32(150)
	y := float32(50)
	
	vector.DrawFilledRect(screen, x, y, float32(barWidth), float32(barHeight), color.RGBA{50, 50, 50, 255}, true)
	
	healthPercent := float32(health) / float32(maxHealth)
	barColor := color.RGBA{255, 50, 50, 255}
	if healthPercent > 0.5 {
		barColor = color.RGBA{50, 200, 50, 255}
	} else if healthPercent > 0.25 {
		barColor = color.RGBA{255, 200, 50, 255}
	}
	
	vector.DrawFilledRect(screen, x+2, y+2, float32(barWidth-4)*healthPercent, float32(barHeight-4), barColor, true)
}

// DrawBossHealthBar - полоска здоровья босса
func (r *Renderer) DrawBossHealthBar(screen *ebiten.Image, health, maxHealth int) {
	barWidth := 500
	barHeight := 25
	x := float32(screen.Bounds().Dx())/2 - float32(barWidth)/2
	y := float32(30)
	
	// Фон
	vector.DrawFilledRect(screen, x, y, float32(barWidth), float32(barHeight), color.RGBA{50, 50, 50, 255}, true)
	
	// Здоровье
	healthPercent := float32(health) / float32(maxHealth)
	vector.DrawFilledRect(screen, x+2, y+2, float32(barWidth-4)*healthPercent, float32(barHeight-4), color.RGBA{100, 150, 50, 255}, true)
	
	// Рамка
	vector.DrawFilledRect(screen, x, y, float32(barWidth), float32(barHeight), color.RGBA{255, 255, 255, 255}, false)
	
	// Текст
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🤢 Гнилой Шеф: %d/%d", health, maxHealth), int(x)+150, int(y)-5)
}

// DrawParticles - отрисовка частиц
func (r *Renderer) DrawParticles(screen *ebiten.Image, particles []Particle, cameraX, cameraY float64) {
	for _, p := range particles {
		screenX := p.X - cameraX
		screenY := p.Y - cameraY
		alpha := uint8(255 * (p.Life / p.MaxLife))
		
		r, gr, b, _ := p.Color.RGBA()
		vector.DrawFilledCircle(
			screen,
			float32(screenX),
			float32(screenY),
			float32(p.Size),
			color.RGBA{uint8(r >> 8), uint8(gr >> 8), uint8(b >> 8), alpha},
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
